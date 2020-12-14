package feed_test

import (
	"path/filepath"
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestExportReports_should_export_all_reports(t *testing.T) {
	exporter, err := getTemporaryReportExporter()
	assert.Nil(t, err)

	viperConfig := viper.New()
	viperConfig.Set("report.format", []string{"PDF"})

	apiClient := api.NewAPIClient("http://localhost:9999", "token")
	initMockFeedsSet1(apiClient.HTTPClient())

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.Nil(t, err)

	fileExists(t, filepath.Join(exporter.ExportPath, "My-Audit.pdf"))
	fileExists(t, filepath.Join(exporter.ExportPath, "audit_2.pdf"))
	fileExists(t, filepath.Join(exporter.ExportPath, "audit_3.pdf"))
}

func TestExportReports_should_return_error_for_unsupported_format(t *testing.T) {
	exporter, err := getTemporaryReportExporter()
	assert.Nil(t, err)

	viperConfig := viper.New()
	viperConfig.Set("report.format", []string{"PNG"})

	apiClient := api.NewAPIClient("http://localhost:9999", "token")
	initMockFeedsSet1(apiClient.HTTPClient())

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.EqualError(t, err, "No valid export format specified")
}
