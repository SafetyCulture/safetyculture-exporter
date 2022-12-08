package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/external/version"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/events"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/dghubble/sling"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	// Default retry configuration
	defaultRetryWaitMin = 1 * time.Second
	defaultRetryWaitMax = 30 * time.Second
	defaultRetryMax     = 4
)

const activityHistoryLogURL = "/accounts/history/v1/activity_log/list"

// Client is used to with SafetyCulture API's.
type Client struct {
	logger              *zap.SugaredLogger
	authorizationHeader string
	baseURL             string
	sling               *sling.Sling
	httpClient          *http.Client
	httpTransport       *http.Transport

	Duration      time.Duration
	CheckForRetry CheckForRetry
	Backoff       Backoff
	RetryMax      int
	RetryWaitMin  time.Duration
	RetryWaitMax  time.Duration
}

// Opt is an option to configure the Client
type Opt func(*Client)

// NewClient creates a new instance of the Client
func NewClient(addr string, authorizationHeader string, opts ...Opt) *Client {
	httpTransport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
	}

	httpClient := &http.Client{
		Timeout:   120 * time.Second,
		Transport: httpTransport,
	}

	a := &Client{
		logger:              util.GetLogger(),
		httpClient:          httpClient,
		baseURL:             addr,
		httpTransport:       httpTransport,
		sling:               sling.New().Client(httpClient).Base(addr),
		authorizationHeader: authorizationHeader,
		Duration:            0,
		CheckForRetry:       DefaultRetryPolicy,
		Backoff:             DefaultBackoff,
		RetryMax:            defaultRetryMax,
		RetryWaitMin:        defaultRetryWaitMin,
		RetryWaitMax:        defaultRetryWaitMax,
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

// HTTPClient returns the http Client used by APIClient
func (a *Client) HTTPClient() *http.Client {
	return a.httpClient
}

// HTTPTransport returns the http Transport used by APIClient
func (a *Client) HTTPTransport() *http.Transport {
	return a.httpTransport
}

// OptSetTimeout sets the timeout for the request
func OptSetTimeout(t time.Duration) Opt {
	return func(a *Client) {
		a.httpClient.Timeout = t
	}
}

// OptSetProxy sets the proxy URL to use for API requests
func OptSetProxy(proxyURL *url.URL) Opt {
	return func(a *Client) {
		a.httpTransport.Proxy = http.ProxyURL(proxyURL)
	}
}

// OptSetInsecureTLS sets whether TLS certs should be verified
func OptSetInsecureTLS(insecureSkipVerify bool) Opt {
	return func(a *Client) {
		if a.httpTransport.TLSClientConfig == nil {
			a.httpTransport.TLSClientConfig = &tls.Config{}
		}

		a.httpTransport.TLSClientConfig.InsecureSkipVerify = insecureSkipVerify
	}
}

// OptAddTLSCert adds a certificate at the supplied path to the cert pool
func OptAddTLSCert(certPath string) Opt {
	return func(a *Client) {
		if certPath == "" {
			return
		}

		if a.httpTransport.TLSClientConfig == nil {
			a.httpTransport.TLSClientConfig = &tls.Config{}
		}

		// Get the SystemCertPool, continue with an empty pool on error
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		// Read in the cert file
		certs, err := os.ReadFile(certPath)
		if err != nil {
			log.Fatalf("Failed to append %q to RootCAs: %v", certPath, err)
		}

		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			log.Println("No certs appended, using system certs only")
		}

		a.httpTransport.TLSClientConfig.RootCAs = rootCAs
	}
}

