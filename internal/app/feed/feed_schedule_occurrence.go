package feed

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
)

// ScheduleOccurrence represents a row from the schedule_occurrences feed
type ScheduleOccurrence struct {
	ID               string     `json:"id" csv:"id" gorm:"primarykey;size:128"`
	ScheduleID       string     `json:"schedule_id" csv:"schedule_id" gorm:"size:45"`
	OccurrenceID     string     `json:"occurrence_id" csv:"occurrence_id" gorm:"size:30"`
	TemplateID       string     `json:"template_id" csv:"template_id" gorm:"size:100"`
	OrganisationID   string     `json:"organisation_id" csv:"organisation_id" gorm:"size:37"`
	StartTime        *time.Time `json:"start_time" csv:"start_time"`
	DueTime          *time.Time `json:"due_time" csv:"due_time"`
	MissTime         *time.Time `json:"miss_time" csv:"miss_time"`
	OccurrenceStatus string     `json:"occurrence_status" csv:"occurrence_status" gorm:"size:20"`
	AuditID          *string    `json:"audit_id" csv:"audit_id" gorm:"size:100"`
	CompletedAt      *time.Time `json:"completed_at" csv:"completed_at"`
	ExportedAt       time.Time  `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
	UserID           string     `json:"user_id" csv:"user_id" gorm:"size:37"`
	AssigneeStatus   string     `json:"assignee_status" csv:"assignee_status" gorm:"size:20"`
}

// ScheduleOccurrenceFeed is a representation of the schedule_occurrences feed
type ScheduleOccurrenceFeed struct {
	TemplateIDs []string
}

// Name is the name of the feed
func (f *ScheduleOccurrenceFeed) Name() string {
	return "schedule_occurrences"
}

// RowsModel returns the model of feed rows
func (f *ScheduleOccurrenceFeed) RowsModel() interface{} {
	return &[]*ScheduleOccurrence{}
}

// Model returns the model of the feed row
func (f *ScheduleOccurrenceFeed) Model() interface{} {
	return ScheduleOccurrence{}
}

// PrimaryKey returns the primary key(s)
func (f *ScheduleOccurrenceFeed) PrimaryKey() []string {
	return []string{"id"}
}

// Columns returns the columns of the row
func (f *ScheduleOccurrenceFeed) Columns() []string {
	return []string{
		"schedule_id",
		"occurrence_id",
		"template_id",
		"organisation_id",
		"start_time",
		"due_time",
		"miss_time",
		"occurrence_status",
		"audit_id",
		"completed_at",
		"user_id",
		"assignee_status",
	}
}

// Order returns the ordering when retrieving an export
func (f *ScheduleOccurrenceFeed) Order() string {
	return "occurrence_id ASC, schedule_id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *ScheduleOccurrenceFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*ScheduleOccurrence{})
}

// Export exports the feed to the supplied exporter
func (f *ScheduleOccurrenceFeed) Export(ctx context.Context, apiClient *api.Client, exporter Exporter, orgID string) error {
	logger := util.GetLogger().With(
		"feed", f.Name(),
		"org_id", orgID,
	)

	logger.Info("exporting")

	exporter.InitFeed(f, &InitFeedOptions{
		// Always truncate. This data must be refreshed in order to be accurate
		Truncate: false,
	})

	err := apiClient.DrainFeed(ctx, &api.GetFeedRequest{
		InitialURL: "/feed/schedule_occurrences",
		Params: api.GetFeedParams{
			TemplateIDs: f.TemplateIDs,
		},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*ScheduleOccurrence{}

		err := json.Unmarshal(resp.Data, &rows)
		util.Check(err, "Failed to unmarshal schedule-occurrences data to struct")

		if len(rows) != 0 {
			// Calculate the size of the batch we can insert into the DB at once. Column count + buffer to account for primary keys
			batchSize := exporter.ParameterLimit() / (len(f.Columns()) + 4)

			for i := 0; i < len(rows); i += batchSize {
				j := i + batchSize
				if j > len(rows) {
					j = len(rows)
				}

				err = exporter.WriteRows(f, rows[i:j])
				util.Check(err, "Failed to write data to exporter")
			}
		}

		logger.With(
			"estimated_remaining", resp.Metadata.RemainingRecords,
			"duration_ms", apiClient.Duration.Milliseconds(),
			"export_duration_ms", exporter.GetDuration().Milliseconds(),
		).Info("export batch complete")
		return nil
	})
	util.Check(err, "Failed to export feed")

	return exporter.FinaliseExport(f, &[]*ScheduleOccurrence{})
}
