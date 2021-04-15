package feed_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"gopkg.in/h2non/gock.v1"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const (
	mockAPIBaseURL            string = "http://localhost:9999"
	initiateReportURL         string = "/audits/.*/report"
	reportExportCompletionURL string = "/audits/.*/report/.*"
	downloadReportURL         string = "/report-exports/abc"
)

func getReportExportCompletionMessage(status string) string {
	return fmt.Sprintf(`{"status": "%s", "url": "%s%s"}`, status, mockAPIBaseURL, downloadReportURL)
}

func TestExportReports_should_export_all_reports(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"PDF", "WORD"}, "", "INSPECTION_TITLE")
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.GetTestClient()
	defer resetMocks(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	gock.New(mockAPIBaseURL).
		Post(initiateReportURL).
		Times(6).
		Reply(200).
		JSON(`{"messageId": "abc"}`)

	gock.New(mockAPIBaseURL).
		Get(reportExportCompletionURL).
		Times(6).
		Reply(200).
		JSON(getReportExportCompletionMessage("SUCCESS"))

	gock.New(mockAPIBaseURL).
		Get(downloadReportURL).
		Times(6).
		Reply(200).
		Body(bytes.NewBuffer([]byte(`file content`)))

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.Nil(t, err)

	fileExists(t, filepath.Join(exporter.ExportPath, "My-Audit.pdf"))
	fileExists(t, filepath.Join(exporter.ExportPath, "audit_2.pdf"))
	fileExists(t, filepath.Join(exporter.ExportPath, "audit_3.pdf"))

	fileExists(t, filepath.Join(exporter.ExportPath, "My-Audit.docx"))
	fileExists(t, filepath.Join(exporter.ExportPath, "audit_2.docx"))
	fileExists(t, filepath.Join(exporter.ExportPath, "audit_3.docx"))
}

func TestExportReports_should_export_all_reports_with_ID_filename(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"PDF", "WORD"}, "", "INSPECTION_ID")
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.GetTestClient()
	defer resetMocks(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	gock.New(mockAPIBaseURL).
		Post(initiateReportURL).
		Times(6).
		Reply(200).
		JSON(`{"messageId": "abc"}`)

	gock.New(mockAPIBaseURL).
		Get(reportExportCompletionURL).
		Times(6).
		Reply(200).
		JSON(getReportExportCompletionMessage("SUCCESS"))

	gock.New(mockAPIBaseURL).
		Get(downloadReportURL).
		Times(6).
		Reply(200).
		Body(bytes.NewBuffer([]byte(`file content`)))

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.Nil(t, err)

	fileExists(t, filepath.Join(exporter.ExportPath, "audit_1.pdf"))
	fileExists(t, filepath.Join(exporter.ExportPath, "audit_2.pdf"))
	fileExists(t, filepath.Join(exporter.ExportPath, "audit_3.pdf"))

	fileExists(t, filepath.Join(exporter.ExportPath, "audit_1.docx"))
	fileExists(t, filepath.Join(exporter.ExportPath, "audit_2.docx"))
	fileExists(t, filepath.Join(exporter.ExportPath, "audit_3.docx"))
}

func TestExportReports_should_not_run_if_all_exported(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"PDF"}, "", "INSPECTION_TITLE")
	assert.Nil(t, err)

	viperConfig := viper.New()
	viperConfig.Set("export.incremental", true)

	apiClient := api.GetTestClient()
	defer resetMocks(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	// Making sure the endpoints have been called only 3 times
	gock.New(mockAPIBaseURL).
		Post(initiateReportURL).
		Times(3).
		Reply(200).
		JSON(`{"messageId": "abc"}`)

	gock.New(mockAPIBaseURL).
		Get(reportExportCompletionURL).
		Times(3).
		Reply(200).
		JSON(getReportExportCompletionMessage("SUCCESS"))

	gock.New(mockAPIBaseURL).
		Get(downloadReportURL).
		Times(3).
		Reply(200).
		Body(bytes.NewBuffer([]byte(`file content`)))

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

func TestExportReports_should_take_care_of_invalid_file_names(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"PDF"}, "", "INSPECTION_TITLE")
	assert.Nil(t, err)

	viperConfig := viper.New()
	viperConfig.Set("export.incremental", true)

	apiClient := api.GetTestClient()
	defer resetMocks(apiClient.HTTPClient())
	gock.InterceptClient(apiClient.HTTPClient())

	gock.New("http://localhost:9999").
		Get("/feed/inspections").
		Reply(200).
		File("mocks/set_4/feed_inspections_1.json")

	// Making sure the endpoints have been called only 3 times
	gock.New(mockAPIBaseURL).
		Post(initiateReportURL).
		Times(3).
		Reply(200).
		JSON(`{"messageId": "abc"}`)

	gock.New(mockAPIBaseURL).
		Get(reportExportCompletionURL).
		Times(3).
		Reply(200).
		JSON(getReportExportCompletionMessage("SUCCESS"))

	gock.New(mockAPIBaseURL).
		Get(downloadReportURL).
		Times(3).
		Reply(200).
		Body(bytes.NewBuffer([]byte(`file content`)))

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.Nil(t, err)

	fileExists(t, filepath.Join(exporter.ExportPath, "My-Audit-1.pdf"))
	fileExists(t, filepath.Join(exporter.ExportPath, "My-Audit-1 (1).pdf"))
	var files []string
	filepath.Walk(exporter.ExportPath, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		assert.LessOrEqual(t, len(path), 255)
		return nil
	})
	assert.Equal(t, 5, len(files))
}

