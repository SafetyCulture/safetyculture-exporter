package feed_test

import (
	"context"
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"github.com/stretchr/testify/assert"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
)

func TestInspectionItemFeedExport_should_export_rows_to_sql_db(t *testing.T) {
	exporter, err := getInmemorySQLExporter()
	assert.Nil(t, err)

	apiClient := api.NewAPIClient("http://localhost:9999", "abc123")
	initMockFeedsSet1(apiClient.HTTPClient())

	inspectionItemFeed := feed.InspectionItemFeed{
		SkipIDs:       []string{},
		ModifiedAfter: "",
		TemplateIDs:   []string{},
		Archived:      "both",
		Completed:     "both",
		Incremental:   true,
	}

	err = inspectionItemFeed.Export(context.Background(), apiClient, exporter)
	assert.Nil(t, err)

	rows := []feed.InspectionItem{}
	resp := exporter.DB.Table("inspection_items").Scan(&rows)
	assert.Nil(t, resp.Error)

	assert.Equal(t, 5, len(rows))
}
