package feed

import (
	"context"
	"sync"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/config"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
	"github.com/spf13/viper"
)

// GetFeeds returns list of all available data feeds
func GetFeeds(v *viper.Viper) []Feed {
	inspectionIncludeInactiveItems := v.GetBool("export.inspection.included_inactive_items")
	templateIDs := getTemplateIDs(v)
	inspectionConfig := config.GetInspectionConfig(v)

	return []Feed{
		&InspectionFeed{
			SkipIDs:       inspectionConfig.SkipIDs,
			ModifiedAfter: inspectionConfig.ModifiedAfter,
			TemplateIDs:   templateIDs,
			Archived:      inspectionConfig.Archived,
			Completed:     inspectionConfig.Completed,
			Incremental:   inspectionConfig.Incremental,
		},
		&InspectionItemFeed{
			SkipIDs:         inspectionConfig.SkipIDs,
			ModifiedAfter:   inspectionConfig.ModifiedAfter,
			TemplateIDs:     templateIDs,
			Archived:        inspectionConfig.Archived,
			Completed:       inspectionConfig.Completed,
			IncludeInactive: inspectionIncludeInactiveItems,
			Incremental:     inspectionConfig.Incremental,
		},
		&TemplateFeed{
			Incremental: inspectionConfig.Incremental,
		},
		&SiteFeed{},
		&UserFeed{},
		&GroupFeed{},
		&GroupUserFeed{},
		&ScheduleFeed{
			TemplateIDs: templateIDs,
		},
		&ScheduleAssigneeFeed{
			TemplateIDs: templateIDs,
		},
		&ScheduleOccurrenceFeed{
			TemplateIDs: templateIDs,
		},
	}
}

func getInspectionFeed(v *viper.Viper, inspectionConfig *config.InspectionConfig, templateIDs []string) *InspectionFeed {
	return &InspectionFeed{
		SkipIDs:       inspectionConfig.SkipIDs,
		ModifiedAfter: inspectionConfig.ModifiedAfter,
		TemplateIDs:   templateIDs,
		Archived:      inspectionConfig.Archived,
		Completed:     inspectionConfig.Completed,
		Incremental:   inspectionConfig.Incremental,
	}
}

func getTemplateIDs(v *viper.Viper) []string {
	return v.GetStringSlice("export.template_ids")
}

// CreateSchemas generates schemas for the data feeds without fetching any data
func CreateSchemas(v *viper.Viper, exporter Exporter) error {
	logger := util.GetLogger()
	logger.Info("Creating schemas started")

	for _, feed := range GetFeeds(v) {
		err := feed.CreateSchema(exporter)
		util.Check(err, "failed to create schema")
	}

	logger.Info("Creating schemas finished")
	return nil
}

// WriteSchemas is used to print the schema of each feed to console output
func WriteSchemas(v *viper.Viper, exporter *SchemaExporter) error {
	logger := util.GetLogger()
	logger.Info("Writing schemas started")

	for _, feed := range GetFeeds(v) {
		err := exporter.CreateSchema(feed, feed.RowsModel())
		util.Check(err, "failed to create schema")

		err = exporter.WriteSchema(feed)
		util.Check(err, "failed to write schema")
	}

	logger.Info("Writing schemas finished")
	return nil
}

// ExportFeeds fetches all the feeds data from server and stores them in the format provided
func ExportFeeds(v *viper.Viper, apiClient api.Client, exporter Exporter) error {
	logger := util.GetLogger()
	ctx := context.Background()

	var wg sync.WaitGroup
	tables := v.GetStringSlice("export.tables")
	tablesMap := map[string]bool{}
	for _, table := range tables {
		tablesMap[table] = true
	}

	// TODO. Should validate auth before doing anything

	for _, feed := range GetFeeds(v) {
		if tablesMap[feed.Name()] || len(tables) == 0 {
			wg.Add(1)

			go func(f Feed) {
				defer wg.Done()

				err := f.Export(ctx, apiClient, exporter)
				util.Check(err, "failed to export")
			}(feed)
		}
	}

	wg.Wait()

	logger.Info("Export finished")

	return nil
}

// ExportInspectionReports download all the reports for inspections and stores them on disk
func ExportInspectionReports(v *viper.Viper, apiClient api.Client, exporter *ReportExporter) error {
	logger := util.GetLogger()
	ctx := context.Background()

	feed := getInspectionFeed(v, config.GetInspectionConfig(v), getTemplateIDs(v))
	err := feed.Export(ctx, apiClient, exporter)
	util.Check(err, "failed to export inspection feed")

	err = exporter.SaveReports(ctx, apiClient, feed)
	if err != nil {
		logger.Info("Export finished")
	}

	return err
}
