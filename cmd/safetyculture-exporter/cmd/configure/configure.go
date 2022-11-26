package configure

import (
	"fmt"
	"os"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/app/util"
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
			_, err := os.Stat(viper.ConfigFileUsed())
			configNoExists := os.IsNotExist(err)
			if configNoExists {
				fmt.Println("Creating a new config file:", viper.ConfigFileUsed())
			} else {
				fmt.Println("Config file found:", viper.ConfigFileUsed())
			}

			// TODO - Validate flags shouldn't be empty.
			// TODO - Optionally accept the flags as a prompt.
			util.Check(viper.WriteConfigAs(viper.ConfigFileUsed()), "while writing viper config to file")

			if configNoExists {
				fmt.Println("Config file created successfully \U0001f389")
			} else {
				fmt.Println("Config file updated successfully \U0001f389")
			}

			fmt.Println("\nVisit \"https://github.com/SafetyCulture/safetyculture-exporter#configure\" to learn more about configuration options.")

			util.Check(os.Chmod(viper.ConfigFileUsed(), 0666), "while updating file permissions")
		},
	}
}
