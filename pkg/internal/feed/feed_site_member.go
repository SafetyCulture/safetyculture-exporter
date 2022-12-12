package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/events"
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
func (f *SiteMemberFeed) Export(ctx context.Context, apiClient *httpapi.Client, exporter Exporter, orgID string) error {
	logger := logger.GetLogger().With("feed", f.Name(), "org_id", orgID)

	if err := exporter.InitFeed(f, &InitFeedOptions{
		// Truncate files if upserts aren't supported.
		// This ensures that the export does not contain duplicate rows
		Truncate: !exporter.SupportsUpsert(),
	}); err != nil {
		return events.WrapEventError(err, "init feed")
	}

	drainFn := func(resp *GetFeedResponse) error {
		var rows []*SiteMember

		if err := json.Unmarshal(resp.Data, &rows); err != nil {
			return events.NewEventErrorWithMessage(err, events.ErrorSeverityError, events.ErrorSubSystemDataIntegrity, false, "map data")
		}

		if len(rows) != 0 {
			// Calculate the size of the batch we can insert into the DB at once. Column count + buffer to account for primary keys
			batchSize := exporter.ParameterLimit() / (len(f.Columns()) + 4)

			for i := 0; i < len(rows); i += batchSize {
				j := i + batchSize
				if j > len(rows) {
					j = len(rows)
				}

				if err := exporter.WriteRows(f, rows[i:j]); err != nil {
					return events.WrapEventError(err, "write rows")
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

	req := &GetFeedRequest{InitialURL: "/feed/site_members"}
	if err := DrainFeed(ctx, apiClient, req, drainFn); err != nil {
		return events.WrapEventError(err, fmt.Sprintf("feed %q", f.Name()))
	}
	return exporter.FinaliseExport(f, &[]*SiteMember{})
}
