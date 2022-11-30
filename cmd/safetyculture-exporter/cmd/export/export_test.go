package export_test

import (
	"testing"

	"github.com/SafetyCulture/safetyculture-exporter/cmd/safetyculture-exporter/cmd/export"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/config"
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
	cfg := config.BuildConfigurationWithDefaults()
	cfg.Export.Action.Limit = 50
	cm, err := config.NewConfigurationManager("test.yaml", false, false, cfg)
	require.Nil(t, err)
	require.NotNil(t, cm)
	assert.EqualValues(t, 50, cm.Configuration.Export.Action.Limit)
}

func TestMapViperConfigToConfigurationOptions_ShouldEnforceActionsLimit(t *testing.T) {
	cfg := config.BuildConfigurationWithDefaults()
	cfg.Export.Action.Limit = 101
	cm, err := config.NewConfigurationManager("test.yaml", false, false, cfg)
	require.Nil(t, err)
	require.NotNil(t, cm)
	assert.EqualValues(t, 100, cm.Configuration.Export.Action.Limit)
}

func TestMapViperConfigToConfigurationOptions_ShouldRespectIssuesLimit(t *testing.T) {
	cfg := config.BuildConfigurationWithDefaults()
	cfg.Export.Issue.Limit = 50
	cm, err := config.NewConfigurationManager("test.yaml", false, false, cfg)
	require.Nil(t, err)
	require.NotNil(t, cm)
	assert.EqualValues(t, 50, cm.Configuration.Export.Issue.Limit)
}

func TestMapViperConfigToConfigurationOptions_ShouldEnforceIssuesLimit(t *testing.T) {
	cfg := config.BuildConfigurationWithDefaults()
	cfg.Export.Issue.Limit = 101
	cm, err := config.NewConfigurationManager("test.yaml", false, false, cfg)
	require.Nil(t, err)
	require.NotNil(t, cm)
	assert.EqualValues(t, 100, cm.Configuration.Export.Issue.Limit)
}
