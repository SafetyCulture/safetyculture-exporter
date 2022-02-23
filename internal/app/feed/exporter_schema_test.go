package feed_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestSchemaWriter_should_write_schema(t *testing.T) {
	var buf bytes.Buffer
	exporter, err := feed.NewSchemaExporter(&buf)
	assert.Nil(t, err)

	viperConfig := viper.New()

	testSchema := func(f feed.Feed) {
		err := exporter.CreateSchema(f, f.RowsModel())
		assert.NotNil(t, err)

		err = exporter.WriteSchema(f)
		assert.NotNil(t, err)

		actual, err := ioutil.ReadFile(fmt.Sprintf("mocks/set_1/schemas/formatted/%s.txt", f.Name()))
		assert.NotNil(t, err, fmt.Sprintf("something is wrong with %s, %v", f.Name(), err))
		assert.Equal(t, strings.TrimSpace(buf.String()), strings.TrimSpace(string(actual)))

		buf.Reset()
	}

	for _, feed := range feed.GetFeeds(viperConfig) {
		testSchema(feed)
	}
}

func TestSchemaWriter_should_write_all_schemas(t *testing.T) {
	var buf bytes.Buffer
	exporter, err := feed.NewSchemaExporter(&buf)
	assert.Nil(t, err)

	viperConfig := viper.New()

	err = feed.WriteSchemas(viperConfig, exporter)
	assert.Nil(t, err)

	assert.NotNil(t, buf.String())
}
