package feed

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
)

// TemplatePermission represents a row from the template_permissions feed
type TemplatePermission struct {
	ID             string `json:"id" csv:"permission_id" gorm:"primarykey;column:permission_id;size:375"`
	TemplateID     string `json:"template_id" csv:"template_id" gorm:"size:100"`
	Permission     string `json:"permission" csv:"permission" gorm:"size:10"`
	AssigneeID     string `json:"assignee_id" csv:"assignee_id" gorm:"size:256"`
	AssigneeType   string `json:"assignee_type" csv:"assignee_type" gorm:"size:10"`
	OrganisationID string `json:"organisation_id" csv:"organisation_id" gorm:"size:37"`
}

// TemplatePermissionFeed is a representation of the template_permissions feed
type TemplatePermissionFeed struct {
	ModifiedAfter time.Time
	Incremental   bool
}

// Name is the name of the feed
func (f *TemplatePermissionFeed) Name() string {
	return "template_permissions"
}

// Model returns the model of the feed row
func (f *TemplatePermissionFeed) Model() interface{} {
	return TemplatePermission{}
}

// RowsModel returns the model of feed rows
func (f *TemplatePermissionFeed) RowsModel() interface{} {
	return &[]*TemplatePermission{}
}

// PrimaryKey returns the primary key(s)
func (f *TemplatePermissionFeed) PrimaryKey() []string {
	return []string{"permission_id"}
}

// Columns returns the columns of the row
func (f *TemplatePermissionFeed) Columns() []string {
	return []string{
		"template_id",
		"permission",
		"assignee_id",
		"assignee_type",
		"organisation_id",
	}
}

// Order returns the ordering when retrieving an export
func (f *TemplatePermissionFeed) Order() string {
	return "permission_id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *TemplatePermissionFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*TemplatePermission{})
}

// Export exports the feed to the supplied exporter
func (f *TemplatePermissionFeed) Export(ctx context.Context, apiClient *api.Client, exporter Exporter, orgID string) error {
	logger := util.GetLogger()
	feedName := f.Name()

	logger.Infof("%s: exporting for org_id: %s", feedName, orgID)

	exporter.InitFeed(f, &InitFeedOptions{
		// Always truncate. This data must be refreshed in order to be accurate
		Truncate: true,
	})

	err := apiClient.DrainFeed(ctx, &api.GetFeedRequest{
		InitialURL: "/feed/template_permissions",
		Params: api.GetFeedParams{
			ModifiedAfter: f.ModifiedAfter,
		},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*TemplatePermission{}

		err := json.Unmarshal(resp.Data, &rows)
		util.Check(err, "Failed to unmarshal templates-permissions data to struct")

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

	return exporter.FinaliseExport(f, &[]*TemplatePermission{})
}
