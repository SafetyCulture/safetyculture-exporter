package feed_test

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"github.com/stretchr/testify/assert"
)

func TestCSVExporterSupportsUpsert_should_return_true(t *testing.T) {
	exporter, err := getTemporaryCSVExporter()
	assert.Nil(t, err)

	assert.True(t, exporter.SupportsUpsert())
}

func TestCSVExporterInitFeed_should_create_table_if_not_exists(t *testing.T) {
	exporter, err := getTemporaryCSVExporter()
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

func TestCSVExporterInitFeed_should_truncate_table_if_truncate_is_true(t *testing.T) {
	exporter, err := getTemporaryCSVExporter()
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

func TestCSVExporterInitFeed_should_not_truncate_table_if_truncate_is_false(t *testing.T) {
	exporter, err := getTemporaryCSVExporter()
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

func TestCSVExporterWriteRows_should_write_rows(t *testing.T) {
	exporter, err := getTemporaryCSVExporter()
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

func TestCSVExporterWriteRows_should_update_rows(t *testing.T) {
	exporter, err := getTemporaryCSVExporter()
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

func TestCSVExporterLastModifiedAt_should_return_latest_modified_at(t *testing.T) {
	exporter, err := getTemporaryCSVExporter()
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

func TestCSVExporterFinaliseExport_should_write_rows_out_to_file(t *testing.T) {
	exporter, err := getTemporaryCSVExporter()
	assert.Nil(t, err)

	userFeed := &feed.UserFeed{}

	err = exporter.InitFeed(userFeed, &feed.InitFeedOptions{
		Truncate: false,
	})
	assert.Nil(t, err)

	users := []feed.User{
		{
			ID:             "user_1",
			OrganisationID: "role_123",
			Email:          "user.1@test.com",
			Firstname:      "User 1",
			Lastname:       "User 1",
		},
		{
			ID:             "user_2",
			OrganisationID: "role_123",
			Email:          "user.2@test.com",
			Firstname:      "User 2",
			Lastname:       "User 2",
		},
	}

	err = exporter.WriteRows(userFeed, users)
	assert.Nil(t, err)

	users = []feed.User{
		{
			ID:             "user_1",
			OrganisationID: "role_123",
			Email:          "user.1@test.com",
			Firstname:      "User 1",
			Lastname:       "User 1",
		},
		{
			ID:             "user_2",
			OrganisationID: "role_123",
			Email:          "user.2@test.com",
			Firstname:      "User 2",
			Lastname:       "User 2",
		},
		{
			ID:             "user_3",
			OrganisationID: "role_123",
			Email:          "user.3@test.com",
			Firstname:      "User 3",
			Lastname:       "User 3",
		},
	}

	err = exporter.WriteRows(userFeed, users)
	assert.Nil(t, err)

	err = exporter.FinaliseExport(userFeed, &[]feed.User{})
	assert.Nil(t, err)

	content, err := ioutil.ReadFile(filepath.Join(exporter.ExportPath, "users.csv"))
	assert.Nil(t, err)

	contentString := dateRegex.ReplaceAllLiteralString(strings.TrimSpace(string(content)), "--date--")

	expected := `user_id,organisation_id,email,firstname,lastname,active,exported_at
user_1,role_123,user.1@test.com,User 1,User 1,false,--date--
user_2,role_123,user.2@test.com,User 2,User 2,false,--date--
user_3,role_123,user.3@test.com,User 3,User 3,false,--date--`
	assert.Equal(t, strings.TrimSpace(expected), contentString)
}
