package export_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/cmd/safetyculture-exporter/cmd/export"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintSchemaCmd(t *testing.T) {
	res := export.PrintSchemaCmd()
	require.NotNil(t, res)
	assert.EqualValues(t, "schema", res.Use)
	assert.EqualValues(t, "Print SafetyCulture table schemas", res.Short)
	assert.EqualValues(t, "safetyculture-exporter schema", res.Example)
}

func TestReportCmd(t *testing.T) {
	res := export.ReportCmd()
	require.NotNil(t, res)
	assert.EqualValues(t, "report", res.Use)
	assert.EqualValues(t, "Export inspection report", res.Short)
}

func TestInspectionJSONCmd(t *testing.T) {
	res := export.InspectionJSONCmd()
	require.NotNil(t, res)
	assert.EqualValues(t, "inspection-json", res.Use)
	assert.EqualValues(t, "Export SafetyCulture inspections to json files", res.Short)
}

func TestCSVCmd(t *testing.T) {
	res := export.CSVCmd()
	require.NotNil(t, res)
	assert.EqualValues(t, "csv", res.Use)
	assert.EqualValues(t, "Export SafetyCulture data to CSV files", res.Short)
}

func TestSQLCmd(t *testing.T) {
	res := export.SQLCmd()
	require.NotNil(t, res)
	assert.EqualValues(t, "sql", res.Use)
	assert.EqualValues(t, "Export SafetyCulture data to SQL database", res.Short)
}

func TestMapViperConfigToConfigurationOptions_ShouldRespectActionsLimit(t *testing.T) {
	viperConfig := viper.New()
	viperConfig.Set("export.action.limit", 50)
	cfg := export.MapViperConfigToConfigurationOptions(viperConfig)
	assert.EqualValues(t, 50, cfg.ExportConfig.ActionConfig.BatchLimit)
}

func TestMapViperConfigToConfigurationOptions_ShouldEnforceActionsLimit(t *testing.T) {
	viperConfig := viper.New()
	viperConfig.Set("export.action.limit", 101)
	cfg := export.MapViperConfigToConfigurationOptions(viperConfig)
	assert.EqualValues(t, 100, cfg.ExportConfig.ActionConfig.BatchLimit)
}

func TestMapViperConfigToConfigurationOptions_ShouldRespectIssuesLimit(t *testing.T) {
	viperConfig := viper.New()
	viperConfig.Set("export.issue.limit", 50)
	cfg := export.MapViperConfigToConfigurationOptions(viperConfig)
	assert.EqualValues(t, 50, cfg.ExportConfig.IssueConfig.BatchLimit)
}

func TestMapViperConfigToConfigurationOptions_ShouldEnforceIssuesLimit(t *testing.T) {
	viperConfig := viper.New()
	viperConfig.Set("export.issue.limit", 101)
	cfg := export.MapViperConfigToConfigurationOptions(viperConfig)
	assert.EqualValues(t, 100, cfg.ExportConfig.IssueConfig.BatchLimit)
}

func TestMapViperConfigToConfigurationOptions_ModifiedAfter(t *testing.T) {
	viperConfig := viper.New()
	viperConfig.Set("export.modified_after", "")
	cfg := export.MapViperConfigToConfigurationOptions(viperConfig)
	fmt.Println(cfg.ExportConfig.ModifiedAfter)

}

func TestMapViperConfigToConfigurationOptions(t *testing.T) {
	viperConfig := viper.New()
	viperConfig.Set("access_token", "sc-api-123")
	viperConfig.Set("export.action.limit", 100)
	viperConfig.Set("export.issue.limit", 100)
	viperConfig.Set("export.incremental", true)
	viperConfig.Set("export.modified_after", "2022-01-20")
	viperConfig.Set("export.template_ids", "A B C")
	viperConfig.Set("export.tables", "A B C")
	viperConfig.Set("export.inspection.included_inactive_items", true)
	viperConfig.Set("export.inspection.archived", "both")
	viperConfig.Set("export.inspection.completed", "true")
	viperConfig.Set("export.inspection.skip_ids", "A B C")
	viperConfig.Set("export.inspection.limit", 10)
	viperConfig.Set("export.inspection.web_report_link", "web_link")
	viperConfig.Set("export.site.include_deleted", true)
	viperConfig.Set("export.site.include_full_hierarchy", true)
	viperConfig.Set("export.media", true)

	cfg := export.MapViperConfigToConfigurationOptions(viperConfig)

	// GENERIC FIELDS
	assert.EqualValues(t, "sc-api-123", cfg.ApiConfig.AccessToken)
	assert.True(t, cfg.ExportConfig.Incremental)
	modifiedAfter, err := time.Parse("2006-01-02", "2022-01-20")
	assert.Nil(t, err)
	assert.Equal(t, modifiedAfter, cfg.ExportConfig.ModifiedAfter)
	assert.EqualValues(t, []string{"A", "B", "C"}, cfg.ExportConfig.FilterByTemplateID)
	assert.EqualValues(t, []string{"A", "B", "C"}, cfg.ExportConfig.FilterByTableName)

	// INSPECTION FIELDS
	assert.True(t, cfg.ExportConfig.InspectionConfig.IncludeInactiveItems)
	assert.EqualValues(t, "both", cfg.ExportConfig.InspectionConfig.Archived)
	assert.EqualValues(t, "true", cfg.ExportConfig.InspectionConfig.Completed)
	assert.EqualValues(t, []string{"A", "B", "C"}, cfg.ExportConfig.InspectionConfig.SkipIDs)
	assert.EqualValues(t, 10, cfg.ExportConfig.InspectionConfig.BatchLimit)
	assert.EqualValues(t, "web_link", cfg.ExportConfig.InspectionConfig.WebReportLink)

	// SITE CONFIG FIELDS
	assert.True(t, cfg.ExportConfig.SiteConfig.IncludeDeleted)
	assert.True(t, cfg.ExportConfig.SiteConfig.IncludeFullHierarchy)

	// MEDIA CONFIG FIELDS
	assert.True(t, cfg.ExportConfig.MediaConfig.Export)

	// ACTION CONFIG FIELDS
	assert.EqualValues(t, 100, cfg.ExportConfig.ActionConfig.BatchLimit)

	// ISSUE CONFIG FIELDS
	assert.EqualValues(t, 100, cfg.ExportConfig.IssueConfig.BatchLimit)
}
