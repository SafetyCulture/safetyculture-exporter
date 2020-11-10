// +build sql

package feed_test

import (
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"github.com/stretchr/testify/assert"
)

func TestIntegrationDbSQLExporterInitFeed_integration_should_not_initialise_schemas_with_auto_migrate_disabled(t *testing.T) {
	exporter, err := getTestingSQLExporter()
	assert.Nil(t, err)
	exporter.AutoMigrate = false

	userFeed := &feed.UserFeed{}

	err = exporter.InitFeed(userFeed, &feed.InitFeedOptions{
		Truncate: false,
	})
	assert.Nil(t, err)

	users := []feed.User{
		{
			ID:        "user_1",
			Firstname: "User One",
			Lastname:  "User One",
		},
	}

	err = exporter.WriteRows(userFeed, users)
	assert.NotNil(t, err, "Should throw an error when attempting to insert to a table that doesn't exist")
}

func TestIntegrationDbSQLExporterInitFeed_integration_should_initialise_schemas(t *testing.T) {
	exporter, err := getTestingSQLExporter()
	assert.Nil(t, err)

	userFeed := &feed.UserFeed{}

	err = exporter.InitFeed(userFeed, &feed.InitFeedOptions{
		Truncate: false,
	})
	assert.Nil(t, err)

	users := []feed.User{
		{
			ID:        "user_1",
			Firstname: "User One",
			Lastname:  "User One",
		},
	}

	err = exporter.WriteRows(userFeed, users)
	assert.Nil(t, err)
}
