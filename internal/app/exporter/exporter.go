package exporter

import (
	"encoding/json"
	"time"
)

// TODO: Move the exporter interface code from 'feed' to here
type Exporter interface {
	WriteRow(name string, row *json.RawMessage)
	SetLastModifiedAt(modifiedAt time.Time)
	GetLastModifiedAt(modifiedAfter time.Time) *time.Time
}
