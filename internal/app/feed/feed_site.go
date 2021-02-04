package feed

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
)

// Site represents a row from the sites feed
type Site struct {
	ID             string    `json:"id" csv:"site_id" gorm:"primarykey;column:site_id"`
	Name           string    `json:"name" csv:"name"`
	CreatorID      string    `json:"creator_id" csv:"creator_id"`
	OrganisationID string    `json:"organisation_id" csv:"organisation_id"`
	ExportedAt     time.Time `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
	Deleted        bool      `json:"deleted" csv:"deleted" gorm:"deleted"`
}

// SiteFeed is a representation of the sites feed
type SiteFeed struct {
	IncludeDeleted bool
}

// Name is the name of the feed
func (f *SiteFeed) Name() string {
	return "sites"
}

// Model returns the model of the feed row
func (f *SiteFeed) Model() interface{} {
	return Site{}
}

// RowsModel returns the model of feed rows
func (f *SiteFeed) RowsModel() interface{} {
	return &[]*Site{}
}

// PrimaryKey returns the primary key(s)
func (f *SiteFeed) PrimaryKey() []string {
	return []string{"site_id"}
}

// Columns returns the columns of the row
func (f *SiteFeed) Columns() []string {
	return []string{
		"name",
		"creator_id",
		"organisation_id",
		"exported_at",
	}
}

// Order returns the ordering when retrieving an export
func (f *SiteFeed) Order() string {
	return "site_id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *SiteFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*Site{})
}

// Export exports the feed to the supplied exporter
func (f *SiteFeed) Export(ctx context.Context, apiClient api.Client, exporter Exporter) error {
	logger := util.GetLogger()
	feedName := f.Name()

	logger.Infof("%s: exporting", feedName)

	exporter.InitFeed(f, &InitFeedOptions{
		// Truncate files if upserts aren't supported.
		// This ensure that the export does not contain duplicate rows
		Truncate: exporter.SupportsUpsert() == false,
	})

	err := apiClient.DrainFeed(ctx, &api.GetFeedRequest{
		InitialURL: "/feed/sites",
		Params: api.GetFeedParams{
			IncludeDeleted: f.IncludeDeleted,
		},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*Site{}

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

	return exporter.FinaliseExport(f, &[]*Site{})
}
