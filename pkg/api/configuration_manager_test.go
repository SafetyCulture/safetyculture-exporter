package api_test

import (
	"os"
	"testing"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigurationManagerFromFile_when_invalid_filename(t *testing.T) {
	cm, err := api.NewConfigurationManagerFromFile("", "fake_file")
	require.Nil(t, cm)
	assert.Equal(t, "invalid file name provided", err.Error())
}

func TestNewConfigurationManagerFromFile_when_empty_filename(t *testing.T) {
	cm, err := api.NewConfigurationManagerFromFile("", "  ")
	require.Nil(t, cm)
	assert.Equal(t, "invalid file name provided", err.Error())
}

func TestNewConfigurationManagerFromFile_when_file_is_missing(t *testing.T) {
	cm, err := api.NewConfigurationManagerFromFile("", "abc.yaml")
	require.Nil(t, cm)
	assert.Equal(t, "read file: open abc.yaml: no such file or directory", err.Error())
}

func TestNewConfigurationManager_should_use_empty_time(t *testing.T) {
	cm := api.NewConfigurationManager("", "fixtures/valid_no_time.yaml")
	assert.NotNil(t, cm)
	assert.NotNil(t, cm.Configuration)
	assert.EqualValues(t, "", cm.Configuration.Db.ConnectionString)
	assert.EqualValues(t, time.Time{}, cm.Configuration.Export.ModifiedAfter.Time)
}

func TestNewConfigurationManagerFromFile_when_filename_exists_with_time(t *testing.T) {
	cm, err := api.NewConfigurationManagerFromFile("", "fixtures/valid_with_time.yaml")
	require.Nil(t, err)
	require.NotNil(t, cm)
	require.NotNil(t, cm.Configuration)

	cfg := cm.Configuration
	exp, _ := time.Parse("2006-01-02", "2022-11-29")
	assert.Equal(t, exp, cfg.Export.ModifiedAfter.Time)
}

func TestNewConfigurationManagerFromFile_should_create_file(t *testing.T) {
	_ = os.Remove("fake_file.yaml")
	cm := api.NewConfigurationManager("", "fake_file.yaml")
	err := cm.SaveConfiguration()
	assert.Nil(t, err)
	assert.NotNil(t, cm)
	assert.NotNil(t, cm.Configuration)
	_, err = os.Stat("fake_file.yaml")
	assert.Nil(t, err)
	_ = os.Remove("fake_file.yaml")
}

func TestNewConfigurationManagerFromFile_when_filename_exists_without_time(t *testing.T) {
	cm, err := api.NewConfigurationManagerFromFile("", "fixtures/valid_no_time.yaml")
	assert.Nil(t, err)
	require.NotNil(t, cm)
	require.NotNil(t, cm.Configuration)

	cfg := cm.Configuration

	// root section
	assert.Equal(t, "fake_token", cfg.AccessToken)
	assert.Equal(t, "fake_company_id", cfg.SheqsyCompanyID)
	assert.Equal(t, "fake_username", cfg.SheqsyUsername)
	assert.Equal(t, "123456", cfg.SheqsyPassword)

	// api section
	assert.Equal(t, "https://fake_proxy.com", cfg.API.ProxyURL)
	assert.Equal(t, "https://app.sheqsy.com", cfg.API.SheqsyURL)
	assert.Equal(t, "https://api.safetyculture.io", cfg.API.URL)
	assert.Equal(t, "", cfg.API.TLSCert)
	assert.False(t, cfg.API.TLSSkipVerify)

	// csv section
	assert.Equal(t, 1000000, cfg.Csv.MaxRowsPerFile)

	// db section
	assert.Equal(t, "fake_connection_string", cfg.Db.ConnectionString)
	assert.Equal(t, "mysql", cfg.Db.Dialect)

	// export section
	assert.Equal(t, 100, cfg.Export.Action.Limit)
	assert.True(t, cfg.Export.Incremental)
	assert.Equal(t, "false", cfg.Export.Inspection.Archived)
	assert.Equal(t, "true", cfg.Export.Inspection.Completed)
	assert.False(t, cfg.Export.Inspection.IncludedInactiveItems)
	assert.Equal(t, 100, cfg.Export.Inspection.Limit)
	assert.Equal(t, []string{"ID1", "ID2"}, cfg.Export.Inspection.SkipIds)
	assert.Equal(t, "private", cfg.Export.Inspection.WebReportLink)
	assert.False(t, cfg.Export.Media)
	assert.Equal(t, "./export/media/", cfg.Export.MediaPath)
	assert.Equal(t, "./export/", cfg.Export.Path)
	assert.Equal(t, time.Time{}, cfg.Export.ModifiedAfter.Time)
	assert.False(t, cfg.Export.Site.IncludeDeleted)
	assert.False(t, cfg.Export.Site.IncludeFullHierarchy)
	assert.Equal(t, []string{"TA1", "TA2", "TA3"}, cfg.Export.Tables)
	assert.Equal(t, []string{}, cfg.Export.TemplateIds)

	// report section
	assert.Equal(t, "INSPECTION_TITLE", cfg.Report.FilenameConvention)
	assert.Equal(t, "", cfg.Report.PreferenceID)
	assert.Equal(t, 15, cfg.Report.RetryTimeout)
	assert.Equal(t, []string{"PDF"}, cfg.Report.Format)
}

func TestConfigurationManager_SaveConfiguration(t *testing.T) {
	_ = os.Remove("fake_file.yaml")
	cm := api.NewConfigurationManager("", "fake_file.yaml")
	require.NotNil(t, cm)
	require.NotNil(t, cm.Configuration)

	// mutate
	cm.Configuration.AccessToken = "new-access-token"
	cm.Configuration.Export.Tables = []string{"users", "inspections"}
	cm.Configuration.Db.Dialect = "sqlserver"
	cm.Configuration.Export.Inspection.Limit = 25
	err := cm.SaveConfiguration()
	assert.Nil(t, err)

	// read the file as new
	newCm, err := api.NewConfigurationManagerFromFile("", "fake_file.yaml")
	require.Nil(t, err)
	require.NotNil(t, newCm)
	require.NotNil(t, newCm.Configuration)

	// changed values
	assert.EqualValues(t, "new-access-token", newCm.Configuration.AccessToken)
	assert.EqualValues(t, []string{"users", "inspections"}, newCm.Configuration.Export.Tables)
	assert.EqualValues(t, "sqlserver", newCm.Configuration.Db.Dialect)
	assert.EqualValues(t, 25, newCm.Configuration.Export.Inspection.Limit)

	// existing defaults
	assert.EqualValues(t, "https://api.safetyculture.io", newCm.Configuration.API.URL)
	assert.EqualValues(t, "https://app.sheqsy.com", newCm.Configuration.API.SheqsyURL)
	assert.EqualValues(t, 1000000, newCm.Configuration.Csv.MaxRowsPerFile)
	assert.EqualValues(t, 100, newCm.Configuration.Export.Action.Limit)
	assert.True(t, newCm.Configuration.Export.Incremental)
	assert.EqualValues(t, "false", newCm.Configuration.Export.Inspection.Archived)
	assert.EqualValues(t, "true", newCm.Configuration.Export.Inspection.Completed)
	assert.EqualValues(t, "private", newCm.Configuration.Export.Inspection.WebReportLink)
	assert.EqualValues(t, "./export/media/", newCm.Configuration.Export.MediaPath)
	assert.EqualValues(t, "./export/", newCm.Configuration.Export.Path)
	assert.EqualValues(t, "INSPECTION_TITLE", newCm.Configuration.Report.FilenameConvention)
	assert.EqualValues(t, []string{"PDF"}, newCm.Configuration.Report.Format)
	assert.EqualValues(t, 15, newCm.Configuration.Report.RetryTimeout)

	_ = os.Remove("fake_file.yaml")
}

func TestMapViperConfigToConfigurationOptions_ShouldRespectLimit(t *testing.T) {
	cm, err := api.NewConfigurationManagerFromFile("", "fixtures/test_limit_50.yaml")
	require.Nil(t, err)
	require.NotNil(t, cm)
	assert.EqualValues(t, 50, cm.Configuration.Export.Action.Limit)
	assert.EqualValues(t, 50, cm.Configuration.Export.Issue.Limit)
}

func TestMapViperConfigToConfigurationOptions_ShouldEnforceLimit(t *testing.T) {
	cm, err := api.NewConfigurationManagerFromFile("", "fixtures/test_limit_101.yaml")
	require.Nil(t, err)
	require.NotNil(t, cm)
	assert.EqualValues(t, 100, cm.Configuration.Export.Action.Limit)
	assert.EqualValues(t, 100, cm.Configuration.Export.Issue.Limit)
}
