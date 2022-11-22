package feed

import (
	"context"
	"errors"
	"sync"

	"github.com/SafetyCulture/safetyculture-exporter/internal/app/api"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/config"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/util"
)

/**
NOTE: these functions were migrated from various feed methods and adapted not to use viper
They are called directly by the CMD from export cmd.package
*/

// SafetyCultureFeedExporter defines the basic action in regard to the exporter
type SafetyCultureFeedExporter interface {
	// CreateSchemas will generate the schema for SafetyCulture feeds, without downloading data
	CreateSchemas(exporter Exporter) error
	// ExportFeeds will export SafetyCulture feeds and Sheqsy feeds
	ExportFeeds(exporter Exporter) error
	// PrintSchemas is used to print the schema of each feed to console output
	PrintSchemas(exporter Exporter) error
}

type ExporterFeedClient struct {
	exportConfig    *config.ExportConfig
	apiClient       *api.Client
	apiConfig       *config.ApiConfig
	sheqsyApiClient *api.Client
	sheqsyApiConfig *config.SheqsyApiConfig
}

// CreateSchemas generates schemas for the data feeds without fetching any data
func (e *ExporterFeedClient) CreateSchemas(exporter Exporter) error {
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

	var wg sync.WaitGroup
	semaphore := make(chan int, maxConcurrentGoRoutines)

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
		util.Check(err, "failed to get details of the current user")

		logger.Infof("Exporting data by user: %s %s", resp.Firstname, resp.Lastname)

		if len(feeds) == 0 {
			return errors.New("no tables selected")
		}

		for _, feed := range feeds {
			semaphore <- 1
			wg.Add(1)

			go func(f Feed) {
				logger.Infof(" ... queueing %s\n", f.Name())
				defer wg.Done()
				err := f.Export(ctx, e.apiClient, exporter, resp.OrganisationID)
				util.Check(err, "failed to export")
				<-semaphore
			}(feed)
		}
	}

	// Run export for SHEQSY data
	if len(e.sheqsyApiConfig.UserName) != 0 {
		atLeastOneRun = true
		logger.Info("exporting SHEQSY data")

		var feeds []Feed
		for _, feed := range GetSheqsyFeeds() {
			if tablesMap[feed.Name()] || len(tables) == 0 {
				feeds = append(feeds, feed)
			}
		}

		resp, err := e.sheqsyApiClient.GetSheqsyCompany(ctx, e.sheqsyApiConfig.CompanyID)
		util.Check(err, "failed to get details of the current user")

		logger.Infof("Exporting data for SHEQSY company: %s %s", resp.Name, resp.CompanyUID)

		if len(feeds) == 0 {
			return errors.New("no tables selected")
		}

		for _, feed := range feeds {
			semaphore <- 1
			wg.Add(1)

			go func(f Feed) {
				logger.Infof(" ... queueing %s\n", f.Name())
				defer wg.Done()
				err := f.Export(ctx, e.sheqsyApiClient, exporter, resp.CompanyUID)
				util.Check(err, "failed to export")
				<-semaphore
			}(feed)
		}
	}

	wg.Wait()

	if !atLeastOneRun {
		return errors.New("no API tokens provided")
	}

	logger.Info("Export finished")

	return nil
}

// GetFeeds returns list of available SafetyCulture feeds
func (e *ExporterFeedClient) GetFeeds() []Feed {
	return []Feed{
		&InspectionFeed{
			SkipIDs:       e.exportConfig.InspectionConfig.SkipIDs,
			ModifiedAfter: e.exportConfig.ModifiedAfter,
			TemplateIDs:   e.exportConfig.FilterByTemplateID,
			Archived:      e.exportConfig.InspectionConfig.Archived,
			Completed:     e.exportConfig.InspectionConfig.Completed,
			Incremental:   e.exportConfig.Incremental,
			Limit:         e.exportConfig.InspectionConfig.BatchLimit,
			WebReportLink: e.exportConfig.InspectionConfig.WebReportLink,
		},
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

// PrintSchemas is used to print the schema of each feed to console output
func (e *ExporterFeedClient) PrintSchemas(exporter *SchemaExporter) error {
	for _, feed := range append(e.GetFeeds(), GetSheqsyFeeds()...) {
		err := exporter.CreateSchema(feed, feed.RowsModel())
		util.Check(err, "failed to create schema")

		err = exporter.WriteSchema(feed)
		util.Check(err, "failed to write schema")
	}
	return nil
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
