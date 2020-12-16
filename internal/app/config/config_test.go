package config_test

import (
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/config"
	"github.com/bmizerany/assert"
	"github.com/spf13/viper"
)

func TestGetInspectionConfig(t *testing.T) {
	viperConfig := viper.New()
	viperConfig.Set("export.inspection.incremental", true)
	viperConfig.Set("export.inspection.modified_after", "2010-12-20")
	viperConfig.Set("export.inspection.archived", "both")
	viperConfig.Set("export.inspection.completed", "both")
	viperConfig.Set("export.inspection.skip_ids", "1 2 3")

	actual := config.GetInspectionConfig(viperConfig)
	assert.Equal(t, true, actual.Incremental)
	assert.Equal(t, "2010-12-20", actual.ModifiedAfter)
	assert.Equal(t, "both", actual.Archived)
	assert.Equal(t, "both", actual.Completed)
	assert.Equal(t, []string{"1", "2", "3"}, actual.SkipIDs)
}
