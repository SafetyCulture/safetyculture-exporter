package feed

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gocarina/gocsv"

	"go.uber.org/zap"
)

// CSVExporter is an interface to export data feeds to CSV files
type CSVExporter struct {
	*SQLExporter

	ExportPath string

	MaxRowsPerFile int

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

	err := e.cleanOldFiles(feed.Name())
	if err != nil {
		return err
	}

	limit := 10000
	if limit > e.MaxRowsPerFile {
		limit = e.MaxRowsPerFile
	}
	offset := 0
	rowsAdded := 0
	var file *os.File
	for {
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

		if file == nil || rowsAdded >= e.MaxRowsPerFile {
			file, err = e.getExportFile(feed.Name())
			if err != nil {
				return err
			}

			err = gocsv.Marshal(rows, file)
			if err != nil {
				return err
			}

			rowsAdded = 0
		} else {
			err = gocsv.MarshalWithoutHeaders(rows, file)
			if err != nil {
				return err
			}
		}

		offset = offset + limit
		rowsAdded += int(resp.RowsAffected)
	}

	return nil
}

func (e *CSVExporter) getExportFile(feedName string) (*os.File, error) {
	exportFilePath := filepath.Join(e.ExportPath, fmt.Sprintf("%s.csv", feedName))

	fileExists, err := fileExists(exportFilePath)
	if err != nil {
		return nil, err
	}

	if fileExists {
		newFilePath := filepath.Join(e.ExportPath, fmt.Sprintf("%s-%s.csv", feedName, time.Now().Format("20060102150405.999999")))
		os.Rename(exportFilePath, newFilePath)
	}

	file, err := os.OpenFile(exportFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (e *CSVExporter) cleanOldFiles(feedName string) error {
	files, err := filepath.Glob(filepath.Join(e.ExportPath, fmt.Sprintf("%s*.csv", feedName)))
	if err != nil {
		return err
	}

	for _, f := range files {
		err = os.Remove(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func fileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}

// NewCSVExporter creates a new instance of CSVExporter
func NewCSVExporter(exportPath, exportMediaPath string, maxRowsPerFile int) (*CSVExporter, error) {
	sqlExporter, err := NewSQLExporter("sqlite", filepath.Join(exportPath, "sqlite.db"), true, exportMediaPath)
	if err != nil {
		return nil, err
	}

	return &CSVExporter{
		SQLExporter:    sqlExporter,
		ExportPath:     exportPath,
		MaxRowsPerFile: maxRowsPerFile,
		Logger:         sqlExporter.Logger,
	}, nil
}
