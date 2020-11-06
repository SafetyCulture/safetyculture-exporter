package feed

import (
	"fmt"
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
	DB     *gorm.DB
	Logger *zap.SugaredLogger
}

// SupportsUpsert returns a bool if the exporter supports upserts
func (e *SQLExporter) SupportsUpsert() bool {
	return true
}

// InitFeed initialises any tables required to export
func (e *SQLExporter) InitFeed(feed Feed, opts *InitFeedOptions) error {
	model := feed.Model()

	err := e.DB.AutoMigrate(model)
	if err != nil {
		return err
	}

	if opts.Truncate {
		e.Logger.Infof("%s: truncating", feed.Name())
		result := e.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(model)
		err = result.Error
		if err != nil {
			return errors.Wrap(err, "Unable to truncate table")
		}
	}

	return nil
}

// SetLastModifiedAt updates the last modified at for the feed. No op for SQL as this is managed automatically.
func (e *SQLExporter) SetLastModifiedAt(feed Feed, ts time.Time) error {
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
	ModifiedAt time.Time
}

// LastModifiedAt returns the latest stored modified at date for the feed
func (e *SQLExporter) LastModifiedAt(feed Feed) (*time.Time, error) {
	latestRow := modifiedAtRow{}

	result := e.DB.Table(feed.Name()).Order("modified_at desc").Limit(1).First(&latestRow)
	if result.RowsAffected != 0 {
		return &latestRow.ModifiedAt, nil
	}

	return nil, nil
}

// FinaliseExport closes out an export
func (e *SQLExporter) FinaliseExport(feed Feed, rows interface{}) error {
	return nil
}

// NewSQLExporter creates a new instance of the SQLExporter
func NewSQLExporter(dialect, connectionString string) (*SQLExporter, error) {
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
		DB:     db,
		Logger: logger,
	}, nil
}
