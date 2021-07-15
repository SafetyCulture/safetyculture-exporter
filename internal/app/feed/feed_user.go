package feed

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
)

// User represents a row from the users feed
type User struct {
	ID             string    `json:"id" csv:"user_id" gorm:"primarykey;column:user_id"`
	OrganisationID string    `json:"organisation_id" csv:"organisation_id" gorm:"index:idx_usr_organisation_id"`
	Email          string    `json:"email" csv:"email"`
	Firstname      string    `json:"firstname" csv:"firstname"`
	Lastname       string    `json:"lastname" csv:"lastname"`
	Active         bool      `json:"active" csv:"active"`
	ExportedAt     time.Time `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
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
	logger := util.GetLogger()
	feedName := f.Name()

	logger.Infof("%s: exporting for org_id: %s", feedName, orgID)

	exporter.InitFeed(f, &InitFeedOptions{
		// Truncate files if upserts aren't supported.
		// This ensure that the export does not contain duplicate rows
		Truncate: exporter.SupportsUpsert() == false,
	})

	err := apiClient.DrainFeed(ctx, &api.GetFeedRequest{
		InitialURL: "/feed/users",
		Params:     api.GetFeedParams{},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*User{}

		err := json.Unmarshal(resp.Data, &rows)
		util.Check(err, "Failed to unmarshal users data to struct")

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

		logger.Infof("%s: %d remaining", feedName, resp.Metadata.RemainingRecords)
		return nil
	})
	util.Check(err, "Failed to export feed")

	return exporter.FinaliseExport(f, &[]*User{})
}
