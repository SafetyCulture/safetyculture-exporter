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
}

// SiteFeed is a representation of the sites feed
type SiteFeed struct{}

// Name is the name of the feed
func (f *SiteFeed) Name() string {
	return "sites"
}

// Model returns the model of the feed row
func (f *SiteFeed) Model() interface{} {
	return Site{}
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

// Create schema of the feed for the supplied exporter
func (f *SiteFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*Site{})
}

// Export exports the feed to the supplied exporter
func (f *SiteFeed) Export(ctx context.Context, apiClient api.APIClient, exporter Exporter) error {
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
		Params:     api.GetFeedParams{},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*Site{}

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

	return exporter.FinaliseExport(f, &[]*Site{})
}
