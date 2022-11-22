package feed

import (
	"context"

	"github.com/SafetyCulture/safetyculture-exporter/internal/app/api"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/config"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/util"
	"github.com/spf13/viper"
)

const maxConcurrentGoRoutines = 10

func getInspectionFeed(inspectionConfig *config.InspectionConfig, templateIDs []string) *InspectionFeed {
	return &InspectionFeed{
		SkipIDs:       inspectionConfig.SkipIDs,
		ModifiedAfter: inspectionConfig.ModifiedAfter,
		TemplateIDs:   templateIDs,
		Archived:      inspectionConfig.Archived,
		Completed:     inspectionConfig.Completed,
		Incremental:   inspectionConfig.Incremental,
		Limit:         inspectionConfig.Limit,
		WebReportLink: inspectionConfig.WebReportLink,
	}
}

func getTemplateIDs(v *viper.Viper) []string {
	return v.GetStringSlice("export.template_ids")
}

// GetSheqsyFeeds returns list of all available data feeds for sheqsy
func GetSheqsyFeeds() []Feed {
	return []Feed{
		&SheqsyEmployeeFeed{},
		&SheqsyDepartmentEmployeeFeed{},
		&SheqsyDepartmentFeed{},
		&SheqsyActivityFeed{},
		&SheqsyShiftFeed{},
	}
}

// ExportInspectionReports download all the reports for inspections and stores them on disk
func ExportInspectionReports(v *viper.Viper, apiClient *api.Client, exporter *ReportExporter) error {
	logger := util.GetLogger()
	ctx := context.Background()

	resp, err := apiClient.WhoAmI(ctx)
	util.Check(err, "failed to get details of the current user")

	logger.Infof("Exporting inspection reports by user: %s %s", resp.Firstname, resp.Lastname)

	feed := getInspectionFeed(config.GetInspectionConfig(v), getTemplateIDs(v))
	err = feed.Export(ctx, apiClient, exporter, resp.OrganisationID)
	util.Check(err, "failed to export inspection feed")

	err = exporter.SaveReports(ctx, apiClient, feed)
	if err != nil {
		logger.Info("Export finished")
	}

	return err
}
