package httpapi

import "time"

// WhoAmIResponse represents the response of  WhoAmI
type WhoAmIResponse struct {
	UserID         string `json:"user_id"`
	OrganisationID string `json:"organisation_id"`
	Firstname      string `json:"firstname"`
	Lastname       string `json:"lastname"`
}

type GetSheqsyCompanyResponse struct {
	CompanyID   int         `json:"companyId"`
	CompanyName string      `json:"companyName"`
	Name        interface{} `json:"name"`
	CompanyUID  string      `json:"companyUId"`
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

// ListInspectionsParams is a list of all parameters we can set when fetching inspections
type ListInspectionsParams struct {
	ModifiedAfter time.Time `url:"modified_after,omitempty"`
	TemplateIDs   []string  `url:"template,omitempty"`
	Archived      string    `url:"archived,omitempty"`
	Completed     string    `url:"completed,omitempty"`
	Limit         int       `url:"limit,omitempty"`
}

// InspectionReportExportCompletionResponse represents the response of report export completion status
type InspectionReportExportCompletionResponse struct {
	Status string `json:"status"`
	URL    string `json:"url,omitempty"`
}

type activityResponse struct {
	Type     string            `json:"type"`
	Metadata map[string]string `json:"metadata"`
}

// GetAccountsActivityLogResponse is the response from activity log history api
type GetAccountsActivityLogResponse struct {
	Activities    []activityResponse
	NextPageToken string `json:"next_page_token"`
}

// GetAccountsActivityLogRequestParams contains fields required to make a post request to activity log history api
type GetAccountsActivityLogRequestParams struct {
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

// NewGetAccountsActivityLogRequest build a request for AccountsActivityLog
// for now it serves the purposes only for inspection.deleted. If we need later, we can change this builder
func NewGetAccountsActivityLogRequest(pageSize int, from time.Time) *GetAccountsActivityLogRequestParams {
	return &GetAccountsActivityLogRequestParams{
		PageSize: pageSize,
		Filters: accountsActivityLogFilter{
			Timeframe: timeFrame{
				From: from,
			},
			Limit:      pageSize,
			EventTypes: []string{"inspection.deleted"},
		},
	}
}
