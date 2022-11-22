package feed_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/SafetyCulture/safetyculture-exporter/cmd/safetyculture-exporter/cmd/export"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/feed"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestSchemaWriter_should_write_schema(t *testing.T) {
	var buf bytes.Buffer
	exporter, err := feed.NewSchemaExporter(&buf)
	assert.NoError(t, err)

	viperConfig := viper.New()

	testSchema := func(f feed.Feed) {
		err := exporter.CreateSchema(f, f.RowsModel())
		assert.Nil(t, err, fmt.Sprintf("something is wrong when creating schema: %s, %v", f.Name(), err))

		err = exporter.WriteSchema(f)
		assert.Nil(t, err, fmt.Sprintf("something is wrong when writing schema %s, %v", f.Name(), err))

		actual, err := os.ReadFile(fmt.Sprintf("mocks/set_1/schemas/formatted/%s.txt", f.Name()))
		assert.Nil(t, err, fmt.Sprintf("something is wrong when reading file %s.txt, %v", f.Name(), err))
		assert.Equal(t, strings.TrimSpace(buf.String()), strings.TrimSpace(string(actual)))

		buf.Reset()
	}

	exporterAppCfg := export.MapViperConfigToConfigurationOptions(viperConfig)
	exporterApp := feed.NewExporterApp(nil, nil, exporterAppCfg)

	for _, f := range exporterApp.GetFeeds() {
		fmt.Printf("TESTING FEED: %s\n", f.Name())
		testSchema(f)
	}

	for _, f := range feed.GetSheqsyFeeds() {
		fmt.Printf("TESTING FEED: %s\n", f.Name())
		testSchema(f)
	}
}

func TestSchemaWriter_should_write_all_schemas(t *testing.T) {
	var buf bytes.Buffer
	exporter, err := feed.NewSchemaExporter(&buf)
	assert.NoError(t, err)

	viperConfig := viper.New()
	exporterAppCfg := export.MapViperConfigToConfigurationOptions(viperConfig)
	exporterApp := feed.NewExporterApp(nil, nil, exporterAppCfg)

	err = exporterApp.PrintSchemas(exporter)
	assert.NoError(t, err)

	assert.NotNil(t, buf.String())
}
