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
	duration        time.Duration
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
func (e *SQLExporter) CreateSchema(feed Feed, _ interface{}) error {
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

// GetDuration will return the duration for exporting a batch
func (e *SQLExporter) GetDuration() time.Duration {
	return e.duration
}

// DeleteRowsIfExist will delete the rows if already exist
func (e *SQLExporter) DeleteRowsIfExist(feed Feed, query string, args ...interface{}) error {
	del := e.DB.Table(feed.Name()).
		Clauses(clause.Where{
			Exprs: []clause.Expression{
				clause.Expr{
					SQL:  query,
					Vars: args,
				},
			},
		}).
		Delete(feed.Model())
	if del.Error != nil {
		return errors.Wrap(del.Error, "Unable to delete rows")
	}

	return nil
}

// WriteRows writes out the rows to the DB
func (e *SQLExporter) WriteRows(feed Feed, rows interface{}) error {
	columns := []clause.Column{}
	for _, column := range feed.PrimaryKey() {
		columns = append(columns, clause.Column{Name: column})
	}

	start := time.Now()
	insert := e.DB.Table(feed.Name()).
		Clauses(clause.OnConflict{
			Columns:   columns,
			DoUpdates: clause.AssignmentColumns(feed.Columns()),
		}).
		Create(rows)
	e.duration = time.Since(start)
	if insert.Error != nil {
		return errors.Wrap(insert.Error, "Unable to insert rows")
	}

	return nil
}

// UpdateRows batch updates. Returns number of rows updated or error. Works with single PKey, not with composed PKeys
func (e *SQLExporter) UpdateRows(feed Feed, primaryKeys []string, element map[string]interface{}) (int64, error) {
	result := e.DB.
		Table(feed.Name()).
		//Where(fmt.Sprintf("%s in ?", feed.PrimaryKey()[0]), primaryKeys).
		Where(primaryKeys).
		Updates(element)
	if result.Error != nil {
		return 0, errors.Wrap(result.Error, "Unable to update rows")
	}

	return result.RowsAffected, nil
}

type modifiedAtRow struct {
	// ExportedAt is here so gorm has an additional field to sort on in SQL Server
	ExportedAt time.Time
	ModifiedAt time.Time
}

// LastModifiedAt returns the latest stored modified at date for the feed
func (e *SQLExporter) LastModifiedAt(feed Feed, modifiedAfter time.Time, orgID string) (time.Time, error) {
	latestRow := modifiedAtRow{}

	var result *gorm.DB
	result = e.DB.Table(feed.Name()).
		Where("organisation_id = ?", orgID).
		Order("modified_at DESC").
		Limit(1).
		First(&latestRow)
	if result.RowsAffected == 0 {
		// This can happen when there is no org_id stored in the existing data.
		// In this case try to get the latest modifiedAt timestamp  from the table
		// where there is no org_id defined.
		result = e.DB.Table(feed.Name()).
			Where("organisation_id IS NULL OR organisation_id = ''").
			Order("modified_at DESC").
			Limit(1).
			First(&latestRow)
	}
	if result.RowsAffected != 0 && modifiedAfter.Before(latestRow.ModifiedAt) {
		return latestRow.ModifiedAt, nil
	}

	return modifiedAfter, nil
}

// FinaliseExport closes out an export
func (e *SQLExporter) FinaliseExport(Feed, interface{}) error {
	return nil
}

// WriteMedia writes the media to a file
func (e *SQLExporter) WriteMedia(auditID, mediaID, contentType string, body []byte) error {

	exportMediaDir := filepath.Join(e.ExportMediaPath, auditID)
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
		SlowThreshold: 30 * time.Second,
	}

	var dialector gorm.Dialector
	switch dialect {
	case "mysql":
		dialector = mysql.Open(connectionString)
	case "postgres":
		dialector = postgres.Open(connectionString)
	case "sqlserver":
		dialector = sqlserver.Open(connectionString)
	case "sqlite":
		dialector = sqlite.Open(connectionString)
	default:
		return nil, fmt.Errorf("invalid database dialect %s", dialect)
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
		duration:        0,
	}, nil
}
