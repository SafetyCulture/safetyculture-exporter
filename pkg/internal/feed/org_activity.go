package feed

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/events"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/version"
)

const activityHistoryLogURL = "/accounts/history/v1/activity_log/list"

// ListOrganisationActivityLog returns response from AccountsActivityLog or error
func ListOrganisationActivityLog(ctx context.Context, apiClient *httpapi.Client, request *GetAccountsActivityLogRequestParams) (*GetAccountsActivityLogResponse, error) {
	sl := apiClient.Sling.New().
		Post(activityHistoryLogURL).
		Set(string(httpapi.Authorization), apiClient.AuthorizationHeader).
		Set(string(httpapi.IntegrationID), "safetyculture-exporter").
		Set(string(httpapi.IntegrationVersion), version.GetVersion()).
		Set(string(httpapi.XRequestID), util.RequestIDFromContext(ctx)).
		BodyJSON(request)

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	var res GetAccountsActivityLogResponse
	var errMsg json.RawMessage
	_, err := apiClient.Do(&util.SlingHTTPDoer{
		Sl:       sl,
		Req:      req,
		SuccessV: &res,
		FailureV: &errMsg,
	})
	if err != nil {
		return nil, events.NewEventErrorWithMessage(err, events.ErrorSeverityError, events.ErrorSubSystemAPI, false, "api request")
	}

	return &res, nil
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

// GetAccountsActivityLogRequestParams contains fields required to make a post request to activity log history api
type GetAccountsActivityLogRequestParams struct {
	OrgID     string                    `json:"org_id"`
	PageSize  int                       `json:"page_size"`
	PageToken string                    `json:"page_token"`
	Filters   accountsActivityLogFilter `json:"filters"`
}

// GetAccountsActivityLogResponse is the response from activity log history api
type GetAccountsActivityLogResponse struct {
	Activities    []activityResponse
	NextPageToken string `json:"next_page_token"`
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

type activityResponse struct {
	Type     string            `json:"type"`
	Metadata map[string]string `json:"metadata"`
}
