package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/version"
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

// Client is used to with iAuditor API's.
type Client struct {
	logger        *zap.SugaredLogger
	accessToken   string
	baseURL       string
	sling         *sling.Sling
	httpClient    *http.Client
	httpTransport *http.Transport

	CheckForRetry CheckForRetry
	Backoff       Backoff
	RetryMax      int
	RetryWaitMin  time.Duration
	RetryWaitMax  time.Duration
}

// Opt is an option to configure the Client
type Opt func(*Client)

// NewClient creates a new instance of the Client
func NewClient(addr string, accessToken string, opts ...Opt) *Client {
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
		logger:        util.GetLogger(),
		httpClient:    httpClient,
		baseURL:       addr,
		httpTransport: httpTransport,
		sling:         sling.New().Client(httpClient).Base(addr),
		accessToken:   accessToken,
		CheckForRetry: DefaultRetryPolicy,
		Backoff:       DefaultBackoff,
		RetryMax:      defaultRetryMax,
		RetryWaitMin:  defaultRetryWaitMin,
		RetryWaitMax:  defaultRetryWaitMax,
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
		certs, err := ioutil.ReadFile(certPath)
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

// FeedMetadata is a representation of the metadata returned when fetching a feed
type FeedMetadata struct {
	NextPage         string `json:"next_page"`
	RemainingRecords int64  `json:"remaining_records"`
}

// GetFeedParams is a list of all parameters we can set when fetching a feed
type GetFeedParams struct {
	ModifiedAfter   time.Time `url:"modified_after,omitempty"`
	TemplateIDs     []string  `url:"template,omitempty"`
	Archived        string    `url:"archived,omitempty"`
	Completed       string    `url:"completed,omitempty"`
	IncludeInactive bool      `url:"include_inactive,omitempty"`
	Limit           int       `url:"limit,omitempty"`

	// Applicable only for sites
	IncludeDeleted bool `url:"include_deleted,omitempty"`
}

// GetFeedRequest has all the data needed to make a request to get a feed
type GetFeedRequest struct {
	URL        string
	InitialURL string
	Params     GetFeedParams
}

// GetFeedResponse is a representation of the data returned when fetching a feed
type GetFeedResponse struct {
	Metadata FeedMetadata `json:"metadata"`

	Data json.RawMessage `json:"data"`
}

// GetMediaRequest has all the data needed to make a request to get a media
type GetMediaRequest struct {
	URL     string
	AuditID string
}

// GetMediaResponse is a representation of the data returned when fetching media
type GetMediaResponse struct {
	ContentType string
	Body        []byte
	MediaID     string
}

// ListInspectionsParams is a list of all parameters we can set when fetching inspections
type ListInspectionsParams struct {
	ModifiedAfter time.Time `url:"modified_after,omitempty"`
	TemplateIDs   []string  `url:"template,omitempty"`
	Archived      string    `url:"archived,omitempty"`
	Completed     string    `url:"completed,omitempty"`
	Limit         int       `url:"limit,omitempty"`
}

// Inspection represents some of the properties present in an inspection
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

func (a *Client) do(doer HTTPDoer) (*http.Response, error) {
	url := doer.URL()

	for iter := 0; ; iter++ {
		a.logger.Debugw("http request",
			"url", url,
		)

		resp, err := doer.Do()
		status := ""
		if resp != nil {
			status = resp.Status
		}

		if err != nil {
			a.logger.Errorw("http request error",
				"url", url,
				"status", status,
				"err", err,
			)
		}

		a.logger.Debugw("http response",
			"url", url,
			"status", status,
		)

		if resp != nil && (resp.StatusCode > 299 || resp.StatusCode < 200) {
			a.logger.Errorw("http request error status",
				"url", url,
				"status", resp.Status,
				"err", doer.Error(),
			)
		}

		// Check if we should continue with the retries
		shouldRetry, checkErr := a.CheckForRetry(resp, err)
		if !shouldRetry {
			if checkErr != nil {
				err = checkErr
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

// GetMedia fetches the media object from iAuditor.
func (a *Client) GetMedia(ctx context.Context, request *GetMediaRequest) (*GetMediaResponse, error) {
	baseURL := strings.TrimPrefix(request.URL, a.baseURL)
	mediaID := strings.TrimPrefix(baseURL, fmt.Sprintf("/audits/%s/media/", request.AuditID))

	sl := a.sling.New().Get(baseURL).
		Set(string(Authorization), fmt.Sprintf("Bearer %s", a.accessToken)).
		Set(string(IntegrationID), "iauditor-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	result, err := a.do(&defaultHTTPDoer{
		req:        req,
		httpClient: a.httpClient,
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	if result.StatusCode == 204 {
		return nil, nil
	}

	contentType, ok := result.Header["Content-Type"]
	if !ok {
		return nil, fmt.Errorf("Failed to get content-type of media")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)

	resp := &GetMediaResponse{
		ContentType: contentType[0],
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

	url := request.InitialURL
	if request.URL != "" {
		url = request.URL
	}

	sl := a.sling.New().Get(url).
		Set(string(Authorization), "Bearer "+a.accessToken).
		Set(string(IntegrationID), "iauditor-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	if request.URL == "" {
		sl.QueryStruct(request.Params)
	}

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	_, err := a.do(&slingHTTPDoer{
		sl:       sl,
		req:      req,
		successV: &result,
		failureV: &errMsg,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed request to API")
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
			return httpErr
		}
		nextURL = resp.Metadata.NextPage

		err := feedFn(resp)
		if err != nil {
			return err
		}
	}

	return nil
}

// ListInspections retrieves the list of inspections from iAuditor
func (a *Client) ListInspections(ctx context.Context, params *ListInspectionsParams) (*ListInspectionsResponse, error) {
	var (
		result *ListInspectionsResponse
		errMsg json.RawMessage
	)

	sl := a.sling.New().Get("/audits/search").
		Set(string(Authorization), fmt.Sprintf("Bearer %s", a.accessToken)).
		Set(string(IntegrationID), "iauditor-exporter").
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
		return nil, errors.Wrap(err, "Failed request to API")
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
		Set(string(Authorization), fmt.Sprintf("Bearer %s", a.accessToken)).
		Set(string(IntegrationID), "iauditor-exporter").
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
		return nil, errors.Wrap(err, "Failed request to API")
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

type initiateInspectionReportExportRequest struct {
	Format       string `json:"format"`
	PreferenceID string `json:"preference_id,omitempty"`
}

type initiateInspectionReportExportResponse struct {
	MessageID string `json:"messageId"`
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
		Set(string(Authorization), "Bearer "+a.accessToken).
		Set(string(IntegrationID), "iauditor-exporter").
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
		return "", errors.Wrap(err, "Failed request to API")
	}

	return result.MessageID, nil
}

// InspectionReportExportCompletionResponse represents the response of report export completion status
type InspectionReportExportCompletionResponse struct {
	Status string `json:"status"`
	URL    string `json:"url,omitempty"`
}

// CheckInspectionReportExportCompletion checks if the report export is complete.
func (a *Client) CheckInspectionReportExportCompletion(ctx context.Context, auditID string, messageID string) (*InspectionReportExportCompletionResponse, error) {
	var (
		result *InspectionReportExportCompletionResponse
		errMsg json.RawMessage
	)

	url := fmt.Sprintf("audits/%s/report/%s", auditID, messageID)

	sl := a.sling.New().Get(url).
		Set(string(Authorization), "Bearer "+a.accessToken).
		Set(string(IntegrationID), "iauditor-exporter").
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
		return nil, errors.Wrap(err, "Failed request to API")
	}

	return result, nil
}

// DownloadInspectionReportFile downloads the report file of the inspection.
func (a *Client) DownloadInspectionReportFile(ctx context.Context, url string) (io.ReadCloser, error) {
	var res *http.Response

	sl := a.sling.New().Get(url).
		Set(string(Authorization), "Bearer "+a.accessToken).
		Set(string(IntegrationID), "iauditor-exporter").
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
