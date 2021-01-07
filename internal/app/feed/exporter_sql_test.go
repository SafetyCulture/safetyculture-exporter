package feed_test

import (
	"testing"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"github.com/stretchr/testify/assert"
)

func TestSQLExporterSupportsUpsert_should_return_true(t *testing.T) {
	exporter, err := getInmemorySQLExporter()
	assert.Nil(t, err)

	assert.True(t, exporter.SupportsUpsert())
}

func TestSQLExporterInitFeed_should_create_table_if_not_exists(t *testing.T) {
	exporter, err := getInmemorySQLExporter()
	assert.Nil(t, err)

	userFeed := &feed.UserFeed{}

	err = exporter.InitFeed(userFeed, &feed.InitFeedOptions{
		Truncate: false,
	})
	assert.Nil(t, err)

	// This query will only work for SQLite
	result := []struct {
		Name string
	}{}

	resp := exporter.DB.Raw("SELECT name FROM sqlite_master WHERE type='table';").Scan(&result)
	assert.Nil(t, resp.Error)

	assert.Len(t, result, 1)
	assert.Equal(t, "users", result[0].Name)
}

func TestSQLExporterInitFeed_should_truncate_table_if_truncate_is_true(t *testing.T) {
	exporter, err := getInmemorySQLExporter()
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

	err = exporter.InitFeed(userFeed, &feed.InitFeedOptions{
		Truncate: true,
	})
	assert.Nil(t, err)

	var rowCount int64
	resp := exporter.DB.Table("users").Count(&rowCount)
	assert.Nil(t, resp.Error)
	assert.Equal(t, int64(0), rowCount)
}

func TestSQLExporterInitFeed_should_not_truncate_table_if_truncate_is_false(t *testing.T) {
	exporter, err := getInmemorySQLExporter()
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

	err = exporter.InitFeed(userFeed, &feed.InitFeedOptions{
		Truncate: false,
	})
	assert.Nil(t, err)

	var rowCount int64
	resp := exporter.DB.Table("users").Count(&rowCount)
	assert.Nil(t, resp.Error)
	assert.Equal(t, int64(1), rowCount)
}

func TestSQLExporterWriteRows_should_write_rows(t *testing.T) {
	exporter, err := getInmemorySQLExporter()
	assert.Nil(t, err)

	userFeed := &feed.UserFeed{}

	err = exporter.InitFeed(userFeed, &feed.InitFeedOptions{
		Truncate: false,
	})
	assert.Nil(t, err)

	users := []feed.User{
		{
			ID:        "user_1",
			Firstname: "User 1",
			Lastname:  "User 1",
		},
		{
			ID:        "user_2",
			Firstname: "User 2",
			Lastname:  "User 2",
		},
	}

	err = exporter.WriteRows(userFeed, users)
	assert.Nil(t, err)

	rows := []feed.User{}
	resp := exporter.DB.Table("users").Scan(&rows)
	assert.Nil(t, resp.Error)

	assert.Equal(t, 2, len(rows))
	assert.Equal(t, "user_1", rows[0].ID)
	assert.Equal(t, "user_2", rows[1].ID)
}

func TestSQLExporterWriteRows_should_update_rows(t *testing.T) {
	exporter, err := getInmemorySQLExporter()
	assert.Nil(t, err)

	userFeed := &feed.UserFeed{}

	err = exporter.InitFeed(userFeed, &feed.InitFeedOptions{
		Truncate: true,
	})
	assert.Nil(t, err)

	users := []feed.User{
		{
			ID:        "user_1",
			Firstname: "User 1",
			Lastname:  "User 1",
		},
	}

	err = exporter.WriteRows(userFeed, users)
	assert.Nil(t, err)

	users = []feed.User{
		{
			ID:        "user_1",
			Firstname: "User One",
			Lastname:  "User One",
		},
	}

	err = exporter.WriteRows(userFeed, users)
	assert.Nil(t, err)

	rows := []feed.User{}
	resp := exporter.DB.Table("users").Scan(&rows)
	assert.Nil(t, resp.Error)

	assert.Equal(t, 1, len(rows))
	assert.Equal(t, "user_1", rows[0].ID)
	assert.Equal(t, "User One", rows[0].Firstname)
	assert.Equal(t, "User One", rows[0].Lastname)
}

func TestSQLExporterLastModifiedAt_should_return_latest_modified_at(t *testing.T) {
	exporter, err := getInmemorySQLExporter()
	assert.Nil(t, err)

	inspectionFeed := &feed.InspectionFeed{}

	err = exporter.InitFeed(inspectionFeed, &feed.InitFeedOptions{
		Truncate: false,
	})
	assert.Nil(t, err)

	now := time.Now()
	inspections := []feed.Inspection{
		{
			ID:         "audit_1",
			ModifiedAt: now,
		},
		{
			ID:         "audit_2",
			ModifiedAt: time.Now().Add(time.Hour * -128),
		},
		{
			ID:         "audit_3",
			ModifiedAt: time.Now().Add(time.Hour * -3000),
		},
		{
			ID:         "audit_4",
			ModifiedAt: time.Now().Add(time.Hour * -2),
		},
	}

	err = exporter.WriteRows(inspectionFeed, inspections)
	assert.Nil(t, err)

	lastModifiedAt, err := exporter.LastModifiedAt(inspectionFeed)
	assert.Nil(t, err)
	// Times are slightly lossy, convery to ISO string
	assert.Equal(t, now.Format(time.RFC3339), lastModifiedAt.Format(time.RFC3339))
}

func TestNewSQLExporter_should_create_exporter_for_sqlite(t *testing.T) {
	sqlExporter, err := feed.NewSQLExporter("sqlite", "file::memory:", true, "")
	assert.Nil(t, err)

	assert.NotNil(t, sqlExporter)
}

func TestNewSQLExporter_should_return_error_for_invalid_dialect(t *testing.T) {
	sqlExporter, err := feed.NewSQLExporter("not-supported", "file::memory:", true, "")
	assert.NotNil(t, err)
	assert.Nil(t, sqlExporter)
}

func TestNewSQLExporter_should_return_error_for_connection_errors(t *testing.T) {
	sqlExporter, err := feed.NewSQLExporter("sqlite", "*$////bad connection string", true, "")
	assert.NotNil(t, err)
	assert.Nil(t, sqlExporter)
}
