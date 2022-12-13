package api

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/feed"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/inspections"
	"gopkg.in/yaml.v3"
)

// ExporterConfiguration is the equivalent struct of YAML
type ExporterConfiguration struct {
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
		Asset struct {
			Limit int `yaml:"limit"`
		} `yaml:"asset"`
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
		SchemaOnly    bool   `yaml:"-"`
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

// mTime wrapper around time.Time in order to have a custom YAML marshaller/un-marshaller
type mTime struct {
	time.Time
}

// UnmarshalYAML custom un-marshaller for time.Time since empty strings throws an error
func (mt *mTime) UnmarshalYAML(value *yaml.Node) error {
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

	mt.Time = t
	return nil
}

// MarshalYAML custom marshaller for time, when is ZERO, marshal as empty string
// note: doesn't work with pointer receiver
func (mt mTime) MarshalYAML() (interface{}, error) {
	if mt.Time.IsZero() {
		return "", nil
	}

	return mt.Time.Format("2006-01-02"), nil
}

// ConfigurationManager wrapper for configuration and fileName
type ConfigurationManager struct {
	path 		  string
	fileName      string
	Configuration *ExporterConfiguration
}

// loadConfiguration will load the specified YAML file if exists and map it
func (c *ConfigurationManager) loadConfiguration() error {
	if len(strings.TrimSpace(c.fileName)) == 0 || !strings.HasSuffix(c.fileName, ".yaml") {
		return fmt.Errorf("invalid file name provided")
	}

	yamlContents, err := os.ReadFile(path.Join(c.path, c.fileName))
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	if err := yaml.Unmarshal(yamlContents, c.Configuration); err != nil {
		return fmt.Errorf("unmarshal file: %w", err)
	}

	return nil
}

// ApplySafetyGuards will adjust certain values to acceptable maximum values
func (c *ConfigurationManager) ApplySafetyGuards() {
	// caps action batch limit to 100
	if c.Configuration.Export.Action.Limit > 100 {
		c.Configuration.Export.Action.Limit = 100
	}
	// caps issue batch limit to 100
	if c.Configuration.Export.Issue.Limit > 100 {
		c.Configuration.Export.Issue.Limit = 100
	}
}

// SaveConfiguration will save the configuration to the file
func (c *ConfigurationManager) SaveConfiguration() error {
	if len(strings.TrimSpace(c.fileName)) == 0 || !strings.HasSuffix(c.fileName, ".yaml") {
		return fmt.Errorf("invalid file name provided")
	}

	data, err := yaml.Marshal(c.Configuration)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(path.Join(c.path, c.fileName), data, 0666); err != nil {
		return fmt.Errorf("writing file %s: %w", c.fileName, err)
	}
	return nil
}

// BuildConfigurationWithDefaults will set up an initial configuration with default values
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
	cfg.Export.Issue.Limit = 100
	cfg.Export.MediaPath = "./export/media/"
	cfg.Export.Path = "./export/"
	cfg.Report.FilenameConvention = "INSPECTION_TITLE"
	cfg.Report.Format = []string{"PDF"}
	cfg.Report.RetryTimeout = 15

	return cfg
}

// NewConfigurationManagerFromFile will create a ConfigurationManager with data from the specified file
func NewConfigurationManagerFromFile(path string, fileName string) (*ConfigurationManager, error) {
	cm := &ConfigurationManager{
		path: path,
		fileName:      fileName,
		Configuration: &ExporterConfiguration{},
	}

	err := cm.loadConfiguration()
	if err != nil {
		return nil, err
	}

	cm.ApplySafetyGuards()
	return cm, nil
}

// NewConfigurationManager will create a ConfigurationManager with default data,
func NewConfigurationManager(path string, fileName string) *ConfigurationManager {
	cm := &ConfigurationManager{
		path: path,
		fileName:      fileName,
		Configuration: BuildConfigurationWithDefaults(),
	}

	cm.ApplySafetyGuards()
	return cm
}

func (ec *ExporterConfiguration) ToExporterConfig() *feed.ExporterFeedCfg {
	return &feed.ExporterFeedCfg{
		AccessToken:                           ec.AccessToken,
		ExportTables:                          ec.Export.Tables,
		SheqsyUsername:                        ec.SheqsyUsername,
		SheqsyCompanyID:                       ec.SheqsyCompanyID,
		ExportInspectionSkipIds:               ec.Export.Inspection.SkipIds,
		ExportModifiedAfterTime:               ec.Export.ModifiedAfter.Time,
		ExportTemplateIds:                     ec.Export.TemplateIds,
		ExportInspectionArchived:              ec.Export.Inspection.Archived,
		ExportInspectionCompleted:             ec.Export.Inspection.Completed,
		ExportInspectionIncludedInactiveItems: ec.Export.Inspection.IncludedInactiveItems,
		ExportInspectionWebReportLink:         ec.Export.Inspection.WebReportLink,
		ExportIncremental:                     ec.Export.Incremental,
		ExportInspectionLimit:                 ec.Export.Inspection.Limit,
		ExportMedia:                           ec.Export.Media,
		ExportSiteIncludeDeleted:              ec.Export.Site.IncludeDeleted,
		ExportActionLimit:                     ec.Export.Action.Limit,
		ExportSiteIncludeFullHierarchy:        ec.Export.Site.IncludeFullHierarchy,
		ExportIssueLimit:                      ec.Export.Issue.Limit,
		ExportAssetLimit:                      ec.Export.Asset.Limit,
	}
}

func (ec *ExporterConfiguration) ToReporterConfig() *ReportExporterCfg {
	return &ReportExporterCfg{
		Format:       ec.Report.Format,
		PreferenceID: ec.Report.PreferenceID,
		Filename:     ec.Report.FilenameConvention,
		RetryTimeout: ec.Report.RetryTimeout,
	}
}

func (ec *ExporterConfiguration) ToInspectionConfig() *inspections.InspectionClientCfg {
	return &inspections.InspectionClientCfg{
		SkipIDs:       ec.Export.Inspection.SkipIds,
		ModifiedAfter: ec.Export.ModifiedAfter.Time,
		TemplateIDs:   ec.Export.TemplateIds,
		Archived:      ec.Export.Inspection.Archived,
		Completed:     ec.Export.Inspection.Completed,
		Incremental:   ec.Export.Incremental,
	}
}
