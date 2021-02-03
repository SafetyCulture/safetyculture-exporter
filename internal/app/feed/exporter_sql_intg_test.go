// +build sql

package feed_test

import (
	"testing"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"github.com/stretchr/testify/assert"
)

func TestIntegrationDbSQLExporterLastModifiedAt_should_return_latest_modified_at(t *testing.T) {
	exporter, err := getTestingSQLExporter()
	assert.Nil(t, err)

	inspectionFeed := &feed.InspectionFeed{}

	err = exporter.InitFeed(inspectionFeed, &feed.InitFeedOptions{
		Truncate: false,
	})
	assert.Nil(t, err)

	now := time.Now()
	inspections := []feed.Inspection{
		{
			ID:           "audit_1",
			DateStarted:  now,
			DateModified: now,
			CreatedAt:    now,
			ExportedAt:   now,
			ModifiedAt:   now,
		},
		{
			ID:           "audit_2",
			DateStarted:  time.Now().Add(time.Hour * -128),
			DateModified: time.Now().Add(time.Hour * -128),
			CreatedAt:    time.Now().Add(time.Hour * -128),
			ExportedAt:   time.Now().Add(time.Hour * -128),
			ModifiedAt:   time.Now().Add(time.Hour * -128),
		},
		{
			ID:           "audit_3",
			DateStarted:  time.Now().Add(time.Hour * -3000),
			DateModified: time.Now().Add(time.Hour * -3000),
			CreatedAt:    time.Now().Add(time.Hour * -3000),
			ExportedAt:   time.Now().Add(time.Hour * -3000),
			ModifiedAt:   time.Now().Add(time.Hour * -3000),
		},
		{
			ID:           "audit_4",
			DateStarted:  time.Now().Add(time.Hour * -2),
			DateModified: time.Now().Add(time.Hour * -2),
			CreatedAt:    time.Now().Add(time.Hour * -2),
			ExportedAt:   time.Now().Add(time.Hour * -2),
			ModifiedAt:   time.Now().Add(time.Hour * -2),
		},
	}

	err = exporter.WriteRows(inspectionFeed, inspections)
	assert.Nil(t, err)

	lastModifiedAt, err := exporter.LastModifiedAt(inspectionFeed)
	assert.Nil(t, err)
	// Times are slightly lossy, convery to ISO string
	assert.Equal(t, now.Format(time.RFC3339), lastModifiedAt.Format(time.RFC3339))
}

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
