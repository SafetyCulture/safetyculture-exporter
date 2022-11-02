//go:build soak
// +build soak

package feed_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/SafetyCulture/safetyculture-exporter/internal/app/api"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/feed"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestIntegrationDbSoakExportFeeds_should_successfully_export_with_significant_data(t *testing.T) {
	exporter, err := getTestingSQLExporter()
	assert.Nil(t, err)
	exporter.AutoMigrate = true

	fmt.Println(err, exporter)

	viperConfig := viper.New()

	apiClient := api.NewClient(os.Getenv("TEST_API_HOST"), os.Getenv("TEST_ACCESS_TOKEN"))

	err = feed.ExportFeeds(viperConfig, apiClient, exporter)
	assert.Nil(t, err)
}
