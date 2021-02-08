package feed

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SQLExporter is an interface to export data feeds to SQL databases
type SQLExporter struct {
	DB              *gorm.DB
	Logger          *zap.SugaredLogger
	AutoMigrate     bool
	ExportMediaPath string
}

// SupportsUpsert returns a bool if the exporter supports upserts
func (e *SQLExporter) SupportsUpsert() bool {
	return true
}

// ParameterLimit returns the number of parameters supported by the target DB
func (e *SQLExporter) ParameterLimit() int {
	switch e.DB.Dialector.Name() {
	case "sqlserver":
		return 2100
	case "sqlite":
		return 32768
	}

	return 65536
}

// CreateSchema creates the schema on the DB for the supplied feed
func (e *SQLExporter) CreateSchema(feed Feed, rows interface{}) error {
	return e.InitFeed(feed, &InitFeedOptions{
		Truncate: false,
	})
}

// InitFeed initialises any tables required to export
func (e *SQLExporter) InitFeed(feed Feed, opts *InitFeedOptions) error {
	model := feed.Model()

	if e.AutoMigrate {
		err := e.DB.AutoMigrate(model)
		if err != nil {
			return err
		}
	}

	if opts.Truncate {
		e.Logger.Infof("%s: truncating", feed.Name())
		result := e.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(model)
		if result.Error != nil {
			return errors.Wrap(result.Error, "Unable to truncate table")
		}
	}

	return nil
}

// WriteRows writes out the rows to the DB
func (e *SQLExporter) WriteRows(feed Feed, rows interface{}) error {
	columns := []clause.Column{}
	for _, column := range feed.PrimaryKey() {
		columns = append(columns, clause.Column{Name: column})
	}

	insert := e.DB.Table(feed.Name()).
		Clauses(clause.OnConflict{
			Columns:   columns,
			DoUpdates: clause.AssignmentColumns(feed.Columns()),
		}).
		Create(rows)
	if insert.Error != nil {
		return errors.Wrap(insert.Error, "Unable to insert rows")
	}

	return nil
}

type modifiedAtRow struct {
	// ExportedAt is here so gorm has an additional field to sort on in SQL Server
	ExportedAt time.Time
	ModifiedAt time.Time
}

// LastModifiedAt returns the latest stored modified at date for the feed
func (e *SQLExporter) LastModifiedAt(feed Feed, modifiedAfter time.Time) (time.Time, error) {
	latestRow := modifiedAtRow{}

	result := e.DB.Table(feed.Name()).Order("modified_at DESC").Limit(1).First(&latestRow)
	if result.RowsAffected != 0 && modifiedAfter.Before(latestRow.ModifiedAt) {
		return latestRow.ModifiedAt, nil
	}

	return modifiedAfter, nil
}

// FinaliseExport closes out an export
func (e *SQLExporter) FinaliseExport(feed Feed, rows interface{}) error {
	return nil
}

// WriteMedia writes the media to a file
func (e *SQLExporter) WriteMedia(auditID, mediaID, contentType string, body []byte) error {

	exportMediaDir := filepath.Join(e.ExportMediaPath, fmt.Sprintf("%s", auditID))
	err := os.MkdirAll(exportMediaDir, os.ModePerm)
	util.Check(err, fmt.Sprintf("Failed to create directory %s", exportMediaDir))

	ext := strings.Split(contentType, "/")
	exportFilePath := filepath.Join(exportMediaDir, fmt.Sprintf("%s.%s", mediaID, ext[1]))

	file, err := os.OpenFile(exportFilePath, os.O_RDWR|os.O_CREATE, 0666)
	util.Check(err, fmt.Sprintf("Failed to open file: %v", exportFilePath))
	defer file.Close()

	_, err = file.WriteAt(body, 0)
	util.Check(err, "Failed to write media to a file")

	return nil
}

// NewSQLExporter creates a new instance of the SQLExporter
func NewSQLExporter(dialect, connectionString string, autoMigrate bool, exportMediaPath string) (*SQLExporter, error) {
	logger := util.GetLogger()
	gormLogger := &util.GormLogger{
		SugaredLogger: logger,
		SlowThreshold: time.Second,
	}

	var dialector gorm.Dialector
	switch dialect {
	case "mysql":
		dialector = mysql.Open(connectionString)
		break
	case "postgres":
		dialector = postgres.Open(connectionString)
		break
	case "sqlserver":
		dialector = sqlserver.Open(connectionString)
		break
	case "sqlite":
		dialector = sqlite.Open(connectionString)
		break
	default:
		return nil, fmt.Errorf("Invalid database dialect %s", dialect)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Unable to connect to DB")
	}

	return &SQLExporter{
		DB:              db,
		Logger:          logger,
		AutoMigrate:     autoMigrate,
		ExportMediaPath: exportMediaPath,
	}, nil
}
