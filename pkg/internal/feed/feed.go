package feed

//go:generate mockery --name Exporter --case underscore --output mocks --outpkg mocks --filename exporter_mock.go

import (
	"context"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
)

// Feed is an interface to a data feed. It provides methods to export the data to an exporter
type Feed interface {
	Name() string
	Model() interface{}
	RowsModel() interface{}

	PrimaryKey() []string
	Columns() []string
	Order() string

	CreateSchema(exporter Exporter) error
	Export(ctx context.Context, apiClient *httpapi.Client, exporter Exporter, orgID string, status *ExportStatus) error
}

// InitFeedOptions contains the options used when initialising a feed
type InitFeedOptions struct {
	Truncate bool
}

// Exporter is an interface to a Feed exporter. It provides methods to write rows out to an implemented format
type Exporter interface {
	InitFeed(feed Feed, opts *InitFeedOptions) error
	CreateSchema(feed Feed, rows interface{}) error

	WriteRows(feed Feed, rows interface{}) error
	UpdateRows(feed Feed, primaryKeys []string, element map[string]interface{}) (int64, error)

	FinaliseExport(feed Feed, rows interface{}, status *ExportStatus) error
	LastModifiedAt(feed Feed, modifiedAfter time.Time, orgID string) (time.Time, error)
	WriteMedia(auditID string, mediaID string, contentType string, body []byte) error
	DeleteRowsIfExist(feed Feed, query string, args ...interface{}) error
	GetDuration() time.Duration

	SupportsUpsert() bool
	ParameterLimit() int
}
