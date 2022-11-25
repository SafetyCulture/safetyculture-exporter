package export

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/internal/app/api"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/config"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/exporter"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/feed"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/inspections"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/util"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
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

// HostedDBCmd is used to export data in csv format
func HostedDBCmd() *cobra.Command {
	return &cobra.Command{
		Hidden:  true,
		Use:     "hosted-db",
		Short:   "Export SafetyCulture data to hosted database",
		Example: `DO NOT USE, THIS IS AN ALPHA FEATURE AND MY RESULT IN A CORRUPTED DATA EXPORT`,
		RunE:    runHostedDB,
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

// MapViperConfigToConfigurationOptions maps Viper config to ConfigurationOptions structure
// doesn't map 100% every field. Just the ones I found needed.
func MapViperConfigToConfigurationOptions(v *viper.Viper) *config.ConfigurationOptions {

	// caps action batch limit to 100
	actionLimit := v.GetInt("export.action.limit")
	if actionLimit > 100 {
		actionLimit = 100
	}

	// caps issue batch limit to 100
	issueLimit := v.GetInt("export.issue.limit")
	if issueLimit > 100 {
		issueLimit = 100
	}

	return &config.ConfigurationOptions{
		ApiConfig: &config.ApiConfig{
			AccessToken: v.GetString("access_token"),
		},
		SheqsyApiConfig: &config.SheqsyApiConfig{
			UserName:  v.GetString("sheqsy_username"),
			CompanyID: v.GetString("sheqsy_company_id"),
		},
		DBConfig: &config.DBConfig{
			Dialect:          v.GetString("db.dialect"),
			ConnectionString: v.GetString("db.connection_string"),
		},
		CSVConfig: &config.CSVConfig{
			MaxRowsPerFile: v.GetInt("csv.max_rows_per_file"),
		},
		ExportConfig: &config.ExportConfig{
			Path:               v.GetString("export.path"),
			Incremental:        v.GetBool("export.incremental"),
			ModifiedAfter:      v.GetTime("export.modified_after"),
			FilterByTemplateID: v.GetStringSlice("export.template_ids"),
			FilterByTableName:  v.GetStringSlice("export.tables"),
			SchemaOnly:         v.GetBool("export.schema_only"),
			InspectionConfig: &config.ExportInspectionConfig{
				IncludeInactiveItems: v.GetBool("export.inspection.included_inactive_items"),
				Archived:             v.GetString("export.inspection.archived"),
				Completed:            v.GetString("export.inspection.completed"),
				SkipIDs:              v.GetStringSlice("export.inspection.skip_ids"),
				BatchLimit:           v.GetInt("export.inspection.limit"),
				WebReportLink:        v.GetString("export.inspection.web_report_link"),
			},
			SiteConfig: &config.ExportSiteConfig{
				IncludeDeleted:       v.GetBool("export.site.include_deleted"),
				IncludeFullHierarchy: v.GetBool("export.site.include_full_hierarchy"),
			},
			MediaConfig: &config.ExportMediaConfig{
				Export: v.GetBool("export.media"),
				Path:   v.GetString("export.media_path"),
			},
			ActionConfig: &config.ExportActionConfig{
				BatchLimit: actionLimit,
			},
			IssueConfig: &config.ExportIssueConfig{
				BatchLimit: issueLimit,
			},
		},
		ReportConfig: &config.ReportConfig{
			Format:             v.GetStringSlice("report.format"),
			PreferenceID:       v.GetString("report.preference_id"),
			FileNameConvention: v.GetString("report.filename_convention"),
			RetryTimeout:       v.GetInt("report.retry_timeout"),
		},
	}
}

func runSQL(cmd *cobra.Command, args []string) error {
	e := NewSafetyCultureExporter(viper.GetViper())
	err := e.RunSQL()
	util.Check(err, "error while exporting SQL")
	return nil
}

func runInspectionJSON(cmd *cobra.Command, args []string) error {
	e := NewSafetyCultureExporter(viper.GetViper())
	err := e.RunInspectionJSON()
	util.Check(err, "error while exporting JSON")
	return nil
}

func runCSV(cmd *cobra.Command, args []string) error {
	e := NewSafetyCultureExporter(viper.GetViper())
	err := e.RunCSV()
	util.Check(err, "error while exporting CSV")
	return nil
}

func runHostedDB(cmd *cobra.Command, args []string) error {
	if os.Getenv("SC_EXPORTER_ALPHA_FEATURES") != "true" {
		return errors.New("invalid command hosted-db")
	}

	logger := util.GetLogger()

	exportPath := viper.GetString("export.path")

	postgres := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Version(embeddedpostgres.V14).
		RuntimePath(path.Join(exportPath, "pg/runtime")).
		DataPath(path.Join(exportPath, "pg/data")).
		BinariesPath(path.Join(exportPath, "pg/bin")).
		Username("safetyculture_exporter").
		Password("safetyculture_exporter").
		Database("safetyculture_exporter").
		Port(35432).
		StartTimeout(45 * time.Second),
	)

	defer func() {
		logger.Info("stopping postgres DB")
		if err := postgres.Stop(); err != nil {
			logger.With("error", err).Error("failed to stop postgres DB")
		}
	}()

	err := postgres.Start()
	if err != nil && !strings.Contains(err.Error(), "server is already started") {
		util.Check(err, fmt.Sprintf("Failed to started DB %s", exportPath))
	}

	err = os.MkdirAll(exportPath, os.ModePerm)
	util.Check(err, fmt.Sprintf("Failed to create directory %s", exportPath))

	var exportMediaPath string
	exportMedia := viper.GetBool("export.media")
	if exportMedia {
		exportMediaPath = viper.GetString("export.media_path")
		err := os.MkdirAll(exportMediaPath, os.ModePerm)
		util.Check(err, fmt.Sprintf("Failed to create directory %s", exportMediaPath))
	}

	exporter, err := feed.NewSQLExporter("postgres", "postgres://safetyculture_exporter:safetyculture_exporter@localhost:35432/safetyculture_exporter", true, exportMediaPath)
	util.Check(err, "unable to create exporter")

	if viper.GetBool("export.schema_only") {
		return feed.CreateSchemas(viper.GetViper(), exporter)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		s := <-sigc
		logger.Infof("received %s, shutting down", s.String())
		postgres.Stop()

		os.Exit(0)
	}()

	for {
		err := feed.ExportFeeds(
			viper.GetViper(),
			getAPIClient(),
			getSheqsyAPIClient(),
			exporter,
		)
		if err != nil {
			postgres.Stop()
			logger.Fatal("error exporting data", err)
		}

		delay := time.Hour

		logger.Infof("export finished, checking for updates in %s (%s)", delay.String(), time.Now().Add(delay).Local().Format(time.RFC822))

		time.Sleep(delay)
	}
}

func printSchema(cmd *cobra.Command, args []string) error {
	e := NewSafetyCultureExporter(viper.GetViper())
	err := e.RunPrintSchema()
	util.Check(err, "error while printing schema")
	return nil
}

func runInspectionReports(cmd *cobra.Command, args []string) error {
	e := NewSafetyCultureExporter(viper.GetViper())
	err := e.RunInspectionReports()
	util.Check(err, "failed to generate reports")
	return nil
}

type SafetyCultureExporter struct {
	cfg *config.ConfigurationOptions
}

func (s *SafetyCultureExporter) RunInspectionJSON() error {
	exportPath := fmt.Sprintf("%s/json/", s.cfg.ExportConfig.Path)
	err := os.MkdirAll(exportPath, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "Failed to create directory %s", exportPath)
	}

	exporter := exporter.NewJSONExporter(exportPath)
	inspectionsClient := inspections.NewInspectionClient(s.cfg, getAPIClient(), exporter)

	err = inspectionsClient.Export(context.Background())
	if err != nil {
		return errors.Wrap(err, "error while exporting JSON")
	}
	return nil
}

