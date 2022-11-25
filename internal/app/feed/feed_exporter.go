package feed

import (
	"context"
	"errors"
	"fmt"

	"github.com/SafetyCulture/safetyculture-exporter/internal/app/api"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/config"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/util"
)

/*
*
NOTE: these functions were migrated from various feed methods and adapted not to use viper
They are called directly by the CMD from export cmd.package
*/
const maxConcurrentGoRoutines = 10

// SafetyCultureFeedExporter defines the basic action in regard to the exporter
type SafetyCultureFeedExporter interface {
	// ExportSchemas will generate the schema for SafetyCulture feeds, without downloading data
	ExportSchemas(exporter Exporter) error
	// ExportFeeds will export SafetyCulture feeds and Sheqsy feeds
	ExportFeeds(exporter Exporter) error
	// PrintSchemas is used to print the schema of each feed to console output
	PrintSchemas(exporter *SchemaExporter) error
	// ExportInspectionReports download all the reports for inspections and stores them on disk
	ExportInspectionReports(exporter *ReportExporter) error
}

type ExporterFeedClient struct {
	exportConfig    *config.ExportConfig
	apiClient       *api.Client
	apiConfig       *config.ApiConfig
	sheqsyApiClient *api.Client
	sheqsyApiConfig *config.SheqsyApiConfig
}

// ExportSchemas generates schemas for the data feeds without fetching any data
func (e *ExporterFeedClient) ExportSchemas(exporter Exporter) error {
	var lastErr error = nil
	feeds := e.GetFeeds()
	for _, feed := range feeds {
		lastErr = feed.CreateSchema(exporter)
	}

	return lastErr
}

// ExportFeeds fetches all the feeds data from server and stores them in the format provided
func (e *ExporterFeedClient) ExportFeeds(exporter Exporter) error {
	logger := util.GetLogger()
	ctx := context.Background()

	tables := e.exportConfig.FilterByTableName
	tablesMap := map[string]bool{}
	for _, table := range tables {
		tablesMap[table] = true
	}

	semaphore := make(chan int, maxConcurrentGoRoutines)
	var lastError error = nil
	var errCh chan error
	var doneCh chan string

	atLeastOneRun := false

	// Run export for SafetyCulture data
	if len(e.apiConfig.AccessToken) != 0 {
		atLeastOneRun = true
		logger.Info("exporting SafetyCulture data")

		var feeds []Feed
		for _, feed := range e.GetFeeds() {
			if tablesMap[feed.Name()] || len(tables) == 0 {
				feeds = append(feeds, feed)
			}
		}

		resp, err := e.apiClient.WhoAmI(ctx)
		if err != nil {
			return fmt.Errorf("failed to get details of the current user: %w", err)
		}

		logger.Infof("Exporting data by user: %s %s", resp.Firstname, resp.Lastname)

		if len(feeds) == 0 {
			return errors.New("no tables selected")
		}

		numberOfTasks := len(feeds)
		errCh = make(chan error, numberOfTasks)
		doneCh = make(chan string, numberOfTasks)

		for _, feed := range feeds {
			semaphore <- 1

			go func(f Feed) {
				logger.Infof(" ... queueing %s\n", f.Name())
				err := f.Export(ctx, e.apiClient, exporter, resp.OrganisationID)
				if err != nil {
					errCh <- err
				} else {
					doneCh <- f.Name()
				}
				<-semaphore
			}(feed)
		}

		var jobsProcessed = 0
		for jobsProcessed != numberOfTasks {
			select {
			case exportError := <-errCh:
				logger.Error(exportError)
				lastError = exportError
				jobsProcessed++
			case name := <-doneCh:
				logger.Infof("finished processing feed %s \n", name)
				jobsProcessed++
			}
		}
	}

	// Run export for SHEQSY data
	//if len(e.sheqsyApiConfig.UserName) != 0 {
	//	atLeastOneRun = true
	//	logger.Info("exporting SHEQSY data")
	//
	//	var feeds []Feed
	//	for _, feed := range e.GetSheqsyFeeds() {
	//		if tablesMap[feed.Name()] || len(tables) == 0 {
	//			feeds = append(feeds, feed)
	//		}
	//	}
	//
	//	resp, err := e.sheqsyApiClient.GetSheqsyCompany(ctx, e.sheqsyApiConfig.CompanyID)
	//	util.Check(err, "failed to get details of the current user")
	//
	//	logger.Infof("Exporting data for SHEQSY company: %s %s", resp.Name, resp.CompanyUID)
	//
	//	if len(feeds) == 0 {
	//		return errors.New("no tables selected")
	//	}
	//
	//	for _, feed := range feeds {
	//		semaphore <- 1
	//
	//		go func(f Feed) {
	//			logger.Infof(" ... queueing %s\n", f.Name())
	//			err := f.Export(ctx, e.sheqsyApiClient, exporter, resp.CompanyUID)
	//			util.Check(err, "failed to export")
	//			<-semaphore
	//		}(feed)
	//	}
	//}

	if !atLeastOneRun {
		return errors.New("no API tokens provided")
	}

	logger.Info("Export finished")

	return lastError
}

