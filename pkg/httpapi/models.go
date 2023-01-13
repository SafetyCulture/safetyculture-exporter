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
