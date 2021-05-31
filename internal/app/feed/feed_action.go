package feed

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
)

// Action represents a row from the actions feed
type Action struct {
	ID              string    `json:"id" csv:"action_id" gorm:"primarykey;column:action_id"`
	Title           string    `json:"title" csv:"title" gorm:"size:64000"`
	Description     string    `json:"description" csv:"description" gorm:"size:64000"`
	SiteID          string    `json:"site_id" csv:"site_id"`
	Priority        string    `json:"priority" csv:"priority"`
	Status          string    `json:"status" csv:"status" gorm:"size:64000"`
	DueDate         time.Time `json:"due_date" csv:"due_date"`
	CreatedAt       time.Time `json:"created_at" csv:"created_at"`
	ModifiedAt      time.Time `json:"modified_at" csv:"modified_at"`
	ExportedAt      time.Time `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
	CreatorUserID   string    `json:"creator_user_id" csv:"creator_user_id"`
	CreatorUserName string    `json:"creator_user_name" csv:"creator_user_name" gorm:"size:64000"`
	TemplateID      string    `json:"template_id" csv:"template_id"`
	AuditID         string    `json:"audit_id" csv:"audit_id"`
	AuditTitle      string    `json:"audit_title" csv:"audit_title" gorm:"size:64000"`
	AuditItemID     string    `json:"audit_item_id" csv:"audit_item_id"`
	AuditItemLabel  string    `json:"audit_item_label" csv:"audit_item_label" gorm:"size:64000"`
}

// ActionFeed is a representation of the actions feed
type ActionFeed struct {
	ModifiedAfter time.Time
	Incremental   bool
}

// Name is the name of the feed
func (f *ActionFeed) Name() string {
	return "actions"
}

// Model returns the model of the feed row
func (f *ActionFeed) Model() interface{} {
	return Action{}
}

// RowsModel returns the model of feed rows
func (f *ActionFeed) RowsModel() interface{} {
	return &[]*Action{}
}

// PrimaryKey returns the primary key(s)
func (f *ActionFeed) PrimaryKey() []string {
	return []string{"action_id"}
}

// Columns returns the columns of the row
func (f *ActionFeed) Columns() []string {
	return []string{
		"title",
		"description",
		"site_id",
		"priority",
		"status",
		"due_date",
		"created_at",
		"modified_at",
		"exported_at",
		"creator_user_id",
		"creator_user_name",
		"template_id",
		"audit_id",
		"audit_title",
		"audit_item_id",
		"audit_item_label",
	}
}

// Order returns the ordering when retrieving an export
func (f *ActionFeed) Order() string {
	return "action_id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *ActionFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*Action{})
}

// Export exports the feed to the supplied exporter
func (f *ActionFeed) Export(ctx context.Context, apiClient *api.Client, exporter Exporter) error {
	logger := util.GetLogger()
	feedName := f.Name()

	exporter.InitFeed(f, &InitFeedOptions{
		// Delete data if incremental refresh is disabled so there is no duplicates
		Truncate: f.Incremental == false,
	})

	var err error
	f.ModifiedAfter, err = exporter.LastModifiedAt(f, f.ModifiedAfter)
	util.Check(err, "unable to load modified after")

	logger.Infof("%s: exporting since %s", feedName, f.ModifiedAfter.Format(time.RFC1123))

	err = apiClient.DrainFeed(ctx, &api.GetFeedRequest{
		InitialURL: "/feed/actions",
		Params: api.GetFeedParams{
			ModifiedAfter: f.ModifiedAfter,
		},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*Action{}

		err := json.Unmarshal(resp.Data, &rows)
		util.Check(err, "Failed to unmarshal actions data to struct")

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

	return exporter.FinaliseExport(f, &[]*Action{})
}
