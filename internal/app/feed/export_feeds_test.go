package feed_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// filesEqualish checks if files are equal enough (ignoring dates)
func filesEqualish(t *testing.T, expectedPath, actualPath string) {
	expectedFile, err := ioutil.ReadFile(expectedPath)
	assert.Nil(t, err)

	actualFile, err := ioutil.ReadFile(actualPath)
	assert.Nil(t, err)

	assert.Equal(t,
		dateRegex.ReplaceAllLiteralString(strings.TrimSpace(string(expectedFile)), "--date--"),
		dateRegex.ReplaceAllLiteralString(strings.TrimSpace(string(actualFile)), "--date--"),
	)
}

func TestExportFeeds_should_export_all_feeds_to_file(t *testing.T) {
	exporter, err := getTemporaryCSVExporter()
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.NewAPIClient("http://localhost:9999", "token")
	initMockFeedsSet1(apiClient.HTTPClient())

	err = feed.ExportFeeds(viperConfig, apiClient, exporter)
	assert.Nil(t, err)

	fmt.Println(exporter.ExportPath)

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
func TestExportFeeds_should_perform_incremental_update_on_second_run(t *testing.T) {
	exporter, err := getTemporaryCSVExporter()
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

	fmt.Println(exporter.ExportPath)

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
