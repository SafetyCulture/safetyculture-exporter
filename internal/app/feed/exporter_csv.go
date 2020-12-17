package feed

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gocarina/gocsv"

	"go.uber.org/zap"
)

// CSVExporter is an interface to export data feeds to CSV files
type CSVExporter struct {
	*SQLExporter

	ExportPath string

	Logger *zap.SugaredLogger
}

// CreateSchema generated schema for a feed in csv format
func (e *CSVExporter) CreateSchema(feed Feed, rows interface{}) error {
	e.Logger.Infof("%s: writing out CSV schema file", feed.Name())

	exportFilePath := filepath.Join(e.ExportPath, fmt.Sprintf("%s.csv", feed.Name()))
	_, err := os.Stat(exportFilePath)

	if os.IsNotExist(err) {
		file, err := os.OpenFile(exportFilePath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return err
		}

		return gocsv.Marshal(rows, file)
	}

	e.Logger.Infof("%s: skipping. CSV file already exists.", feed.Name())

	return nil
}

// FinaliseExport closes out an export
func (e *CSVExporter) FinaliseExport(feed Feed, rows interface{}) error {
	e.Logger.Infof("%s: writing out CSV file", feed.Name())

	exportFilePath := filepath.Join(e.ExportPath, fmt.Sprintf("%s.csv", feed.Name()))
	file, err := os.OpenFile(exportFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	first := true
	limit := 10000
	offset := 0
	for true {
		resp := e.DB.Table(feed.Name()).
			Order(feed.Order()).
			Limit(limit).
			Offset(offset).
			Scan(rows)
		if resp.Error != nil {
			return resp.Error
		}

		if resp.RowsAffected == 0 || resp.RowsAffected == -1 {
			break
		}

		if first {
			err = gocsv.Marshal(rows, file)
		} else {
			err = gocsv.MarshalWithoutHeaders(rows, file)
		}
		if err != nil {
			return err
		}
		offset = offset + limit
		first = false
	}

	return nil
}

// NewCSVExporter creates a new instance of CSVExporter
func NewCSVExporter(exportPath string) (*CSVExporter, error) {
	sqlExporter, err := NewSQLExporter("sqlite", filepath.Join(exportPath, "sqlite.db"), true)
	if err != nil {
		return nil, err
	}

	return &CSVExporter{
		SQLExporter: sqlExporter,
		ExportPath:  exportPath,
		Logger:      sqlExporter.Logger,
	}, nil
}
