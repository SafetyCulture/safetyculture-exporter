package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/SafetyCulture/iauditor-exporter/cmd/iauditor-exporter/cmd/configure"
	"github.com/SafetyCulture/iauditor-exporter/cmd/iauditor-exporter/cmd/export"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/version"
	"github.com/SafetyCulture/iauditor-exporter/internal/update"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"

	flag "github.com/spf13/pflag"
)

var cfgFile string
var connectionFlags, dbFlags, exportFlags, mediaFlags, inspectionFlags, templatesFlag, tablesFlag, schemasFlag, reportFlags *flag.FlagSet

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Version: version.GetVersion(),
	Use:     "iauditor-exporter",
	Short:   "A CLI tool for extracting your iAuditor data",
	Long:    "A CLI tool for extracting your iAuditor data",
}

var disclaimer = `THIS SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.`

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	fmt.Printf("\033[1;33m%s\033[0m\n", disclaimer)
	updateMsgChan := make(chan *update.ReleaseInfo)

	go func() {
		res := update.Check(version.GetVersion())
		updateMsgChan <- res
	}()

	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	if newRelease := <-updateMsgChan; newRelease != nil {
		yellow := color.FgYellow.Render
		cyan := color.FgCyan.Render
		fmt.Fprintf(os.Stderr, "\n\n%s %s → %s\n%s\n\n",
			yellow("A new version of iauditor-exporter is available"),
			cyan(version.GetVersion()),
			cyan(newRelease.Version),
			yellow(newRelease.ChangelogURL),
		)
	}
}

func writeDocs(cmd *cobra.Command, args []string) error {
	return doc.GenMarkdownTree(RootCmd, "docs/")
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config-path", "./iauditor-exporter.yaml", "config file")

	configFlags()
	bindFlags()

	// Add sub-commands
	addCmd(export.SQLCmd(), connectionFlags, exportFlags, dbFlags, inspectionFlags, templatesFlag, tablesFlag, schemasFlag, mediaFlags)
	addCmd(export.CSVCmd(), connectionFlags, exportFlags, inspectionFlags, templatesFlag, tablesFlag, schemasFlag, mediaFlags)
	addCmd(export.InspectionJSONCmd(), exportFlags, connectionFlags, inspectionFlags, templatesFlag)
	addCmd(export.ReportCmd(), connectionFlags, exportFlags, inspectionFlags, templatesFlag, reportFlags)
	addCmd(export.PrintSchemaCmd())
	addCmd(configure.Cmd(), connectionFlags, dbFlags, exportFlags, inspectionFlags, templatesFlag, tablesFlag)
	RootCmd.AddCommand(&cobra.Command{
		Hidden: true,
		Use:    "docs",
		RunE:   writeDocs,
	})

	initConfig()
}

func configFlags() {
	connectionFlags = flag.NewFlagSet("connection", flag.ContinueOnError)
	// TODO - Can we validate these tokens and throw error if they are wrong?
	connectionFlags.StringP("access-token", "t", "", "API Access Token")
	connectionFlags.String("api-url", "https://api.safetyculture.io", "API URL")
	connectionFlags.Bool("tls-skip-verify", false, "Skip verification of API TLS certificates")
	connectionFlags.String("tls-cert", "", "Custom root CA certificate to use when making API requests")
	connectionFlags.String("proxy-url", "", "Proxy URL for making API requests through")

	dbFlags = flag.NewFlagSet("db", flag.ContinueOnError)
	dbFlags.String("db-dialect", "mysql", "Database dialect. mysql, postgres and sqlserver are the only valid options.")
	dbFlags.String("db-connection-string", "", "Database connection string")

	exportFlags = flag.NewFlagSet("export", flag.ContinueOnError)
	exportFlags.String("export-path", "./export/", "File Export Path")
	exportFlags.Bool("incremental", true, "Update inspections, inspection_items and templates tables incrementally")
	exportFlags.String("modified-after", "", "Return inspections modified after this date (see readme for supported formats)")

	mediaFlags = flag.NewFlagSet("media", flag.ContinueOnError)
	mediaFlags.Bool("export-media", false, "Export media")
	mediaFlags.String("export-media-path", "./export/media/", "Media Export Path")

	inspectionFlags = flag.NewFlagSet("inspection", flag.ContinueOnError)
	inspectionFlags.StringSlice("inspection-skip-ids", []string{}, "Skip storing these inspection IDs")
	inspectionFlags.Bool("inspection-include-inactive-items", false, "Include inactive items in the inspection_items table (default false)")
	inspectionFlags.String("inspection-archived", "false", "Return archived inspections, false, true or both")
	inspectionFlags.String("inspection-completed", "both", "Return completed inspections, false, true or both")

	templatesFlag = flag.NewFlagSet("templates", flag.ContinueOnError)
	templatesFlag.StringSlice("template-ids", []string{}, "Template IDs to filter inspections and schedules by (default all)")

	tablesFlag = flag.NewFlagSet("tables", flag.ContinueOnError)
	tablesFlag.StringSlice("tables", []string{}, "Tables to export (default all)")

	schemasFlag = flag.NewFlagSet("schemas", flag.ContinueOnError)
	schemasFlag.Bool("create-schema-only", false, "Create schema only (default false)")

	reportFlags = flag.NewFlagSet("report", flag.ContinueOnError)
	reportFlags.StringSlice("format", []string{}, "Export format (PDF,WORD)")
	reportFlags.String("preference-id", "", "The report preference to apply to the document")
}