func (a *Client) do(doer HTTPDoer) (*http.Response, error) {
	url := doer.URL()

	for iter := 0; ; iter++ {
		a.logger.Debugw("http request", "url", url)

		start := time.Now()
		resp, err := doer.Do()
		a.Duration = time.Since(start)

		status := ""
		if resp != nil {
			status = resp.Status
		}

		if err != nil {
			a.logger.Errorw("http request error", "url", url, "status", status, "err", err)
		}

		a.logger.Debugw("http response", "url", url, "status", status)

		// Check if we should continue with the retries
		shouldRetry, _ := a.CheckForRetry(resp, err)
		if !shouldRetry {
			if resp != nil {
				switch status := resp.StatusCode; {
				case status >= 200 && status <= 299:
					return resp, nil

				case status == http.StatusNotFound:
					a.logger.Errorw("http request error status",
						"url", url,
						"status", status,
					)
					return resp, nil

				case status == http.StatusForbidden:
					a.logger.Errorw("no access to this resource", "url", url, "status", status)
					return resp, nil

				default:
					a.logger.Errorw("http request error status",
						"url", url,
						"status", status,
						"err", doer.Error(),
					)
					return resp, errors.Errorf("request error status: %d", status)
				}
			}
			return resp, err
		}

		remain := a.RetryMax - iter
		if remain == 0 {
			break
		}

		wait := a.Backoff(a.RetryWaitMin, a.RetryWaitMax, iter, resp)
		a.logger.Infof("retrying URL %s after %v", url, wait)

		time.Sleep(wait)
	}

	return nil, fmt.Errorf("%s giving up after %d attempt(s)", url, a.RetryMax+1)
}

// GetMedia fetches the media object from SafetyCulture.
func (a *Client) GetMedia(ctx context.Context, request *GetMediaRequest) (*GetMediaResponse, error) {
	baseURL := strings.TrimPrefix(request.URL, a.baseURL)

	// The mediaURL will be in the following format:
	// https://api.eu.safetyculture.com/audits/audit_xxx/media/4c83fcf2-180b-4d3e-958f-389f7ac49777
	// The string that is after the word "media/" is the ID of it.
	mediaIDURL := strings.Split(request.URL, "/")
	mediaID := mediaIDURL[len(mediaIDURL)-1]

	sl := a.sling.New().Get(baseURL).
		Set(string(Authorization), a.authorizationHeader).
		Set(string(IntegrationID), "safetyculture-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	result, err := a.do(&defaultHTTPDoer{
		req:        req,
		httpClient: a.httpClient,
	})
	if err != nil {
		// Ignore forbidden errors for media objects.
		if result != nil && result.StatusCode == http.StatusForbidden {
			return nil, nil
		}
		return nil, err
	}
	defer result.Body.Close()

	if result.StatusCode == 204 {
		return nil, nil
	}

	contentType := result.Header.Get("Content-Type")
	if contentType == "" {
		return nil, fmt.Errorf("failed to get content-type of media")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)

	resp := &GetMediaResponse{
		ContentType: contentType,
		Body:        buf.Bytes(),
		MediaID:     mediaID,
	}
	return resp, nil
}

// GetFeed executes the feed request and
func (a *Client) GetFeed(ctx context.Context, request *GetFeedRequest) (*GetFeedResponse, error) {
	var (
		result *GetFeedResponse
		errMsg json.RawMessage
	)

	initialURL := request.InitialURL
	if request.URL != "" {
		initialURL = request.URL
	}

	sl := a.sling.New().
		Get(initialURL).
		Set(string(Authorization), a.authorizationHeader).
		Set(string(IntegrationID), "safetyculture-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	if request.URL == "" {
		sl.QueryStruct(request.Params)
	}

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	httpRes, err := a.do(&slingHTTPDoer{
		sl:       sl,
		req:      req,
		successV: &result,
		failureV: &errMsg,
	})

	if err != nil {
		return nil, events.NewEventErrorWithMessage(err, events.ErrorSeverityError, events.ErrorSubSystemAPI, false, "api request")
	}

	if httpRes != nil && (httpRes.StatusCode < 200 || httpRes.StatusCode > 299) {
		//TODO?
		return nil, util.HTTPError{
			StatusCode: httpRes.StatusCode,
			Resource:   request.InitialURL,
			Message:    string(errMsg),
		}
	}

	return result, nil
}

