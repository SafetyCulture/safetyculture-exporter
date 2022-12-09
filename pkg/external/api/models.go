package api

import (
	"time"
)

type ReportExporterCfg struct {
	Format       []string
	PreferenceID string
	Filename     string
	RetryTimeout int
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

// Inspection represents some properties present in an inspection
type Inspection struct {
	ID         string    `json:"audit_id"`
	ModifiedAt time.Time `json:"modified_at"`
}
