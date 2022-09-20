package feed

import (
	"context"
	"encoding/json"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
	"time"
)

const (
	feedName = "issues"
	feedURL  = "/feed/issues"
)

// Issue represents a row from the issues feed
type Issue struct {
	ID              string     `json:"id" csv:"id" gorm:"primarykey;column:id;size:36"`
	Title           string     `json:"title" csv:"title"`
	Description     string     `json:"description" csv:"description"`
	CreatorID       string     `json:"creator_id" csv:"creator_id"`
	CreatorUserName string     `json:"creator_user_name" csv:"creator_user_name"`
	CreatedAt       time.Time  `json:"created_at" csv:"created_at"`
	DueAt           *time.Time `json:"due_at,omitempty" csv:"due_at"`
	Priority        string     `json:"priority" csv:"priority"`
	Status          string     `json:"status" csv:"status"`
	TemplateID      string     `json:"template_id" csv:"template_id"`
	InspectionID    string     `json:"inspection_id" csv:"inspection_id"`
	InspectionName  string     `json:"inspection_name" csv:"inspection_name"`
	SiteID          string     `json:"site_id" csv:"site_id"`
	SiteName        string     `json:"site_name" csv:"site_name"`
	LocationName    string     `json:"location_name" csv:"location_name"`
	CategoryID      string     `json:"category_id" csv:"category_id"`
	CategoryLabel   string     `json:"category_label" csv:"category_label"`
	ModifiedAt      time.Time  `json:"modified_at" csv:"modified_at"`
}

// IssueFeed is a representation of the issues feed
type IssueFeed struct {
	Limit       int
	Incremental bool
}

// Name returns the name of the feed
func (f *IssueFeed) Name() string {
	return feedName
}

// Model returns the model of the feed row
func (f *IssueFeed) Model() interface{} {
	return Issue{}
}

// RowsModel returns the model of the feed rows
func (f *IssueFeed) RowsModel() interface{} {
	return &[]*Issue{}
}

// PrimaryKey return the primary key
func (f *IssueFeed) PrimaryKey() []string {
	return []string{"id"}
}

// Columns returns the columns of the row
func (f *IssueFeed) Columns() []string {
	return []string{
		"id", "title", "description", "creator_id", "creator_user_name",
		"created_at", "due_at", "priority", "status", "template_id",
		"inspection_id", "inspection_name", "site_id", "site_name",
		"location_name", "category_id", "category_label", "modified_at",
	}
}

// Order returns the ordering when retrieving an export
func (f *IssueFeed) Order() string {
	return "id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *IssueFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*Issue{})
}

// Export exports the feed to the supplied exporter
func (f *IssueFeed) Export(ctx context.Context, apiClient *api.Client, exporter Exporter, orgID string) error {
	logger := util.GetLogger()

	_ = exporter.InitFeed(f, &InitFeedOptions{
		// Delete data if incremental refresh is disabled so there is no duplicates
		Truncate: !f.Incremental,
	})

	var request = &api.GetFeedRequest{
		InitialURL: feedURL,
		Params: api.GetFeedParams{
			Limit: f.Limit,
		},
	}

	var feedFn = func(resp *api.GetFeedResponse) error {
		var rows []*Issue
		err := json.Unmarshal(resp.Data, &rows)
		util.Check(err, "Failed to unmarshal actions data to struct")

		if len(rows) != 0 {
			// Calculate the size of the batch we can insert into the DB at once.
			// Column count + buffer to account for primary keys
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

		logger.Infof("%s: %d remaining. Last call was %dms", f.Name(), resp.Metadata.RemainingRecords, apiClient.Duration.Milliseconds())
		return nil
	}

	err := apiClient.DrainFeed(ctx, request, feedFn)
	util.Check(err, "Failed to export feed")
	return exporter.FinaliseExport(f, &[]*Issue{})
}