// DrainFeed fetches the data in batches and triggers the callback for each batch.
func (a *Client) DrainFeed(ctx context.Context, request *GetFeedRequest, feedFn func(*GetFeedResponse) error) error {
	var nextURL string
	// Used to both ensure the fetchFn is called at least once
	first := true
	for nextURL != "" || first {
		first = false
		request.URL = nextURL
		resp, httpErr := a.GetFeed(ctx, request)
		if httpErr != nil {
			return events.NewEventError(httpErr, events.ErrorSeverityError, events.ErrorSubSystemAPI, false)
		}
		nextURL = resp.Metadata.NextPage

		err := feedFn(resp)
		if err != nil {
			return events.NewEventError(err, events.ErrorSeverityError, events.ErrorSubSystemAPI, false)
		}
	}

	return nil
}

// ListOrganisationActivityLog returns response from AccountsActivityLog or error
func (a *Client) ListOrganisationActivityLog(ctx context.Context, request *GetAccountsActivityLogRequestParams) (*GetAccountsActivityLogResponse, error) {
	sl := a.sling.New().
		Post(activityHistoryLogURL).
		Set(string(Authorization), a.authorizationHeader).
		Set(string(IntegrationID), "safetyculture-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx)).
		BodyJSON(request)

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	var res GetAccountsActivityLogResponse
	var errMsg json.RawMessage
	_, err := a.do(&slingHTTPDoer{
		sl:       sl,
		req:      req,
		successV: &res,
		failureV: &errMsg,
	})
	if err != nil {
		return nil, events.NewEventErrorWithMessage(err, events.ErrorSeverityError, events.ErrorSubSystemAPI, false, "api request")
	}

	return &res, nil
}

// DrainAccountActivityHistoryLog cycle through GetAccountsActivityLogResponse and adapts the filter while there is a next page
func (a *Client) DrainAccountActivityHistoryLog(ctx context.Context, req *GetAccountsActivityLogRequestParams, feedFn func(*GetAccountsActivityLogResponse) error) error {
	for {
		res, err := a.ListOrganisationActivityLog(ctx, req)
		if err != nil {
			return err
		}

		err = feedFn(res)
		if err != nil {
			return err
		}

		if res.NextPageToken != "" {
			req.PageToken = res.NextPageToken
		} else {
			break
		}
	}
	return nil
}

// ListInspections retrieves the list of inspections from SafetyCulture
func (a *Client) ListInspections(ctx context.Context, params *ListInspectionsParams) (*ListInspectionsResponse, error) {
	var (
		result *ListInspectionsResponse
		errMsg json.RawMessage
	)

	sl := a.sling.New().Get("/audits/search").
		Set(string(Authorization), a.authorizationHeader).
		Set(string(IntegrationID), "safetyculture-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	sl.QueryStruct(params)
	req, _ := sl.Request()
	req = req.WithContext(ctx)

	_, err := a.do(&slingHTTPDoer{
		sl:       sl,
		req:      req,
		successV: &result,
		failureV: &errMsg,
	})
	if err != nil {
		return nil, errors.Wrap(err, "api request")
	}

	return result, nil
}

// GetInspection retrieves the inspection of the given id.
func (a *Client) GetInspection(ctx context.Context, id string) (*json.RawMessage, error) {
	var (
		result *json.RawMessage
		errMsg json.RawMessage
	)

	sl := a.sling.New().Get(fmt.Sprintf("/audits/%s", id)).
		Set(string(Authorization), a.authorizationHeader).
		Set(string(IntegrationID), "safetyculture-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	_, err := a.do(&slingHTTPDoer{
		sl:       sl,
		req:      req,
		successV: &result,
		failureV: &errMsg,
	})
	if err != nil {
		return nil, errors.Wrap(err, "api request")
	}

	return result, nil
}

// Get makes a get request
func (a *Client) Get(ctx context.Context, url string) (*json.RawMessage, error) {
	var (
		result *json.RawMessage
		errMsg json.RawMessage
	)

	sl := a.sling.New().Get(url).
		Set(string(Authorization), a.authorizationHeader).
		Set(string(IntegrationID), "safetyculture-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	_, err := a.do(&slingHTTPDoer{
		sl:       sl,
		req:      req,
		successV: &result,
		failureV: &errMsg,
	})
	if err != nil {
		return nil, errors.Wrap(err, "api request")
	}

	return result, nil
}

