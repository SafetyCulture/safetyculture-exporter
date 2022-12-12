package inspections

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/version"
	"github.com/pkg/errors"
)

// ListInspections retrieves the list of inspections from SafetyCulture
func ListInspections(ctx context.Context, apiClient *httpapi.Client, params *ListInspectionsParams) (*ListInspectionsResponse, error) {
	var (
		result *ListInspectionsResponse
		errMsg json.RawMessage
	)

	sl := apiClient.Sling.New().Get("/audits/search").
		Set(string(httpapi.Authorization), apiClient.AuthorizationHeader).
		Set(string(httpapi.IntegrationID), "safetyculture-exporter").
		Set(string(httpapi.IntegrationVersion), version.GetVersion()).
		Set(string(httpapi.XRequestID), util.RequestIDFromContext(ctx))

	sl.QueryStruct(params)
	req, _ := sl.Request()
	req = req.WithContext(ctx)

	_, err := apiClient.Do(&util.SlingHTTPDoer{
		Sl:       sl,
		Req:      req,
		SuccessV: &result,
		FailureV: &errMsg,
	})
	if err != nil {
		return nil, errors.Wrap(err, "api request")
	}

	return result, nil
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
