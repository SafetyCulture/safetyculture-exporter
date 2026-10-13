package mcp

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	exporterAPI "github.com/SafetyCulture/mitti-exporter/pkg/api"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SEP-1865 UI resources for reports and sync progress.
const (
	reportViewerURI = "ui://mitti/report-viewer"
	syncStatusURI   = "ui://mitti/sync-status"
)

//go:embed ui/report-viewer.html
var reportViewerHTML string

//go:embed ui/sync-status.html
var syncStatusHTML string

// maxFileBytes caps files read for download/media tools (25 MB).
const maxFileBytes = 25 * 1024 * 1024

type generateReportArgs struct {
	Formats []string `json:"formats,omitempty" jsonschema:"report formats to generate: PDF, WORD (default from config)"`
}

type readFileArgs struct {
	Path string `json:"path" jsonschema:"path of a file inside the export directory (as returned by generate_report)"`
}

type readMediaArgs struct {
	Path string `json:"path" jsonschema:"path of a media file relative to the export media directory"`
}

// registerExtendedTools adds report, media, progress and app-only tools.
func registerExtendedTools(srv *mcp.Server, cfg *Config) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "generate_report",
		Description: "Generate inspection reports (PDF/WORD) into the export directory, using the " +
			"exporter's configured filters (templates, date range, completed/archived). " +
			"Returns the list of report files written.",
		Meta: mcp.Meta{"ui": map[string]any{"resourceUri": reportViewerURI}},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args generateReportArgs) (*mcp.CallToolResult, any, error) {
		if cfg.GenerateReports == nil {
			return errorResult(fmt.Errorf("report generation is not configured"))
		}
		files, err := cfg.GenerateReports(ctx, args.Formats)
		if err != nil {
			return errorResult(fmt.Errorf("report generation failed: %w", err))
		}
		out := map[string]any{"files": files, "count": len(files), "export_path": cfg.ExportPath}
		return structuredResult(out, fmt.Sprintf("%d report files generated in %s", len(files), cfg.ExportPath))
	})

	// App-only: the report viewer calls this to fetch file bytes for
	// ui/download-file. Hidden from the model (visibility: ["app"]).
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "read_export_file",
		Description: "Read a file from the export directory as base64 (for the report viewer UI).",
		Meta:        mcp.Meta{"ui": map[string]any{"resourceUri": reportViewerURI, "visibility": []string{"app"}}},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args readFileArgs) (*mcp.CallToolResult, any, error) {
		data, name, mime, err := readJailedFile(cfg.ExportPath, args.Path, []string{".pdf", ".docx", ".csv", ".json", ".md"})
		if err != nil {
			return errorResult(err)
		}
		out := map[string]any{
			"name":      name,
			"mime_type": mime,
			"size":      len(data),
			"data":      base64.StdEncoding.EncodeToString(data),
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("%s (%d bytes)", name, len(data))}},
		}, out, nil
	})

	// App-only: the sync-status view polls this for live progress.
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "sync_progress",
		Description: "Current export/sync progress per feed (for the sync status UI).",
		Meta:        mcp.Meta{"ui": map[string]any{"resourceUri": syncStatusURI, "visibility": []string{"app"}}},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyArgs) (*mcp.CallToolResult, any, error) {
		status := exporterAPI.GetGlobalExportStatus()
		return structuredResult(status, progressSummary(status))
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name: "read_media",
		Description: "Read an inspection media file (photo) from the export media directory and return it " +
			"as an image. Paths are relative to the media directory (e.g. from inspection_items media fields). " +
			"Requires media export to be enabled (export.media: true).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args readMediaArgs) (*mcp.CallToolResult, any, error) {
		mediaPath := cfg.MediaPath
		if mediaPath == "" {
			mediaPath = filepath.Join(cfg.ExportPath, "media")
		}
		data, _, mime, err := readJailedFile(mediaPath, args.Path, []string{".jpg", ".jpeg", ".png", ".gif", ".webp"})
		if err != nil {
			return errorResult(err)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.ImageContent{Data: data, MIMEType: mime}},
		}, nil, nil
	})
}

// progressSummary renders a compact text summary of export progress.
func progressSummary(s *exporterAPI.ExportStatusResponse) string {
	if !s.ExportStarted {
		return "no sync has run yet in this session"
	}
	done, total := 0, len(s.Feeds)
	for _, f := range s.Feeds {
		if f.Finished {
			done++
		}
	}
	if s.ExportCompleted {
		return fmt.Sprintf("sync complete: %d/%d feeds", done, total)
	}
	return fmt.Sprintf("sync in progress: %d/%d feeds finished", done, total)
}

