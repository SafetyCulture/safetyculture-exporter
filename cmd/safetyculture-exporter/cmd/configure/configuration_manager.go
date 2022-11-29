package configure

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	AccessToken string            `yaml:"access_token"`
	Api         *ConfigurationAPI `yaml:"api"`
	Csv         *ConfigurationCSV `yaml:"csv"`
}

type ConfigurationAPI struct {
	ProxyURL       string `yaml:"proxy_url"`
	SheqsyURL      string `yaml:"sheqsy_url"`
	URL            string `yaml:"url"`
	TLSCertificate string `yaml:"tls_cert"`
	TLSSkipVerify  bool   `yaml:"tls_skip_verify"`
}

type ConfigurationCSV struct {
	MaxRowsPerFile int `yaml:"max_rows_per_file"`
}

type ConfigurationManager struct {
	fileName      string
	Configuration *Configuration
}

func (c *ConfigurationManager) CreateEmptyConfiguration() error {
	// check if file already exists
	// create file
	// update permissions
	return nil
}

func (c *ConfigurationManager) LoadConfiguration() error {
	yamlContents, err := os.ReadFile(c.fileName)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	if err := yaml.Unmarshal(yamlContents, c.Configuration); err != nil {
		return fmt.Errorf("unmarshal file: %w", err)
	}
	return nil
}

func (c *ConfigurationManager) SaveConfiguration() error {
	return nil
}

// NewConfigurationManager creates a new ConfigurationManager.
// If the configuration file exists, will be loaded
func NewConfigurationManager(fn string) (error, *ConfigurationManager) {
	if len(strings.TrimSpace(fn)) == 0 || !strings.HasSuffix(fn, ".yaml") {
		return fmt.Errorf("invalid file name provided"), nil
	}

	cm := &ConfigurationManager{
		fileName: fn,
		Configuration: &Configuration{
			Api: &ConfigurationAPI{},
			Csv: &ConfigurationCSV{},
		},
	}

	_, err := os.Stat(fn)
	if err == nil {
		// file exists
		err := cm.LoadConfiguration()
		if err != nil {
			return err, nil
		}
	}
	//else if errors.Is(err, os.ErrNotExist) {
	//	// file will be created (?)
	//	file, err := os.Create(fn)
	//	defer file.Close()
	//	if err != nil {
	//		return err, nil
	//	}
	//}

	return nil, cm
}
