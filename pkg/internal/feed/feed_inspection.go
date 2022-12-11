package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
	"strings"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/events"
)

// Inspection represents a row from the inspections feed
type Inspection struct {
	ID              string     `json:"id" csv:"audit_id" gorm:"primarykey;column:audit_id;size:100"`
	Name            string     `json:"name" csv:"name"`
	Archived        bool       `json:"archived" csv:"archived"`
	OwnerName       string     `json:"owner_name" csv:"owner_name"`
	OwnerID         string     `json:"owner_id" csv:"owner_id" gorm:"size:37"`
	AuthorName      string     `json:"author_name" csv:"author_name"`
	AuthorID        string     `json:"author_id" csv:"author_id" gorm:"size:37"`
	Score           float32    `json:"score" csv:"score"`
	MaxScore        float32    `json:"max_score" csv:"max_score"`
	ScorePercentage float32    `json:"score_percentage" csv:"score_percentage"`
	Duration        int64      `json:"duration" csv:"duration"`
	TemplateID      string     `json:"template_id" csv:"template_id" gorm:"size:100"`
	OrganisationID  string     `json:"organisation_id" csv:"organisation_id" gorm:"index:idx_ins_modified_at;size:37"`
	TemplateName    string     `json:"template_name" csv:"template_name"`
	TemplateAuthor  string     `json:"template_author" csv:"template_author"`
	SiteID          string     `json:"site_id" csv:"site_id" gorm:"size:41"`
	DateStarted     time.Time  `json:"date_started" csv:"date_started"`
	DateCompleted   *time.Time `json:"date_completed" csv:"date_completed"`
	DateModified    time.Time  `json:"date_modified" csv:"date_modified"`
	CreatedAt       time.Time  `json:"created_at" csv:"created_at"`
	ModifiedAt      time.Time  `json:"modified_at" csv:"modified_at" gorm:"index:idx_ins_modified_at,sort:desc"`
	ExportedAt      time.Time  `json:"exported_at" csv:"exported_at" gorm:"index:idx_ins_modified_at;autoUpdateTime"`
	DocumentNo      string     `json:"document_no" csv:"document_no"`
	PreparedBy      string     `json:"prepared_by" csv:"prepared_by"`
	Location        string     `json:"location" csv:"location"`
	ConductedOn     *time.Time `json:"conducted_on" csv:"conducted_on"`
	Personnel       string     `json:"personnel" csv:"personnel"`
	ClientSite      string     `json:"client_site" csv:"client_site"`
	Latitude        *float64   `json:"latitude" csv:"latitude"`
	Longitude       *float64   `json:"longitude" csv:"longitude"`
	WebReportLink   string     `json:"web_report_link" csv:"web_report_link"`
	Deleted         bool       `json:"deleted" csv:"deleted"`
}

// InspectionFeed is a representation of the inspections feed
type InspectionFeed struct {
	SkipIDs       []string
	ModifiedAfter time.Time
	TemplateIDs   []string
	Archived      string
	Completed     string
	Incremental   bool
	Limit         int
	WebReportLink string
}

// Name is the name of the feed
func (f *InspectionFeed) Name() string {
	return "inspections"
}

// Model returns the model of the feed row
func (f *InspectionFeed) Model() interface{} {
	return Inspection{}
}

// RowsModel returns the model of feed rows
func (f *InspectionFeed) RowsModel() interface{} {
	return &[]*Inspection{}
}

// PrimaryKey returns the primary key(s)
func (f *InspectionFeed) PrimaryKey() []string {
	return []string{"audit_id"}
}

// Columns returns the columns of the row
func (f *InspectionFeed) Columns() []string {
	return []string{
		"name",
		"archived",
		"owner_name",
		"owner_id",
		"author_name",
		"author_id",
		"score",
		"max_score",
		"score_percentage",
		"duration",
		"template_id",
		"organisation_id",
		"template_name",
		"template_author",
		"site_id",
		"date_started",
		"date_completed",
		"date_modified",
		"created_at",
		"modified_at",
		"exported_at",
		"document_no",
		"prepared_by",
		"location",
		"conducted_on",
		"personnel",
		"client_site",
		"latitude",
		"longitude",
		"web_report_link",
		"deleted",
	}
}

// Order returns the ordering when retrieving an export
func (f *InspectionFeed) Order() string {
	return "modified_at ASC, audit_id"
}

