// +build sql

package feed_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

// getTestingSQLExporter creates a temporary DB on the target SQL Database
func getTestingSQLExporter() (*feed.SQLExporter, error) {
	dialect := os.Getenv("TEST_DB_DIALECT")
	connectionString := os.Getenv("TEST_DB_CONN_STRING")

	fmt.Println(connectionString)
	exporter, err := feed.NewSQLExporter(dialect, connectionString, true)
	if err != nil {
		return nil, err
	}

	dbName := strings.ReplaceAll(fmt.Sprintf("iaud_exporter_%s", uuid.Must(uuid.NewV4()).String()), "-", "")

	switch dialect {
	case "postgres":
		dbResp := exporter.DB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		err = dbResp.Error
		break
	case "mysql":
		dbResp := exporter.DB.Exec(fmt.Sprintf(`CREATE DATABASE %s;`, dbName))
		err = dbResp.Error
		break
	case "sqlserver":
		dbResp := exporter.DB.Exec(fmt.Sprintf(`CREATE DATABASE %s;`, dbName))
		err = dbResp.Error
		break
	default:
		return nil, fmt.Errorf("Invalid DB dialect %s", dialect)
	}
	if err != nil {
		return nil, err
	}

	connectionString = strings.Replace(connectionString, "iauditor_exporter_db", dbName, 1)
	connectionString = strings.Replace(connectionString, "master", dbName, 1)

	return feed.NewSQLExporter(dialect, connectionString, true)
}

func TestIntegrationDbExportFeeds_integration_should_not_initialise_schemas_with_auto_migrate_disabled(t *testing.T) {
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

func TestIntegrationDbExportFeeds_integration_should_initialise_schemas(t *testing.T) {
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
