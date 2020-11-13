package feed

import (
	"context"
	"encoding/json"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
)

// ActionAssignee represents a row from the action_assignees feed
type ActionAssignee struct {
	ID         string `json:"id" csv:"id" gorm:"primarykey"`
	ActionID   string `json:"action_id" csv:"action_id"`
	AssigneeID string `json:"assignee_id" csv:"assignee_id"`
	Type       string `json:"type" csv:"type"`
	Name       string `json:"name" csv:"name"`
}

// ActionAssigneeFeed is a representation of the action_assignees feed
type ActionAssigneeFeed struct{}

// Name is the name of the feed
func (f *ActionAssigneeFeed) Name() string {
	return "action_assignees"
}

// Model returns the model of the feed row
func (f *ActionAssigneeFeed) Model() interface{} {
	return ActionAssignee{}
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
	}
}

// Order returns the ordering when retrieving an export
func (f *ActionAssigneeFeed) Order() string {
	return "action_id, assignee_id"
}

// Export exports the feed to the supplied exporter
func (f *ActionAssigneeFeed) Export(ctx context.Context, apiClient api.APIClient, exporter Exporter) error {
	logger := util.GetLogger()
	feedName := f.Name()

	logger.Infof("%s: exporting", feedName)

	exporter.InitFeed(f, &InitFeedOptions{
		// Truncate files if upserts aren't supported.
		// This ensure that the export does not contain duplicate rows
		Truncate: exporter.SupportsUpsert() == false,
	})

	err := apiClient.DrainFeed(ctx, &api.GetFeedRequest{
		InitialURL: "/feed/action_assignees",
		Params:     api.GetFeedParams{},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*ActionAssignee{}

		err := json.Unmarshal(resp.Data, &rows)
		util.Check(err, "Failed to unmarshal data to struct")

		if len(rows) != 0 {
			err = exporter.WriteRows(f, rows)
			util.Check(err, "Failed to write data to exporter")
		}

		logger.Infof("%s: %d remaining", feedName, resp.Metadata.RemainingRecords)
		return nil
	})
	util.Check(err, "Failed to export feed")

	return exporter.FinaliseExport(f, &[]*ActionAssignee{})
}
