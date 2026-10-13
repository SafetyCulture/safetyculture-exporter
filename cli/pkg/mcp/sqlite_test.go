package mcp

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateReadOnlySQL(t *testing.T) {
	valid := []string{
		"SELECT * FROM inspections",
		"SELECT status, COUNT(*) FROM inspections GROUP BY status;",
		"WITH x AS (SELECT 1) SELECT * FROM x",
		"  select id from users  ",
		"EXPLAIN QUERY PLAN SELECT * FROM sites",
		"SELECT name FROM templates WHERE name LIKE '%updated%'", // keyword inside string literal-ish word
		"SELECT * FROM inspections WHERE notes = 'drop it'",
		"SELECT * FROM inspections WHERE notes = 'it''s a delete word'",
		"SELECT updated_at, deleted_at FROM inspections", // column names containing keywords
	}
	for _, q := range valid {
		assert.NoError(t, validateReadOnlySQL(q), q)
	}

	invalid := []string{
		"",
		"DELETE FROM inspections",
		"SELECT 1; DROP TABLE inspections",
		"UPDATE users SET name='x'",
		"PRAGMA journal_mode=WAL",
		"ATTACH DATABASE 'x' AS y",
		"WITH x AS (DELETE FROM t RETURNING *) SELECT * FROM x",
		"CREATE TABLE foo (id INT)",
		"SELECT * FROM t WHERE x = 'ok'; DROP TABLE t", // stacked even with literal
		"SELECT 1 FROM t WHERE a = 'x' OR 1=1; DELETE FROM t",
	}
	for _, q := range invalid {
		assert.Error(t, validateReadOnlySQL(q), q)
	}
}

func setupTestDB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	db, err := sql.Open("sqlite3", filepath.Join(dir, sqliteDBFile))
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE inspections (id TEXT PRIMARY KEY, status TEXT, score INTEGER);
		INSERT INTO inspections VALUES ('a1', 'completed', 90), ('a2', 'failed', 40), ('a3', 'completed', 75);`)
	require.NoError(t, err)
	return dir
}

func TestRunQuery(t *testing.T) {
	dir := setupTestDB(t)
	ctx := context.Background()

	res, err := runQuery(ctx, dir, "SELECT id, status FROM inspections ORDER BY id", 0)
	require.NoError(t, err)
	assert.Equal(t, []string{"id", "status"}, res.Columns)
	assert.Equal(t, 3, res.RowCount)
	assert.False(t, res.Truncated)
	assert.Equal(t, "a1", res.Rows[0][0])
}

func TestRunQueryTruncates(t *testing.T) {
	dir := setupTestDB(t)

	res, err := runQuery(context.Background(), dir, "SELECT * FROM inspections", 2)
	require.NoError(t, err)
	assert.Equal(t, 2, res.RowCount)
	assert.True(t, res.Truncated)
}

func TestRunQueryRejectsWrites(t *testing.T) {
	dir := setupTestDB(t)

	_, err := runQuery(context.Background(), dir, "DELETE FROM inspections", 0)
	assert.Error(t, err)

	// Defence in depth: even if the guard missed it, the connection is read-only.
	_, err = runQuery(context.Background(), dir, "SELECT 1", 0)
	require.NoError(t, err)
}

func TestRunQueryMissingDB(t *testing.T) {
	_, err := runQuery(context.Background(), t.TempDir(), "SELECT 1", 0)
	require.ErrorContains(t, err, "sync_data")
}

func TestListTables(t *testing.T) {
	dir := setupTestDB(t)

	tables, err := listTables(context.Background(), dir)
	require.NoError(t, err)
	require.Len(t, tables, 1)
	assert.Equal(t, "inspections", tables[0].Name)
	assert.Equal(t, 3, tables[0].Columns)
	assert.Equal(t, 3, tables[0].Rows)
}

func TestDescribeTable(t *testing.T) {
	dir := setupTestDB(t)

	cols, sample, err := describeTable(context.Background(), dir, "inspections")
	require.NoError(t, err)
	require.Len(t, cols, 3)
	assert.Equal(t, "id", cols[0].Name)
	assert.True(t, cols[0].Primary)
	require.NotNil(t, sample)
	assert.Equal(t, 3, sample.RowCount)

	_, _, err = describeTable(context.Background(), dir, "nope")
	assert.Error(t, err)

	_, _, err = describeTable(context.Background(), dir, "bad'; DROP TABLE x--")
	assert.Error(t, err)
}

func TestStructuredResultRoundTrip(t *testing.T) {
	res, out, err := structuredResult(map[string]any{"ok": true}, "done")
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Len(t, res.Content, 1)
	assert.False(t, res.IsError)
	assert.Contains(t, fmt.Sprint(res.Content[0]), "done")
}
