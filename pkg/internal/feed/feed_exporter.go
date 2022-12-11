package feed

import (
	"context"
	"errors"
	"fmt"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
	"sync"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/events"
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
	configuration   *ExporterFeedCfg
	apiClient       *httpapi.Client
	sheqsyApiClient *httpapi.Client
	errMu           sync.Mutex
	errs            []error
}

type ExporterFeedCfg struct {
	AccessToken                           string
	ExportTables                          []string
	SheqsyUsername                        string
	SheqsyCompanyID                       string
	ExportInspectionSkipIds               []string
	ExportModifiedAfterTime               time.Time
	ExportTemplateIds                     []string
	ExportInspectionArchived              string
	ExportInspectionCompleted             string
	ExportInspectionIncludedInactiveItems bool
	ExportInspectionWebReportLink         string
	ExportIncremental                     bool
	ExportInspectionLimit                 int
	ExportMedia                           bool
	ExportSiteIncludeDeleted              bool
	ExportActionLimit                     int
	ExportSiteIncludeFullHierarchy        bool
	ExportIssueLimit                      int
	ExportAssetLimit                      int
}

func NewExporterApp(scApiClient *httpapi.Client, sheqsyApiClient *httpapi.Client, cfg *ExporterFeedCfg) *ExporterFeedClient {
	return &ExporterFeedClient{
		configuration:   cfg,
		apiClient:       scApiClient,
		sheqsyApiClient: sheqsyApiClient,
	}
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
	logger := logger.GetLogger()
	ctx := context.Background()

	tables := e.configuration.ExportTables
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

		resp, err := GetSheqsyCompany(ctx, e.sheqsyApiClient, e.configuration.SheqsyCompanyID)
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
			SkipIDs:         e.configuration.ExportInspectionSkipIds,
			ModifiedAfter:   e.configuration.ExportModifiedAfterTime,
			TemplateIDs:     e.configuration.ExportTemplateIds,
			Archived:        e.configuration.ExportInspectionArchived,
			Completed:       e.configuration.ExportInspectionCompleted,
			IncludeInactive: e.configuration.ExportInspectionIncludedInactiveItems,
			Incremental:     e.configuration.ExportIncremental,
			Limit:           e.configuration.ExportInspectionLimit,
			ExportMedia:     e.configuration.ExportMedia,
		},
		&TemplateFeed{
			Incremental: e.configuration.ExportIncremental,
		},
		&TemplatePermissionFeed{
			Incremental: e.configuration.ExportIncremental,
		},
		&SiteFeed{
			IncludeDeleted:       e.configuration.ExportSiteIncludeDeleted,
			IncludeFullHierarchy: e.configuration.ExportSiteIncludeFullHierarchy,
		},
		&SiteMemberFeed{},
		&UserFeed{},
		&GroupFeed{},
		&GroupUserFeed{},
		&ScheduleFeed{
			TemplateIDs: e.configuration.ExportTemplateIds,
		},
		&ScheduleAssigneeFeed{
			TemplateIDs: e.configuration.ExportTemplateIds,
		},
		&ScheduleOccurrenceFeed{
			TemplateIDs: e.configuration.ExportTemplateIds,
		},
		&ActionFeed{
			ModifiedAfter: e.configuration.ExportModifiedAfterTime,
			Incremental:   e.configuration.ExportIncremental,
			Limit:         e.configuration.ExportActionLimit,
		},
		&ActionAssigneeFeed{
			ModifiedAfter: e.configuration.ExportModifiedAfterTime,
			Incremental:   e.configuration.ExportIncremental,
		},
		&IssueFeed{
			Incremental: false, // this was disabled on request. Issues API doesn't support modified After filters
			Limit:       e.configuration.ExportIssueLimit,
		},
		&AssetFeed{
			Incremental: false, // Assets API doesn't support modified after filters
			Limit:       e.configuration.ExportAssetLimit,
		},
	}
}

func (e *ExporterFeedClient) getInspectionFeed() *InspectionFeed {
	return &InspectionFeed{
		SkipIDs:       e.configuration.ExportInspectionSkipIds,
		ModifiedAfter: e.configuration.ExportModifiedAfterTime,
		TemplateIDs:   e.configuration.ExportTemplateIds,
		Archived:      e.configuration.ExportInspectionArchived,
		Completed:     e.configuration.ExportInspectionCompleted,
		Incremental:   e.configuration.ExportIncremental,
		Limit:         e.configuration.ExportInspectionLimit,
		WebReportLink: e.configuration.ExportInspectionWebReportLink,
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
	logger := logger.GetLogger()
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
