package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/exporter"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/feed"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/inspections"
	"github.com/pkg/errors"
)

func NewSafetyCultureExporter(cfg *ExporterConfiguration, apiClient *httpapi.Client, sheqsyApiClient *httpapi.Client) *SafetyCultureExporter {
	return &SafetyCultureExporter{
		apiClient:       apiClient,
		sheqsyApiClient: sheqsyApiClient,
		cfg:             cfg,
	}
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

type SafetyCultureExporter struct {
	apiClient       *httpapi.Client
	sheqsyApiClient *httpapi.Client
	cfg             *ExporterConfiguration
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

	exporterApp := feed.NewExporterApp(s.apiClient, s.sheqsyApiClient, s.cfg.ToExporterConfig())
	if s.cfg.Export.SchemaOnly {
		return exporterApp.ExportSchemas(e)
	}

	if len(s.cfg.AccessToken) != 0 {
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

	exporterApp := feed.NewExporterApp(s.apiClient, s.sheqsyApiClient, s.cfg.ToExporterConfig())
	if s.cfg.Export.SchemaOnly {
		return exporterApp.ExportSchemas(e)
	}

	if len(s.cfg.AccessToken) != 0 {
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

	exporterApp := feed.NewExporterApp(s.apiClient, s.sheqsyApiClient, s.cfg.ToExporterConfig())
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

	exporterApp := feed.NewExporterApp(s.apiClient, s.sheqsyApiClient, s.cfg.ToExporterConfig())
	err = exporterApp.PrintSchemas(e)
	if err != nil {
		return errors.Wrap(err, "error while printing schema")
	}

	return nil
}
