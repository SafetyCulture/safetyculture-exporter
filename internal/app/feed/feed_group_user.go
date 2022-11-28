package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/internal/app/api"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/util"
)

// GroupUser represents a row from the group_users feed
type GroupUser struct {
	UserID         string    `json:"user_id" csv:"user_id" gorm:"primaryKey;size:37"`
	GroupID        string    `json:"group_id" csv:"group_id" gorm:"primaryKey;size:37"`
	OrganisationID string    `json:"organisation_id" csv:"organisation_id" gorm:"size:37"`
	ExportedAt     time.Time `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
}

// GroupUserFeed is a representation of the group_users feed
type GroupUserFeed struct{}

// Name is the name of the feed
func (f *GroupUserFeed) Name() string {
	return "group_users"
}

// Model returns the model of the feed row
func (f *GroupUserFeed) Model() interface{} {
	return GroupUser{}
}

// RowsModel returns the model of feed rows
func (f *GroupUserFeed) RowsModel() interface{} {
	return &[]*GroupUser{}
}

// PrimaryKey returns the primary key(s)
func (f *GroupUserFeed) PrimaryKey() []string {
	return []string{"user_id", "group_id"}
}

// Columns returns the columns of the row
func (f *GroupUserFeed) Columns() []string {
	return []string{
		"user_id",
		"group_id",
		"organisation_id",
		"exported_at",
	}
}

// Order returns the ordering when retrieving an export
func (f *GroupUserFeed) Order() string {
	return "group_id, user_id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *GroupUserFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*GroupUser{})
}

// Export exports the feed to the supplied exporter
func (f *GroupUserFeed) Export(ctx context.Context, apiClient *api.Client, exporter Exporter, orgID string) error {
	logger := util.GetLogger().With("feed", f.Name(), "org_id", orgID)

	if err := exporter.InitFeed(f, &InitFeedOptions{
		// Truncate files if upserts aren't supported.
		// This ensures that the export does not contain duplicate rows
		Truncate: false,
	}); err != nil {
		return fmt.Errorf("init feed: %w", err)
	}

	// Delete the actions if already exist
	if err := exporter.DeleteRowsIfExist(f, "organisation_id = ?", orgID); err != nil {
		return fmt.Errorf("delete row: %w", err)
	}

	drainFn := func(resp *api.GetFeedResponse) error {
		var rows []*GroupUser

		if err := json.Unmarshal(resp.Data, &rows); err != nil {
			return fmt.Errorf("map users data: %w", err)
		}

		// deduplicate rows (hotfix) because the feed Api GetUserGroups returns duplicates and this creates PK violations issues
		deDupedRows := DeduplicateList(func(row *GroupUser) string {
			return fmt.Sprintf("pk__%s_%s", row.UserID, row.GroupID)
		}, rows)

		if len(deDupedRows) != 0 {
			// Calculate the size of the batch we can insert into the DB at once. Column count + buffer to account for primary keys
			batchSize := exporter.ParameterLimit() / (len(f.Columns()) + 4)

			for i := 0; i < len(deDupedRows); i += batchSize {
				j := i + batchSize
				if j > len(deDupedRows) {
					j = len(deDupedRows)
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

	req := &api.GetFeedRequest{
		InitialURL: "/feed/group_users",
		Params:     api.GetFeedParams{},
	}

	if err := apiClient.DrainFeed(ctx, req, drainFn); err != nil {
		return fmt.Errorf("failed to export feed %q: %w", f.Name(), err)
	}
	return exporter.FinaliseExport(f, &[]*GroupUser{})
}
