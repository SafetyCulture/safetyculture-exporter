package export

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/exporter"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/feed"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/inspections"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SQLCmd is used to export data in sql format
func SQLCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sql",
		Short: "Export iAuditor data to SQL database",
		Example: `// Limit inspections and schedules to these templates
iauditor-exporter sql --template-ids template_F492E54D87F2419E9398F7BDCA0FA5D9,template_d54e06808d2f11e2893e83a731dba0ca

// Customise export location
iauditor-exporter sql --export-path /path/to/export/to`,
		RunE: runSQL,
	}
}

// CSVCmd is used to export data in csv format
func CSVCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "csv",
		Short: "Export iAuditor data to CSV files",
		Example: `// Limit inspections and schedules to these templates
iauditor-exporter csv --template-ids template_F492E54D87F2419E9398F7BDCA0FA5D9,template_d54e06808d2f11e2893e83a731dba0ca

// Customise export location
iauditor-exporter csv --export-path /path/to/export/to`,
		RunE: runCSV,
	}
}

// InspectionJSONCmd is used to export inspections to json files
func InspectionJSONCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inspection-json",
		Short: "Export iAuditor inspections to json files",
		Example: `// Limit inspections to these templates
iauditor-exporter inspection-json --template-ids template_F492E54D87F2419E9398F7BDCA0FA5D9,template_d54e06808d2f11e2893e83a731dba0ca

// Customise export location
iauditor-exporter inspection-json --export-path /path/to/export/to`,
		RunE: runInspectionJSON,
	}
}

// PrintSchemaCmd is used to print table schemas
func PrintSchemaCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "schema",
		Short:   "Print iAuditor table schemas",
		Example: `iauditor-exporter schema`,
		RunE:    printSchema,
	}
}

func getAPIClient() api.Client {
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
