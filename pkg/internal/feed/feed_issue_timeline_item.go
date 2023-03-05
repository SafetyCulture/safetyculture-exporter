package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
)

// IssueTimelineItem represents a row from the issue timeline items feed
type IssueTimelineItem struct {
	ID              string    `json:"id" csv:"item_id" gorm:"primarykey;column:item_id;size:36"`
	TaskID          string    `json:"task_id" csv:"task_id"`
	OrganisationID  string    `json:"organisation_id" csv:"organisation_id"`
	TaskCreatorID   string    `json:"task_creator_id" csv:"task_creator_id"`
	TaskCreatorName string    `json:"task_creator_name" csv:"task_creator_name"`
	Timestamp       time.Time `json:"timestamp" csv:"timestamp" gorm:"index:idx_iss_tim_timestamp,sort:desc"`
	CreatorID       string    `json:"creator_id" csv:"creator_id"`
	CreatorName     string    `json:"creator_name" csv:"creator_name"`
	ItemType        string    `json:"item_type" csv:"item_type"`
	ItemData        string    `json:"item_data" csv:"item_data"`
}

// IssueTimelineItemFeed is a representation of the issue timeline items feed
type IssueTimelineItemFeed struct {
	Limit       int
	Incremental bool
}

// Name is the name of the feed
func (f *IssueTimelineItemFeed) Name() string {
	return "issue_timeline_items"
}

// Model returns the model of the feed row
func (f *IssueTimelineItemFeed) Model() interface{} {
	return IssueTimelineItem{}
}

// RowsModel returns the model of feed rows
func (f *IssueTimelineItemFeed) RowsModel() interface{} {
	return &[]*IssueTimelineItem{}
}

// PrimaryKey returns the primary key(s)
func (f *IssueTimelineItemFeed) PrimaryKey() []string {
	return []string{"item_id"}
}

// Columns returns the columns of the row
func (f *IssueTimelineItemFeed) Columns() []string {
	return []string{
		"task_id",
		"organisation_id",
		"task_creator_id",
		"task_creator_name",
		"timestamp",
		"creator_id",
		"creator_name",
		"item_type",
		"item_data",
	}
}

// Order returns the ordering when retrieving an export
func (f *IssueTimelineItemFeed) Order() string {
	return "item_id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *IssueTimelineItemFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*IssueTimelineItem{})
}

// Export exports the feed to the supplied exporter
func (f *IssueTimelineItemFeed) Export(ctx context.Context, apiClient *httpapi.Client, exporter Exporter, orgID string) error {
	l := logger.GetLogger().With("feed", f.Name(), "org_id", orgID)
	status := GetExporterStatus()

	if err := exporter.InitFeed(f, &InitFeedOptions{
		// Delete data if incremental refresh is disabled so there is no duplicates
		Truncate: !f.Incremental,
	}); err != nil {
		return fmt.Errorf("init feed: %w", err)
	}

	drainFn := func(resp *GetFeedResponse) error {
		var rows []*IssueTimelineItem

		if err := json.Unmarshal(resp.Data, &rows); err != nil {
			return fmt.Errorf("map data: %w", err)
		}

		if len(rows) != 0 {
			// Calculate the size of the batch we can insert into the DB at once.
			// Column count + buffer to account for primary keys
			batchSize := exporter.ParameterLimit() / (len(f.Columns()) + 4)

			for i := 0; i < len(rows); i += batchSize {
				j := i + batchSize
				if j > len(rows) {
					j = len(rows)
				}

				if err := exporter.WriteRows(f, rows[i:j]); err != nil {
					return fmt.Errorf("exporter: %w", err)
				}
			}
		}

		status.UpdateStatus(f.Name(), resp.Metadata.RemainingRecords, exporter.GetDuration().Milliseconds())

		l.With(
			"estimated_remaining", resp.Metadata.RemainingRecords,
			"duration_ms", apiClient.Duration.Milliseconds(),
			"export_duration_ms", exporter.GetDuration().Milliseconds(),
		).Info("export batch complete")
		return nil
	}

	req := &GetFeedRequest{
		InitialURL: "/feed/issue_timeline_items",
		Params: GetFeedParams{
			Limit: f.Limit,
		},
	}

	if err := DrainFeed(ctx, apiClient, req, drainFn); err != nil {
		return fmt.Errorf("feed %q: %w", f.Name(), err)
	}
	return exporter.FinaliseExport(f, &[]*IssueTimelineItem{})
}
