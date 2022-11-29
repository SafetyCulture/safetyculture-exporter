package configure

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	AccessToken string `yaml:"access_token"`
	API         struct {
		ProxyURL      string `yaml:"proxy_url"`
		SheqsyURL     string `yaml:"sheqsy_url"`
		TLSCert       string `yaml:"tls_cert"`
		TLSSkipVerify bool   `yaml:"tls_skip_verify"`
		URL           string `yaml:"url"`
	} `yaml:"api"`
	Csv struct {
		MaxRowsPerFile int `yaml:"max_rows_per_file"`
	} `yaml:"csv"`
	Db struct {
		ConnectionString string `yaml:"connection_string"`
		Dialect          string `yaml:"dialect"`
	} `yaml:"db"`
	Export struct {
		Action struct {
			Limit int `yaml:"limit"`
		} `yaml:"action"`
		Incremental bool `yaml:"incremental"`
		Inspection  struct {
			Archived              string   `yaml:"archived"`
			Completed             string   `yaml:"completed"`
			IncludedInactiveItems bool     `yaml:"included_inactive_items"`
			Limit                 int      `yaml:"limit"`
			SkipIds               []string `yaml:"skip_ids"`
			WebReportLink         string   `yaml:"web_report_link"`
		} `yaml:"inspection"`
		Media         bool     `yaml:"media"`
		MediaPath     string   `yaml:"media_path"`
		ModifiedAfter yamlTime `yaml:"modified_after"`
		Path          string   `yaml:"path"`
		Site          struct {
			IncludeDeleted       bool `yaml:"include_deleted"`
			IncludeFullHierarchy bool `yaml:"include_full_hierarchy"`
		} `yaml:"site"`
		Tables      []string `yaml:"tables"`
		TemplateIds []string `yaml:"template_ids"`
	} `yaml:"export"`
	Report struct {
		FilenameConvention string   `yaml:"filename_convention"`
		Format             []string `yaml:"format"`
		PreferenceID       string   `yaml:"preference_id"`
		RetryTimeout       int      `yaml:"retry_timeout"`
	} `yaml:"report"`
	SheqsyCompanyID string `yaml:"sheqsy_company_id"`
	SheqsyPassword  string `yaml:"sheqsy_password"`
	SheqsyUsername  string `yaml:"sheqsy_username"`
}

type yamlTime time.Time

func (yt *yamlTime) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var timeString string
	err := unmarshal(&timeString)
	if err != nil {
		return err
	}
	timeString = strings.TrimSpace(timeString)

	var t time.Time
	switch timeString {
	case "":
		t = time.Time{}
	default:
		t, err = time.Parse("2006-01-02", timeString)
		if err != nil {
			return fmt.Errorf("failed to parse '%s' to time.Time: %v", timeString, err)
		}
	}

	*yt = yamlTime(t)
	return nil
}

func (yt *yamlTime) Time() time.Time {
	return time.Time(*yt)
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
// If the configuration file doesn't exist, will be created (WIP)
func NewConfigurationManager(fn string) (error, *ConfigurationManager) {
	if len(strings.TrimSpace(fn)) == 0 || !strings.HasSuffix(fn, ".yaml") {
		return fmt.Errorf("invalid file name provided"), nil
	}

	cm := &ConfigurationManager{
		fileName:      fn,
		Configuration: &Configuration{},
	}

	_, err := os.Stat(fn)
	if err == nil {
		// file exists
		err := cm.LoadConfiguration()
		if err != nil {
			return err, nil
		}
	}
	return nil, cm
}
