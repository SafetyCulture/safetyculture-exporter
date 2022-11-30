package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ExporterConfiguration is the equivalent struct of YAML
type ExporterConfiguration struct {
	AccessToken string `yaml:"access_token"`
	SchemaOnly  bool   `yaml:"schema_only"`
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
		Issue struct {
			Limit int `yaml:"limit"`
		} `yaml:"issue"`
		Media         bool   `yaml:"media"`
		MediaPath     string `yaml:"media_path"`
		ModifiedAfter mTime  `yaml:"modified_after"`
		Path          string `yaml:"path"`
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

type mTime struct {
	time.Time
}

// UnmarshalYAML custom unmarshaler for time.Time since empty strings throws an error
func (yt *mTime) UnmarshalYAML(value *yaml.Node) error {
	var timeString string
	err := value.Decode(&timeString)
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

	yt.Time = t
	return nil
}

func (yt mTime) MarshalYAML() (interface{}, error) {
	if yt.Time.IsZero() {
		return "", nil
	}

	return yt.Time.Format("2006-01-02"), nil
}

// ConfigurationManager wrapper for configuration and fileName
type ConfigurationManager struct {
	fileName      string
	Configuration *ExporterConfiguration
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

	c.applySafetyGuards()
	return nil
}

func (c *ConfigurationManager) applySafetyGuards() {
	// caps action batch limit to 100
	if c.Configuration.Export.Action.Limit > 100 {
		c.Configuration.Export.Action.Limit = 100
	}
	// caps issue batch limit to 100
	if c.Configuration.Export.Issue.Limit > 100 {
		c.Configuration.Export.Issue.Limit = 100
	}
}

func (c *ConfigurationManager) SaveConfiguration() error {
	return nil
}

// ConfigurationManagerDefaults specifies the defaults for the configuration file
type ConfigurationManagerDefaults struct {
	DefaultSheqsyURL                     string
	DefaultSafetyCultureURL              string
	DefaultCsvMaxRowsPerFile             int
	DefaultDbDialect                     string
	DefaultExportActionLimit             int
	DefaultExportIncremental             bool
	DefaultExportInspectionArchived      string
	DefaultExportInspectionCompleted     string
	DefaultExportInspectionLimit         int
	DefaultExportInspectionWebReportLink string
	DefaultExportMediaPath               string
	DefaultExportPath                    string
	DefaultReportFilenameConvention      string
	DefaultReportFormat                  []string
	DefaultReportRetryTimeout            int
}

func BuildConfigurationWithDefaults() *ExporterConfiguration {
	cfg := &ExporterConfiguration{}
	cfg.API.SheqsyURL = "https://app.sheqsy.com"
	cfg.API.URL = "https://api.safetyculture.io"
	cfg.Csv.MaxRowsPerFile = 1000000
	cfg.Db.Dialect = "mysql"
	cfg.Export.Action.Limit = 100
	cfg.Export.Incremental = true
	cfg.Export.Inspection.Archived = "false"
	cfg.Export.Inspection.Completed = "true"
	cfg.Export.Inspection.Limit = 100
	cfg.Export.Inspection.WebReportLink = "private"
	cfg.Export.MediaPath = "./export/media/"
	cfg.Export.Path = "./export/"
	cfg.Report.FilenameConvention = "INSPECTION_TITLE"
	cfg.Report.Format = []string{"PDF"}
	cfg.Report.RetryTimeout = 15
	return cfg

}

// NewConfigurationManager creates a new ConfigurationManager.
func NewConfigurationManager(fn string, autoLoad bool, autoCreate bool, defaultCfg *ExporterConfiguration) (*ConfigurationManager, error) {
	if len(strings.TrimSpace(fn)) == 0 || !strings.HasSuffix(fn, ".yaml") {
		return nil, fmt.Errorf("invalid file name provided")
	}

	var cfg *ExporterConfiguration = nil
	if defaultCfg != nil {
		cfg = defaultCfg
	} else {
		cfg = &ExporterConfiguration{}
	}

	cm := &ConfigurationManager{
		fileName:      fn,
		Configuration: cfg,
	}

	_, err := os.Stat(fn)

	switch {
	case err == nil:
		// file exists
		if autoLoad {
			err := cm.loadConfiguration()
			if err != nil {
				return nil, err
			}
		}
		cm.applySafetyGuards()
		return cm, nil

	case errors.Is(err, os.ErrNotExist):
		if autoCreate {
			// create the configuration
			err := cm.createEmptyConfiguration()
			if err != nil {
				return nil, err
			}
		}
		cm.applySafetyGuards()
		return cm, nil
	default:
		return nil, err
	}
}
