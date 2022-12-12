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

// ScheduleAssignee represents a row from the schedule_assignees feed
type ScheduleAssignee struct {
	ID             string    `json:"id" csv:"id" gorm:"primarykey;size:100"`
	ScheduleID     string    `json:"schedule_id" csv:"schedule_id" gorm:"size:45"`
	AssigneeID     string    `json:"assignee_id" csv:"assignee_id" gorm:"size:37"`
	OrganisationID string    `json:"organisation_id" csv:"organisation_id" gorm:"size:37"`
	Type           string    `json:"type" csv:"type" gorm:"size:10"`
	Name           string    `json:"name" csv:"name"`
	ExportedAt     time.Time `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
}

// ScheduleAssigneeFeed is a representation of the schedule_assignees feed
type ScheduleAssigneeFeed struct {
	TemplateIDs []string
}

// Name is the name of the feed
func (f *ScheduleAssigneeFeed) Name() string {
	return "schedule_assignees"
}

// Model returns the model of the feed row
func (f *ScheduleAssigneeFeed) Model() interface{} {
	return ScheduleAssignee{}
}

// RowsModel returns the model of feed rows
func (f *ScheduleAssigneeFeed) RowsModel() interface{} {
	return &[]*ScheduleAssignee{}
}

// PrimaryKey returns the primary key(s)
func (f *ScheduleAssigneeFeed) PrimaryKey() []string {
	return []string{"id"}
}

// Columns returns the columns of the row
func (f *ScheduleAssigneeFeed) Columns() []string {
	return []string{
		"schedule_id",
		"assignee_id",
		"organisation_id",
		"type",
		"name",
		"exported_at",
	}
}

// Order returns the ordering when retrieving an export
func (f *ScheduleAssigneeFeed) Order() string {
	return "schedule_id, assignee_id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *ScheduleAssigneeFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*ScheduleAssignee{})
}

// Export exports the feed to the supplied exporter
func (f *ScheduleAssigneeFeed) Export(ctx context.Context, apiClient *httpapi.Client, exporter Exporter, orgID string) error {
	logger := logger.GetLogger().With("feed", f.Name(), "org_id", orgID)

	if err := exporter.InitFeed(f, &InitFeedOptions{
		// Always truncate. This data must be refreshed in order to be accurate
		Truncate: true,
	}); err != nil {
		return events.WrapEventError(err, "init feed")
	}

	drainFn := func(resp *GetFeedResponse) error {
		var rows []*ScheduleAssignee

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
		InitialURL: "/feed/schedule_assignees",
		Params: GetFeedParams{
			TemplateIDs: f.TemplateIDs,
		},
	}

	if err := DrainFeed(ctx, apiClient, req, drainFn); err != nil {
		return events.WrapEventError(err, fmt.Sprintf("feed %q", f.Name()))
	}
	return exporter.FinaliseExport(f, &[]*ScheduleAssignee{})
}
