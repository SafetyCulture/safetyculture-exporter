package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
	"strings"
	"sync"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/events"
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
	MediaIDs                string    `json:"media_ids" csv:"media_ids"`
	MediaHypertextReference string    `json:"media_hypertext_reference" csv:"media_hypertext_reference"`
	Score                   float32   `json:"score" csv:"score"`
	MaxScore                float32   `json:"max_score" csv:"max_score"`
	ScorePercentage         float32   `json:"score_percentage" csv:"score_percentage"`
	CombinedScore           float32   `json:"combined_score" csv:"combined_score"`
	CombinedMaxScore        float32   `json:"combined_max_score" csv:"combined_max_score"`
	CombinedScorePercentage float32   `json:"combined_score_percentage" csv:"combined_score_percentage"`
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
		"media_ids",
		"media_hypertext_reference",
		"score",
		"max_score",
		"score_percentage",
		"combined_score",
		"combined_max_score",
		"combined_score_percentage",
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

func fetchAndWriteMedia(ctx context.Context, apiClient *httpapi.Client, exporter Exporter, auditID, mediaURL string) error {
	resp, err := GetMedia(
		ctx,
		apiClient,
		&GetMediaRequest{
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

func (f *InspectionItemFeed) writeRows(ctx context.Context, exporter Exporter, rows []*InspectionItem, apiClient *httpapi.Client) error {
	logger := logger.GetLogger()
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
		var rowsToInsert []*InspectionItem
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
			if len(mediaURLList) > 0 {
				logger.Infof(" downloading media for inspection item %s", row.ItemID)
			}

			for _, mediaURL := range mediaURLList {
				wg.Add(1)

				go func(mediaURL string) error {
					defer wg.Done()
					buffers <- true

					if err := fetchAndWriteMedia(ctx, apiClient, exporter, row.AuditID, mediaURL); err != nil {
						return events.WrapEventError(err, "write media")
					}

					<-buffers
					return nil
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
func (f *InspectionItemFeed) Export(ctx context.Context, apiClient *httpapi.Client, exporter Exporter, orgID string) error {
	logger := logger.GetLogger().With("feed", f.Name(), "org_id", orgID)

	exporter.InitFeed(f, &InitFeedOptions{
		// Delete data if incremental refresh is disabled so there is no duplicates
		Truncate: !f.Incremental,
	})

	var err error
	f.ModifiedAfter, err = exporter.LastModifiedAt(f, f.ModifiedAfter, orgID)
	if err != nil {
		return events.NewEventErrorWithMessage(err, events.ErrorSeverityError, events.ErrorSubSystemDB, false, "unable to load modified after")
	}

	drainFn := func(resp *GetFeedResponse) error {
		var rows []*InspectionItem

		if err := json.Unmarshal(resp.Data, &rows); err != nil {
			return events.NewEventErrorWithMessage(err, events.ErrorSeverityError, events.ErrorSubSystemDataIntegrity, false, "map data")
		}

		if len(rows) != 0 {
			if err := f.writeRows(ctx, exporter, rows, apiClient); err != nil {
				return err
			}
		}

		logger.With(
			"estimated_remaining", resp.Metadata.RemainingRecords,
			"duration_ms", apiClient.Duration.Milliseconds(),
			"export_duration_ms", exporter.GetDuration().Milliseconds(),
		).Info("export batch complete")
		return nil
	}

	req := &GetFeedRequest{
		InitialURL: "/feed/inspection_items",
		Params: GetFeedParams{
			ModifiedAfter:   f.ModifiedAfter,
			TemplateIDs:     f.TemplateIDs,
			Archived:        f.Archived,
			Completed:       f.Completed,
			IncludeInactive: f.IncludeInactive,
			Limit:           f.Limit,
		},
	}

	if err := DrainFeed(ctx, apiClient, req, drainFn); err != nil {
		return events.WrapEventError(err, fmt.Sprintf("feed %q", f.Name()))
	}
	return exporter.FinaliseExport(f, &[]*InspectionItem{})
}
