package api_test

import (
	"bytes"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	exporterAPI "github.com/SafetyCulture/safetyculture-exporter/pkg/external/api"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/feed"
	"github.com/stretchr/testify/assert"
)

// GetTestClient creates a new test apiClient
func GetTestClient(opts ...httpapi.Opt) *httpapi.Client {
	apiClient := httpapi.NewClient("http://localhost:9999", "abc123", opts...)
	apiClient.RetryWaitMin = 10 * time.Millisecond
	apiClient.RetryWaitMax = 10 * time.Millisecond
	apiClient.CheckForRetry = httpapi.DefaultRetryPolicy
	apiClient.RetryMax = 1
	return apiClient
}

// getInmemorySQLExporter creates a SQLExporter that uses an inmemory DB
func getInmemorySQLExporter(exportMediaPath string) (*feed.SQLExporter, error) {
	return feed.NewSQLExporter("sqlite", "file::memory:", true, exportMediaPath)
}

// getTemporaryReportExporter creates a ReportExporter that writes to a temp folder
func getTemporaryReportExporter(format []string, preferenceID string, filename string) (*feed.ReportExporter, error) {
	dir, err := os.MkdirTemp("", "export")
	if err != nil {
		log.Fatal(err)
	}

	cfg := &exporterAPI.ExporterConfiguration{}
	cfg.Report.Format = format
	cfg.Report.PreferenceID = preferenceID
	cfg.Report.FilenameConvention = filename
	cfg.Report.RetryTimeout = 10
	return exporterAPI.NewReportExporter(dir, cfg.ToReporterConfig())
}

// getTemporaryCSVExporter creates a CSVExporter that writes to a temp folder
func getTemporaryCSVExporter() (*feed.CSVExporter, error) {
	dir, err := os.MkdirTemp("", "export")
	if err != nil {
		log.Fatal(err)
	}

	return feed.NewCSVExporter(dir, "", 100000)
}

func fileExists(t *testing.T, expectedPath string) {
	_, err := os.Stat(expectedPath)
	assert.NoError(t, err)
}

func getFileModTime(filePath string) (time.Time, error) {
	file, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, err
	}
	return file.ModTime(), nil
}

// filesEqualish checks if files are equal enough (ignoring dates)
func filesEqualish(t *testing.T, expectedPath, actualPath string) {
	expectedFile, err := os.ReadFile(expectedPath)
	assert.NoError(t, err)

	actualFile, err := os.ReadFile(actualPath)
	assert.NoError(t, err)

	assert.Equal(t,
		dateRegex.ReplaceAllLiteralString(strings.TrimSpace(string(expectedFile)), "--date--"),
		dateRegex.ReplaceAllLiteralString(strings.TrimSpace(string(actualFile)), "--date--"),
	)
}

func countFileLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}

	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := file.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

// getTemporaryCSVExporterWithMaxRowsLimit creates a CSVExporter that writes to a temp folder with row limit
func getTemporaryCSVExporterWithMaxRowsLimit(maxRowsPerFile int) (*feed.CSVExporter, error) {
	dir, err := os.MkdirTemp("", "export")
	if err != nil {
		log.Fatal(err)
	}

	return feed.NewCSVExporter(dir, "", maxRowsPerFile)
}

var dateRegex = regexp.MustCompile(`(?m)(-?(?:[1-9][0-9]*)?[0-9]{4})-(1[0-2]|0[1-9])-(3[01]|0[1-9]|[12][0-9])T(2[0-3]|[01][0-9]):([0-5][0-9]):([0-5][0-9])(\.[0-9]+)?(\+|Z)(2[0-3]|[01][0-9])?:?([0-5][0-9])?`)
