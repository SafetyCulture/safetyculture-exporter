package feed

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
)

// SiteMember represents a row from the site members feed
type SiteMember struct {
	SiteID     string    `json:"site_id" csv:"site_id" gorm:"primarykey;column:site_id;size:41"`
	MemberID   string    `json:"member_id" csv:"member_id" gorm:"primarykey;column:member_id;size:37"`
	ExportedAt time.Time `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
}

// SiteMemberFeed is a representation of the sites feed
type SiteMemberFeed struct {
}

// Name is the name of the feed
func (f *SiteMemberFeed) Name() string {
	return "site_members"
}

// Model returns the model of the feed row
func (f *SiteMemberFeed) Model() interface{} {
	return SiteMember{}
}

// RowsModel returns the model of feed rows
func (f *SiteMemberFeed) RowsModel() interface{} {
	return &[]*SiteMember{}
}

// PrimaryKey returns the primary key(s)
func (f *SiteMemberFeed) PrimaryKey() []string {
	return []string{"site_id", "member_id"}
}

// Columns returns the columns of the row
func (f *SiteMemberFeed) Columns() []string {
	return []string{"exported_at"}
}

// Order returns the ordering when retrieving an export
func (f *SiteMemberFeed) Order() string {
	return "site_id,member_id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *SiteMemberFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*SiteMember{})
}

// Export exports the feed to the supplied exporter
func (f *SiteMemberFeed) Export(ctx context.Context, apiClient *api.Client, exporter Exporter, orgID string) error {
	logger := util.GetLogger()
	feedName := f.Name()

	logger.Infof("%s: exporting for org_id: %s", feedName, orgID)

	exporter.InitFeed(f, &InitFeedOptions{
		// Truncate files if upserts aren't supported.
		// This ensure that the export does not contain duplicate rows
		Truncate: !exporter.SupportsUpsert(),
	})

	err := apiClient.DrainFeed(ctx, &api.GetFeedRequest{
		InitialURL: "/feed/site_members",
	}, func(resp *api.GetFeedResponse) error {
		rows := []*SiteMember{}

		err := json.Unmarshal(resp.Data, &rows)
		util.Check(err, "Failed to unmarshal sites data to struct")

		if len(rows) != 0 {
			// Calculate the size of the batch we can insert into the DB at once. Column count + buffer to account for primary keys
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

		logger.Infof("%s: %d remaining. Last call was %dms", feedName, resp.Metadata.RemainingRecords, apiClient.Duration.Milliseconds())
		return nil
	})
	util.Check(err, "Failed to export feed")

	return exporter.FinaliseExport(f, &[]*SiteMember{})
}
