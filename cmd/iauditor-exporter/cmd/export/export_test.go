package export_test

import (
	"github.com/SafetyCulture/iauditor-exporter/cmd/iauditor-exporter/cmd/export"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPrintSchemaCmd(t *testing.T) {
	res := export.PrintSchemaCmd()
	require.NotNil(t, res)
	assert.EqualValues(t, "schema", res.Use)
	assert.EqualValues(t, "Print iAuditor table schemas", res.Short)
	assert.EqualValues(t, "iauditor-exporter schema", res.Example)
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
	assert.EqualValues(t, "Export iAuditor inspections to json files", res.Short)
}

func TestCSVCmd(t *testing.T) {
	res := export.CSVCmd()
	require.NotNil(t, res)
	assert.EqualValues(t, "csv", res.Use)
	assert.EqualValues(t, "Export iAuditor data to CSV files", res.Short)
}

func TestSQLCmd(t *testing.T) {
	res := export.SQLCmd()
	require.NotNil(t, res)
	assert.EqualValues(t, "sql", res.Use)
	assert.EqualValues(t, "Export iAuditor data to SQL database", res.Short)
}
