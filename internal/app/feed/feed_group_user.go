package feed

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
)

// GroupUser represents a row from the group_users feed
type GroupUser struct {
	UserID     string    `json:"user_id" csv:"user_id" gorm:"primaryKey"`
	GroupID    string    `json:"group_id" csv:"group_id" gorm:"primaryKey"`
	ExportedAt time.Time `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
}

// GroupUserFeed is a representation of the group_users feed
type GroupUserFeed struct{}

// Name is the name of the feed
func (f *GroupUserFeed) Name() string {
	return "group_users"
}

// Model returns the model of the feed row
func (f *GroupUserFeed) Model() interface{} {
	return GroupUser{}
}

// PrimaryKey returns the primary key(s)
func (f *GroupUserFeed) PrimaryKey() []string {
	return []string{"user_id", "group_id"}
}

// Columns returns the columns of the row
func (f *GroupUserFeed) Columns() []string {
	return []string{
		"user_id",
		"group_id",
		"exported_at",
	}
}

// Order returns the ordering when retrieving an export
func (f *GroupUserFeed) Order() string {
	return "group_id, user_id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *GroupUserFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*GroupUser{})
}

// Export exports the feed to the supplied exporter
func (f *GroupUserFeed) Export(ctx context.Context, apiClient api.APIClient, exporter Exporter) error {
	logger := util.GetLogger()
	feedName := f.Name()

	logger.Infof("%s: exporting", feedName)

	exporter.InitFeed(f, &InitFeedOptions{
		// Clear this table before loading data.
		// This table does not receive updates, it is only refreshed.
		Truncate: true,
	})

	err := apiClient.DrainFeed(ctx, &api.GetFeedRequest{
		InitialURL: "/feed/group_users",
		Params:     api.GetFeedParams{},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*GroupUser{}

		err := json.Unmarshal(resp.Data, &rows)
		util.Check(err, "Failed to unmarshal data to struct")

		if len(rows) != 0 {
			// Calculate the size of the batch we can insert into the DB at once. Column count + buffer
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

		logger.Infof("%s: %d remaining", feedName, resp.Metadata.RemainingRecords)
		return nil
	})
	util.Check(err, "Failed to export feed")

	return exporter.FinaliseExport(f, &[]*GroupUser{})
}
