package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/templates"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/spf13/viper"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/exporter"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/feed"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/inspections"
	"github.com/pkg/errors"
)

// NewSafetyCultureExporter builds a SafetyCultureExporter with clients inferred from own configuration
func NewSafetyCultureExporter(cfg *ExporterConfiguration, version *AppVersion) (*SafetyCultureExporter, error) {
	apiClient, err := getAPIClient(cfg.ToApiConfig())
	if err != nil {
		return nil, err
	}
	apiClient.SetVersion(version.IntegrationID, version.IntegrationVersion)

	sheqsyApiClient, err := getSheqsyAPIClient(cfg.ToApiConfig())
	if err != nil {
		return nil, err
	}
	sheqsyApiClient.SetVersion(version.IntegrationID, version.IntegrationVersion)

	return &SafetyCultureExporter{
		apiClient:       apiClient,
		sheqsyApiClient: sheqsyApiClient,
		cfg:             cfg,
		exportStatus:    feed.NewExportStatus(),
	}, nil
}

func getAPIClient(cfg *HttpApiCfg) (*httpapi.Client, error) {
	var apiOpts []httpapi.Opt

	if cfg.tlsSkipVerify {
		apiOpts = append(apiOpts, httpapi.OptSetInsecureTLS(true))
	}
	if cfg.tlsCert != "" {
		apiOpts = append(apiOpts, httpapi.OptAddTLSCert(viper.GetString("api.tls_cert")))
	}
	if cfg.proxyUrl != "" {
		proxyURL, err := url.Parse(cfg.proxyUrl)
		if err != nil {
			return nil, fmt.Errorf("unable to parse proxy URL")
		}
		apiOpts = append(apiOpts, httpapi.OptSetProxy(proxyURL))
	}

	return httpapi.NewClient(
		cfg.apiUrl,
		fmt.Sprintf("Bearer %s", cfg.accessToken),
		apiOpts...,
	), nil
}

func getSheqsyAPIClient(cfg *HttpApiCfg) (*httpapi.Client, error) {
	var apiOpts []httpapi.Opt
	if cfg.tlsSkipVerify {
		apiOpts = append(apiOpts, httpapi.OptSetInsecureTLS(true))
	}
	if cfg.tlsCert != "" {
		apiOpts = append(apiOpts, httpapi.OptAddTLSCert(viper.GetString("api.tls_cert")))
	}
	if viper.GetString("api.proxy_url") != "" {
		proxyURL, err := url.Parse(cfg.proxyUrl)
		if err != nil {
			return nil, fmt.Errorf("unable to parse proxy URL")
		}
		apiOpts = append(apiOpts, httpapi.OptSetProxy(proxyURL))
	}

	token := base64.StdEncoding.EncodeToString(
		[]byte(
			fmt.Sprintf(
				"%s:%s",
				cfg.sheqsyUsername,
				cfg.sheqsyPassword,
			),
		),
	)

	return httpapi.NewClient(
		cfg.sheqsyApiUrl,
		fmt.Sprintf("Basic %s", token),
		apiOpts...,
	), nil
}

// NewReportExporter returns a new instance of ReportExporter
func NewReportExporter(exportPath string, reportCfg *ReportExporterCfg) (*feed.ReportExporter, error) {
	sqlExporter, err := feed.NewSQLExporter("sqlite", filepath.Join(exportPath, "reports.db"), true, "")
	if err != nil {
		return nil, err
	}

	return &feed.ReportExporter{
		SQLExporter:  sqlExporter,
		Logger:       sqlExporter.Logger,
		ExportPath:   exportPath,
		Format:       reportCfg.Format,
		PreferenceID: reportCfg.PreferenceID,
		Filename:     reportCfg.Filename,
		RetryTimeout: reportCfg.RetryTimeout,
	}, nil
}

type ReportExporterCfg struct {
	Format       []string
	PreferenceID string
	Filename     string
	RetryTimeout int
}

type HttpApiCfg struct {
	tlsSkipVerify  bool
	tlsCert        string
	proxyUrl       string
	apiUrl         string
	accessToken    string
	sheqsyApiUrl   string
	sheqsyUsername string
	sheqsyPassword string
}

type SafetyCultureExporter struct {
	apiClient       *httpapi.Client
	sheqsyApiClient *httpapi.Client
	cfg             *ExporterConfiguration
	exportStatus    *feed.ExportStatus
}

func (s *SafetyCultureExporter) SetApiClient(apiClient *httpapi.Client) {
	s.apiClient = apiClient
}

func (s *SafetyCultureExporter) SetSheqsyApiClient(apiClient *httpapi.Client) {
	s.sheqsyApiClient = apiClient
}

