package api

import (
	"time"
)

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
