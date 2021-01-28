package config

import (
	"time"

	"github.com/spf13/viper"
)

// InspectionConfig includes all the configurations available when fetching inspections
type InspectionConfig struct {
	SkipIDs       []string
	ModifiedAfter time.Time
	Archived      string
	Completed     string
	Incremental   bool
}

// GetInspectionConfig returns configurations that have been set for fetching inspections
func GetInspectionConfig(v *viper.Viper) *InspectionConfig {
	return &InspectionConfig{
		SkipIDs:       v.GetStringSlice("export.inspection.skip_ids"),
		ModifiedAfter: v.GetTime("export.inspection.modified_after"),
		Archived:      v.GetString("export.inspection.archived"),
		Completed:     v.GetString("export.inspection.completed"),
		Incremental:   v.GetBool("export.inspection.incremental"),
	}
}
