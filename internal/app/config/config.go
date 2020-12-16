package config

import (
	"github.com/spf13/viper"
)

type inspectionConfig struct {
	SkipIDs       []string
	ModifiedAfter string
	Archived      string
	Completed     string
	Incremental   bool
}

func GetInspectionConfig(v *viper.Viper) *inspectionConfig {
	return &inspectionConfig{
		SkipIDs:       v.GetStringSlice("export.inspection.skip_ids"),
		ModifiedAfter: v.GetString("export.inspection.modified_after"),
		Archived:      v.GetString("export.inspection.archived"),
		Completed:     v.GetString("export.inspection.completed"),
		Incremental:   v.GetBool("export.inspection.incremental"),
	}
}
