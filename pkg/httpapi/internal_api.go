package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
)

// WhoAmI returns the details for the user who is making the request
func WhoAmI(ctx context.Context, apiClient *Client) (*WhoAmIResponse, error) {
	return ExecuteGet[WhoAmIResponse](ctx, apiClient, "accounts/user/v1/user:WhoAmI", nil)
}

// GetSheqsyCompany returns the details for the selected company
func GetSheqsyCompany(ctx context.Context, apiClient *Client, companyID string) (*GetSheqsyCompanyResponse, error) {
	return ExecuteGet[GetSheqsyCompanyResponse](ctx, apiClient, fmt.Sprintf("/SheqsyIntegrationApi/api/v3/companies/%s", companyID), nil)
}

// GetRawInspection returns the JSON Raw Message of an inspection response
func GetRawInspection(ctx context.Context, apiClient *Client, id string) (*json.RawMessage, error) {
	return ExecuteGet[json.RawMessage](ctx, apiClient, fmt.Sprintf("/audits/%s", id), nil)
}

// ListInspections retrieves the list of inspections from SafetyCulture
func ListInspections(ctx context.Context, apiClient *Client, params *ListInspectionsParams) (*ListInspectionsResponse, error) {
	return ExecuteGet[ListInspectionsResponse](ctx, apiClient, "/audits/search", params)
}
