package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/events"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
)

func ExecuteGet[T any](ctx context.Context, apiClient *Client, url string, params any) (*T, error) {
	sl := apiClient.Sling.New().Get(url).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	if params != nil {
		sl.QueryStruct(params)
	}

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	var res = new(T)
	var errMsg json.RawMessage

	httpRes, err := apiClient.Do(&util.SlingHTTPDoer{
		Sl:       sl,
		Req:      req,
		SuccessV: &res,
		FailureV: &errMsg,
	})

	if err != nil {
		return nil, events.NewEventErrorWithMessage(err, events.ErrorSeverityError, events.ErrorSubSystemAPI, false, "api request")
	}

	if httpRes != nil && (httpRes.StatusCode < 200 || httpRes.StatusCode > 299) {
		return nil, util.HTTPError{
			StatusCode: httpRes.StatusCode,
			Resource:   url,
			Message:    string(errMsg),
		}
	}

	return res, nil
}

func ExecutePost[T any](ctx context.Context, apiClient *Client, url string, body any) (*T, error) {
	sl := apiClient.Sling.New().Post(url).
		Set(string(XRequestID), util.RequestIDFromContext(ctx)).
		BodyJSON(body)

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	var res = new(T)
	var errMsg json.RawMessage

	httpRes, err := apiClient.Do(&util.SlingHTTPDoer{
		Sl:       sl,
		Req:      req,
		SuccessV: &res,
		FailureV: &errMsg,
	})

	if err != nil {
		return nil, events.NewEventErrorWithMessage(err, events.ErrorSeverityError, events.ErrorSubSystemAPI, false, "api request")
	}

	if httpRes != nil && (httpRes.StatusCode < 200 || httpRes.StatusCode > 299) {
		return nil, util.HTTPError{
			StatusCode: httpRes.StatusCode,
			Resource:   url,
			Message:    string(errMsg),
		}
	}

	return res, nil
}

func ExecuteRawGet(ctx context.Context, apiClient *Client, url string) (*http.Response, error) {
	sl := apiClient.Sling.New().Get(url).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	httpRes, err := apiClient.Do(&util.DefaultHTTPDoer{
		Req:        req,
		HttpClient: apiClient.HttpClient,
	})
	if err != nil {
		return nil, err
	}

	return httpRes, nil
}
