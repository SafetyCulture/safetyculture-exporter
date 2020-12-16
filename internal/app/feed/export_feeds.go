package feed

import (
	"context"
	"sync"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
	"github.com/spf13/viper"
)

// GetFeeds returns list of all available data feeds
func GetFeeds(v *viper.Viper) []Feed {
	inspectionSkipIDs := v.GetStringSlice("export.inspection.skip_ids")
	inspectionModifiedAfter := v.GetString("export.inspection.modified_after")
	inspectionArchived := v.GetString("export.inspection.archived")
	inspectionCompleted := v.GetString("export.inspection.completed")
	inspectionIncremental := v.GetBool("export.inspection.incremental")
	inspectionIncludeInactiveItems := v.GetBool("export.inspection.included_inactive_items")
	templateIDs := v.GetStringSlice("export.template_ids")

	return []Feed{
		&InspectionFeed{
			SkipIDs:       inspectionSkipIDs,
			ModifiedAfter: inspectionModifiedAfter,
			TemplateIDs:   templateIDs,
			Archived:      inspectionArchived,
			Completed:     inspectionCompleted,
			Incremental:   inspectionIncremental,
		},
		&InspectionItemFeed{
			SkipIDs:         inspectionSkipIDs,
			ModifiedAfter:   inspectionModifiedAfter,
			TemplateIDs:     templateIDs,
			Archived:        inspectionArchived,
			Completed:       inspectionCompleted,
			IncludeInactive: inspectionIncludeInactiveItems,
			Incremental:     inspectionIncremental,
		},
		&TemplateFeed{
			Incremental: inspectionIncremental,
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
