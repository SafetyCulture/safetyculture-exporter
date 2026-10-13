package mcp

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// tableViewerURI is the SEP-1865 (MCP Apps) UI resource used to render query results.
const tableViewerURI = "ui://mitti/table-viewer"

//go:embed ui/table-viewer.html
var tableViewerHTML string

// Config wires the MCP server into the exporter.
type Config struct {
	// Name/Version identify the server to MCP hosts.
	Name    string
	Version string

	// ExportPath is where the SQLite export database lives (export.path).
	ExportPath string

	// MediaPath is where exported media files live (export.media_path).
	// Defaults to <ExportPath>/media when empty.
	MediaPath string

	// Sync runs an incremental SQLite export (equivalent to
	// `mitti-exporter sqlite`). Called by the sync_data tool.
	// Must be safe to call more than once.
	Sync func(ctx context.Context) error

	// SyncFull runs a non-incremental export. Optional — when nil, sync_data
	// always uses Sync.
	SyncFull func(ctx context.Context) error

	// GenerateReports generates inspection reports in the given formats
	// ("PDF", "WORD") and returns the files written. Optional — the
	// generate_report tool errors clearly when nil.
	GenerateReports func(ctx context.Context, formats []string) ([]string, error)
}

// NewServer builds the MCP server with all tools and UI resources registered.
func NewServer(cfg *Config) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    cfg.Name,
		Version: cfg.Version,
	}, nil)

	registerTools(srv, cfg)
	registerExtendedTools(srv, cfg)
	registerPrompts(srv)
	registerResources(srv)

	return srv
}

// Run starts the server over stdio until the host disconnects.
func Run(ctx context.Context, cfg *Config) error {
	return NewServer(cfg).Run(ctx, &mcp.StdioTransport{})
}

// --- Tool argument types ---

type queryArgs struct {
	SQL     string `json:"sql" jsonschema:"read-only SQL SELECT to run against the local export database"`
	MaxRows int    `json:"max_rows,omitempty" jsonschema:"maximum rows to return (default 200, hard cap 1000)"`
}

type describeTableArgs struct {
	Table string `json:"table" jsonschema:"name of the table to describe (see list_tables)"`
}

type emptyArgs struct{}

// --- Tool registration ---

func registerTools(srv *mcp.Server, cfg *Config) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "list_tables",
		Description: "List all tables in the local Mitti export database with column and row counts. " +
			"Start here to discover what data is available before writing queries.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyArgs) (*mcp.CallToolResult, any, error) {
		tables, err := listTables(ctx, cfg.ExportPath)
		if err != nil {
			return errorResult(err)
		}
		return structuredResult(tables, fmt.Sprintf("%d tables available", len(tables)))
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name: "describe_table",
		Description: "Show the schema (columns, types, primary keys) of one table plus 5 sample rows. " +
			"Use this before writing a query against a table you haven't seen.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args describeTableArgs) (*mcp.CallToolResult, any, error) {
		cols, sample, err := describeTable(ctx, cfg.ExportPath, args.Table)
		if err != nil {
			return errorResult(err)
		}
		out := map[string]any{"columns": cols, "sample_rows": sample}
		return structuredResult(out, fmt.Sprintf("table %s: %d columns", args.Table, len(cols)))
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name: "query",
		Description: "Run a read-only SQL SELECT against the local Mitti export database (SQLite). " +
			"Returns JSON rows, capped at max_rows. Use list_tables/describe_table first to learn the schema. " +
			"Example: SELECT status, COUNT(*) AS n FROM inspections GROUP BY status",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
		// SEP-1865: hosts that support MCP Apps render results in the table viewer UI.
		// "app" visibility lets the view re-run queries (pagination/sorting)
		// without spending model tokens.
		Meta: mcp.Meta{"ui": map[string]any{"resourceUri": tableViewerURI, "visibility": []string{"model", "app"}}},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args queryArgs) (*mcp.CallToolResult, any, error) {
		if args.MaxRows > 1000 {
			args.MaxRows = 1000
		}
		res, err := runQuery(ctx, cfg.ExportPath, args.SQL, args.MaxRows)
		if err != nil {
			return errorResult(err)
		}
		summary := fmt.Sprintf("%d rows returned", res.RowCount)
		if res.Truncated {
			summary += fmt.Sprintf(" (truncated at %d — refine the query with WHERE/LIMIT)", args.MaxRows)
		}
		return structuredResult(res, summary)
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name: "sync_data",
		Description: "Sync the latest Mitti data into the local export database (same as `mitti-exporter sqlite`). " +
			"Incremental by default: only fetches changes since the last sync. Run this when the user asks " +
			"for fresh data or when list_tables reports the database is missing.",
		Meta: mcp.Meta{"ui": map[string]any{"resourceUri": syncStatusURI}},
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ emptyArgs) (*mcp.CallToolResult, any, error) {
		if cfg.Sync == nil {
			return errorResult(fmt.Errorf("sync is not configured"))
		}
		syncFn := cfg.Sync
		mode := "incremental"
		// Ask the user when the host supports elicitation and a full sync is available.
		if cfg.SyncFull != nil && elicitSyncMode(ctx, req) == "full" {
			syncFn = cfg.SyncFull
			mode = "full"
		}
		if err := syncFn(ctx); err != nil {
			return errorResult(fmt.Errorf("sync failed: %w", err))
		}
		tables, _ := listTables(ctx, cfg.ExportPath)
		return structuredResult(map[string]any{"synced": true, "mode": mode, "tables": tables}, fmt.Sprintf("sync complete (%s)", mode))
	})
}

