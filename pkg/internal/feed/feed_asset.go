package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
	"time"
)

// Asset represents a row from the assets feed
type Asset struct {
	ID         string    `json:"id" csv:"asset_id" gorm:"primarykey;column:asset_id;size:36"`
	Code       string    `json:"code" csv:"code"`
	TypeID     string    `json:"type_id" csv:"type_id"`
	TypeName   string    `json:"type_name" csv:"type_name"`
	Fields     string    `json:"fields" csv:"fields"`
	CreatedAt  time.Time `json:"created_at" csv:"created_at"`
	ModifiedAt time.Time `json:"modified_at" csv:"modified_at" gorm:"index:idx_ast_modified_at,sort:desc"`
	SiteID     string    `json:"site_id" csv:"site_id" gorm:"size:41"`
	State      string    `json:"state" csv:"state"`
}

// AssetFeed is a representation of the assets feed
type AssetFeed struct {
	Limit       int
	Incremental bool
}

// Name is the name of the feed
func (f *AssetFeed) Name() string {
	return "assets"
}

// Model returns the model of the feed row
func (f *AssetFeed) Model() interface{} {
	return Asset{}
}

// RowsModel returns the model of feed rows
func (f *AssetFeed) RowsModel() interface{} {
	return &[]*Asset{}
}

// PrimaryKey returns the primary key(s)
func (f *AssetFeed) PrimaryKey() []string {
	return []string{"asset_id"}
}

// Columns returns the columns of the row
func (f *AssetFeed) Columns() []string {
	return []string{
		"code",
		"type_id",
		"type_name",
		"fields",
		"created_at",
		"modified_at",
		"site_id",
		"state",
	}
}

// Order returns the ordering when retrieving an export
func (f *AssetFeed) Order() string {
	return "asset_id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *AssetFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*Asset{})
}

// Export exports the feed to the supplied exporter
func (f *AssetFeed) Export(ctx context.Context, apiClient *httpapi.Client, exporter Exporter, orgID string) error {
	logger := logger.GetLogger().With("feed", f.Name(), "org_id", orgID)

	if err := exporter.InitFeed(f, &InitFeedOptions{
		// Delete data if incremental refresh is disabled so there is no duplicates
		Truncate: !f.Incremental,
	}); err != nil {
		return fmt.Errorf("init feed: %w", err)
	}

	drainFn := func(resp *GetFeedResponse) error {
		var rows []*Asset

		if err := json.Unmarshal(resp.Data, &rows); err != nil {
			return fmt.Errorf("map data: %w", err)
		}

		if len(rows) != 0 {
			// Calculate the size of the batch we can insert into the DB at once.
			// Column count + buffer to account for primary keys
			batchSize := exporter.ParameterLimit() / (len(f.Columns()) + 4)

			for i := 0; i < len(rows); i += batchSize {
				j := i + batchSize
				if j > len(rows) {
					j = len(rows)
				}

				if err := exporter.WriteRows(f, rows[i:j]); err != nil {
					return fmt.Errorf("exporter: %w", err)
				}
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
		InitialURL: "/feed/assets",
		Params: GetFeedParams{
			Limit: f.Limit,
		},
	}

	if err := DrainFeed(ctx, apiClient, req, drainFn); err != nil {
		return fmt.Errorf("assets feed %q: %w", f.Name(), err)
	}
	return exporter.FinaliseExport(f, &[]*Asset{})
}
