package feed

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
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
		for k := i; k < (i+j) || k < len(rows); k++ {
			actionIDs = append(actionIDs, rows[k].ActionID)
		}

		// Delete the actions if already exist
		err := exporter.DeleteRowsIfExist(f, "action_id IN ?", actionIDs)
		util.Check(err, "Failed to delete rows in exporter")

		err = exporter.WriteRows(f, rows[i:j])
		util.Check(err, "Failed to write data to exporter")
	}

	return nil
}

// Export exports the feed to the supplied exporter
func (f *ActionAssigneeFeed) Export(ctx context.Context, apiClient *api.Client, exporter Exporter, orgID string) error {
	logger := util.GetLogger()
	feedName := f.Name()

	exporter.InitFeed(f, &InitFeedOptions{
		// Delete data if incremental refresh is disabled so there is no duplicates
		Truncate: !f.Incremental,
	})

	var err error
	f.ModifiedAfter, err = exporter.LastModifiedAt(f, f.ModifiedAfter, orgID)
	util.Check(err, "unable to load modified after")

	logger.Infof("%s: exporting for org_id: %s since: %s", feedName, orgID, f.ModifiedAfter.Format(time.RFC1123))

	err = apiClient.DrainFeed(ctx, &api.GetFeedRequest{
		InitialURL: "/feed/action_assignees",
		Params: api.GetFeedParams{
			ModifiedAfter: f.ModifiedAfter,
		},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*ActionAssignee{}

		err := json.Unmarshal(resp.Data, &rows)
		util.Check(err, "Failed to unmarshal action-assignees data to struct")

		if len(rows) != 0 {
			err = f.writeRows(ctx, exporter, rows)
			util.Check(err, "Failed to write data to exporter")
		}

		logger.Infof("%s: %d remaining", feedName, resp.Metadata.RemainingRecords)
		return nil
	})
	util.Check(err, "Failed to export feed")

	return exporter.FinaliseExport(f, &[]*ActionAssignee{})
}