// readJailedFile reads a file strictly inside baseDir, rejecting path
// traversal and disallowing extensions not in allowExts.
func readJailedFile(baseDir, rel string, allowExts []string) (data []byte, name, mime string, err error) {
	if rel == "" {
		return nil, "", "", fmt.Errorf("path is required")
	}
	clean := filepath.Clean(rel)
	if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		return nil, "", "", fmt.Errorf("path %q escapes the allowed directory", rel)
	}
	full := filepath.Join(baseDir, clean)
	// Resolve symlinks on both sides before comparing.
	baseResolved, err := filepath.EvalSymlinks(baseDir)
	if err != nil {
		return nil, "", "", fmt.Errorf("directory %s not found — run an export first", baseDir)
	}
	fullResolved, err := filepath.EvalSymlinks(full)
	if err != nil {
		return nil, "", "", fmt.Errorf("file not found: %s", rel)
	}
	if fullResolved != baseResolved && !strings.HasPrefix(fullResolved, baseResolved+string(os.PathSeparator)) {
		return nil, "", "", fmt.Errorf("path %q escapes the allowed directory", rel)
	}

	ext := strings.ToLower(filepath.Ext(fullResolved))
	allowed := false
	for _, e := range allowExts {
		if ext == e {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, "", "", fmt.Errorf("file type %s is not allowed", ext)
	}

	info, err := os.Stat(fullResolved)
	if err != nil {
		return nil, "", "", err
	}
	if info.Size() > maxFileBytes {
		return nil, "", "", fmt.Errorf("file is %d bytes; limit is %d", info.Size(), maxFileBytes)
	}

	data, err = os.ReadFile(fullResolved)
	if err != nil {
		return nil, "", "", err
	}
	return data, filepath.Base(fullResolved), mimeForExt(ext), nil
}

func mimeForExt(ext string) string {
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".csv":
		return "text/csv"
	case ".json":
		return "application/json"
	case ".md":
		return "text/markdown"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	}
	return "application/octet-stream"
}

// ListReportFiles returns report files (PDF/DOCX) in the export dir, newest first.
func ListReportFiles(exportPath string) ([]string, error) {
	entries, err := os.ReadDir(exportPath)
	if err != nil {
		return nil, err
	}
	type fileEnt struct {
		name string
		mod  int64
	}
	var files []fileEnt
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext != ".pdf" && ext != ".docx" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileEnt{e.Name(), info.ModTime().Unix()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].mod > files[j].mod })
	var names []string
	for _, f := range files {
		if len(names) >= 50 {
			break
		}
		names = append(names, f.name)
	}
	return names, nil
}

// registerPrompts adds ready-made prompts so users can kick off common
// reporting flows ("weekly safety review") without writing anything.
func registerPrompts(srv *mcp.Server) {
	srv.AddPrompt(&mcp.Prompt{
		Name:        "weekly_safety_review",
		Description: "Summarise the last 7 days of safety activity: inspections, failures, actions, issues",
	}, func(_ context.Context, _ *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return promptResult("Weekly safety review", `Using the Mitti data tools, produce a weekly safety review:
1. Call sync_data if the data hasn't been synced today.
2. Inspections completed in the last 7 days vs the 7 before (count, average score).
3. Failed inspection items this week, grouped by template, top 10.
4. Actions created vs completed this week, and how many are now overdue.
5. New issues raised this week by site.
Finish with a short "what needs attention" list. Present as a concise report.`), nil
	})

	srv.AddPrompt(&mcp.Prompt{
		Name:        "overdue_actions",
		Description: "List overdue actions grouped by site and assignee, with ageing",
	}, func(_ context.Context, _ *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return promptResult("Overdue actions", `Using the Mitti data tools, report on overdue actions:
1. Query the actions and action_assignees tables for actions past their due date that are not completed.
2. Group by site, then by assignee, showing count and oldest overdue age in days.
3. Highlight the 10 oldest overdue actions with their titles and due dates.
Suggest follow-ups for the worst offenders.`), nil
	})

	srv.AddPrompt(&mcp.Prompt{
		Name:        "site_scorecard",
		Description: "Scorecard for one site: inspection volume, scores, failures, actions",
		Arguments:   []*mcp.PromptArgument{{Name: "site", Description: "Site name or ID", Required: true}},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		site := req.Params.Arguments["site"]
		return promptResult("Site scorecard", fmt.Sprintf(`Using the Mitti data tools, build a scorecard for the site "%s":
1. Find the site in the sites table (match by name if needed).
2. Inspection volume and average score for the last 90 days.
3. Most-failed inspection items for this site.
4. Open and overdue actions for this site, with ageing.
5. Schedule compliance: scheduled inspections missed in the last 30 days.
Present as a one-page scorecard with a short trend commentary.`, site)), nil
	})
}

func promptResult(description, text string) *mcp.GetPromptResult {
	return &mcp.GetPromptResult{
		Description: description,
		Messages: []*mcp.PromptMessage{{
			Role:    "user",
			Content: &mcp.TextContent{Text: text},
		}},
	}
}

// elicitSyncMode asks the user whether to run a full or incremental sync,
// when the host supports elicitation. Defaults to incremental on any error.
func elicitSyncMode(ctx context.Context, req *mcp.CallToolRequest) string {
	if req == nil || req.Session == nil {
		return "incremental"
	}
	res, err := req.Session.Elicit(ctx, &mcp.ElicitParams{
		Message: "Run an incremental sync (only changes since last export) or a full sync (re-download everything)?",
		RequestedSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"mode": map[string]any{
					"type":        "string",
					"title":       "Sync mode",
					"enum":        []string{"incremental", "full"},
					"default":     "incremental",
					"description": "Incremental is faster; full re-downloads all data.",
				},
			},
		},
	})
	if err != nil || res.Action != "accept" {
		return "incremental"
	}
	if mode, ok := res.Content["mode"].(string); ok && mode == "full" {
		return "full"
	}
	return "incremental"
}
