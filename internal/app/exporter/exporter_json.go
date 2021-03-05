package exporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
)

const (
	lastModified = "last-modified"
	layout       = time.RFC3339Nano
)

// JSONExporter is an interface to export data feeds to json files
type JSONExporter struct {
	exportPath       string
	lastModifiedFile *os.File
}

// NewJSONExporter creates new instance of JSONExporter
func NewJSONExporter(exportPath string) Exporter {
	return &JSONExporter{
		exportPath: exportPath,
	}
}

// SetLastModifiedAt writes last modified date to a file
func (e *JSONExporter) SetLastModifiedAt(modifiedAt time.Time) {

	exportFilePath := filepath.Join(e.exportPath, fmt.Sprintf("%s", lastModified))
	if e.lastModifiedFile == nil {
		var err error
		e.lastModifiedFile, err = os.OpenFile(exportFilePath, os.O_RDWR|os.O_CREATE, 0666)
		util.Check(err, "Failed to open last-modified file")
	}

	str := fmt.Sprintf("%s", modifiedAt.Format(layout))
	_, err := e.lastModifiedFile.WriteAt([]byte(str), 0)
	util.Check(err, "Failed to write last-modified to a file")

	err = e.lastModifiedFile.Sync()
	util.Check(err, "Failed to write last-modified to a file")

	return
}

// GetLastModifiedAt returns last modified timestamp
// Value from config-path(modifiedAfter) -> (A)
// Value from last-modified file -> (B)
func (e *JSONExporter) GetLastModifiedAt(modifiedAfter time.Time) *time.Time {
	exportFilePath := filepath.Join(e.exportPath, fmt.Sprintf("%s", lastModified))
	_, err := os.Stat(exportFilePath)
	if os.IsNotExist(err) {
		return &modifiedAfter
	}

	if e.lastModifiedFile == nil {
		var err error
		e.lastModifiedFile, err = os.OpenFile(exportFilePath, os.O_RDWR|os.O_CREATE, 0666)
		util.Check(err, "Failed to open last-modified file")
	}

	b := make([]byte, 50)
	_, err = e.lastModifiedFile.Read([]byte(b))
	util.Check(err, "Failed to read last-modified")

	modifiedAt, err := time.Parse(layout, strings.TrimSpace(string(bytes.Trim(b, "\x00"))))
	util.Check(err, "Failed to convert last-modified to iso format")

	// If (A) is less than (B) then return (B)
	if !modifiedAt.IsZero() && modifiedAfter.Before(modifiedAt) {
		return &modifiedAt
	}

	return &modifiedAfter
}

// WriteRow writes the json response into a file
func (e *JSONExporter) WriteRow(name string, row *json.RawMessage) {
	str, err := json.MarshalIndent(row, "", " ")

	exportFilePath := filepath.Join(e.exportPath, fmt.Sprintf("%s.json", name))
	file, err := os.OpenFile(exportFilePath, os.O_RDWR|os.O_CREATE, 0666)
	util.Check(err, "Failed to open file")
	defer file.Close()

	_, err = file.WriteAt([]byte(str), 0)
	util.Check(err, "Failed to write inspection to a file")

	return
}
