package configure_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/SafetyCulture/safetyculture-exporter/cmd/safetyculture-exporter/cmd"
	"github.com/SafetyCulture/safetyculture-exporter/cmd/safetyculture-exporter/cmd/configure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandConfigure_should_not_throw_error(t *testing.T) {
	b := bytes.NewBufferString("")
	cmd.RootCmd.SetOut(b)
	cmd.RootCmd.SetArgs([]string{"configure"})
	cmd.Execute()
	_, err := io.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewConfiguration_when_invalid_filename(t *testing.T) {
	err, cm := configure.NewConfigurationManager("fake_file")
	require.Nil(t, cm)
	assert.Equal(t, "invalid file name provided", err.Error())
}

func TestNewConfiguration_when_empty_filename(t *testing.T) {
	err, cm := configure.NewConfigurationManager(" ")
	require.Nil(t, cm)
	assert.Equal(t, "invalid file name provided", err.Error())
}

func TestNewConfiguration_when_valid_filename(t *testing.T) {
	err, cm := configure.NewConfigurationManager("fake_file.yaml")
	require.Nil(t, err)
	require.NotNil(t, cm)
	require.NotNil(t, cm.Configuration)
}

func TestNewConfigurationManager_when_filename_exists(t *testing.T) {
	err, cm := configure.NewConfigurationManager("fixtures/valid.yaml")
	require.Nil(t, err)
	require.NotNil(t, cm)
	require.NotNil(t, cm.Configuration)

	cfg := cm.Configuration
	assert.Equal(t, "fake_token", cfg.AccessToken)
	assert.Equal(t, "https://fake_proxy.com", cfg.Api.ProxyURL)
	assert.Equal(t, "https://app.sheqsy.com", cfg.Api.SheqsyURL)
	assert.Equal(t, "https://api.safetyculture.io", cfg.Api.URL)
	assert.Equal(t, "", cfg.Api.TLSCertificate)
	assert.False(t, cfg.Api.TLSSkipVerify)
	assert.Equal(t, 1000000, cfg.Csv.MaxRowsPerFile)

}
