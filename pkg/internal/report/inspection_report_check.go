package report

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/version"
	"github.com/pkg/errors"
)

// CheckInspectionReportExportCompletion checks if the report export is complete.
func CheckInspectionReportExportCompletion(ctx context.Context, apiClient *httpapi.Client, auditID string, messageID string) (*InspectionReportExportCompletionResponse, error) {
	var (
		result *InspectionReportExportCompletionResponse
		errMsg json.RawMessage
	)

	url := fmt.Sprintf("audits/%s/report/%s", auditID, messageID)

	sl := apiClient.Sling.New().Get(url).
		Set(string(httpapi.Authorization), apiClient.AuthorizationHeader).
		Set(string(httpapi.IntegrationID), "safetyculture-exporter").
		Set(string(httpapi.IntegrationVersion), version.GetVersion()).
		Set(string(httpapi.XRequestID), util.RequestIDFromContext(ctx))

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

// InspectionReportExportCompletionResponse represents the response of report export completion status
type InspectionReportExportCompletionResponse struct {
	Status string `json:"status"`
	URL    string `json:"url,omitempty"`
}
