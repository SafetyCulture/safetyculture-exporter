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

// ActionAssignee represents a row from the action_assignees feed
type ActionAssignee struct {
	ID             string    `json:"id" csv:"id" gorm:"primarykey;size:375"`
	ActionID       string    `json:"action_id" csv:"action_id" gorm:"index:idx_act_action_id;size:36"`
	AssigneeID     string    `json:"assignee_id" csv:"assignee_id" gorm:"size:256"`
	Type           string    `json:"type" csv:"type" gorm:"size:10"`
	Name           string    `json:"name" csv:"name"`
	OrganisationID string    `json:"organisation_id" csv:"organisation_id" gorm:"index:idx_act_asg_modified_at;size:37"`
	ModifiedAt     time.Time `json:"modified_at" csv:"modified_at" gorm:"index:idx_act_asg_modified_at,sort:desc"`
	ExportedAt     time.Time `json:"exported_at" csv:"exported_at" gorm:"index:idx_act_asg_modified_at;autoUpdateTime"`
}

// ActionAssigneeFeed is a representation of the action_assignees feed
type ActionAssigneeFeed struct {
	ModifiedAfter time.Time
	Incremental   bool
}

// Name is the name of the feed
func (f *ActionAssigneeFeed) Name() string {
	return "action_assignees"
}

// Model returns the model of the feed row
func (f *ActionAssigneeFeed) Model() interface{} {
	return ActionAssignee{}
}

// RowsModel returns the model of feed rows
func (f *ActionAssigneeFeed) RowsModel() interface{} {
	return &[]*ActionAssignee{}
}

// PrimaryKey returns the primary key(s)
func (f *ActionAssigneeFeed) PrimaryKey() []string {
	return []string{"id"}
}

// Columns returns the columns of the row
func (f *ActionAssigneeFeed) Columns() []string {
	return []string{
		"action_id",
		"assignee_id",
		"type",
		"name",
		"organisation_id",
		"modified_at",
		"exported_at",
	}
}

// Order returns the ordering when retrieving an export
func (f *ActionAssigneeFeed) Order() string {
	return "action_id, assignee_id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *ActionAssigneeFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*ActionAssignee{})
}

func (f *ActionAssigneeFeed) writeRows(ctx context.Context, exporter Exporter, rows []*ActionAssignee) error {
	// Calculate the size of the batch we can insert into the DB at once. Column count + buffer to account for primary keys
	batchSize := exporter.ParameterLimit() / (len(f.Columns()) + 4)

	for i := 0; i < len(rows); i += batchSize {
		j := i + batchSize
		if j > len(rows) {
			j = len(rows)
		}
		var actionIDs []string
		for k := range rows[i:j] {
			actionIDs = append(actionIDs, rows[k].ActionID)
		}

		// Delete the actions if already exist
		if err := exporter.DeleteRowsIfExist(f, "action_id IN ?", actionIDs); err != nil {
			return fmt.Errorf("delete rows: %w", err)
		}

		if err := exporter.WriteRows(f, rows[i:j]); err != nil {
			return events.WrapEventError(err, "write rows")
		}
	}

	return nil
}

// Export exports the feed to the supplied exporter
func (f *ActionAssigneeFeed) Export(ctx context.Context, apiClient *httpapi.Client, exporter Exporter, orgID string) error {
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
		var rows []*ActionAssignee

		if err := json.Unmarshal(resp.Data, &rows); err != nil {
			return events.NewEventErrorWithMessage(err, events.ErrorSeverityError, events.ErrorSubSystemDataIntegrity, false, "map data")
		}

		if len(rows) != 0 {
			if err := f.writeRows(ctx, exporter, rows); err != nil {
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

	req := &GetFeedRequest{
		InitialURL: "/feed/action_assignees",
		Params: GetFeedParams{
			ModifiedAfter: f.ModifiedAfter,
		},
	}

	if err := DrainFeed(ctx, apiClient, req, drainFn); err != nil {
		return events.WrapEventError(err, fmt.Sprintf("feed %q", f.Name()))
	}
	return exporter.FinaliseExport(f, &[]*ActionAssignee{})
}
