package feed_test

import (
	"context"
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"github.com/stretchr/testify/assert"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
)

func TestInspectionFeedExport_should_export_rows_to_sql_db(t *testing.T) {
	exporter, err := getInmemorySQLExporter()
	assert.Nil(t, err)

	apiClient := api.NewAPIClient("http://localhost:9999", "abc123")
	initMockFeedsSet1(apiClient.HTTPClient())

	inspectionsFeed := feed.InspectionFeed{
		SkipIDs:       []string{},
		ModifiedAfter: "",
		TemplateIDs:   []string{},
		Archived:      "both",
		Completed:     "both",
		Incremental:   true,
	}

	err = inspectionsFeed.Export(context.Background(), apiClient, exporter)
	assert.Nil(t, err)

	rows := []feed.Inspection{}
	resp := exporter.DB.Table("inspections").Scan(&rows)
	assert.Nil(t, resp.Error)

	assert.Equal(t, 3, len(rows))
	assert.Equal(t, "audit_1", rows[0].ID)
}
