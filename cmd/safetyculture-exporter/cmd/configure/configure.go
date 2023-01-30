package configure

import (
	"fmt"

	util "github.com/SafetyCulture/safetyculture-exporter/cmd/safetyculture-exporter/cmd/utils"
	exporterAPI "github.com/SafetyCulture/safetyculture-exporter/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Cmd is used to initialize configuration file
func Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Initialises your configuration file.",
		Long:  `Initialises your config file with your access token and other configuration options.`,
		RunE:  generateYamlConfiguration,
	}
}

func generateYamlConfiguration(cmd *cobra.Command, args []string) error {
	cm := exporterAPI.NewConfigurationManager("", viper.ConfigFileUsed())
	err := cm.SaveConfiguration()
	util.Check(err, "while writing config to file")
	fmt.Println("Config file created successfully \U0001f389")
	return nil
}
