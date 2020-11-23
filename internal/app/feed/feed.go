package feed

import (
	"context"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
)

// Feed is an interface to a data feed. It provides methods to export the data to an exporter
type Feed interface {
	Name() string
	Model() interface{}

	PrimaryKey() []string
	Columns() []string
	Order() string

	CreateSchema(exporter Exporter) error
	Export(ctx context.Context, apiClient api.APIClient, exporter Exporter) error
}

type InitFeedOptions struct {
	Truncate bool
}

// Exporter is an interface to a Feed exporter. It provides methods to write rows out to a implemented format
type Exporter interface {
	InitFeed(feed Feed, opts *InitFeedOptions) error
	CreateSchema(feed Feed, rows interface{}) error

	WriteRows(feed Feed, rows interface{}) error
	FinaliseExport(feed Feed, rows interface{}) error
	LastModifiedAt(feed Feed) (*time.Time, error)

	SupportsUpsert() bool
	ParameterLimit() int
}
