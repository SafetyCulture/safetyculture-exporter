package export

import (
	"net/url"
	"os"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Cmd implements the sts commands.
func Cmds() []*cobra.Command {
	return []*cobra.Command{
		&cobra.Command{
			Use:   "csv",
			Short: "Export iAuditor data to CSV files",
			Example: `// Limit inspections and schedules to these templates
iauditor-exporter csv --template-ids template_F492E54D87F2419E9398F7BDCA0FA5D9,template_d54e06808d2f11e2893e83a731dba0ca

// Customise export location
iauditor-exporter csv --export-path /path/to/export/to`,
			RunE: runCSV,
		},
		&cobra.Command{
			Use:   "sql",
			Short: "Export iAuditor data to SQL database",
			Example: `// Limit inspections and schedules to these templates
iauditor-exporter sql --template-ids template_F492E54D87F2419E9398F7BDCA0FA5D9,template_d54e06808d2f11e2893e83a731dba0ca

// Customise export location
iauditor-exporter sql --export-path /path/to/export/to`,
			RunE: runSQL,
		},
		&cobra.Command{
			Use:     "schema",
			Short:   "Print iAuditor table schemas",
			Example: `iauditor-exporter schema`,
			RunE:    printSchema,
		},
	}
}

func getAPIClient() api.APIClient {
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

	return api.NewAPIClient(
		viper.GetString("api.url"),
		viper.GetString("access_token"),
		apiOpts...,
	)
}

func runSQL(cmd *cobra.Command, args []string) error {

	exporter, err := feed.NewSQLExporter(viper.GetString("db.dialect"), viper.GetString("db.connection_string"), true)
	util.Check(err, "unable to create exporter")

	if viper.GetBool("export.schema_only") {
		return feed.CreateSchemas(viper.GetViper(), exporter)
	}

	apiClient := getAPIClient()

	return feed.ExportFeeds(viper.GetViper(), apiClient, exporter)
}

func runCSV(cmd *cobra.Command, args []string) error {

	exportPath := viper.GetString("export.path")
	os.MkdirAll(exportPath, os.ModePerm)

	exporter, err := feed.NewCSVExporter(exportPath)
	util.Check(err, "unable to create exporter")

	if viper.GetBool("export.schema_only") {
		return feed.CreateSchemas(viper.GetViper(), exporter)
	}

	apiClient := getAPIClient()

	return feed.ExportFeeds(viper.GetViper(), apiClient, exporter)
}

func printSchema(cmd *cobra.Command, args []string) error {
	exporter, err := feed.NewSchemaExporter(os.Stdout)
	util.Check(err, "unable to create exporter")

	return feed.WriteSchemas(viper.GetViper(), exporter)
}

func ExportReports(cmd *cobra.Command, args []string) error {

	exportPath := viper.GetString("export.path")
	os.MkdirAll(exportPath, os.ModePerm)

	exporter, err := feed.NewReportExporter(exportPath)
	util.Check(err, "unable to create exporter")

	apiClient := getAPIClient()

	err = feed.ExportInspectionReports(viper.GetViper(), apiClient, exporter)
	util.Check(err, "failed to generate reports")

	return nil
}
