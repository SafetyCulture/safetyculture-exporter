package feed

import (
	"context"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/internal/app/api"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/config"
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

// Exporter is an interface to a Feed exporter. It provides methods to write rows out to an implemented format
type Exporter interface {
	InitFeed(feed Feed, opts *InitFeedOptions) error
	CreateSchema(feed Feed, rows interface{}) error

	WriteRows(feed Feed, rows interface{}) error
	UpdateRows(feed Feed, primaryKeys []string, element map[string]interface{}) (int64, error)

	FinaliseExport(feed Feed, rows interface{}) error
	LastModifiedAt(feed Feed, modifiedAfter time.Time, orgID string) (time.Time, error)
	WriteMedia(auditID string, mediaID string, contentType string, body []byte) error
	DeleteRowsIfExist(feed Feed, query string, args ...interface{}) error
	GetDuration() time.Duration

	SupportsUpsert() bool
	ParameterLimit() int
}

// SafetyCultureExporter defines the basic action in regard to the exporter
type SafetyCultureExporter interface {
	// CreateSchemas will generate the schema for SafetyCulture feeds, without downloading data
	CreateSchemas(exporter Exporter) error
	// ExportFeeds will export SafetyCulture feeds and Sheqsy feeds
	ExportFeeds(exporter Exporter) error
}

type ExporterApp struct {
	exportConfig    *config.ExportConfig
	apiClient       *api.Client
	apiConfig       *config.ApiConfig
	sheqsyApiClient *api.Client
	sheqsyApiConfig *config.SheqsyApiConfig
}

func NewExporterApp(scApiClient *api.Client, sheqsyApiClient *api.Client, cfg *config.ConfigurationOptions) *ExporterApp {
	return &ExporterApp{
		exportConfig:    cfg.ExportConfig,
		apiClient:       scApiClient,
		apiConfig:       cfg.ApiConfig,
		sheqsyApiClient: sheqsyApiClient,
		sheqsyApiConfig: cfg.SheqsyApiConfig,
	}
}

// DeduplicateList a list of T type and maintains the latest value
func DeduplicateList[T any](pkFun func(element *T) string, elements []*T) []*T {
	var dMap = map[string]*T{}
	var filteredValues []*T

	if len(elements) == 0 {
		return filteredValues
	}

	for _, row := range elements {
		mapPk := pkFun(row)
		dMap[mapPk] = row
	}

	for _, row := range dMap {
		filteredValues = append(filteredValues, row)
	}
	return filteredValues
}
