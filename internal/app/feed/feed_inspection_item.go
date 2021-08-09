package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
)

const maxGoRoutines = 10

// InspectionItem represents a row from the inspection_items feed
type InspectionItem struct {
	ID                      string    `json:"id" csv:"id" gorm:"primarykey;size:150"`
	ItemID                  string    `json:"item_id" csv:"item_id" gorm:"size:100"`
	AuditID                 string    `json:"audit_id" csv:"audit_id" gorm:"size:100"`
	ItemIndex               int64     `json:"item_index" csv:"item_index"`
	TemplateID              string    `json:"template_id" csv:"template_id" gorm:"size:100"`
	ParentID                string    `json:"parent_id" csv:"parent_id" gorm:"size:100"`
	CreatedAt               time.Time `json:"created_at" csv:"created_at"`
	ModifiedAt              time.Time `json:"modified_at" csv:"modified_at" gorm:"index:idx_ins_itm_modified_at,sort:desc"`
	ExportedAt              time.Time `json:"exported_at" csv:"exported_at" gorm:"index:idx_ins_itm_modified_at;autoUpdateTime"`
	Type                    string    `json:"type" csv:"type" gorm:"size:20"`
	Category                string    `json:"category" csv:"category"`
	CategoryID              string    `json:"category_id" csv:"category_id" gorm:"size:100"`
	OrganisationID          string    `json:"organisation_id" csv:"organisation_id" gorm:"index:idx_ins_itm_modified_at;size:37"`
	ParentIDs               string    `json:"parent_ids" csv:"parent_ids"`
	Label                   string    `json:"label" csv:"label"`
	Response                string    `json:"response" csv:"response"`
	ResponseID              string    `json:"response_id" csv:"response_id" gorm:"size:100"`
	ResponseSetID           string    `json:"response_set_id" csv:"response_set_id" gorm:"size:100"`
	IsFailedResponse        bool      `json:"is_failed_response" csv:"is_failed_response"`
	Comment                 string    `json:"comment" csv:"comment"`
	MediaFiles              string    `json:"media_files" csv:"media_files"`
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
	ModifiedAfter   time.Time
	TemplateIDs     []string
	Archived        string
	Completed       string
	IncludeInactive bool
	Incremental     bool
	ExportMedia     bool
	Limit           int
}

// Name is the name of the feed
func (f *InspectionItemFeed) Name() string {
	return "inspection_items"
}

// Model returns the model of the feed row
func (f *InspectionItemFeed) Model() interface{} {
	return InspectionItem{}
}

// RowsModel returns the model of feed rows
func (f *InspectionItemFeed) RowsModel() interface{} {
	return &[]*InspectionItem{}
}

// PrimaryKey returns the primary key(s)
func (f *InspectionItemFeed) PrimaryKey() []string {
	return []string{"id"}
}

// Columns returns the columns of the row
func (f *InspectionItemFeed) Columns() []string {
	return []string{
		"item_index",
		"template_id",
		"parent_id",
		"created_at",
		"modified_at",
		"type",
		"category",
		"category_id",
		"organisation_id",
		"parent_ids",
		"label",
		"response",
		"response_id",
		"response_set_id",
		"is_failed_response",
		"comment",
		"media_files",
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

func fetchAndWriteMedia(ctx context.Context, apiClient *api.Client, exporter Exporter, auditID, mediaURL string) error {
	resp, err := apiClient.GetMedia(
		ctx,
		&api.GetMediaRequest{
			URL:     mediaURL,
			AuditID: auditID,
		},
	)
	if err != nil {
		return err
	}

	// If the response is empty, then ignore this media object
	if resp == nil {
		return nil
	}

	err = exporter.WriteMedia(auditID, resp.MediaID, resp.ContentType, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (f *InspectionItemFeed) writeRows(ctx context.Context, exporter Exporter, rows []*InspectionItem, apiClient *api.Client) error {
	skipIDs := map[string]bool{}
	for _, id := range f.SkipIDs {
		skipIDs[id] = true
	}

	// Calculate the size of the batch we can insert into the DB at once. Column count + buffer to account for primary keys
	batchSize := exporter.ParameterLimit() / (len(f.Columns()) + 5)
	for i := 0; i < len(rows); i += batchSize {
		j := i + batchSize
		if j > len(rows) {
			j = len(rows)
		}

		// you can specify level of concurrency by increasing channel size
		buffers := make(chan bool, maxGoRoutines)
		var wg sync.WaitGroup

		// Some audits in production have the same item ID multiple times
		// We can't insert them simultaneously. This means we are dropping data, which sucks.
		rowsToInsert := []*InspectionItem{}
		idSeen := map[string]bool{}
		for _, row := range rows[i:j] {
			skip := skipIDs[row.AuditID]
			seen := idSeen[row.ID]
			if seen || skip {
				continue
			}
			idSeen[row.ID] = true
			rowsToInsert = append(rowsToInsert, row)

			if !f.ExportMedia || len(row.MediaHypertextReference) == 0 {
				continue
			}

			mediaURLList := strings.Split(row.MediaHypertextReference, "\n")
			for _, mediaURL := range mediaURLList {
				wg.Add(1)

				go func(mediaURL string) {
					defer wg.Done()
					buffers <- true

					err := fetchAndWriteMedia(ctx, apiClient, exporter, row.AuditID, mediaURL)
					util.Check(err, fmt.Sprintf("Failed to write media of inspection: %s", row.AuditID))

					<-buffers
				}(mediaURL)
			}
			wg.Wait()
		}

		err := exporter.WriteRows(f, rowsToInsert)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *InspectionItemFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*InspectionItem{})
}

// Export exports the feed to the supplied exporter
func (f *InspectionItemFeed) Export(ctx context.Context, apiClient *api.Client, exporter Exporter, orgID string) error {
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
		InitialURL: "/feed/inspection_items",
		Params: api.GetFeedParams{
			ModifiedAfter:   f.ModifiedAfter,
			TemplateIDs:     f.TemplateIDs,
			Archived:        f.Archived,
			Completed:       f.Completed,
			IncludeInactive: f.IncludeInactive,
			Limit:           f.Limit,
		},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*InspectionItem{}

		err := json.Unmarshal(resp.Data, &rows)
		util.Check(err, "Failed to unmarshal inspection-items data to struct")

		if len(rows) != 0 {
			err = f.writeRows(ctx, exporter, rows, apiClient)
			util.Check(err, "Failed to write data to exporter")
		}

		logger.Infof("%s: %d remaining", feedName, resp.Metadata.RemainingRecords)
		return nil
	})
	util.Check(err, "Failed to export feed")

	return exporter.FinaliseExport(f, &[]*InspectionItem{})
}