func (f *InspectionFeed) writeRows(exporter Exporter, rows []Inspection) error {
	skipIDs := map[string]bool{}
	for _, id := range f.SkipIDs {
		skipIDs[id] = true
	}

	// Calculate the size of the batch we can insert into the DB at once. Column count + buffer to account for primary keys
	batchSize := exporter.ParameterLimit() / (len(f.Columns()) + 4)
	for i := 0; i < len(rows); i += batchSize {
		j := i + batchSize
		if j > len(rows) {
			j = len(rows)
		}

		// Some audits in production have the same item ID multiple times
		// We can't insert them simultaneously. This means we are dropping data, which sucks.
		var rowsToInsert []Inspection
		for _, row := range rows[i:j] {
			skip := skipIDs[row.ID]
			if !skip {
				rowsToInsert = append(rowsToInsert, row)
			}
		}

		if err := exporter.WriteRows(f, rowsToInsert); err != nil {
			return err
		}
	}

	return nil
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *InspectionFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*Inspection{})
}

// Export exports the feed to the supplied exporter
func (f *InspectionFeed) Export(ctx context.Context, apiClient *httpapi.Client, exporter Exporter, orgID string) error {
	if err := exporter.InitFeed(f, &InitFeedOptions{
		// Delete data if incremental refresh is disabled so there is no duplicates
		Truncate: !f.Incremental,
	}); err != nil {
		return events.WrapEventError(err, "init feed")
	}

	var err error
	f.ModifiedAfter, err = exporter.LastModifiedAt(f, f.ModifiedAfter, orgID)
	if err != nil {
		return events.NewEventErrorWithMessage(err, events.ErrorSeverityError, events.ErrorSubSystemDB, false, "unable to load modified after")
	}

	// Process Inspections
	if err := f.processNewInspections(ctx, apiClient, exporter, orgID); err != nil {
		return events.WrapEventError(err, "export")
	}

	// Process Deleted Inspections
	if err := f.processDeletedInspections(ctx, apiClient, exporter); err != nil {
		return events.WrapEventError(err, "process deleted inspections")
	}

	return exporter.FinaliseExport(f, &[]*Inspection{})
}

func (f *InspectionFeed) processNewInspections(ctx context.Context, apiClient *httpapi.Client, exporter Exporter, orgID string) error {
	logger := logger.GetLogger().With("feed", f.Name(), "org_id", orgID)
	req := GetFeedRequest{
		InitialURL: "/feed/inspections",
		Params: GetFeedParams{
			ModifiedAfter: f.ModifiedAfter,
			TemplateIDs:   f.TemplateIDs,
			Archived:      f.Archived,
			Completed:     f.Completed,
			Limit:         f.Limit,
			WebReportLink: f.WebReportLink,
		},
	}
	feedFn := func(resp *GetFeedResponse) error {
		var rows []Inspection

		if err := json.Unmarshal(resp.Data, &rows); err != nil {
			return events.NewEventErrorWithMessage(err, events.ErrorSeverityError, events.ErrorSubSystemDataIntegrity, false, "map data")
		}

		if len(rows) != 0 {
			err := f.writeRows(exporter, rows)
			if err != nil {
				return err
			}
		}

		logger.With(
			"estimated_remaining", resp.Metadata.RemainingRecords,
			"duration_ms", apiClient.Duration.Milliseconds(),
			"export_duration_ms", exporter.GetDuration().Milliseconds(),
		).Info("export batch complete")

		return nil
	}

	return DrainFeed(ctx, apiClient, &req, feedFn)
}

func (f *InspectionFeed) processDeletedInspections(ctx context.Context, apiClient *httpapi.Client, exporter Exporter) error {
	lg := logger.GetLogger()
	dreq := NewGetAccountsActivityLogRequest(f.Limit, f.ModifiedAfter)
	delFn := func(resp *GetAccountsActivityLogResponse) error {
		var pkeys = make([]string, 0, len(resp.Activities))
		for _, a := range resp.Activities {
			uid := getPrefixID(a.Metadata["inspection_id"])
			if uid != "" {
				pkeys = append(pkeys, uid)
			}
		}
		if len(pkeys) > 0 {
			rowsUpdated, err := exporter.UpdateRows(f, pkeys, map[string]interface{}{"deleted": true})
			if err != nil {
				return err
			}
			lg.Infof("there were %d rows marked as deleted", rowsUpdated)
		}

		return nil
	}
	return DrainAccountActivityHistoryLog(ctx, apiClient, dreq, delFn)
}

func getPrefixID(id string) string {
	return fmt.Sprintf("audit_%s", strings.ReplaceAll(id, "-", ""))
}
