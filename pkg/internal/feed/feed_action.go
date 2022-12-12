package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/events"
)

// Action represents a row from the actions feed
type Action struct {
	ID              string     `json:"id" csv:"action_id" gorm:"primarykey;column:action_id;size:36"`
	Title           string     `json:"title" csv:"title"`
	Description     string     `json:"description" csv:"description"`
	SiteID          string     `json:"site_id" csv:"site_id" gorm:"size:41"`
	Priority        string     `json:"priority" csv:"priority" gorm:"size:20"`
	Status          string     `json:"status" csv:"status" gorm:"size:20"`
	DueDate         time.Time  `json:"due_date" csv:"due_date"`
	CreatedAt       time.Time  `json:"created_at" csv:"created_at"`
	ModifiedAt      time.Time  `json:"modified_at" csv:"modified_at" gorm:"index:idx_act_modified_at,sort:desc"`
	ExportedAt      time.Time  `json:"exported_at" csv:"exported_at" gorm:"index:idx_act_modified_at;autoUpdateTime"`
	CreatorUserID   string     `json:"creator_user_id" csv:"creator_user_id" gorm:"size:37"`
	CreatorUserName string     `json:"creator_user_name" csv:"creator_user_name"`
	TemplateID      string     `json:"template_id" csv:"template_id" gorm:"size:100"`
	AuditID         string     `json:"audit_id" csv:"audit_id" gorm:"size:100"`
	AuditTitle      string     `json:"audit_title" csv:"audit_title"`
	AuditItemID     string     `json:"audit_item_id" csv:"audit_item_id" gorm:"size:100"`
	AuditItemLabel  string     `json:"audit_item_label" csv:"audit_item_label"`
	OrganisationID  string     `json:"organisation_id" csv:"organisation_id" gorm:"index:idx_act_modified_at;size:37"`
	CompletedAt     *time.Time `json:"completed_at" csv:"completed_at"`
}

// ActionFeed is a representation of the actions feed
type ActionFeed struct {
	ModifiedAfter time.Time
	Incremental   bool
	Limit         int
}

// Name is the name of the feed
func (f *ActionFeed) Name() string {
	return "actions"
}

// Model returns the model of the feed row
func (f *ActionFeed) Model() interface{} {
	return Action{}
}

// RowsModel returns the model of feed rows
func (f *ActionFeed) RowsModel() interface{} {
	return &[]*Action{}
}

// PrimaryKey returns the primary key(s)
func (f *ActionFeed) PrimaryKey() []string {
	return []string{"action_id"}
}

// Columns returns the columns of the row
func (f *ActionFeed) Columns() []string {
	return []string{
		"title",
		"description",
		"site_id",
		"priority",
		"status",
		"due_date",
		"created_at",
		"modified_at",
		"exported_at",
		"creator_user_id",
		"creator_user_name",
		"template_id",
		"audit_id",
		"audit_title",
		"audit_item_id",
		"audit_item_label",
		"organisation_id",
		"completed_at",
	}
}

// Order returns the ordering when retrieving an export
func (f *ActionFeed) Order() string {
	return "action_id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *ActionFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*Action{})
}

// Export exports the feed to the supplied exporter
func (f *ActionFeed) Export(ctx context.Context, apiClient *httpapi.Client, exporter Exporter, orgID string) error {
	logger := logger.GetLogger().With("feed", f.Name(), "org_id", orgID)

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

	drainFn := func(resp *GetFeedResponse) error {
		var rows []*Action

		if err := json.Unmarshal(resp.Data, &rows); err != nil {
			return events.NewEventErrorWithMessage(err, events.ErrorSeverityError, events.ErrorSubSystemDataIntegrity, false, "map data")
		}

		if len(rows) != 0 {
			// Calculate the size of the batch we can insert into the DB at once. Column count + buffer to account for primary keys
			batchSize := exporter.ParameterLimit() / (len(f.Columns()) + 4)

			for i := 0; i < len(rows); i += batchSize {
				j := i + batchSize
				if j > len(rows) {
					j = len(rows)
				}

				if err := exporter.WriteRows(f, rows[i:j]); err != nil {
					return events.WrapEventError(err, "write rows")
				}
			}
		}

		logger.With(
			"estimated_remaining", resp.Metadata.RemainingRecords,
			"duration_ms", apiClient.Duration.Milliseconds(),
			"export_duration_ms", exporter.GetDuration().Milliseconds(),
		).Info("export batch complete")
		return nil
	}

	req := &GetFeedRequest{
		InitialURL: "/feed/actions",
		Params: GetFeedParams{
			ModifiedAfter: f.ModifiedAfter,
			Limit:         f.Limit,
		},
	}

	if err := DrainFeed(ctx, apiClient, req, drainFn); err != nil {
		return events.WrapEventError(err, fmt.Sprintf("feed %q", f.Name()))
	}
	return exporter.FinaliseExport(f, &[]*Action{})
}
