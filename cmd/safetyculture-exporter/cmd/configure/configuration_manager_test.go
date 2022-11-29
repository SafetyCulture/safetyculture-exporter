package configure_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/cmd/safetyculture-exporter/cmd/configure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfiguration_when_invalid_filename(t *testing.T) {
	err, cm := configure.NewConfigurationManager("fake_file", true, true)
	require.Nil(t, cm)
	assert.Equal(t, "invalid file name provided", err.Error())
}

func TestNewConfiguration_when_empty_filename(t *testing.T) {
	err, cm := configure.NewConfigurationManager(" ", true, true)
	require.Nil(t, cm)
	assert.Equal(t, "invalid file name provided", err.Error())
}

func TestNewConfigurationManager_should_not_read_file(t *testing.T) {
	err, cm := configure.NewConfigurationManager("fixtures/valid_no_time.yaml", false, false)
	require.Nil(t, err)
	assert.NotNil(t, cm)
	assert.NotNil(t, cm.Configuration)
}

func TestNewConfiguration_should_create_when_valid_filename_does_not_exist(t *testing.T) {
	os.Remove("fake_file.yaml")
	err, cm := configure.NewConfigurationManager("fake_file.yaml", true, true)
	assert.Nil(t, err)
	assert.NotNil(t, cm)
	assert.NotNil(t, cm.Configuration)
	_, err = os.Stat("fake_file.yaml")
	assert.Nil(t, err)
	os.Remove("fake_file.yaml")
}

func TestNewConfiguration_should_not_create_when_valid_filename_does_not_exist(t *testing.T) {
	err, cm := configure.NewConfigurationManager("fake_file.yaml", true, false)
	assert.Nil(t, err)
	assert.NotNil(t, cm)
	assert.NotNil(t, cm.Configuration)
	_, err = os.Stat("fake_file.yaml")
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, os.ErrNotExist))
}

func TestNewConfigurationManager_when_filename_exists_without_time(t *testing.T) {
	err, cm := configure.NewConfigurationManager("fixtures/valid_no_time.yaml", true, true)
	require.Nil(t, err)
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
	assert.Equal(t, time.Time{}, cfg.Export.ModifiedAfter.Time())
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

func TestNewConfigurationManager_when_filename_exists_with_time(t *testing.T) {
	err, cm := configure.NewConfigurationManager("fixtures/valid_with_time.yaml", true, true)
	require.Nil(t, err)
	require.NotNil(t, cm)
	require.NotNil(t, cm.Configuration)

	cfg := cm.Configuration
	exp, _ := time.Parse("2006-01-02", "2022-11-29")
	assert.Equal(t, exp, cfg.Export.ModifiedAfter.Time())
}
