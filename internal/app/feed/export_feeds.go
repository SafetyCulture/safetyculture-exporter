package feed

import (
	"context"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/config"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
	"github.com/spf13/viper"
	"sync"
)

const maxConcurrentGoRoutines = 10

// GetFeeds returns list of all available data feeds
func GetFeeds(v *viper.Viper) []Feed {
	inspectionIncludeInactiveItems := v.GetBool("export.inspection.included_inactive_items")
	templateIDs := getTemplateIDs(v)
	actionLimit := GetActionLimit(v)
	issueLimit := GetIssueLimit(v)
	inspectionConfig := config.GetInspectionConfig(v)
	exportMedia := viper.GetBool("export.media")
	sitesIncludeDeleted := viper.GetBool("export.site.include_deleted")

	return []Feed{
		getInspectionFeed(v, inspectionConfig, templateIDs),
		&InspectionItemFeed{
			SkipIDs:         inspectionConfig.SkipIDs,
			ModifiedAfter:   inspectionConfig.ModifiedAfter,
			TemplateIDs:     templateIDs,
			Archived:        inspectionConfig.Archived,
			Completed:       inspectionConfig.Completed,
			IncludeInactive: inspectionIncludeInactiveItems,
			Incremental:     inspectionConfig.Incremental,
			Limit:           inspectionConfig.Limit,
			ExportMedia:     exportMedia,
		},
		&TemplateFeed{
			Incremental: inspectionConfig.Incremental,
		},
		&TemplatePermissionFeed{
			Incremental: inspectionConfig.Incremental,
		},
		&SiteFeed{
			IncludeDeleted: sitesIncludeDeleted,
		},
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
		&ActionFeed{
			ModifiedAfter: inspectionConfig.ModifiedAfter,
			Incremental:   inspectionConfig.Incremental,
			Limit:         actionLimit,
		},
		&ActionAssigneeFeed{
			ModifiedAfter: inspectionConfig.ModifiedAfter,
			Incremental:   inspectionConfig.Incremental,
		},
		&IssueFeed{
			Incremental: false, //this was disabled on request. Issues API doesn't support modified After filters
			Limit:       issueLimit,
		},
	}
}

// GetActionLimit Return 100 (upper bound value) if param value > 100
func GetActionLimit(v *viper.Viper) int {
	actionLimit := v.GetInt("export.action.limit")
	if actionLimit <= 100 {
		return actionLimit
	}
	return 100
}

// GetIssueLimit Return 100 (upper bound value) if param value > 100
func GetIssueLimit(v *viper.Viper) int {
	issueLimit := v.GetInt("export.issue.limit")
	if issueLimit <= 100 {
		return issueLimit
	}
	return 100
}

func getInspectionFeed(v *viper.Viper, inspectionConfig *config.InspectionConfig, templateIDs []string) *InspectionFeed {
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
func ExportFeeds(v *viper.Viper, apiClient *api.Client, exporter Exporter) error {
	logger := util.GetLogger()
	ctx := context.Background()

	var wg sync.WaitGroup
	tables := v.GetStringSlice("export.tables")
	tablesMap := map[string]bool{}
	for _, table := range tables {
		tablesMap[table] = true
	}

	// TODO. Should validate auth before doing anything

	resp, err := apiClient.WhoAmI(ctx)
	util.Check(err, "failed to get details of the current user")

	logger.Infof("Exporting data by user: %s %s", resp.Firstname, resp.Lastname)

	semaphore := make(chan int, maxConcurrentGoRoutines)
	for _, feed := range GetFeeds(v) {
		if tablesMap[feed.Name()] || len(tables) == 0 {
			semaphore <- 1
			wg.Add(1)

			go func(f Feed) {
				logger.Infof(" ... queueing %s\n", f.Name())
				defer wg.Done()
				err := f.Export(ctx, apiClient, exporter, resp.OrganisationID)
				util.Check(err, "failed to export")
				<-semaphore
			}(feed)
		}
	}

	wg.Wait()

	logger.Info("Export finished")

	return nil
}

// ExportInspectionReports download all the reports for inspections and stores them on disk
func ExportInspectionReports(v *viper.Viper, apiClient *api.Client, exporter *ReportExporter) error {
	logger := util.GetLogger()
	ctx := context.Background()

	resp, err := apiClient.WhoAmI(ctx)
	util.Check(err, "failed to get details of the current user")

	logger.Infof("Exporting inspection reports by user: %s %s", resp.Firstname, resp.Lastname)

	feed := getInspectionFeed(v, config.GetInspectionConfig(v), getTemplateIDs(v))
	err = feed.Export(ctx, apiClient, exporter, resp.OrganisationID)
	util.Check(err, "failed to export inspection feed")

	err = exporter.SaveReports(ctx, apiClient, feed)
	if err != nil {
		logger.Info("Export finished")
	}

	return err
}
