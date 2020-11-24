package feed_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"github.com/stretchr/testify/assert"
)

func TestSchemaWriter_should_write_schema(t *testing.T) {
	var buf bytes.Buffer
	exporter, err := feed.NewSchemaExporter(&buf)
	assert.Nil(t, err)

	testSchema := func(f feed.Feed) {
		exporter.CreateSchema(f, f.RowsModel())
		exporter.WriteSchema(f)

		actual, _ := ioutil.ReadFile(fmt.Sprintf("mocks/set_1/schemas/formatted/%s.txt", f.Name()))
		assert.Equal(t, strings.TrimSpace(buf.String()), strings.TrimSpace(string(actual)))

		buf.Reset()
	}

	for _, feed := range feed.GetFeeds() {
		testSchema(feed)
	}
}