func TestExportReports_should_fail_after_retries(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"PDF"}, "", "INSPECTION_TITLE")
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.GetTestClient()
	defer resetMocks(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	gock.New(mockAPIBaseURL).
		Post(initiateReportURL).
		Times(3).
		Reply(200).
		JSON(`{"messageId": "abc"}`)

	// Making sure the endpoints is called 15 times for each inspection
	gock.New(mockAPIBaseURL).
		Get(reportExportCompletionURL).
		Times(45).
		Reply(200).
		JSON(getReportExportCompletionMessage("IN_PROGRESS"))

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Failed to generate 3 PDF reports and 0 WORD reports")
}

func TestExportReports_should_fail_if_report_status_fails(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"WORD"}, "", "INSPECTION_TITLE")
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.GetTestClient()
	defer resetMocks(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	gock.New(mockAPIBaseURL).
		Post(initiateReportURL).
		Times(3).
		Reply(200).
		JSON(`{"messageId": "abc"}`)

	gock.New(mockAPIBaseURL).
		Get(reportExportCompletionURL).
		Times(3).
		Reply(200).
		JSON(getReportExportCompletionMessage("FAILED"))

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Failed to generate 0 PDF reports and 3 WORD reports")
}

func TestExportReports_should_fail_if_init_report_reply_is_not_success(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"WORD"}, "", "INSPECTION_TITLE")
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.GetTestClient()
	defer resetMocks(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	gock.New(mockAPIBaseURL).
		Post(initiateReportURL).
		Times(3).
		Reply(500).
		JSON(`{"error": "something went wrong"}`)

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Failed to generate 0 PDF reports and 3 WORD reports")
}

func TestExportReports_should_fail_if_report_completion_reply_is_not_success(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"WORD"}, "", "INSPECTION_TITLE")
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.GetTestClient()
	defer resetMocks(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	gock.New(mockAPIBaseURL).
		Post(initiateReportURL).
		Times(3).
		Reply(200).
		JSON(`{"messageId": "abc"}`)

	gock.New(mockAPIBaseURL).
		Get(reportExportCompletionURL).
		Times(3).
		Reply(500).
		JSON(`{"error": "something went wrong"}`)

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Failed to generate 0 PDF reports and 3 WORD reports")
}

func TestExportReports_should_fail_if_download_report_reply_is_not_success(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"PDF"}, "", "INSPECTION_TITLE")
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.GetTestClient()
	defer resetMocks(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	gock.New(mockAPIBaseURL).
		Post(initiateReportURL).
		Times(3).
		Reply(200).
		JSON(`{"messageId": "abc"}`)

	gock.New(mockAPIBaseURL).
		Get(reportExportCompletionURL).
		Times(3).
		Reply(200).
		JSON(getReportExportCompletionMessage("SUCCESS"))

	gock.New(mockAPIBaseURL).
		Get(downloadReportURL).
		Times(3).
		Reply(500).
		JSON(`{"error": "something went wrong"}`)

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Failed to generate 3 PDF reports and 0 WORD reports")
}

func TestExportReports_should_return_error_for_unsupported_format(t *testing.T) {
	exporter, err := getTemporaryReportExporter([]string{"PNG"}, "", "INSPECTION_TITLE")
	assert.Nil(t, err)

	viperConfig := viper.New()

	apiClient := api.GetTestClient()
	initMockFeedsSet1(apiClient.HTTPClient())

	err = feed.ExportInspectionReports(viperConfig, apiClient, exporter)
	assert.EqualError(t, err, "No valid export format specified")
}
