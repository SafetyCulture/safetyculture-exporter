package feed_test

import (
	"context"
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"github.com/stretchr/testify/assert"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
)

func TestActionFeedExport_should_export_rows_to_sql_db(t *testing.T) {
	exporter, err := getInmemorySQLExporter("")
	assert.Nil(t, err)

	apiClient := api.GetTestClient()
	initMockFeedsSet1(apiClient.HTTPClient())

	actionsFeed := feed.ActionFeed{
		Limit: 100,
	}

	err = actionsFeed.Export(context.Background(), apiClient, exporter, "")
	assert.Nil(t, err)

	rows := []feed.Action{}
	resp := exporter.DB.Table("actions").Scan(&rows)
	assert.Nil(t, resp.Error)

	assert.Equal(t, 2, len(rows))
	assert.Equal(t, "123", rows[0].ID)
}
