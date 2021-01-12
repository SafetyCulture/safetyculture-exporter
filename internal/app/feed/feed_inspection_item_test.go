package feed_test

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
)

func TestInspectionItemFeedExport_should_export_rows_to_sql_db(t *testing.T) {
	exporter, err := getInmemorySQLExporter("")
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

func TestInspectionItemFeedExportWithMedia(t *testing.T) {
	dir, err := ioutil.TempDir("", "export")
	assert.Nil(t, err)

	exporter, err := getInmemorySQLExporter(dir)
	assert.Nil(t, err)

	result := `{id:"test-id"}`
	req := gock.New("http://localhost:9999").
		Get("/audits/audit_1/media/12345").
		Reply(200).
		BodyString(result)
	req.SetHeader("Content-Type", "image/test-content")

	apiClient := api.NewAPIClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())
	initMockFeedsSet1(apiClient.HTTPClient())

	inspectionItemFeed := feed.InspectionItemFeed{
		SkipIDs:       []string{},
		ModifiedAfter: "",
		TemplateIDs:   []string{},
		Archived:      "both",
		Completed:     "both",
		Incremental:   true,
		ExportMedia:   true,
	}

	err = inspectionItemFeed.Export(context.Background(), apiClient, exporter)
	assert.Nil(t, err)
}
