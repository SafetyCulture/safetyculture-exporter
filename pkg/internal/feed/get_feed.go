package feed

import (
	"context"
	"encoding/json"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/events"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/version"
)

// GetFeed executes the feed request and
func GetFeed(ctx context.Context, apiClient *httpapi.Client, request *GetFeedRequest) (*GetFeedResponse, error) {
	var (
		result *GetFeedResponse
		errMsg json.RawMessage
	)

	initialURL := request.InitialURL
	if request.URL != "" {
		initialURL = request.URL
	}

	sl := apiClient.Sling.New().
		Get(initialURL).
		Set(string(httpapi.Authorization), apiClient.AuthorizationHeader).
		Set(string(httpapi.IntegrationID), "safetyculture-exporter").
		Set(string(httpapi.IntegrationVersion), version.GetVersion()).
		Set(string(httpapi.XRequestID), util.RequestIDFromContext(ctx))

	if request.URL == "" {
		sl.QueryStruct(request.Params)
	}

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	httpRes, err := apiClient.Do(&util.SlingHTTPDoer{
		Sl:       sl,
		Req:      req,
		SuccessV: &result,
		FailureV: &errMsg,
	})

	if err != nil {
		return nil, events.NewEventErrorWithMessage(err, events.ErrorSeverityError, events.ErrorSubSystemAPI, false, "api request")
	}

	if httpRes != nil && (httpRes.StatusCode < 200 || httpRes.StatusCode > 299) {
		return nil, util.HTTPError{
			StatusCode: httpRes.StatusCode,
			Resource:   request.InitialURL,
			Message:    string(errMsg),
		}
	}

	return result, nil
}
