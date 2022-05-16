package feed

import (
	"context"
	"fmt"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
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
	Export(ctx context.Context, apiClient *api.Client, exporter Exporter, orgID string) error
}

// InitFeedOptions contains the options used when initialising a feed
type InitFeedOptions struct {
	Truncate bool
}

// Exporter is an interface to a Feed exporter. It provides methods to write rows out to a implemented format
type Exporter interface {
	InitFeed(feed Feed, opts *InitFeedOptions) error
	CreateSchema(feed Feed, rows interface{}) error

	WriteRows(feed Feed, rows interface{}) error
	FinaliseExport(feed Feed, rows interface{}) error
	LastModifiedAt(feed Feed, modifiedAfter time.Time, orgID string) (time.Time, error)
	WriteMedia(auditID string, mediaID string, contentType string, body []byte) error
	DeleteRowsIfExist(feed Feed, query string, args ...interface{}) error
	GetDuration() time.Duration

	SupportsUpsert() bool
	ParameterLimit() int
}

// LogStringConfig is the config for GetLogString function
type LogStringConfig struct {
	RemainingRecords int64
	HttpDuration     time.Duration
	ExporterDuration time.Duration
}

// GetLogString build a log string based on input arguments
func GetLogString(feedName string, cfg *LogStringConfig) string {
	var args = []any{feedName}
	var format = "%s: "

	if cfg != nil {
		format = format + "%d remaining."
		args = append(args, cfg.RemainingRecords)

		if cfg.HttpDuration.Milliseconds() != 0 {
			format = format + " Last http call was %dms."
			args = append(args, cfg.HttpDuration.Milliseconds())
		}

		if cfg.ExporterDuration.Milliseconds() != 0 {
			format = format + " Last export operation was %dms."
			args = append(args, cfg.ExporterDuration.Milliseconds())
		}
	}

	return fmt.Sprintf(format, args...)
}
