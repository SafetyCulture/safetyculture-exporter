package api

import (
	"encoding/json"
	"time"
)

// NewGetAccountsActivityLogRequest build a request for AccountsActivityLog
// for now it serves the purposes only for inspection.deleted. If we need later, we can change this builder
func NewGetAccountsActivityLogRequest(pageSize int, from time.Time) *GetAccountsActivityLogRequest {
	return &GetAccountsActivityLogRequest{
		URL: "/accounts/history/v1/activity_log/list",
		Params: accountsActivityLogRequestParams{
			PageSize: pageSize,
			Filters: accountsActivityLogFilter{
				Timeframe: timeFrame{
					From: from,
				},
				Limit:      pageSize,
				EventTypes: []string{"inspection.deleted"},
			},
		},
	}
}

// FeedMetadata is a representation of the metadata returned when fetching a feed
type FeedMetadata struct {
	NextPage         string `json:"next_page"`
	RemainingRecords int64  `json:"remaining_records"`
}

// GetFeedParams is a list of all parameters we can set when fetching a feed
type GetFeedParams struct {
	ModifiedAfter   time.Time `url:"modified_after,omitempty"`
	TemplateIDs     []string  `url:"template,omitempty"`
	Archived        string    `url:"archived,omitempty"`
	Completed       string    `url:"completed,omitempty"`
	IncludeInactive bool      `url:"include_inactive,omitempty"`
	Limit           int       `url:"limit,omitempty"`
	WebReportLink   string    `url:"web_report_link,omitempty"`

	// Applicable only for sites
	IncludeDeleted    bool  `url:"include_deleted,omitempty"`
	ShowOnlyLeafNodes *bool `url:"show_only_leaf_nodes,omitempty"`
}

// GetFeedRequest has all the data needed to make a request to get a feed
type GetFeedRequest struct {
	URL        string
	InitialURL string
	Params     GetFeedParams
}

// GetFeedResponse is a representation of the data returned when fetching a feed
type GetFeedResponse struct {
	Metadata FeedMetadata `json:"metadata"`

	Data json.RawMessage `json:"data"`
}

// GetAccountsActivityLogRequest contains fields required to make a post request to activity log history api
type GetAccountsActivityLogRequest struct {
	URL    string
	Params accountsActivityLogRequestParams
}

// accountsActivityLogRequestParams params used for POST request of AccountsActivityLog
type accountsActivityLogRequestParams struct {
	OrgID     string                    `json:"org_id"`
	PageSize  int                       `json:"page_size"`
	PageToken string                    `json:"page_token"`
	Filters   accountsActivityLogFilter `json:"filters"`
}

// accountsActivityLogFilter filter for AccountsActivityLog
type accountsActivityLogFilter struct {
	Timeframe  timeFrame `json:"timeframe,omitempty"`
	EventTypes []string  `json:"event_types"`
	Limit      int       `json:"limit"`
}

type timeFrame struct {
	From time.Time `json:"from"`
}

// GetAccountsActivityLogResponse is the response from activity log history api
type GetAccountsActivityLogResponse struct {
	Activities    []activityResponse
	NextPageToken string `json:"next_page_token"`
}

type activityResponse struct {
	Type     string            `json:"type"`
	Metadata map[string]string `json:"metadata"`
}

// GetMediaRequest has all the data needed to make a request to get a media
type GetMediaRequest struct {
	URL     string
	AuditID string
}

// GetMediaResponse is a representation of the data returned when fetching media
type GetMediaResponse struct {
	ContentType string
	Body        []byte
	MediaID     string
}

// ListInspectionsParams is a list of all parameters we can set when fetching inspections
type ListInspectionsParams struct {
	ModifiedAfter time.Time `url:"modified_after,omitempty"`
	TemplateIDs   []string  `url:"template,omitempty"`
	Archived      string    `url:"archived,omitempty"`
	Completed     string    `url:"completed,omitempty"`
	Limit         int       `url:"limit,omitempty"`
}

// Inspection represents some properties present in an inspection
type Inspection struct {
	ID         string    `json:"audit_id"`
	ModifiedAt time.Time `json:"modified_at"`
}

// ListInspectionsResponse represents the response of listing inspections
type ListInspectionsResponse struct {
	Count       int          `json:"count"`
	Total       int          `json:"total"`
	Inspections []Inspection `json:"audits"`
}

// WhoAmIResponse represents the response of  WhoAmI
type WhoAmIResponse struct {
	UserID         string `json:"user_id"`
	OrganisationID string `json:"organisation_id"`
	Firstname      string `json:"firstname"`
	Lastname       string `json:"lastname"`
}

// InspectionReportExportCompletionResponse represents the response of report export completion status
type InspectionReportExportCompletionResponse struct {
	Status string `json:"status"`
	URL    string `json:"url,omitempty"`
}

type initiateInspectionReportExportRequest struct {
	Format       string `json:"format"`
	PreferenceID string `json:"preference_id,omitempty"`
}

type initiateInspectionReportExportResponse struct {
	MessageID string `json:"messageId"`
}
