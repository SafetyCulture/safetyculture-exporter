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
	Title           string    `json:"title" csv:"title"`
	Description     string    `json:"description" csv:"description"`
	SiteID          string    `json:"site_id" csv:"site_id"`
	Priority        string    `json:"priority" csv:"priority"`
	Status          string    `json:"status" csv:"status"`
	DueDate         time.Time `json:"due_date" csv:"due_date"`
	CreatedAt       time.Time `json:"created_at" csv:"created_at"`
	ModifiedAt      time.Time `json:"modified_at" csv:"modified_at"`
	CreatorUserID   string    `json:"creator_user_id" csv:"creator_user_id"`
	CreatorUserName string    `json:"creator_user_name" csv:"creator_user_name"`
	TemplateID      string    `json:"template_id" csv:"template_id"`
	AuditID         string    `json:"audit_id" csv:"audit_id"`
	AuditTitle      string    `json:"audit_title" csv:"audit_title"`
	AuditItemID     string    `json:"audit_item_id" csv:"audit_item_id"`
	AuditItemLabel  string    `json:"audit_item_label" csv:"audit_item_label"`
}

// ActionFeed is a representation of the actions feed
type ActionFeed struct{}

// Name is the name of the feed
func (f *ActionFeed) Name() string {
	return "actions"
}

// Model returns the model of the feed row
func (f *ActionFeed) Model() interface{} {
	return Action{}
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

// Export exports the feed to the supplied exporter
func (f *ActionFeed) Export(ctx context.Context, apiClient api.APIClient, exporter Exporter) error {
	logger := util.GetLogger()
	feedName := f.Name()

	logger.Infof("%s: exporting", feedName)

	exporter.InitFeed(f, &InitFeedOptions{
		// Truncate files if upserts aren't supported.
		// This ensure that the export does not contain duplicate rows
		Truncate: exporter.SupportsUpsert() == false,
	})

	err := apiClient.DrainFeed(ctx, &api.GetFeedRequest{
		InitialURL: "/feed/actions",
		Params:     api.GetFeedParams{},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*Action{}

		err := json.Unmarshal(resp.Data, &rows)
		util.Check(err, "Failed to unmarshal data to struct")

		if len(rows) != 0 {
			err = exporter.WriteRows(f, rows)
			util.Check(err, "Failed to write data to exporter")
		}

		logger.Infof("%s: %d remaining", feedName, resp.Metadata.RemainingRecords)
		return nil
	})
	util.Check(err, "Failed to export feed")

	return exporter.FinaliseExport(f, &[]*Action{})
}
