package feed

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
)

// InspectionItem represents a row from the inspection_items feed
type InspectionItem struct {
	ID                      string    `json:"id" csv:"id" gorm:"primarykey"`
	ItemID                  string    `json:"item_id" csv:"item_id"`
	AuditID                 string    `json:"audit_id" csv:"audit_id"`
	TemplateID              string    `json:"template_id" csv:"template_id"`
	ParentID                string    `json:"parent_id" csv:"parent_id"`
	CreatedAt               time.Time `json:"created_at" csv:"created_at"`
	ModifiedAt              time.Time `json:"modified_at" csv:"modified_at"`
	ExportedAt              time.Time `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
	Type                    string    `json:"type" csv:"type"`
	Category                string    `json:"category" csv:"category"`
	CategoryID              string    `json:"category_id" csv:"category_id"`
	ParentIDs               string    `json:"parent_ids" csv:"parent_ids"`
	Label                   string    `json:"label" csv:"label"`
	Response                string    `json:"response" csv:"response"`
	ResponseID              string    `json:"response_id" csv:"response_id"`
	ResponseSetID           string    `json:"response_set_id" csv:"response_set_id"`
	IsFailedResponse        bool      `json:"is_failed_response" csv:"is_failed_response"`
	Comment                 string    `json:"comment" csv:"comment"`
	MediaHypertextReference string    `json:"media_hypertext_reference" csv:"media_hypertext_reference"`
	Score                   float32   `json:"score" csv:"score"`
	MaxScore                float32   `json:"max_score" csv:"max_score"`
	ScorePercentage         float32   `json:"score_percentage" csv:"score_percentage"`
	Mandatory               bool      `json:"mandatory" csv:"mandatory"`
	Inactive                bool      `json:"inactive" csv:"inactive"`
	LocationLatitude        *float32  `json:"location_latitude" csv:"location_latitude"`
	LocationLongitude       *float32  `json:"location_longitude" csv:"location_longitude"`
}

// InspectionItemFeed is a representation of the inspection_items feed
type InspectionItemFeed struct {
	SkipIDs         []string
	ModifiedAfter   string
	TemplateIDs     []string
	Archived        string
	Completed       string
	IncludeInactive bool
	Incremental     bool
}

// Name is the name of the feed
func (f *InspectionItemFeed) Name() string {
	return "inspection_items"
}

// Model returns the model of the feed row
func (f *InspectionItemFeed) Model() interface{} {
	return InspectionItem{}
}

// PrimaryKey returns the primary key(s)
func (f *InspectionItemFeed) PrimaryKey() []string {
	return []string{"id"}
}

// Columns returns the columns of the row
func (f *InspectionItemFeed) Columns() []string {
	return []string{
		"template_id",
		"parent_id",
		"created_at",
		"modified_at",
		"type",
		"category",
		"category_id",
		"parent_ids",
		"label",
		"response",
		"response_id",
		"response_set_id",
		"is_failed_response",
		"comment",
		"media_hypertext_reference",
		"score",
		"max_score",
		"score_percentage",
		"mandatory",
		"inactive",
		"location_latitude",
		"location_longitude",
	}
}

// Order returns the ordering when retrieving an export
func (f *InspectionItemFeed) Order() string {
	return "modified_at ASC, id"
}

func (f *InspectionItemFeed) writeRows(exporter Exporter, rows []*InspectionItem) error {
	skipIDs := map[string]bool{}
	for _, id := range f.SkipIDs {
		skipIDs[id] = true
	}

	// Calculate the size of the batch we can insert into the DB at once. Column count + buffer
	batchSize := exporter.ParameterLimit() / (len(f.Columns()) + 4)
	for i := 0; i < len(rows); i += batchSize {
		j := i + batchSize
		if j > len(rows) {
			j = len(rows)
		}

		// Some audits in production have the same item ID multiple times
		// We can't insert them simultaneously. This means we are dropping data, which sucks.
		rowsToInsert := []*InspectionItem{}
		idSeen := map[string]bool{}
		for _, row := range rows[i:j] {
			skip := skipIDs[row.AuditID]
			seen := idSeen[row.ID]
			if !seen && !skip {
				idSeen[row.ID] = true
				rowsToInsert = append(rowsToInsert, row)
			}
		}

		return exporter.WriteRows(f, rowsToInsert)
	}

	return nil
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *InspectionItemFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*InspectionItem{})
}

// Export exports the feed to the supplied exporter
func (f *InspectionItemFeed) Export(ctx context.Context, apiClient api.APIClient, exporter Exporter) error {
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
		InitialURL: "/feed/inspection_items",
		Params: api.GetFeedParams{
			ModifiedAfter:   f.ModifiedAfter,
			TemplateIDs:     f.TemplateIDs,
			Archived:        f.Archived,
			Completed:       f.Completed,
			IncludeInactive: f.IncludeInactive,
		},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*InspectionItem{}

		err := json.Unmarshal(resp.Data, &rows)
		util.Check(err, "Failed to unmarshal data to struct")

		if len(rows) != 0 {
			err = f.writeRows(exporter, rows)
			util.Check(err, "Failed to write data to exporter")
		}

		logger.Infof("%s: %d remaining", feedName, resp.Metadata.RemainingRecords)
		return nil
	})
	util.Check(err, "Failed to export feed")

	return exporter.FinaliseExport(f, &[]*InspectionItem{})
}