func bindFlags() {
	util.Check(viper.BindPFlag("access_token", connectionFlags.Lookup("access-token")), "while binding flag")

	util.Check(viper.BindPFlag("api.url", connectionFlags.Lookup("api-url")), "while binding flag")
	util.Check(viper.BindPFlag("api.tls_skip_verify", connectionFlags.Lookup("tls-skip-verify")), "while binding flag")
	util.Check(viper.BindPFlag("api.tls_cert", connectionFlags.Lookup("tls-cert")), "while binding flag")
	util.Check(viper.BindPFlag("api.proxy_url", connectionFlags.Lookup("proxy-url")), "while binding flag")

	util.Check(viper.BindPFlag("db.dialect", dbFlags.Lookup("db-dialect")), "while binding flag")
	util.Check(viper.BindPFlag("db.connection_string", dbFlags.Lookup("db-connection-string")), "while binding flag")

	util.Check(viper.BindPFlag("export.path", exportFlags.Lookup("export-path")), "while binding flag")
	util.Check(viper.BindPFlag("export.incremental", exportFlags.Lookup("incremental-update")), "while binding flag")
	util.Check(viper.BindPFlag("export.modified_after", exportFlags.Lookup("modified-after")), "while binding flag")

	util.Check(viper.BindPFlag("export.media", mediaFlags.Lookup("export-media")), "while binding flag")
	util.Check(viper.BindPFlag("export.media_path", mediaFlags.Lookup("export-media-path")), "while binding flag")
	util.Check(viper.BindPFlag("export.template_ids", templatesFlag.Lookup("template-ids")), "while binding flag")
	util.Check(viper.BindPFlag("export.tables", tablesFlag.Lookup("tables")), "while binding flag")

	util.Check(viper.BindPFlag("export.inspection.included_inactive_items", inspectionFlags.Lookup("inspection-include-inactive-items")), "while binding flag")
	util.Check(viper.BindPFlag("export.inspection.archived", inspectionFlags.Lookup("inspection-archived")), "while binding flag")
	util.Check(viper.BindPFlag("export.inspection.completed", inspectionFlags.Lookup("inspection-completed")), "while binding flag")
	util.Check(viper.BindPFlag("export.inspection.skip_ids", inspectionFlags.Lookup("inspection-skip-ids")), "while binding flag")

	util.Check(viper.BindPFlag("report.format", reportFlags.Lookup("format")), "while binding flag")
	util.Check(viper.BindPFlag("report.preference_id", reportFlags.Lookup("preference-id")), "while binding flag")
}

func addCmd(cmd *cobra.Command, flags ...*flag.FlagSet) {
	for _, f := range flags {
		cmd.PersistentFlags().AddFlagSet(f)
	}

	RootCmd.AddCommand(cmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigFile("iauditor-exporter.yaml")
	}

	viper.SetEnvPrefix("IAUD")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Config file not found:", viper.ConfigFileUsed())
	}
}