// GetFeeds returns list of available SafetyCulture feeds
func (e *ExporterFeedClient) GetFeeds() []Feed {
	return []Feed{
		e.getInspectionFeed(),
		&InspectionItemFeed{
			SkipIDs:         e.exportConfig.InspectionConfig.SkipIDs,
			ModifiedAfter:   e.exportConfig.ModifiedAfter,
			TemplateIDs:     e.exportConfig.FilterByTemplateID,
			Archived:        e.exportConfig.InspectionConfig.Archived,
			Completed:       e.exportConfig.InspectionConfig.Completed,
			IncludeInactive: e.exportConfig.InspectionConfig.IncludeInactiveItems,
			Incremental:     e.exportConfig.Incremental,
			Limit:           e.exportConfig.InspectionConfig.BatchLimit,
			ExportMedia:     e.exportConfig.MediaConfig.Export,
		},
		&TemplateFeed{
			Incremental: e.exportConfig.Incremental,
		},
		&TemplatePermissionFeed{
			Incremental: e.exportConfig.Incremental,
		},
		&SiteFeed{
			IncludeDeleted:       e.exportConfig.SiteConfig.IncludeDeleted,
			IncludeFullHierarchy: e.exportConfig.SiteConfig.IncludeFullHierarchy,
		},
		&SiteMemberFeed{},
		&UserFeed{},
		&GroupFeed{},
		&GroupUserFeed{},
		&ScheduleFeed{
			TemplateIDs: e.exportConfig.FilterByTemplateID,
		},
		&ScheduleAssigneeFeed{
			TemplateIDs: e.exportConfig.FilterByTemplateID,
		},
		&ScheduleOccurrenceFeed{
			TemplateIDs: e.exportConfig.FilterByTemplateID,
		},
		&ActionFeed{
			ModifiedAfter: e.exportConfig.ModifiedAfter,
			Incremental:   e.exportConfig.Incremental,
			Limit:         e.exportConfig.ActionConfig.BatchLimit,
		},
		&ActionAssigneeFeed{
			ModifiedAfter: e.exportConfig.ModifiedAfter,
			Incremental:   e.exportConfig.Incremental,
		},
		&IssueFeed{
			Incremental: false, //this was disabled on request. Issues API doesn't support modified After filters
			Limit:       e.exportConfig.ActionConfig.BatchLimit,
		},
	}
}

func (e *ExporterFeedClient) getInspectionFeed() *InspectionFeed {
	return &InspectionFeed{
		SkipIDs:       e.exportConfig.InspectionConfig.SkipIDs,
		ModifiedAfter: e.exportConfig.ModifiedAfter,
		TemplateIDs:   e.exportConfig.FilterByTemplateID,
		Archived:      e.exportConfig.InspectionConfig.Archived,
		Completed:     e.exportConfig.InspectionConfig.Completed,
		Incremental:   e.exportConfig.Incremental,
		Limit:         e.exportConfig.InspectionConfig.BatchLimit,
		WebReportLink: e.exportConfig.InspectionConfig.WebReportLink,
	}
}

// GetSheqsyFeeds returns list of all available data feeds for sheqsy
func (e *ExporterFeedClient) GetSheqsyFeeds() []Feed {
	return []Feed{
		&SheqsyEmployeeFeed{},
		&SheqsyDepartmentEmployeeFeed{},
		&SheqsyDepartmentFeed{},
		&SheqsyActivityFeed{},
		&SheqsyShiftFeed{},
	}
}

// PrintSchemas is used to print the schema of each feed to console output
func (e *ExporterFeedClient) PrintSchemas(exporter *SchemaExporter) error {
	for _, feed := range append(e.GetFeeds(), e.GetSheqsyFeeds()...) {
		err := exporter.CreateSchema(feed, feed.RowsModel())
		util.Check(err, "failed to create schema")

		err = exporter.WriteSchema(feed)
		util.Check(err, "failed to write schema")
	}
	return nil
}

func (e *ExporterFeedClient) ExportInspectionReports(exporter *ReportExporter) error {
	logger := util.GetLogger()
	ctx := context.Background()

	resp, err := e.apiClient.WhoAmI(ctx)
	util.Check(err, "failed to get details of the current user")

	logger.Infof("Exporting inspection reports by user: %s %s", resp.Firstname, resp.Lastname)

	feed := e.getInspectionFeed()
	err = feed.Export(ctx, e.apiClient, exporter, resp.OrganisationID)
	util.Check(err, "failed to export inspection feed")

	err = exporter.SaveReports(ctx, e.apiClient, feed)
	if err != nil {
		logger.Info("Export finished")
	}

	return err
}

func NewExporterApp(scApiClient *api.Client, sheqsyApiClient *api.Client, cfg *config.ConfigurationOptions) *ExporterFeedClient {
	return &ExporterFeedClient{
		exportConfig:    cfg.ExportConfig,
		apiClient:       scApiClient,
		apiConfig:       cfg.ApiConfig,
		sheqsyApiClient: sheqsyApiClient,
		sheqsyApiConfig: cfg.SheqsyApiConfig,
	}
}
