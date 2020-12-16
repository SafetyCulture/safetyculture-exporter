package feed

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
)

// Group represents a row from the groups feed
type Group struct {
	ID         string    `json:"id" csv:"group_id" gorm:"primarykey;column:group_id"`
	Name       string    `json:"name" csv:"name"`
	ExportedAt time.Time `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
}

// GroupFeed is a representation of the groups feed
type GroupFeed struct{}

// Name is the name of the feed
func (f *GroupFeed) Name() string {
	return "groups"
}

// Model returns the model of the feed row
func (f *GroupFeed) Model() interface{} {
	return Group{}
}

// RowsModel returns the model of feed rows
func (f *GroupFeed) RowsModel() interface{} {
	return &[]*Group{}
}

// PrimaryKey returns the primary key(s)
func (f *GroupFeed) PrimaryKey() []string {
	return []string{"group_id"}
}

// Columns returns the columns of the row
func (f *GroupFeed) Columns() []string {
	return []string{
		"name",
		"exported_at",
	}
}

// Order returns the ordering when retrieving an export
func (f *GroupFeed) Order() string {
	return "group_id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *GroupFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*Group{})
}

// Export exports the feed to the supplied exporter
func (f *GroupFeed) Export(ctx context.Context, apiClient api.Client, exporter Exporter) error {
	logger := util.GetLogger()
	feedName := f.Name()

	logger.Infof("%s: exporting", feedName)

	exporter.InitFeed(f, &InitFeedOptions{
		// Truncate files if upserts aren't supported.
		// This ensure that the export does not contain duplicate rows
		Truncate: exporter.SupportsUpsert() == false,
	})

	err := apiClient.DrainFeed(ctx, &api.GetFeedRequest{
		InitialURL: "/feed/groups",
		Params:     api.GetFeedParams{},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*Group{}

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

	return exporter.FinaliseExport(f, &[]*Group{})
}
