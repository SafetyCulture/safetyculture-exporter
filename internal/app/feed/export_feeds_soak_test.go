//go:build soak
// +build soak

package feed_test

import (
	"os"
	"testing"

	"github.com/SafetyCulture/safetyculture-exporter/internal/app/api"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/config"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/feed"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationDbSoakExportFeeds_should_successfully_export_with_significant_data(t *testing.T) {
	exporter, err := getTestingSQLExporter()
	assert.NoError(t, err)
	exporter.AutoMigrate = true

	apiClient := api.NewClient(os.Getenv("TEST_API_HOST"), os.Getenv("TEST_ACCESS_TOKEN"))

	cfg := &config.ExporterConfiguration{}
	cfg.AccessToken = "token-123"

	exporterApp := feed.NewExporterApp(apiClient, nil, cfg)
	err = exporterApp.ExportFeeds(exporter)
	assert.NoError(t, err)
}
