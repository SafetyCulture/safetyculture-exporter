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
		ExportConfig: &config.ExportConfig{
			Incremental:        v.GetBool("export.incremental"),
			ModifiedAfter:      v.GetTime("export.modified_after"),
			FilterByTemplateID: v.GetStringSlice("export.template_ids"),
			FilterByTableName:  v.GetStringSlice("export.tables"),
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
			},
			ActionConfig: &config.ExportActionConfig{
				BatchLimit: actionLimit,
			},
		},
	}
}

func runSQL(cmd *cobra.Command, args []string) error {

	var exportMediaPath string
	exportMedia := viper.GetBool("export.media")
	if exportMedia {
		exportMediaPath = viper.GetString("export.media_path")

		err := os.MkdirAll(exportMediaPath, os.ModePerm)
		util.Check(err, fmt.Sprintf("Failed to create directory %s", exportMediaPath))
	}

	exporter, err := feed.NewSQLExporter(viper.GetString("db.dialect"), viper.GetString("db.connection_string"), true, exportMediaPath)
	util.Check(err, "unable to create exporter")

	exporterAppCfg := MapViperConfigToConfigurationOptions(viper.GetViper())
	exporterApp := feed.NewExporterApp(getAPIClient(), getSheqsyAPIClient(), exporterAppCfg)
	if viper.GetBool("export.schema_only") {
		return exporterApp.CreateSchemas(exporter)
	}

	return exporterApp.ExportFeeds(exporter)
}

func runInspectionJSON(cmd *cobra.Command, args []string) error {

	exportPath := fmt.Sprintf("%s/json/", viper.GetString("export.path"))
	err := os.MkdirAll(exportPath, os.ModePerm)
	util.Check(err, fmt.Sprintf("Failed to create directory %s", exportPath))

	inspectionsClient := inspections.NewInspectionClient(
		viper.GetViper(),
		getAPIClient(),
		exporter.NewJSONExporter(exportPath),
	)

	return inspectionsClient.Export(context.Background())
}

func runCSV(cmd *cobra.Command, args []string) error {

	exportPath := viper.GetString("export.path")

	err := os.MkdirAll(exportPath, os.ModePerm)
	util.Check(err, fmt.Sprintf("Failed to create directory %s", exportPath))

	var exportMediaPath string
	exportMedia := viper.GetBool("export.media")
	if exportMedia {
		exportMediaPath = viper.GetString("export.media_path")
		err := os.MkdirAll(exportMediaPath, os.ModePerm)
		util.Check(err, fmt.Sprintf("Failed to create directory %s", exportMediaPath))
	}

	maxRowsPerFile := viper.GetInt("csv.max_rows_per_file")

	exporter, err := feed.NewCSVExporter(exportPath, exportMediaPath, maxRowsPerFile)
	util.Check(err, "unable to create exporter")

	exporterAppCfg := MapViperConfigToConfigurationOptions(viper.GetViper())
	exporterApp := feed.NewExporterApp(getAPIClient(), getSheqsyAPIClient(), exporterAppCfg)
	if viper.GetBool("export.schema_only") {
		return exporterApp.CreateSchemas(exporter)
	}

	if len(viper.GetViper().GetString("access_token")) != 0 {
		return exporterApp.ExportFeeds(exporter)
	}

	return nil
}

func printSchema(cmd *cobra.Command, args []string) error {
	exporter, err := feed.NewSchemaExporter(os.Stdout)
	util.Check(err, "unable to create exporter")

	return feed.WriteSchemas(viper.GetViper(), exporter)
}

func runInspectionReports(cmd *cobra.Command, args []string) error {

	exportPath := viper.GetString("export.path")
	err := os.MkdirAll(exportPath, os.ModePerm)
	util.Check(err, "unable to create export directory")

	format := viper.GetStringSlice("report.format")
	preferenceID := viper.GetString("report.preference_id")
	filenameConvention := viper.GetString("report.filename_convention")

	exporter, err := feed.NewReportExporter(exportPath, format, preferenceID, filenameConvention)
	util.Check(err, "unable to create exporter")

	apiClient := getAPIClient()

	err = feed.ExportInspectionReports(viper.GetViper(), apiClient, exporter)
	util.Check(err, "failed to generate reports")

	return nil
}
