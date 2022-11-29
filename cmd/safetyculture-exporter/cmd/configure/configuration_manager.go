package configure

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Configuration is the equivalent struct of YAML
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

// UnmarshalYAML custom unmarshaler for time.Time since empty strings throws an error
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

// ConfigurationManager wrapper for configuration and fileName
type ConfigurationManager struct {
	fileName      string
	Configuration *Configuration
}

func (c *ConfigurationManager) createEmptyConfiguration() error {
	data, err := yaml.Marshal(c.Configuration)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(c.fileName, data, 0666); err != nil {
		return fmt.Errorf("writing file %s: %w", c.fileName, err)
	}
	return nil
}

func (c *ConfigurationManager) loadConfiguration() error {
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
// fn - filename
// autoLoad - If the configuration file exists, will be loaded
// autoCreate - If the configuration file doesn't exist, will be created
func NewConfigurationManager(fn string, autoLoad bool, autoCreate bool) (error, *ConfigurationManager) {
	if len(strings.TrimSpace(fn)) == 0 || !strings.HasSuffix(fn, ".yaml") {
		return fmt.Errorf("invalid file name provided"), nil
	}

	cm := &ConfigurationManager{
		fileName:      fn,
		Configuration: &Configuration{},
	}

	_, err := os.Stat(fn)

	switch {
	case err == nil:
		// file exists
		if autoLoad {
			err := cm.loadConfiguration()
			if err != nil {
				return err, nil
			}
		}
		return nil, cm

	case errors.Is(err, os.ErrNotExist):
		if autoCreate {
			// create the configuration
			err := cm.createEmptyConfiguration()
			if err != nil {
				return err, nil
			}
		}
		return nil, cm
	default:
		return err, nil
	}
}
