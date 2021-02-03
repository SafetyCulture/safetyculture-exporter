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

var lastModifiedFile *os.File

// JSONExporter is an interface to export data feeds to json files
type JSONExporter struct {
	exportPath string
}

// NewJSONExporter creates new instance of JSONExporter
func NewJSONExporter(exportPath string) Exporter {
	return &JSONExporter{
		exportPath: exportPath,
	}
}

// SetLastModifiedFile is used to set the last modified file pointer.
// Currently used only for tests.
func SetLastModifiedFile(f *os.File) {
	lastModifiedFile = f
}

// SetLastModifiedAt writes last modified date to a file
func (e *JSONExporter) SetLastModifiedAt(modifiedAt time.Time) {

	exportFilePath := filepath.Join(e.exportPath, fmt.Sprintf("%s", lastModified))
	if lastModifiedFile == nil {
		var err error
		lastModifiedFile, err = os.OpenFile(exportFilePath, os.O_RDWR|os.O_CREATE, 0666)
		util.Check(err, "Failed to open last-modified file")
	}

	str := fmt.Sprintf("%s", modifiedAt.Format(layout))
	_, err := lastModifiedFile.WriteAt([]byte(str), 0)
	util.Check(err, "Failed to write last-modified to a file")

	err = lastModifiedFile.Sync()
	util.Check(err, "Failed to write last-modified to a file")

	return
}

// GetLastModifiedAt reads last modified date from a file
func (e *JSONExporter) GetLastModifiedAt() *time.Time {
	exportFilePath := filepath.Join(e.exportPath, fmt.Sprintf("%s", lastModified))
	_, err := os.Stat(exportFilePath)
	if os.IsNotExist(err) {
		return nil
	}

	if lastModifiedFile == nil {
		var err error
		lastModifiedFile, err = os.OpenFile(exportFilePath, os.O_RDWR|os.O_CREATE, 0666)
		util.Check(err, "Failed to open last-modified file")
	}

	b := make([]byte, 50)
	_, err = lastModifiedFile.Read([]byte(b))
	util.Check(err, "Failed to read last-modified")

	modifiedAt, err := time.Parse(layout, strings.TrimSpace(string(bytes.Trim(b, "\x00"))))
	util.Check(err, "Failed to convert last-modified to iso format")
	return &modifiedAt
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