func (s *SafetyCultureExporter) RunInspectionJSON() error {
	exportPath := fmt.Sprintf("%s/json/", s.cfg.Export.Path)
	err := os.MkdirAll(exportPath, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "Failed to create directory %s", exportPath)
	}

	e := exporter.NewJSONExporter(exportPath)
	cfg := inspections.InspectionClientCfg{
		SkipIDs:       s.cfg.Export.Inspection.SkipIds,
		ModifiedAfter: s.cfg.Export.ModifiedAfter.Time,
		TemplateIDs:   s.cfg.Export.TemplateIds,
		Archived:      s.cfg.Export.Inspection.Archived,
		Completed:     s.cfg.Export.Inspection.Completed,
		Incremental:   s.cfg.Export.Incremental,
	}
	inspectionsClient := inspections.NewInspectionClient(&cfg, s.apiClient, e)

	err = inspectionsClient.Export(context.Background())
	if err != nil {
		return errors.Wrap(err, "error while exporting JSON")
	}
	return nil
}

func (s *SafetyCultureExporter) RunSQL() error {
	if s.cfg.Export.Media {
		err := os.MkdirAll(s.cfg.Export.MediaPath, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "Failed to create directory %s", s.cfg.Export.MediaPath)
		}
	}

	e, err := feed.NewSQLExporter(s.cfg.Db.Dialect, s.cfg.Db.ConnectionString, true, s.cfg.Export.MediaPath)
	if err != nil {
		return errors.Wrap(err, "create sql exporter")
	}

	exporterApp := feed.NewExporterApp(s.apiClient, s.sheqsyApiClient, s.cfg.ToExporterConfig(), s.exportStatus)
	if s.cfg.Export.SchemaOnly {
		return exporterApp.ExportSchemas(e)
	}

	if len(s.cfg.AccessToken) != 0 || len(s.cfg.SheqsyUsername) != 0 {
		err = exporterApp.ExportFeeds(e)
		if err != nil {
			return errors.Wrap(err, "exporting feeds")
		}
	}

	return nil
}

func (s *SafetyCultureExporter) RunCSV() error {
	exportPath := s.cfg.Export.Path

	err := os.MkdirAll(exportPath, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "Failed to create directory %s", exportPath)
	}

	if s.cfg.Export.Media {
		err := os.MkdirAll(s.cfg.Export.MediaPath, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "Failed to create directory %s", s.cfg.Export.MediaPath)
		}
	}

	e, err := feed.NewCSVExporter(exportPath, s.cfg.Export.MediaPath, s.cfg.Csv.MaxRowsPerFile)
	if err != nil {
		return errors.Wrap(err, "unable to create csv exporter")
	}

	exporterApp := feed.NewExporterApp(s.apiClient, s.sheqsyApiClient, s.cfg.ToExporterConfig(), s.exportStatus)
	if s.cfg.Export.SchemaOnly {
		return exporterApp.ExportSchemas(e)
	}

	if len(s.cfg.AccessToken) != 0 || len(s.cfg.SheqsyUsername) != 0 {
		err = exporterApp.ExportFeeds(e)
		if err != nil {
			return errors.Wrap(err, "exporting feeds")
		}
	}

	return nil
}

func (s *SafetyCultureExporter) RunInspectionReports() error {
	err := os.MkdirAll(s.cfg.Export.Path, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "Failed to create directory %s", s.cfg.Export.Path)
	}

	e, err := NewReportExporter(s.cfg.Export.Path, s.cfg.ToReporterConfig())
	if err != nil {
		return errors.Wrap(err, "unable to create report exporter")
	}

	exporterApp := feed.NewExporterApp(s.apiClient, s.sheqsyApiClient, s.cfg.ToExporterConfig(), s.exportStatus)
	err = exporterApp.ExportInspectionReports(e)
	if err != nil {
		return errors.Wrap(err, "generate reports")
	}

	return nil
}

func (s *SafetyCultureExporter) RunPrintSchema() error {
	e, err := feed.NewSchemaExporter(os.Stdout)
	if err != nil {
		return errors.Wrap(err, "unable to create exporter")
	}

	exporterApp := feed.NewExporterApp(s.apiClient, s.sheqsyApiClient, s.cfg.ToExporterConfig(), s.exportStatus)
	err = exporterApp.PrintSchemas(e)
	if err != nil {
		return errors.Wrap(err, "error while printing schema")
	}

	return nil
}

func (s *SafetyCultureExporter) GetTemplateList() ([]TemplateResponseItem, error) {
	client := templates.NewTemplatesClient(s.apiClient)
	res, err := client.GetTemplateList(context.Background(), 1000)
	if err != nil {
		return nil, err
	}

	transformer := func(data templates.TemplateResponseItem) TemplateResponseItem {
		return TemplateResponseItem{
			ID:         data.ID,
			Name:       data.Name,
			ModifiedAt: data.ModifiedAt,
		}
	}

	return util.GenericCollectionMapper(res, transformer), nil
}

func (s *SafetyCultureExporter) GetExportStatus() *ExportStatusResponse {
	data := s.exportStatus.ReadStatus()
	var res []ExportStatusResponseItem

	for _, v := range data {
		res = append(res, ExportStatusResponseItem{
			FeedName:    v.Name,
			Started:     v.Started,
			DebugString: fmt.Sprintf("remaining %d", v.EstRemaining),
		})
	}

	return &ExportStatusResponse{
		ExportStarted:   s.exportStatus.GetExportStarted(),
		ExportCompleted: s.exportStatus.GetExportCompleted(),
		Feeds:           res,
	}
}