// --- SEP-1865 UI resources ---

func registerResources(srv *mcp.Server) {
	uiMeta := mcp.Meta{"ui": map[string]any{
		// Fully self-contained: no network, no external resources.
		"csp":           map[string]any{},
		"prefersBorder": true,
	}}

	srv.AddResource(&mcp.Resource{
		URI:         tableViewerURI,
		Name:        "Mitti Query Results",
		Description: "Interactive table view for query tool results",
		MIMEType:    "text/html;profile=mcp-app",
		Meta:        uiMeta,
	}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      req.Params.URI,
				MIMEType: "text/html;profile=mcp-app",
				Text:     tableViewerHTML,
				Meta:     uiMeta,
			}},
		}, nil
	})

	srv.AddResource(&mcp.Resource{
		URI:         reportViewerURI,
		Name:        "Mitti Reports",
		Description: "Generated inspection reports with download buttons",
		MIMEType:    "text/html;profile=mcp-app",
		Meta:        uiMeta,
	}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      req.Params.URI,
				MIMEType: "text/html;profile=mcp-app",
				Text:     reportViewerHTML,
				Meta:     uiMeta,
			}},
		}, nil
	})

	srv.AddResource(&mcp.Resource{
		URI:         syncStatusURI,
		Name:        "Mitti Sync Status",
		Description: "Live progress while data syncs",
		MIMEType:    "text/html;profile=mcp-app",
		Meta:        uiMeta,
	}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      req.Params.URI,
				MIMEType: "text/html;profile=mcp-app",
				Text:     syncStatusHTML,
				Meta:     uiMeta,
			}},
		}, nil
	})
}

// --- Result helpers ---

// maxTextPayload caps the JSON embedded in the text content block. The full
// payload always goes to structuredContent; the text copy is a bounded preview
// so the model's context isn't flooded by large result sets.
const maxTextPayload = 4000

// structuredResult packs a structured payload plus a human/model-readable text
// summary. The text keeps non-MCP-Apps hosts (and the model) fully functional.
func structuredResult(payload any, summary string) (*mcp.CallToolResult, any, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return errorResult(err)
	}
	text := summary
	if len(raw) <= maxTextPayload {
		text += "\n\n" + string(raw)
	} else {
		text += fmt.Sprintf("\n\n(result is %d bytes of JSON — see structured content; refine the query to reduce size)", len(raw))
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, payload, nil
}

func errorResult(err error) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
		IsError: true,
	}, nil, nil
}
