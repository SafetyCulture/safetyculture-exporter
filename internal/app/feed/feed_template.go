package feed

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
)

// Template represents a row from the templates feed
type Template struct {
	ID          string    `json:"id" csv:"template_id" gorm:"primarykey;column:template_id"`
	Archived    bool      `json:"archived" csv:"archived"`
	Name        string    `json:"name" csv:"name"`
	Description string    `json:"description" csv:"description"`
	OwnerName   string    `json:"owner_name" csv:"owner_name"`
	OwnerID     string    `json:"owner_id" csv:"owner_id"`
	AuthorName  string    `json:"author_name" csv:"author_name"`
	AuthorID    string    `json:"author_id" csv:"author_id"`
	CreatedAt   time.Time `json:"created_at" csv:"created_at"`
	ModifiedAt  time.Time `json:"modified_at" csv:"modified_at"`
	ExportedAt  time.Time `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
}

// TemplateFeed is a representation of the templates feed
type TemplateFeed struct {
	ModifiedAfter string
	Incremental   bool
}

// Name is the name of the feed
func (f *TemplateFeed) Name() string {
	return "templates"
}

// Model returns the model of the feed row
func (f *TemplateFeed) Model() interface{} {
	return Template{}
}

// PrimaryKey returns the primary key(s)
func (f *TemplateFeed) PrimaryKey() []string {
	return []string{"template_id"}
}

// Columns returns the columns of the row
func (f *TemplateFeed) Columns() []string {
	return []string{
		"archived",
		"name",
		"description",
		"owner_name",
		"owner_id",
		"author_name",
		"author_id",
		"created_at",
		"modified_at",
		"exported_at",
	}
}

// Order returns the ordering when retrieving an export
func (f *TemplateFeed) Order() string {
	return "modified_at ASC, template_id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *TemplateFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*Template{})
}

// Export exports the feed to the supplied exporter
func (f *TemplateFeed) Export(ctx context.Context, apiClient api.APIClient, exporter Exporter) error {
	logger := util.GetLogger()
	feedName := f.Name()

	logger.Infof("%s: exporting", feedName)

	exporter.InitFeed(f, &InitFeedOptions{
		// Delete data if incremental refresh is disabled so there is no duplicates
		Truncate: f.Incremental == false,
	})

	lastModifiedAt, err := exporter.LastModifiedAt(f)
	util.Check(err, "unable to load modified after")
	if lastModifiedAt != nil {
		f.ModifiedAfter = lastModifiedAt.Format(time.RFC3339Nano)
	}

	err = apiClient.DrainFeed(ctx, &api.GetFeedRequest{
		InitialURL: "/feed/templates",
		Params: api.GetFeedParams{
			ModifiedAfter: f.ModifiedAfter,
		},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*Template{}

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

	return exporter.FinaliseExport(f, &[]*Template{})
}
