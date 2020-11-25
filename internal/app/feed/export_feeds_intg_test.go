// +build sql

package feed_test

import (
	"path/filepath"
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

/*
	For these tests we use the CSV Exporter, but instead of using a SQLite DB as an intermediary layer
	we write the data to a real DB. So we get to test the SQL exporting logic, and compare the results easily.
*/

func TestIntegrationDbCreateSchema_should_create_all_schemas(t *testing.T) {
	sqlExporter, err := getTestingSQLExporter()
	assert.Nil(t, err)
	exporter, err := getTemporaryCSVExporterWithRealSQLExporter(sqlExporter)
	assert.Nil(t, err)

	viperConfig := viper.New()

	err = feed.CreateSchemas(viperConfig, exporter)
	assert.Nil(t, err)

	filesEqualish(t, "mocks/set_1/schemas/inspections.csv", filepath.Join(exporter.ExportPath, "inspections.csv"))
	filesEqualish(t, "mocks/set_1/schemas/inspection_items.csv", filepath.Join(exporter.ExportPath, "inspection_items.csv"))
	filesEqualish(t, "mocks/set_1/schemas/templates.csv", filepath.Join(exporter.ExportPath, "templates.csv"))

	filesEqualish(t, "mocks/set_1/schemas/sites.csv", filepath.Join(exporter.ExportPath, "sites.csv"))

	filesEqualish(t, "mocks/set_1/schemas/users.csv", filepath.Join(exporter.ExportPath, "users.csv"))
	filesEqualish(t, "mocks/set_1/schemas/groups.csv", filepath.Join(exporter.ExportPath, "groups.csv"))
	filesEqualish(t, "mocks/set_1/schemas/group_users.csv", filepath.Join(exporter.ExportPath, "group_users.csv"))

	filesEqualish(t, "mocks/set_1/schemas/schedules.csv", filepath.Join(exporter.ExportPath, "schedules.csv"))
	filesEqualish(t, "mocks/set_1/schemas/schedule_assignees.csv", filepath.Join(exporter.ExportPath, "schedule_assignees.csv"))
	filesEqualish(t, "mocks/set_1/schemas/schedule_occurrences.csv", filepath.Join(exporter.ExportPath, "schedule_occurrences.csv"))
}

func TestIntegrationDbExportFeeds_should_export_all_feeds_to_file(t *testing.T) {
	sqlExporter, err := getTestingSQLExporter()
	assert.Nil(t, err)
	exporter, err := getTemporaryCSVExporterWithRealSQLExporter(sqlExporter)
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.NewAPIClient("http://localhost:9999", "token")
	initMockFeedsSet1(apiClient.HTTPClient())

	err = feed.ExportFeeds(viperConfig, apiClient, exporter)
	assert.Nil(t, err)

	filesEqualish(t, "mocks/set_1/outputs/inspections.csv", filepath.Join(exporter.ExportPath, "inspections.csv"))
	filesEqualish(t, "mocks/set_1/outputs/inspection_items.csv", filepath.Join(exporter.ExportPath, "inspection_items.csv"))
	filesEqualish(t, "mocks/set_1/outputs/templates.csv", filepath.Join(exporter.ExportPath, "templates.csv"))

	filesEqualish(t, "mocks/set_1/outputs/sites.csv", filepath.Join(exporter.ExportPath, "sites.csv"))

	filesEqualish(t, "mocks/set_1/outputs/users.csv", filepath.Join(exporter.ExportPath, "users.csv"))
	filesEqualish(t, "mocks/set_1/outputs/groups.csv", filepath.Join(exporter.ExportPath, "groups.csv"))
	filesEqualish(t, "mocks/set_1/outputs/group_users.csv", filepath.Join(exporter.ExportPath, "group_users.csv"))

	filesEqualish(t, "mocks/set_1/outputs/schedules.csv", filepath.Join(exporter.ExportPath, "schedules.csv"))
	filesEqualish(t, "mocks/set_1/outputs/schedule_assignees.csv", filepath.Join(exporter.ExportPath, "schedule_assignees.csv"))
	filesEqualish(t, "mocks/set_1/outputs/schedule_occurrences.csv", filepath.Join(exporter.ExportPath, "schedule_occurrences.csv"))

}

// Expectation of this test is that group_users and schedule_assignees are truncated and refreshed
// and that other tables are incrementally updated
func TestIntegrationDbExportFeeds_should_perform_incremental_update_on_second_run(t *testing.T) {
	sqlExporter, err := getTestingSQLExporter()
	assert.Nil(t, err)
	exporter, err := getTemporaryCSVExporterWithRealSQLExporter(sqlExporter)
	assert.Nil(t, err)

	viperConfig := viper.New()
	viperConfig.Set("export.inspection.incremental", true)

	apiClient := api.NewAPIClient("http://localhost:9999", "token")
	initMockFeedsSet1(apiClient.HTTPClient())

	err = feed.ExportFeeds(viperConfig, apiClient, exporter)
	assert.Nil(t, err)

	initMockFeedsSet2(apiClient.HTTPClient())

	err = feed.ExportFeeds(viperConfig, apiClient, exporter)
	assert.Nil(t, err)

	filesEqualish(t, "mocks/set_2/outputs/inspections.csv", filepath.Join(exporter.ExportPath, "inspections.csv"))
	filesEqualish(t, "mocks/set_2/outputs/inspection_items.csv", filepath.Join(exporter.ExportPath, "inspection_items.csv"))
	filesEqualish(t, "mocks/set_2/outputs/templates.csv", filepath.Join(exporter.ExportPath, "templates.csv"))

	filesEqualish(t, "mocks/set_2/outputs/sites.csv", filepath.Join(exporter.ExportPath, "sites.csv"))

	filesEqualish(t, "mocks/set_2/outputs/users.csv", filepath.Join(exporter.ExportPath, "users.csv"))
	filesEqualish(t, "mocks/set_2/outputs/groups.csv", filepath.Join(exporter.ExportPath, "groups.csv"))
	filesEqualish(t, "mocks/set_2/outputs/group_users.csv", filepath.Join(exporter.ExportPath, "group_users.csv"))

	filesEqualish(t, "mocks/set_2/outputs/schedules.csv", filepath.Join(exporter.ExportPath, "schedules.csv"))
	filesEqualish(t, "mocks/set_2/outputs/schedule_assignees.csv", filepath.Join(exporter.ExportPath, "schedule_assignees.csv"))
	filesEqualish(t, "mocks/set_2/outputs/schedule_occurrences.csv", filepath.Join(exporter.ExportPath, "schedule_occurrences.csv"))
}

func TestIntegrationDbExportFeeds_should_handle_lots_of_rows_ok(t *testing.T) {
	sqlExporter, err := getTestingSQLExporter()
	assert.Nil(t, err)
	exporter, err := getTemporaryCSVExporterWithRealSQLExporter(sqlExporter)
	assert.Nil(t, err)

	viperConfig := viper.New()
	viperConfig.Set("export.inspection.incremental", true)

	apiClient := api.NewAPIClient("http://localhost:9999", "token")
	initMockFeedsSet3(apiClient.HTTPClient())

	err = feed.ExportFeeds(viperConfig, apiClient, exporter)
	assert.Nil(t, err)

	inspectionsLines, err := countFileLines(filepath.Join(exporter.ExportPath, "inspections.csv"))
	assert.Nil(t, err)
	assert.Equal(t, 97, inspectionsLines)

	inspectionItemsLines, err := countFileLines(filepath.Join(exporter.ExportPath, "inspection_items.csv"))
	assert.Nil(t, err)
	assert.Equal(t, 501, inspectionItemsLines)
}
