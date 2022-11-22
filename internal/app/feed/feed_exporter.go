package feed

import (
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/api"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/config"
)

// SafetyCultureFeedExporter defines the basic action in regard to the exporter
type SafetyCultureFeedExporter interface {
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
