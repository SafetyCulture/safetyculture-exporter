package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/internal/app/api"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/util"
)

// User represents a row from the users feed
type User struct {
	ID             string     `json:"id" csv:"user_id" gorm:"primarykey;column:user_id;size:37"`
	OrganisationID string     `json:"organisation_id" csv:"organisation_id" gorm:"size:37"`
	Email          string     `json:"email" csv:"email" gorm:"size:256"`
	Firstname      string     `json:"firstname" csv:"firstname"`
	Lastname       string     `json:"lastname" csv:"lastname"`
	Active         bool       `json:"active" csv:"active"`
	LastSeenAt     *time.Time `json:"last_seen_at" csv:"last_seen_at"`
	ExportedAt     time.Time  `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
}

// UserFeed is a representation of the users feed
type UserFeed struct{}

// Name is the name of the feed
func (f *UserFeed) Name() string {
	return "users"
}

// Model returns the model of the feed row
func (f *UserFeed) Model() interface{} {
	return User{}
}

// RowsModel returns the model of feed rows
func (f *UserFeed) RowsModel() interface{} {
	return &[]*User{}
}

// PrimaryKey returns the primary key(s)
func (f *UserFeed) PrimaryKey() []string {
	return []string{"user_id"}
}

// Columns returns the columns of the row
func (f *UserFeed) Columns() []string {
	return []string{
		"organisation_id",
		"email",
		"firstname",
		"lastname",
		"active",
		"last_seen_at",
		"exported_at",
	}
}

// Order returns the ordering when retrieving an export
func (f *UserFeed) Order() string {
	return "user_id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *UserFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*User{})
}

// Export exports the feed to the supplied exporter
func (f *UserFeed) Export(ctx context.Context, apiClient *api.Client, exporter Exporter, orgID string) error {
	logger := util.GetLogger().With("feed", f.Name(), "org_id", orgID)

	if err := exporter.InitFeed(f, &InitFeedOptions{
		// Truncate files if upserts aren't supported.
		// This ensures that the export does not contain duplicate rows
		Truncate: !exporter.SupportsUpsert(),
	}); err != nil {
		return fmt.Errorf("init feed: %w", err)
	}

	drainFn := func(resp *api.GetFeedResponse) error {
		var rows []*User

		if err := json.Unmarshal(resp.Data, &rows); err != nil {
			return fmt.Errorf("map users data: %w", err)
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

	req := &api.GetFeedRequest{InitialURL: "/feed/users", Params: api.GetFeedParams{}}
	if err := apiClient.DrainFeed(ctx, req, drainFn); err != nil {
		return fmt.Errorf("failed to export feed %q: %w", f.Name(), err)
	}
	return exporter.FinaliseExport(f, &[]*User{})
}
