package export

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"

	"github.com/SafetyCulture/safetyculture-exporter/internal/app/api"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/config"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/exporter"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/feed"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/inspections"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SQLCmd is used to export data in sql format
func SQLCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sql",
		Short: "Export SafetyCulture data to SQL database",
		Example: `// Limit inspections and schedules to these templates
safetyculture-exporter sql --template-ids template_F492E54D87F2419E9398F7BDCA0FA5D9,template_d54e06808d2f11e2893e83a731dba0ca

// Customise export location
safetyculture-exporter sql --export-path /path/to/export/to`,
		RunE: runSQL,
	}
}

// CSVCmd is used to export data in csv format
func CSVCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "csv",
		Short: "Export SafetyCulture data to CSV files",
		Example: `// Limit inspections and schedules to these templates
safetyculture-exporter csv --template-ids template_F492E54D87F2419E9398F7BDCA0FA5D9,template_d54e06808d2f11e2893e83a731dba0ca

// Customise export location
safetyculture-exporter csv --export-path /path/to/export/to`,
		RunE: runCSV,
	}
}

// InspectionJSONCmd is used to export inspections to json files
func InspectionJSONCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inspection-json",
		Short: "Export SafetyCulture inspections to json files",
		Example: `// Limit inspections to these templates
safetyculture-exporter inspection-json --template-ids template_F492E54D87F2419E9398F7BDCA0FA5D9,template_d54e06808d2f11e2893e83a731dba0ca

// Customise export location
safetyculture-exporter inspection-json --export-path /path/to/export/to`,
		RunE: runInspectionJSON,
	}
}

// PrintSchemaCmd is used to print table schemas
func PrintSchemaCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "schema",
		Short:   "Print SafetyCulture table schemas",
		Example: `safetyculture-exporter schema`,
		RunE:    printSchema,
	}
}

// ReportCmd is used to download inspection reports
func ReportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "Export inspection report",
		Example: `// Export PDF and Word inspection reports
		safetyculture-exporter report --export-path /path/to/export/to --format PDF,WORD
		// Export PDF inspection reports with a custom report layout
		safetyculture-exporter report --export-path /path/to/export/to --format PDF --preference-id abc`,
		RunE: runInspectionReports,
	}
}

func getAPIClient() *api.Client {
	apiOpts := []api.Opt{}
	if viper.GetBool("api.tls_skip_verify") {
		apiOpts = append(apiOpts, api.OptSetInsecureTLS(true))
	}
	if viper.GetString("api.tls_cert") != "" {
		apiOpts = append(apiOpts, api.OptAddTLSCert(viper.GetString("api.tls_cert")))
	}
	if viper.GetString("api.proxy_url") != "" {
		proxyURL, err := url.Parse(viper.GetString("api.proxy_url"))
		util.Check(err, "Unable to parse proxy URL")
		apiOpts = append(apiOpts, api.OptSetProxy(proxyURL))
	}

	return api.NewClient(
		viper.GetString("api.url"),
		fmt.Sprintf("Bearer %s", viper.GetString("access_token")),
		apiOpts...,
	)
}

func getSheqsyAPIClient() *api.Client {
	apiOpts := []api.Opt{}
	if viper.GetBool("api.tls_skip_verify") {
		apiOpts = append(apiOpts, api.OptSetInsecureTLS(true))
	}
	if viper.GetString("api.tls_cert") != "" {
		apiOpts = append(apiOpts, api.OptAddTLSCert(viper.GetString("api.tls_cert")))
	}
	if viper.GetString("api.proxy_url") != "" {
		proxyURL, err := url.Parse(viper.GetString("api.proxy_url"))
		util.Check(err, "Unable to parse proxy URL")
		apiOpts = append(apiOpts, api.OptSetProxy(proxyURL))
	}

	token := base64.StdEncoding.EncodeToString(
		[]byte(
			fmt.Sprintf(
				"%s:%s",
				viper.GetString("sheqsy_username"),
				viper.GetString("sheqsy_password"),
			),
		),
	)

	return api.NewClient(
		viper.GetString("api.sheqsy_url"),
		fmt.Sprintf("Basic %s", token),
		apiOpts...,
	)
}

func runSQL(cmd *cobra.Command, args []string) error {
	exp := NewSafetyCultureExporter(viper.GetViper())
	err := exp.RunSQL()
	util.Check(err, "error while exporting SQL")
	return nil
}

func runInspectionJSON(cmd *cobra.Command, args []string) error {
	exp := NewSafetyCultureExporter(viper.GetViper())
	err := exp.RunInspectionJSON()
	util.Check(err, "error while exporting JSON")
	return nil
}

func runCSV(cmd *cobra.Command, args []string) error {
	exp := NewSafetyCultureExporter(viper.GetViper())
	err := exp.RunCSV()
	util.Check(err, "error while exporting CSV")
	return nil
}