func (s *SafetyCultureExporter) RunSQL() error {
	if s.cfg.ExportConfig.MediaConfig.Export {
		err := os.MkdirAll(s.cfg.ExportConfig.MediaConfig.Path, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "Failed to create directory %s", s.cfg.ExportConfig.MediaConfig.Path)
		}
	}

	exporter, err := feed.NewSQLExporter(s.cfg.DBConfig.Dialect, s.cfg.DBConfig.ConnectionString, true, s.cfg.ExportConfig.MediaConfig.Path)
	if err != nil {
		return errors.Wrap(err, "unable to create sql exporter")
	}

	exporterApp := feed.NewExporterApp(getAPIClient(), getSheqsyAPIClient(), s.cfg)
	if s.cfg.ExportConfig.SchemaOnly {
		return exporterApp.ExportSchemas(exporter)
	}

	if len(s.cfg.ApiConfig.AccessToken) != 0 {
		err = exporterApp.ExportFeeds(exporter)
		if err != nil {
			return errors.Wrap(err, "error while exporting feeds")
		}
	}

	return nil
}

func (s *SafetyCultureExporter) RunCSV() error {
	exportPath := s.cfg.ExportConfig.Path

	err := os.MkdirAll(exportPath, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "Failed to create directory %s", exportPath)
	}

	if s.cfg.ExportConfig.MediaConfig.Export {
		err := os.MkdirAll(s.cfg.ExportConfig.MediaConfig.Path, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "Failed to create directory %s", s.cfg.ExportConfig.MediaConfig.Path)
		}
	}

	exporter, err := feed.NewCSVExporter(exportPath, s.cfg.ExportConfig.MediaConfig.Path, s.cfg.CSVConfig.MaxRowsPerFile)
	if err != nil {
		return errors.Wrap(err, "unable to create csv exporter")
	}

	exporterApp := feed.NewExporterApp(getAPIClient(), getSheqsyAPIClient(), s.cfg)
	if s.cfg.ExportConfig.SchemaOnly {
		return exporterApp.ExportSchemas(exporter)
	}

	if len(s.cfg.ApiConfig.AccessToken) != 0 {
		err = exporterApp.ExportFeeds(exporter)
		if err != nil {
			return errors.Wrap(err, "error while exporting feeds")
		}
	}

	return nil
}

