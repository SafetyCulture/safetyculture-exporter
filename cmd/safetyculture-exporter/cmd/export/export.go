package export

import (
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"

	util "github.com/SafetyCulture/safetyculture-exporter/cmd/safetyculture-exporter/cmd/utils"
	exporterAPI "github.com/SafetyCulture/safetyculture-exporter/pkg/api"
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

// NewSafetyCultureExporter create a new SafetyCultureExporter with configuration from Viper
func NewSafetyCultureExporter(v *viper.Viper) *exporterAPI.SafetyCultureExporter {
	cm, err := exporterAPI.NewConfigurationManagerFromFile("", v.ConfigFileUsed())
	MapViperConfigToExporterConfiguration(v, cm.Configuration)
	cm.ApplySafetyGuards()
	util.Check(err, "while loading config file")

	return exporterAPI.NewSafetyCultureExporter(cm.Configuration, getAPIClient(), getSheqsyAPIClient())
}

// MapViperConfigToExporterConfiguration maps Viper config to ExporterConfiguration structure
func MapViperConfigToExporterConfiguration(v *viper.Viper, cfg *exporterAPI.ExporterConfiguration) {
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
}

func getAPIClient() *httpapi.Client {
	apiOpts := []httpapi.Opt{}
	if viper.GetBool("api.tls_skip_verify") {
		apiOpts = append(apiOpts, httpapi.OptSetInsecureTLS(true))
	}
	if viper.GetString("api.tls_cert") != "" {
		apiOpts = append(apiOpts, httpapi.OptAddTLSCert(viper.GetString("api.tls_cert")))
	}
	if viper.GetString("api.proxy_url") != "" {
		proxyURL, err := url.Parse(viper.GetString("api.proxy_url"))
		util.Check(err, "Unable to parse proxy URL")
		apiOpts = append(apiOpts, httpapi.OptSetProxy(proxyURL))
	}

	return httpapi.NewClient(
		viper.GetString("api.url"),
		fmt.Sprintf("Bearer %s", viper.GetString("access_token")),
		apiOpts...,
	)
}

func getSheqsyAPIClient() *httpapi.Client {
	var apiOpts []httpapi.Opt
	if viper.GetBool("api.tls_skip_verify") {
		apiOpts = append(apiOpts, httpapi.OptSetInsecureTLS(true))
	}
	if viper.GetString("api.tls_cert") != "" {
		apiOpts = append(apiOpts, httpapi.OptAddTLSCert(viper.GetString("api.tls_cert")))
	}
	if viper.GetString("api.proxy_url") != "" {
		proxyURL, err := url.Parse(viper.GetString("api.proxy_url"))
		util.Check(err, "Unable to parse proxy URL")
		apiOpts = append(apiOpts, httpapi.OptSetProxy(proxyURL))
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

	return httpapi.NewClient(
		viper.GetString("api.sheqsy_url"),
		fmt.Sprintf("Basic %s", token),
		apiOpts...,
	)
}
