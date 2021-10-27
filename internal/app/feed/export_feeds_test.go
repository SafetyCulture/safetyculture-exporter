package feed_test

import (
	"path/filepath"
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"gopkg.in/h2non/gock.v1"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestCreateSchemas_should_create_all_schemas_to_file(t *testing.T) {
	exporter, err := getTemporaryCSVExporter()
	assert.Nil(t, err)

	viperConfig := viper.New()
	viperConfig.Set("export.site.include_deleted", true)

	err = feed.CreateSchemas(viperConfig, exporter)
	assert.Nil(t, err)

	filesEqualish(t, "mocks/set_1/schemas/inspections.csv", filepath.Join(exporter.ExportPath, "inspections.csv"))
	filesEqualish(t, "mocks/set_1/schemas/inspection_items.csv", filepath.Join(exporter.ExportPath, "inspection_items.csv"))
	filesEqualish(t, "mocks/set_1/schemas/templates.csv", filepath.Join(exporter.ExportPath, "templates.csv"))
	filesEqualish(t, "mocks/set_1/schemas/template_permissions.csv", filepath.Join(exporter.ExportPath, "template_permissions.csv"))

	filesEqualish(t, "mocks/set_1/schemas/sites.csv", filepath.Join(exporter.ExportPath, "sites.csv"))

	filesEqualish(t, "mocks/set_1/schemas/users.csv", filepath.Join(exporter.ExportPath, "users.csv"))
	filesEqualish(t, "mocks/set_1/schemas/groups.csv", filepath.Join(exporter.ExportPath, "groups.csv"))
	filesEqualish(t, "mocks/set_1/schemas/group_users.csv", filepath.Join(exporter.ExportPath, "group_users.csv"))

	filesEqualish(t, "mocks/set_1/schemas/schedules.csv", filepath.Join(exporter.ExportPath, "schedules.csv"))
	filesEqualish(t, "mocks/set_1/schemas/schedule_assignees.csv", filepath.Join(exporter.ExportPath, "schedule_assignees.csv"))
	filesEqualish(t, "mocks/set_1/schemas/schedule_occurrences.csv", filepath.Join(exporter.ExportPath, "schedule_occurrences.csv"))
}

func TestExportFeeds_should_export_all_feeds_to_file(t *testing.T) {
	defer gock.Off()

	exporter, err := getTemporaryCSVExporter()
	assert.Nil(t, err)

	viperConfig := viper.New()
	viperConfig.Set("export.site.include_deleted", true)

	apiClient := api.GetTestClient()
	initMockFeedsSet1(apiClient.HTTPClient())

	gock.New("http://localhost:9999").
		Get("/accounts/user/v1/user:WhoAmI").
		Reply(200).
		BodyString(`
		{
			"user_id": "user_123",
			"organisation_id": "role_123",
			"firstname": "Test",
			"lastname": "Test"
		  }
		`)

	err = feed.ExportFeeds(viperConfig, apiClient, exporter)
	assert.Nil(t, err)

	filesEqualish(t, "mocks/set_1/outputs/inspections.csv", filepath.Join(exporter.ExportPath, "inspections.csv"))
	filesEqualish(t, "mocks/set_1/outputs/inspection_items.csv", filepath.Join(exporter.ExportPath, "inspection_items.csv"))
	filesEqualish(t, "mocks/set_1/outputs/templates.csv", filepath.Join(exporter.ExportPath, "templates.csv"))
	filesEqualish(t, "mocks/set_1/outputs/template_permissions.csv", filepath.Join(exporter.ExportPath, "template_permissions.csv"))

	filesEqualish(t, "mocks/set_1/outputs/sites.csv", filepath.Join(exporter.ExportPath, "sites.csv"))

	filesEqualish(t, "mocks/set_1/outputs/users.csv", filepath.Join(exporter.ExportPath, "users.csv"))
	filesEqualish(t, "mocks/set_1/outputs/groups.csv", filepath.Join(exporter.ExportPath, "groups.csv"))
	filesEqualish(t, "mocks/set_1/outputs/group_users.csv", filepath.Join(exporter.ExportPath, "group_users.csv"))

	filesEqualish(t, "mocks/set_1/outputs/schedules.csv", filepath.Join(exporter.ExportPath, "schedules.csv"))
	filesEqualish(t, "mocks/set_1/outputs/schedule_assignees.csv", filepath.Join(exporter.ExportPath, "schedule_assignees.csv"))
	filesEqualish(t, "mocks/set_1/outputs/schedule_occurrences.csv", filepath.Join(exporter.ExportPath, "schedule_occurrences.csv"))

	filesEqualish(t, "mocks/set_1/outputs/actions.csv", filepath.Join(exporter.ExportPath, "actions.csv"))
	filesEqualish(t, "mocks/set_1/outputs/action_assignees.csv", filepath.Join(exporter.ExportPath, "action_assignees.csv"))
}

// Expectation of this test is that group_users and schedule_assignees are truncated and refreshed
// and that other tables are incrementally updated
func TestExportFeeds_should_perform_incremental_update_on_second_run(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/accounts/user/v1/user:WhoAmI").
		Times(2).
		Reply(200).
		BodyString(`
		{
			"user_id": "user_123",
			"organisation_id": "role_123",
			"firstname": "Test",
			"lastname": "Test"
		  }
		`)

	exporter, err := getTemporaryCSVExporter()
	assert.Nil(t, err)

	viperConfig := viper.New()
	viperConfig.Set("export.incremental", true)
	viperConfig.Set("export.site.include_deleted", true)

	apiClient := api.GetTestClient()
	initMockFeedsSet1(apiClient.HTTPClient())

	err = feed.ExportFeeds(viperConfig, apiClient, exporter)
	assert.Nil(t, err)

	initMockFeedsSet2(apiClient.HTTPClient())

	err = feed.ExportFeeds(viperConfig, apiClient, exporter)
	assert.Nil(t, err)

	filesEqualish(t, "mocks/set_2/outputs/inspections.csv", filepath.Join(exporter.ExportPath, "inspections.csv"))
	filesEqualish(t, "mocks/set_2/outputs/inspection_items.csv", filepath.Join(exporter.ExportPath, "inspection_items.csv"))
	filesEqualish(t, "mocks/set_2/outputs/templates.csv", filepath.Join(exporter.ExportPath, "templates.csv"))
	filesEqualish(t, "mocks/set_2/outputs/template_permissions.csv", filepath.Join(exporter.ExportPath, "template_permissions.csv"))

	filesEqualish(t, "mocks/set_2/outputs/sites.csv", filepath.Join(exporter.ExportPath, "sites.csv"))

	filesEqualish(t, "mocks/set_2/outputs/users.csv", filepath.Join(exporter.ExportPath, "users.csv"))
	filesEqualish(t, "mocks/set_2/outputs/groups.csv", filepath.Join(exporter.ExportPath, "groups.csv"))
	filesEqualish(t, "mocks/set_2/outputs/group_users.csv", filepath.Join(exporter.ExportPath, "group_users.csv"))

	filesEqualish(t, "mocks/set_2/outputs/schedules.csv", filepath.Join(exporter.ExportPath, "schedules.csv"))
	filesEqualish(t, "mocks/set_2/outputs/schedule_assignees.csv", filepath.Join(exporter.ExportPath, "schedule_assignees.csv"))
	filesEqualish(t, "mocks/set_2/outputs/schedule_occurrences.csv", filepath.Join(exporter.ExportPath, "schedule_occurrences.csv"))

	filesEqualish(t, "mocks/set_2/outputs/actions.csv", filepath.Join(exporter.ExportPath, "actions.csv"))
	filesEqualish(t, "mocks/set_2/outputs/action_assignees.csv", filepath.Join(exporter.ExportPath, "action_assignees.csv"))
}

func TestGetActionLimit(t *testing.T) {
	viperConfig := viper.New()

	viperConfig.Set("export.action.limit", 200)
	assert.Equal(t, 100, feed.GetActionLimit(viperConfig))

	viperConfig.Set("export.action.limit", 20)
	assert.Equal(t, 20, feed.GetActionLimit(viperConfig))

	viperConfig.Set("export.action.limit", 100)
	assert.Equal(t, 100, feed.GetActionLimit(viperConfig))
}

func TestExportFeeds_should_handle_lots_of_rows_ok(t *testing.T) {
	defer gock.Off()

	exporter, err := getTemporaryCSVExporter()
	assert.Nil(t, err)

	viperConfig := viper.New()
	viperConfig.Set("export.incremental", true)

	apiClient := api.GetTestClient()
	initMockFeedsSet3(apiClient.HTTPClient())

	gock.New("http://localhost:9999").
		Get("/accounts/user/v1/user:WhoAmI").
		Reply(200).
		BodyString(`
		{
			"user_id": "user_123",
			"organisation_id": "role_123",
			"firstname": "Test",
			"lastname": "Test"
		  }
		`)

	err = feed.ExportFeeds(viperConfig, apiClient, exporter)
	assert.Nil(t, err)

	inspectionsLines, err := countFileLines(filepath.Join(exporter.ExportPath, "inspections.csv"))
	assert.Nil(t, err)
	assert.Equal(t, 97, inspectionsLines)

	inspectionItemsLines, err := countFileLines(filepath.Join(exporter.ExportPath, "inspection_items.csv"))
	assert.Nil(t, err)
	assert.Equal(t, 230, inspectionItemsLines)
}
