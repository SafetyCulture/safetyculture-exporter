package feed_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

var dateRegex = regexp.MustCompile(`(?m)(-?(?:[1-9][0-9]*)?[0-9]{4})-(1[0-2]|0[1-9])-(3[01]|0[1-9]|[12][0-9])T(2[0-3]|[01][0-9]):([0-5][0-9]):([0-5][0-9])(\.[0-9]+)?(\+|Z)(2[0-3]|[01][0-9])?:?([0-5][0-9])?`)

// getInmemorySQLExporter creates a SQLExporter that uses an inmemory DB
func getInmemorySQLExporter() (*feed.SQLExporter, error) {
	return feed.NewSQLExporter("sqlite", "file::memory:", true)
}

// getTemporaryCSVExporter creates a CSVExporter that writes to a temp folder
func getTemporaryCSVExporter() (*feed.CSVExporter, error) {
	dir, err := ioutil.TempDir("", "export")
	if err != nil {
		log.Fatal(err)
	}

	return feed.NewCSVExporter(dir)
}

// getTemporaryCSVExporterWithRealSQLExporter creates a CSV exporter that writes a temporary folder
// but also uses a real DB as an intermediary
func getTemporaryCSVExporterWithRealSQLExporter(sqlExporter *feed.SQLExporter) (*feed.CSVExporter, error) {
	dir, err := ioutil.TempDir("", "export")
	if err != nil {
		return nil, err
	}

	exporter, err := feed.NewCSVExporter(dir)
	if err != nil {
		return nil, err
	}

	exporter.SQLExporter = sqlExporter

	return exporter, err
}

// getTestingSQLExporter creates a temporary DB on the target SQL Database
func getTestingSQLExporter() (*feed.SQLExporter, error) {
	dialect := os.Getenv("TEST_DB_DIALECT")
	connectionString := os.Getenv("TEST_DB_CONN_STRING")

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

// filesEqualish checks if files are equal enough (ignoring dates)
func filesEqualish(t *testing.T, expectedPath, actualPath string) {
	expectedFile, err := ioutil.ReadFile(expectedPath)
	assert.Nil(t, err)

	actualFile, err := ioutil.ReadFile(actualPath)
	assert.Nil(t, err)

	assert.Equal(t,
		dateRegex.ReplaceAllLiteralString(strings.TrimSpace(string(expectedFile)), "--date--"),
		dateRegex.ReplaceAllLiteralString(strings.TrimSpace(string(actualFile)), "--date--"),
	)
}
