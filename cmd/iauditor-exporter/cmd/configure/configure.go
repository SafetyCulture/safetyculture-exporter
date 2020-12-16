package configure

import (
	"os"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Cmd is used to initialize configuration file
func Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Initialises your configuration file.",
		Long:  `Initialises your config file with your access token and other configuration options.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger := util.GetLogger()

			// TODO - Validate flags shouldn't be empty.
			// TODO - Optionally accept the flags as a prompt.
			util.Check(viper.WriteConfigAs(viper.ConfigFileUsed()), "while writing viper config to file")

			logger.Infof("Updating config file: %s", viper.ConfigFileUsed())

			util.Check(os.Chmod(viper.ConfigFileUsed(), 0666), "while updating file permissions")
		},
	}
}
