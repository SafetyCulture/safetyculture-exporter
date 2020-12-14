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

type JSONExporter struct {
	exportPath string
}

func NewJSONExporter(exportPath string) Exporter {
	return &JSONExporter{
		exportPath: exportPath,
	}
}

func (e *JSONExporter) SetLastModifiedAt(modifiedAt time.Time) {
	exportFilePath := filepath.Join(e.exportPath, fmt.Sprintf("%s", lastModified))
	if lastModifiedFile == nil {
		var err error
		lastModifiedFile, err = os.OpenFile(exportFilePath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			util.Check(err, "Failed to open last-modified file")
		}
	}

	str := fmt.Sprintf("%s", modifiedAt.Format(layout))
	if _, err := lastModifiedFile.WriteAt([]byte(str), 0); err != nil {
		util.Check(err, "Failed to write last-modified to a file")
	}

	if err := lastModifiedFile.Sync(); err != nil {
		util.Check(err, "Failed to write last-modified to a file")
	}

	return
}

func (e *JSONExporter) GetLastModifiedAt() *time.Time {

	exportFilePath := filepath.Join(e.exportPath, fmt.Sprintf("%s", lastModified))
	_, err := os.Stat(exportFilePath)
	if os.IsNotExist(err) {
		return nil
	}

	if lastModifiedFile == nil {
		var err error
		lastModifiedFile, err = os.OpenFile(exportFilePath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			util.Check(err, "Failed to open last-modified file")
		}
	}

	b := make([]byte, 50)
	if _, err := lastModifiedFile.Read([]byte(b)); err != nil {
		util.Check(err, "Failed to read last-modified")
	}
	modifiedAt, err := time.Parse(layout, strings.TrimSpace(string(bytes.Trim(b, "\x00"))))
	util.Check(err, "Failed to convert last-modified to iso format")

	return &modifiedAt
}

func (e *JSONExporter) WriteRow(name string, row *json.RawMessage) {
	str, err := json.MarshalIndent(row, "", " ")

	exportFilePath := filepath.Join(e.exportPath, fmt.Sprintf("%s.json", name))
	file, err := os.OpenFile(exportFilePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		util.Check(err, "Failed to open file")
	}
	defer file.Close()

	if _, err := file.WriteAt([]byte(str), 0); err != nil {
		util.Check(err, "Failed to write inspection to a file")
	}

	return
}
