package feed_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/coreexporter/feed"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/coreexporter/api"
)

func TestInspectionItemFeedExport_should_export_rows_to_sql_db(t *testing.T) {
	exporter, err := getInmemorySQLExporter("")
	assert.NoError(t, err)

	apiClient := api.GetTestClient()
	initMockFeedsSet1(apiClient.HTTPClient())

	inspectionItemFeed := feed.InspectionItemFeed{
		SkipIDs:       []string{},
		ModifiedAfter: time.Now(),
		TemplateIDs:   []string{},
		Archived:      "both",
		Completed:     "both",
		Incremental:   true,
	}

	err = inspectionItemFeed.Export(context.Background(), apiClient, exporter, "")
	assert.NoError(t, err)

	var rows []feed.InspectionItem
	resp := exporter.DB.Table("inspection_items").Scan(&rows)
	assert.Nil(t, resp.Error)

	assert.Equal(t, 5, len(rows))
}

func TestInspectionItemFeedExportWithMedia(t *testing.T) {
	dir, err := os.MkdirTemp("", "export")
	assert.NoError(t, err)

	exporter, err := getInmemorySQLExporter(dir)
	assert.NoError(t, err)

	result := `{id:"test-id"}`
	req := gock.New("http://localhost:9999").
		Get("/audits/audit_1/media/12345").
		Reply(200).
		BodyString(result)
	req.SetHeader("Content-Type", "image/test-content")

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	inspectionItemFeed := feed.InspectionItemFeed{
		SkipIDs:       []string{},
		ModifiedAfter: time.Now(),
		TemplateIDs:   []string{},
		Archived:      "both",
		Completed:     "both",
		Incremental:   true,
		ExportMedia:   true,
	}

	err = inspectionItemFeed.Export(context.Background(), apiClient, exporter, "")
	assert.NoError(t, err)
}
