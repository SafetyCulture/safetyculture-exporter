package exporter

import (
	"encoding/json"
	"time"
)

// TODO: Move the exporter interface code from 'feed' to here

// SafetyCultureJSONExporter interface used by JSON exporter
type SafetyCultureJSONExporter interface {
	WriteRow(name string, row *json.RawMessage)
	SetLastModifiedAt(modifiedAt time.Time)
	GetLastModifiedAt(modifiedAfter time.Time) *time.Time
}
