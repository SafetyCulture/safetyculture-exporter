package feed_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	exporterAPI "github.com/SafetyCulture/safetyculture-exporter/pkg/api"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/feed"
	"github.com/stretchr/testify/assert"
)

func TestSchemaWriter_should_write_schema(t *testing.T) {
	var buf bytes.Buffer
	exporter, err := feed.NewSchemaExporter(&buf)
	assert.NoError(t, err)

	testSchema := func(f feed.Feed) {
		err := exporter.CreateSchema(f, f.RowsModel())
		assert.Nil(t, err, fmt.Sprintf("something is wrong when creating schema: %s, %v", f.Name(), err))

		err = exporter.WriteSchema(f)
		assert.Nil(t, err, fmt.Sprintf("something is wrong when writing schema %s, %v", f.Name(), err))

		actual, err := os.ReadFile(fmt.Sprintf("fixtures/schemas/formatted/%s.txt", f.Name()))
		assert.Nil(t, err, fmt.Sprintf("something is wrong when reading file %s.txt, %v", f.Name(), err))
		assert.Equal(t, strings.TrimSpace(buf.String()), strings.TrimSpace(string(actual)))

		buf.Reset()
	}

	cfg := &exporterAPI.ExporterConfiguration{}
	exporterApp := feed.NewExporterApp(nil, nil, cfg.ToExporterConfig())

	for _, f := range exporterApp.GetFeeds() {
		fmt.Printf("TESTING FEED: %s\n", f.Name())
		testSchema(f)
	}

	for _, f := range exporterApp.GetSheqsyFeeds() {
		fmt.Printf("TESTING FEED: %s\n", f.Name())
		testSchema(f)
	}
}

func TestSchemaWriter_should_write_all_schemas(t *testing.T) {
	var buf bytes.Buffer
	exporter, err := feed.NewSchemaExporter(&buf)
	assert.NoError(t, err)

	cfg := &exporterAPI.ExporterConfiguration{}
	exporterApp := feed.NewExporterApp(nil, nil, cfg.ToExporterConfig())

	err = exporterApp.PrintSchemas(exporter)
	assert.NoError(t, err)

	assert.NotNil(t, buf.String())
}
