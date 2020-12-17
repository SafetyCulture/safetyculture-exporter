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
	exporter, err := getTemporaryReportExporter([]string{"PDF", "WORD"}, "")
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.NewAPIClient("http://localhost:9999", "token")
	defer resetMocks(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	reportExportMockOptions := &InspectionReportExportMockOptions{
		ReportExportCompletionStatus: "SUCCESS",
		InitiateReportReply:          200,
		ReportExportCompletionReply:  200,
		DownloadReportReply:          200,
	}
	initMockReportExport(apiClient.HTTPClient(), reportExportMockOptions)

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.Nil(t, err)

	fileExists(t, filepath.Join(exporter.ExportPath, "My-Audit.pdf"))
	fileExists(t, filepath.Join(exporter.ExportPath, "audit_2.pdf"))
	fileExists(t, filepath.Join(exporter.ExportPath, "audit_3.pdf"))

	fileExists(t, filepath.Join(exporter.ExportPath, "My-Audit.docx"))
	fileExists(t, filepath.Join(exporter.ExportPath, "audit_2.docx"))
	fileExists(t, filepath.Join(exporter.ExportPath, "audit_3.docx"))
}

func TestExportReports_should_not_run_if_all_exported(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"PDF"}, "")
	assert.Nil(t, err)

	viperConfig := viper.New()
	viperConfig.Set("export.inspection.incremental", true)

	apiClient := api.NewAPIClient("http://localhost:9999", "token")
	defer resetMocks(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	reportExportMockOptions := &InspectionReportExportMockOptions{
		ReportExportCompletionStatus: "SUCCESS",
		InitiateReportReply:          200,
		ReportExportCompletionReply:  200,
		DownloadReportReply:          200,
	}
	initMockReportExport(apiClient.HTTPClient(), reportExportMockOptions)

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.Nil(t, err)

	file1ModTime1, _ := getFileModTime(filepath.Join(exporter.ExportPath, "My-Audit.pdf"))
	file2ModTime1, _ := getFileModTime(filepath.Join(exporter.ExportPath, "audit_2.pdf"))
	file3ModTime1, _ := getFileModTime(filepath.Join(exporter.ExportPath, "audit_3.pdf"))

	// run the export process again
	initMockFeedsSet1(apiClient.HTTPClient())
	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.Nil(t, err)

	file1ModTime2, _ := getFileModTime(filepath.Join(exporter.ExportPath, "My-Audit.pdf"))
	file2ModTime2, _ := getFileModTime(filepath.Join(exporter.ExportPath, "audit_2.pdf"))
	file3ModTime2, _ := getFileModTime(filepath.Join(exporter.ExportPath, "audit_3.pdf"))

	assert.Equal(t, file1ModTime1, file1ModTime2)
	assert.Equal(t, file2ModTime1, file2ModTime2)
	assert.Equal(t, file3ModTime1, file3ModTime2)
}

func TestExportReports_should_fail_after_retries(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"PDF"}, "")
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.NewAPIClient("http://localhost:9999", "token")
	defer resetMocks(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	reportExportMockOptions := &InspectionReportExportMockOptions{
		ReportExportCompletionStatus: "IN_PROGRESS",
		InitiateReportReply:          200,
		ReportExportCompletionReply:  200,
		DownloadReportReply:          200,
	}
	initMockReportExport(apiClient.HTTPClient(), reportExportMockOptions)

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.NotNil(t, err)
}

func TestExportReports_should_fail_if_report_status_fails(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"WORD"}, "")
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.NewAPIClient("http://localhost:9999", "token")
	defer resetMocks(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	reportExportMockOptions := &InspectionReportExportMockOptions{
		ReportExportCompletionStatus: "FAILED",
		InitiateReportReply:          200,
		ReportExportCompletionReply:  200,
		DownloadReportReply:          200,
	}
	initMockReportExport(apiClient.HTTPClient(), reportExportMockOptions)

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.NotNil(t, err)
}

func TestExportReports_should_fail_if_init_report_reply_is_not_success(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"WORD"}, "")
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.NewAPIClient("http://localhost:9999", "token")
	defer resetMocks(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	reportExportMockOptions := &InspectionReportExportMockOptions{
		ReportExportCompletionStatus: "SUCCESS",
		InitiateReportReply:          500,
		ReportExportCompletionReply:  200,
		DownloadReportReply:          200,
	}
	initMockReportExport(apiClient.HTTPClient(), reportExportMockOptions)

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.NotNil(t, err)
}

func TestExportReports_should_fail_if_report_completion_reply_is_not_success(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"WORD"}, "")
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.NewAPIClient("http://localhost:9999", "token")
	defer resetMocks(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	reportExportMockOptions := &InspectionReportExportMockOptions{
		ReportExportCompletionStatus: "SUCCESS",
		InitiateReportReply:          200,
		ReportExportCompletionReply:  500,
		DownloadReportReply:          200,
	}
	initMockReportExport(apiClient.HTTPClient(), reportExportMockOptions)

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.NotNil(t, err)
}

func TestExportReports_should_fail_if_download_report_reply_is_not_success(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"WORD"}, "")
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.NewAPIClient("http://localhost:9999", "token")
	defer resetMocks(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	reportExportMockOptions := &InspectionReportExportMockOptions{
		ReportExportCompletionStatus: "SUCCESS",
		InitiateReportReply:          200,
		ReportExportCompletionReply:  200,
		DownloadReportReply:          500,
	}
	initMockReportExport(apiClient.HTTPClient(), reportExportMockOptions)

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.NotNil(t, err)
}

func TestExportReports_should_return_error_for_unsupported_format(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"PNG"}, "")
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.NewAPIClient("http://localhost:9999", "token")
	initMockFeedsSet1(apiClient.HTTPClient())

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.EqualError(t, err, "No valid export format specified")
}
