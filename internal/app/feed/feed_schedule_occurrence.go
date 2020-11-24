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
	ID               string     `json:"id" csv:"id" gorm:"primarykey"`
	ScheduleID       string     `json:"schedule_id" csv:"schedule_id"`
	OccurrenceID     string     `json:"occurrence_id" csv:"occurrence_id"`
	TemplateID       string     `json:"template_id" csv:"template_id"`
	MissTime         time.Time  `json:"miss_time" csv:"miss_time"`
	OccurrenceStatus string     `json:"occurrence_status" csv:"occurrence_status"`
	AuditID          *string    `json:"audit_id" csv:"audit_id"`
	CompletedAt      *time.Time `json:"completed_at" csv:"completed_at"`
	ExportedAt       time.Time  `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
	UserID           string     `json:"user_id" csv:"user_id"`
	AssigneeStatus   string     `json:"assignee_status" csv:"assignee_status"`
}

// ScheduleOccurrenceFeed is a representation of the schedule_occurrences feed
type ScheduleOccurrenceFeed struct {
	TemplateIDs []string
}

// Name is the name of the feed
func (f *ScheduleOccurrenceFeed) Name() string {
	return "schedule_occurrences"
}

// Model returns the model of feed rows
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

// Create schema of the feed for the supplied exporter
func (f *ScheduleOccurrenceFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*ScheduleOccurrence{})
}

func (f *ScheduleOccurrenceFeed) writeRows(exporter Exporter, rows []*ScheduleOccurrence) error {
	// DB parameters are limited to 1000 params per query.
	// Limit the batch size to prevent queries from failing
	batchSize := 1000
	for i := 0; i < len(rows); i += batchSize {
		j := i + batchSize
		if j > len(rows) {
			j = len(rows)
		}

		return exporter.WriteRows(f, rows[i:j])
	}

	return nil
}

// Export exports the feed to the supplied exporter
func (f *ScheduleOccurrenceFeed) Export(ctx context.Context, apiClient api.APIClient, exporter Exporter) error {
	logger := util.GetLogger()
	feedName := f.Name()

	logger.Infof("%s: exporting", feedName)

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
		util.Check(err, "Failed to unmarshal data to struct")

		if len(rows) != 0 {
			err = f.writeRows(exporter, rows)
			util.Check(err, "Failed to write data to exporter")
		}

		logger.Infof("%s: %d remaining", feedName, resp.Metadata.RemainingRecords)
		return nil
	})
	util.Check(err, "Failed to export feed")

	return exporter.FinaliseExport(f, &[]*ScheduleOccurrence{})
}
