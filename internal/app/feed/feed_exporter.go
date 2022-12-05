package feed

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/SafetyCulture/safetyculture-exporter/internal/app/api"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/config"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/events"
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
	configuration   *config.ExporterConfiguration
	apiClient       *api.Client
	sheqsyApiClient *api.Client
	errMu           sync.Mutex
	errs            []error
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

func (e *ExporterFeedClient) addError(err error) {
	e.errMu.Lock()
	e.errs = append(e.errs, err)
	e.errMu.Unlock()
}

// ExportFeeds fetches all the feeds data from server and stores them in the format provided
func (e *ExporterFeedClient) ExportFeeds(exporter Exporter) error {
	logger := util.GetLogger()
	ctx := context.Background()

	tables := e.configuration.Export.Tables
	tablesMap := map[string]bool{}
	for _, table := range tables {
		tablesMap[table] = true
	}

	var wg sync.WaitGroup
	semaphore := make(chan int, maxConcurrentGoRoutines)

	atLeastOneRun := false

	// Run export for SafetyCulture data
	if len(e.configuration.AccessToken) != 0 {
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
			return fmt.Errorf("get details of the current user: %w", err)
		}

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
				if err != nil {
					logger.Errorf("exporting feeds: %v", err)
					e.addError(err)
				}
				<-semaphore
			}(feed)
		}

	}

	// Run export for SHEQSY data
	if len(e.configuration.SheqsyUsername) != 0 {
		atLeastOneRun = true
		logger.Info("exporting SHEQSY data")

		var feeds []Feed
		for _, feed := range e.GetSheqsyFeeds() {
			if tablesMap[feed.Name()] || len(tables) == 0 {
				feeds = append(feeds, feed)
			}
		}

		resp, err := e.sheqsyApiClient.GetSheqsyCompany(ctx, e.configuration.SheqsyCompanyID)
		if err != nil {
			return fmt.Errorf("get details of the current user: %w", err)
		}

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
				if err != nil {
					e.addError(err)
				}
				<-semaphore
			}(feed)
		}
	}

	wg.Wait()

	if !atLeastOneRun {
		return errors.New("no API tokens provided")
	}

	logger.Info("Export finished")
	if len(e.errs) != 0 {
		logger.Warn("There were errors during the export:")
		for _, ee := range e.errs {
			switch theError := ee.(type) {
			case *events.EventError:
				theError.Log(logger)
			default:
				logger.Infof(" > %s", theError.Error())
			}

		}
		// this is temporary code until we finish a follow-up ticket that will use structured errors
		return e.errs[0]
	}

	return nil
}

// GetFeeds returns list of available SafetyCulture feeds
func (e *ExporterFeedClient) GetFeeds() []Feed {
	return []Feed{
		e.getInspectionFeed(),
		&InspectionItemFeed{
			SkipIDs:         e.configuration.Export.Inspection.SkipIds,
			ModifiedAfter:   e.configuration.Export.ModifiedAfter.Time,
			TemplateIDs:     e.configuration.Export.TemplateIds,
			Archived:        e.configuration.Export.Inspection.Archived,
			Completed:       e.configuration.Export.Inspection.Completed,
			IncludeInactive: e.configuration.Export.Inspection.IncludedInactiveItems,
			Incremental:     e.configuration.Export.Incremental,
			Limit:           e.configuration.Export.Inspection.Limit,
			ExportMedia:     e.configuration.Export.Media,
		},
		&TemplateFeed{
			Incremental: e.configuration.Export.Incremental,
		},
		&TemplatePermissionFeed{
			Incremental: e.configuration.Export.Incremental,
		},
		&SiteFeed{
			IncludeDeleted:       e.configuration.Export.Site.IncludeDeleted,
			IncludeFullHierarchy: e.configuration.Export.Site.IncludeFullHierarchy,
		},
		&SiteMemberFeed{},
		&UserFeed{},
		&GroupFeed{},
		&GroupUserFeed{},
		&ScheduleFeed{
			TemplateIDs: e.configuration.Export.TemplateIds,
		},
		&ScheduleAssigneeFeed{
			TemplateIDs: e.configuration.Export.TemplateIds,
		},
		&ScheduleOccurrenceFeed{
			TemplateIDs: e.configuration.Export.TemplateIds,
		},
		&ActionFeed{
			ModifiedAfter: e.configuration.Export.ModifiedAfter.Time,
			Incremental:   e.configuration.Export.Incremental,
			Limit:         e.configuration.Export.Action.Limit,
		},
		&ActionAssigneeFeed{
			ModifiedAfter: e.configuration.Export.ModifiedAfter.Time,
			Incremental:   e.configuration.Export.Incremental,
		},
		&IssueFeed{
			Incremental: false, //this was disabled on request. Issues API doesn't support modified After filters
			Limit:       e.configuration.Export.Issue.Limit,
		},
		&AssetFeed{
			Incremental: false, // Assets API doesn't support modified after filters
			Limit:       e.configuration.Export.Asset.Limit,
		},
	}
}

func (e *ExporterFeedClient) getInspectionFeed() *InspectionFeed {
	return &InspectionFeed{
		SkipIDs:       e.configuration.Export.Inspection.SkipIds,
		ModifiedAfter: e.configuration.Export.ModifiedAfter.Time,
		TemplateIDs:   e.configuration.Export.TemplateIds,
		Archived:      e.configuration.Export.Inspection.Archived,
		Completed:     e.configuration.Export.Inspection.Completed,
		Incremental:   e.configuration.Export.Incremental,
		Limit:         e.configuration.Export.Inspection.Limit,
		WebReportLink: e.configuration.Export.Inspection.WebReportLink,
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
		if err := exporter.CreateSchema(feed, feed.RowsModel()); err != nil {
			return fmt.Errorf("create schema: %w", err)
		}

		if err := exporter.WriteSchema(feed); err != nil {
			return fmt.Errorf("write schema: %w", err)
		}
	}
	return nil
}

func (e *ExporterFeedClient) ExportInspectionReports(exporter *ReportExporter) error {
	logger := util.GetLogger()
	ctx := context.Background()

	resp, err := e.apiClient.WhoAmI(ctx)
	if err != nil {
		return fmt.Errorf("get details of the current user: %w", err)
	}

	logger.Infof("Exporting inspection reports by user: %s %s", resp.Firstname, resp.Lastname)

	feed := e.getInspectionFeed()
	if err := feed.Export(ctx, e.apiClient, exporter, resp.OrganisationID); err != nil {
		return fmt.Errorf("export inspection feed: %w", err)
	}

	err = exporter.SaveReports(ctx, e.apiClient, feed)
	if err != nil {
		return fmt.Errorf("save reports: %w", err)
	}

	return err
}

func NewExporterApp(scApiClient *api.Client, sheqsyApiClient *api.Client, cfg *config.ExporterConfiguration) *ExporterFeedClient {
	return &ExporterFeedClient{
		configuration:   cfg,
		apiClient:       scApiClient,
		sheqsyApiClient: sheqsyApiClient,
	}
}