// DrainInspections fetches the inspections in batches and triggers the callback
// for each batch.
func (a *Client) DrainInspections(
	ctx context.Context,
	params *ListInspectionsParams,
	callback func(*ListInspectionsResponse) error,
) error {
	modifiedAfter := params.ModifiedAfter

	for {
		resp, err := a.ListInspections(
			ctx,
			&ListInspectionsParams{
				ModifiedAfter: modifiedAfter,
				TemplateIDs:   params.TemplateIDs,
				Archived:      params.Archived,
				Completed:     params.Completed,
			},
		)
		if err != nil {
			return err
		}

		if err := callback(resp); err != nil {
			return err
		}

		if (resp.Total - resp.Count) == 0 {
			break
		}
		modifiedAfter = resp.Inspections[resp.Count-1].ModifiedAt
	}

	return nil
}

// InitiateInspectionReportExport export the report of the given auditID.
func (a *Client) InitiateInspectionReportExport(ctx context.Context, auditID string, format string, preferenceID string) (string, error) {
	var (
		result *initiateInspectionReportExportResponse
		errMsg json.RawMessage
	)

	url := fmt.Sprintf("audits/%s/report", auditID)
	body := &initiateInspectionReportExportRequest{
		Format:       format,
		PreferenceID: preferenceID,
	}

	sl := a.sling.New().Post(url).
		Set(string(Authorization), a.authorizationHeader).
		Set(string(IntegrationID), "safetyculture-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx)).
		BodyJSON(body)

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	_, err := a.do(&slingHTTPDoer{
		sl:       sl,
		req:      req,
		successV: &result,
		failureV: &errMsg,
	})
	if err != nil {
		return "", errors.Wrap(err, "api request")
	}

	return result.MessageID, nil
}

// CheckInspectionReportExportCompletion checks if the report export is complete.
func (a *Client) CheckInspectionReportExportCompletion(ctx context.Context, auditID string, messageID string) (*InspectionReportExportCompletionResponse, error) {
	var (
		result *InspectionReportExportCompletionResponse
		errMsg json.RawMessage
	)

	url := fmt.Sprintf("audits/%s/report/%s", auditID, messageID)

	sl := a.sling.New().Get(url).
		Set(string(Authorization), a.authorizationHeader).
		Set(string(IntegrationID), "safetyculture-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	_, err := a.do(&slingHTTPDoer{
		sl:       sl,
		req:      req,
		successV: &result,
		failureV: &errMsg,
	})
	if err != nil {
		return nil, errors.Wrap(err, "api request")
	}

	return result, nil
}

// DownloadInspectionReportFile downloads the report file of the inspection.
func (a *Client) DownloadInspectionReportFile(ctx context.Context, url string) (io.ReadCloser, error) {
	var res *http.Response

	sl := a.sling.New().Get(url).
		Set(string(Authorization), a.authorizationHeader).
		Set(string(IntegrationID), "safetyculture-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	res, err := a.do(&defaultHTTPDoer{
		req:        req,
		httpClient: a.httpClient,
	})
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

// WhoAmI returns the details for the user who is making the request
func (a *Client) WhoAmI(ctx context.Context) (*WhoAmIResponse, error) {
	var (
		result *WhoAmIResponse
		errMsg json.RawMessage
	)

	sl := a.sling.New().Get("accounts/user/v1/user:WhoAmI").
		Set(string(Authorization), a.authorizationHeader).
		Set(string(IntegrationID), "safetyculture-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	_, err := a.do(&slingHTTPDoer{
		sl:       sl,
		req:      req,
		successV: &result,
		failureV: &errMsg,
	})
	if err != nil {
		return nil, errors.Wrap(err, "api request")
	}

	return result, nil
}