func printSchema(cmd *cobra.Command, args []string) error {
	exp := NewSafetyCultureExporter(viper.GetViper())
	err := exp.RunPrintSchema()
	util.Check(err, "error while printing schema")
	return nil
}

func runInspectionReports(cmd *cobra.Command, args []string) error {
	exp := NewSafetyCultureExporter(viper.GetViper())
	err := exp.RunInspectionReports()
	util.Check(err, "failed to generate reports")
	return nil
}

type SafetyCultureExporter struct {
	cfg *config.ExporterConfiguration
}

func (s *SafetyCultureExporter) RunInspectionJSON() error {
	exportPath := fmt.Sprintf("%s/json/", s.cfg.Export.Path)
	err := os.MkdirAll(exportPath, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "Failed to create directory %s", exportPath)
	}

	e := exporter.NewJSONExporter(exportPath)
	inspectionsClient := inspections.NewInspectionClient(s.cfg, getAPIClient(), e)

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

	exporterApp := feed.NewExporterApp(getAPIClient(), getSheqsyAPIClient(), s.cfg)
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

	exporterApp := feed.NewExporterApp(getAPIClient(), getSheqsyAPIClient(), s.cfg)
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

	e, err := feed.NewReportExporter(s.cfg.Export.Path, s.cfg)
	if err != nil {
		return errors.Wrap(err, "unable to create report exporter")
	}

	exporterApp := feed.NewExporterApp(getAPIClient(), getSheqsyAPIClient(), s.cfg)
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

	exporterApp := feed.NewExporterApp(getAPIClient(), getSheqsyAPIClient(), s.cfg)
	err = exporterApp.PrintSchemas(e)
	if err != nil {
		return errors.Wrap(err, "error while printing schema")
	}

	return nil
}

// NewSafetyCultureExporter create a new SafetyCultureExporter with configuration from Viper
func NewSafetyCultureExporter(v *viper.Viper) *SafetyCultureExporter {
	cfg := MapViperConfigToExporterConfiguration(v)
	cm, err := config.NewConfigurationManager(v.ConfigFileUsed(), true, false, cfg)
	util.Check(err, "while loading config file")

	return &SafetyCultureExporter{
		cfg: cm.Configuration,
	}
}

// MapViperConfigToExporterConfiguration maps Viper config to ExporterConfiguration structure
func MapViperConfigToExporterConfiguration(v *viper.Viper) *config.ExporterConfiguration {
	cfg := config.BuildConfigurationWithDefaults()
	cfg.AccessToken = v.GetString("access_token")
	cfg.SheqsyUsername = v.GetString("sheqsy_username")
	cfg.SheqsyCompanyID = v.GetString("sheqsy_company_id")
	cfg.Db.Dialect = v.GetString("db.dialect")
	cfg.Db.ConnectionString = v.GetString("db.connection_string")
	cfg.Csv.MaxRowsPerFile = v.GetInt("csv.max_rows_per_file")
	cfg.Export.Path = v.GetString("export.path")
	cfg.Export.Incremental = v.GetBool("export.incremental")
	cfg.Export.ModifiedAfter.Time = v.GetTime("export.modified_after")
	cfg.Export.TemplateIds = v.GetStringSlice("export.template_ids")
	cfg.Export.Tables = v.GetStringSlice("export.tables")
	cfg.Export.SchemaOnly = v.GetBool("export.schema_only")
	cfg.Export.Inspection.IncludedInactiveItems = v.GetBool("export.inspection.included_inactive_items")
	cfg.Export.Inspection.Archived = v.GetString("export.inspection.archived")
	cfg.Export.Inspection.Completed = v.GetString("export.inspection.completed")
	cfg.Export.Inspection.SkipIds = v.GetStringSlice("export.inspection.skip_ids")
	cfg.Export.Inspection.Limit = v.GetInt("export.inspection.limit")
	cfg.Export.Inspection.WebReportLink = v.GetString("export.inspection.web_report_link")
	cfg.Export.Site.IncludeDeleted = v.GetBool("export.site.include_deleted")
	cfg.Export.Site.IncludeFullHierarchy = v.GetBool("export.site.include_full_hierarchy")
	cfg.Export.Media = v.GetBool("export.media")
	cfg.Export.MediaPath = v.GetString("export.media_path")
	cfg.Export.Action.Limit = v.GetInt("export.action.limit")
	cfg.Export.Issue.Limit = v.GetInt("export.issue.limit")
	cfg.Report.Format = v.GetStringSlice("report.format")
	cfg.Report.PreferenceID = v.GetString("report.preference_id")
	cfg.Report.FilenameConvention = v.GetString("report.filename_convention")
	cfg.Report.RetryTimeout = v.GetInt("report.retry_timeout")

	return cfg
}
