package feed_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/feed"
	"github.com/gofrs/uuid"
)

// getTemporaryCSVExporterWithRealSQLExporter creates a CSV exporter that writes a temporary folder
// but also uses a real DB as an intermediary
func getTemporaryCSVExporterWithRealSQLExporter(sqlExporter *feed.SQLExporter) (*feed.CSVExporter, error) {
	dir, err := os.MkdirTemp("", "export")
	if err != nil {
		return nil, err
	}

	exporter, err := feed.NewCSVExporter(dir, "", 100000)
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

	exporter, err := feed.NewSQLExporter(dialect, connectionString, true, "")
	if err != nil {
		return nil, err
	}

	dbName := strings.ReplaceAll(fmt.Sprintf("iaud_exporter_%s", uuid.Must(uuid.NewV4()).String()), "-", "")

	switch dialect {
	case "postgres", "mysql", "sqlserver":
		dbResp := exporter.DB.Exec(fmt.Sprintf(`CREATE DATABASE %s;`, dbName))
		err = dbResp.Error
	case "sqlite":
		return exporter, nil
	default:
		return nil, fmt.Errorf("Invalid DB dialect %s", dialect)
	}
	if err != nil {
		return nil, err
	}

	connectionString = strings.Replace(connectionString, "safetyculture_exporter_db", dbName, 1)
	connectionString = strings.Replace(connectionString, "master", dbName, 1)

	return feed.NewSQLExporter(dialect, connectionString, true, "")
}
