// Package mcp implements an MCP (Model Context Protocol) server that exposes
// the exporter's local SQLite database to AI agents, plus SEP-1865 (MCP Apps)
// UI resources for interactive results.
package mcp

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// sqliteDBFile is the file produced by `mitti-exporter sqlite`.
const sqliteDBFile = "sqlite_export.db"

// defaultMaxRows caps rows returned by a single query so agents don't blow
// their context window on a full table scan.
const defaultMaxRows = 200

// queryTimeout bounds any single agent-issued query.
const queryTimeout = 30 * time.Second

// dbPath returns the path of the SQLite export database for an export path.
func dbPath(exportPath string) string {
	return filepath.Join(exportPath, sqliteDBFile)
}

// openReadOnly opens the export database in read-only, immutable-ish mode.
// Opening with mode=ro means even a buggy query guard cannot write.
func openReadOnly(exportPath string) (*sql.DB, error) {
	path := dbPath(exportPath)
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("no export database found at %s — run the sync_data tool (or `mitti-exporter sqlite`) first", path)
	}
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro&_busy_timeout=5000", path))
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}
	return db, nil
}

// validateReadOnlySQL enforces a single, read-only SELECT statement.
// Defence in depth: the connection itself is opened mode=ro, but rejecting
// obviously-wrong input early gives the agent a much better error message.
func validateReadOnlySQL(query string) error {
	q := strings.TrimSpace(query)
	if q == "" {
		return fmt.Errorf("query is empty")
	}
	// Strip a single trailing semicolon; reject anything after it (no stacked statements).
	q = strings.TrimSuffix(q, ";")
	if strings.Contains(q, ";") {
		return fmt.Errorf("only a single SELECT statement is allowed (no stacked statements)")
	}
	upper := strings.ToUpper(q)
	if !strings.HasPrefix(upper, "SELECT") && !strings.HasPrefix(upper, "WITH") && !strings.HasPrefix(upper, "EXPLAIN") {
		return fmt.Errorf("only read-only SELECT queries are allowed")
	}
	// Keyword screen for writes hidden in CTEs/subqueries. String literals are
	// blanked first so text like WHERE note = 'drop it' doesn't false-positive.
	scan := blankStringLiterals(upper)
	for _, kw := range []string{"INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "CREATE", "REPLACE", "ATTACH", "DETACH", "PRAGMA", "VACUUM"} {
		// Word-boundary-ish check: keyword surrounded by non-letters.
		for i := 0; i+len(kw) <= len(scan); i++ {
			if scan[i:i+len(kw)] != kw {
				continue
			}
			beforeOK := i == 0 || !isLetter(scan[i-1])
			afterOK := i+len(kw) == len(scan) || !isLetter(scan[i+len(kw)])
			if beforeOK && afterOK {
				return fmt.Errorf("keyword %s is not allowed in read-only queries", kw)
			}
		}
	}
	return nil
}

// blankStringLiterals replaces the contents of single-quoted string literals
// with spaces (preserving length and positions) so keyword scans don't trip
// over user text. Handles '' escaping. Doubled quotes inside literals are
// consumed as part of the literal.
func blankStringLiterals(s string) string {
	b := []byte(s)
	for i := 0; i < len(b); i++ {
		if b[i] != '\'' {
			continue
		}
		// Inside a literal: blank until the closing quote.
		for j := i + 1; j < len(b); j++ {
			if b[j] == '\'' {
				if j+1 < len(b) && b[j+1] == '\'' { // escaped quote ''
					b[j] = ' '
					b[j+1] = ' '
					j++
					continue
				}
				// Closing quote; blank everything between i and j.
				for k := i + 1; k < j; k++ {
					b[k] = ' '
				}
				i = j
				break
			}
			b[j] = ' '
		}
	}
	return string(b)
}

func isLetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || b == '_'
}

// queryResult is the structured payload returned to agents and UI resources.
type queryResult struct {
	Columns   []string        `json:"columns"`
	Rows      [][]interface{} `json:"rows"`
	RowCount  int             `json:"row_count"`
	Truncated bool            `json:"truncated"`
}

// runQuery executes a read-only query against the export database, capping
// the result at maxRows.
func runQuery(ctx context.Context, exportPath, query string, maxRows int) (*queryResult, error) {
	if err := validateReadOnlySQL(query); err != nil {
		return nil, err
	}
	if maxRows <= 0 {
		maxRows = defaultMaxRows
	}

	db, err := openReadOnly(exportPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, strings.TrimSuffix(strings.TrimSpace(query), ";"))
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	res := &queryResult{Columns: cols, Rows: [][]interface{}{}}
	for rows.Next() {
		if len(res.Rows) >= maxRows {
			res.Truncated = true
			break
		}
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		// Normalise SQLite driver types into JSON-friendly values.
		for i, v := range vals {
			switch t := v.(type) {
			case []byte:
				vals[i] = string(t)
			case time.Time:
				vals[i] = t.Format(time.RFC3339)
			}
		}
		res.Rows = append(res.Rows, vals)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	res.RowCount = len(res.Rows)
	return res, nil
}

// tableInfo describes one table in the export database.
type tableInfo struct {
	Name    string `json:"name"`
	Columns int    `json:"columns"`
	Rows    int    `json:"rows"`
}

// listTables returns every user table with approximate row counts.
func listTables(ctx context.Context, exportPath string) ([]tableInfo, error) {
	db, err := openReadOnly(exportPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx,
		`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []tableInfo
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, tableInfo{Name: name})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i, t := range tables {
		var colCount, rowCount int
		if err := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM pragma_table_info('%s')`, t.Name)).Scan(&colCount); err != nil {
			continue
		}
		if err := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, t.Name)).Scan(&rowCount); err != nil {
			continue
		}
		tables[i].Columns = colCount
		tables[i].Rows = rowCount
	}
	return tables, nil
}

// columnInfo describes one column of a table.
type columnInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Primary  bool   `json:"primary_key"`
}

// describeTable returns the schema plus a small sample of rows for a table.
func describeTable(ctx context.Context, exportPath, table string) ([]columnInfo, *queryResult, error) {
	if strings.TrimSpace(table) == "" || strings.ContainsAny(table, `"'%;`) {
		return nil, nil, fmt.Errorf("invalid table name %q", table)
	}

	db, err := openReadOnly(exportPath)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, fmt.Sprintf(`SELECT name, type, "notnull", pk FROM pragma_table_info('%s')`, table))
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var cols []columnInfo
	for rows.Next() {
		var c columnInfo
		var notNull, pk int
		if err := rows.Scan(&c.Name, &c.Type, &notNull, &pk); err != nil {
			return nil, nil, err
		}
		c.Nullable = notNull == 0
		c.Primary = pk > 0
		cols = append(cols, c)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	if len(cols) == 0 {
		return nil, nil, fmt.Errorf("table %q not found in export database", table)
	}

	sample, err := runQuery(ctx, exportPath, fmt.Sprintf(`SELECT * FROM "%s" LIMIT 5`, table), 5)
	if err != nil {
		return cols, nil, nil // schema is still useful without the sample
	}
	return cols, sample, nil
}
