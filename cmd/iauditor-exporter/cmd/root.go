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
)

var cfgFile string

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

	// TODO - Can we validate these tokens and throw error if they are wrong?
	RootCmd.PersistentFlags().StringP("access-token", "t", "", "API Access Token")
	RootCmd.PersistentFlags().String("api-url", "https://api.safetyculture.io", "API URL")
	RootCmd.PersistentFlags().Bool("tls-skip-verify", false, "Skip verification of API TLS certificates")
	RootCmd.PersistentFlags().String("tls-cert", "", "Custom root CA certificate to use when making API requests")
	RootCmd.PersistentFlags().String("proxy-url", "", "Proxy URL for making API requests through")

	RootCmd.PersistentFlags().String("db-dialect", "mysql", "Database dialect. mysql, postgres and sqlserver are the only valid options.")
	RootCmd.PersistentFlags().String("db-connection-string", "", "Database connection string")

	RootCmd.PersistentFlags().String("export-path", "./export/", "CSV Export Path")
	RootCmd.PersistentFlags().StringSlice("template-ids", []string{}, "Template IDs to filter inspections and schedules by (default all)")
	RootCmd.PersistentFlags().StringSlice("inspection-skip-ids", []string{}, "Skip storing these inspection IDs")

	RootCmd.PersistentFlags().Bool("inspection-incremental-update", true, "Update inspections, inspection_items and templates tables incrementally")
	RootCmd.PersistentFlags().Bool("inspection-include-inactive-items", false, "Include inactive items in the inspection_items table (default false)")
	RootCmd.PersistentFlags().String("inspection-archived", "false", "Return archived inspections, false, true or both")
	RootCmd.PersistentFlags().String("inspection-completed", "both", "Return completed inspections, false, true or both")
	RootCmd.PersistentFlags().StringSlice("tables", []string{}, "Tables to export (default all)")

	RootCmd.PersistentFlags().Bool("create-schema-only", false, "Create schema only (default false)")

	util.Check(viper.BindPFlag("access_token", RootCmd.PersistentFlags().Lookup("access-token")), "while binding flag")

	util.Check(viper.BindPFlag("api.url", RootCmd.PersistentFlags().Lookup("api-url")), "while binding flag")
	util.Check(viper.BindPFlag("api.tls_skip_verify", RootCmd.PersistentFlags().Lookup("tls-skip-verify")), "while binding flag")
	util.Check(viper.BindPFlag("api.tls_cert", RootCmd.PersistentFlags().Lookup("tls-cert")), "while binding flag")
	util.Check(viper.BindPFlag("api.proxy_url", RootCmd.PersistentFlags().Lookup("proxy-url")), "while binding flag")

	util.Check(viper.BindPFlag("db.dialect", RootCmd.PersistentFlags().Lookup("db-dialect")), "while binding flag")
	util.Check(viper.BindPFlag("db.connection_string", RootCmd.PersistentFlags().Lookup("db-connection-string")), "while binding flag")

	util.Check(viper.BindPFlag("export.path", RootCmd.PersistentFlags().Lookup("export-path")), "while binding flag")
	util.Check(viper.BindPFlag("export.template_ids", RootCmd.PersistentFlags().Lookup("template-ids")), "while binding flag")
	util.Check(viper.BindPFlag("export.tables", RootCmd.PersistentFlags().Lookup("tables")), "while binding flag")

	util.Check(viper.BindPFlag("export.inspection.incremental", RootCmd.PersistentFlags().Lookup("inspection-incremental-update")), "while binding flag")
	util.Check(viper.BindPFlag("export.inspection.included_inactive_items", RootCmd.PersistentFlags().Lookup("inspection-include-inactive-items")), "while binding flag")
	util.Check(viper.BindPFlag("export.inspection.archived", RootCmd.PersistentFlags().Lookup("inspection-archived")), "while binding flag")
	util.Check(viper.BindPFlag("export.inspection.completed", RootCmd.PersistentFlags().Lookup("inspection-completed")), "while binding flag")
	util.Check(viper.BindPFlag("export.inspection.skip_ids", RootCmd.PersistentFlags().Lookup("inspection-skip-ids")), "while binding flag")

	util.Check(viper.BindPFlag("export.schema_only", RootCmd.PersistentFlags().Lookup("create-schema-only")), "while binding flag")

	// Add sub-commands
	RootCmd.AddCommand(export.Cmds()...)
	RootCmd.AddCommand(configure.Cmd())
	RootCmd.AddCommand(&cobra.Command{
		Hidden: true,
		Use:    "docs",
		RunE:   writeDocs,
	})

	reportCmd := &cobra.Command{
		Use:     "report",
		Short:   "Export inspection reports",
		Example: `iauditor-exporter reports`,
		RunE:    export.ExportReports,
	}
	reportCmd.PersistentFlags().StringSlice("format", []string{}, "Export format (PDF,WORD)")
	util.Check(viper.BindPFlag("report.format", reportCmd.PersistentFlags().Lookup("format")), "while binding flag")
	RootCmd.AddCommand(reportCmd)

	initConfig()
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