func (s *SafetyCultureExporter) RunInspectionReports() error {
	err := os.MkdirAll(s.cfg.ExportConfig.Path, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "Failed to create directory %s", s.cfg.ExportConfig.Path)
	}

	exporter, err := feed.NewReportExporter(s.cfg.ExportConfig.Path, s.cfg.ReportConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create report exporter")
	}

	exporterApp := feed.NewExporterApp(getAPIClient(), getSheqsyAPIClient(), s.cfg)
	err = exporterApp.ExportInspectionReports(exporter)
	if err != nil {
		return errors.Wrap(err, "failed to generate reports")
	}

	return nil
}

func (s *SafetyCultureExporter) RunPrintSchema() error {
	exporter, err := feed.NewSchemaExporter(os.Stdout)
	if err != nil {
		return errors.Wrap(err, "unable to create exporter")
	}

	exporterApp := feed.NewExporterApp(getAPIClient(), getSheqsyAPIClient(), s.cfg)
	err = exporterApp.PrintSchemas(exporter)
	if err != nil {
		return errors.Wrap(err, "error while printing schema")
	}

	return nil
}

func NewSafetyCultureExporter(v *viper.Viper) *SafetyCultureExporter {
	return &SafetyCultureExporter{
		cfg: MapViperConfigToConfigurationOptions(v),
	}
}
