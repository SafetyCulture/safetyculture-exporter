package api

import (
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
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/version"
	"github.com/dghubble/sling"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/pkg/errors"
)

// Client is an interface to the iAuditor API
type Client interface {
	HTTPClient() *http.Client
	GetFeed(ctx context.Context, request *GetFeedRequest) (*GetFeedResponse, error)
	DrainFeed(ctx context.Context, request *GetFeedRequest, feedFn func(*GetFeedResponse) error) error
	ListInspections(ctx context.Context, params *ListInspectionsParams) (*ListInspectionsResponse, error)
	GetInspection(ctx context.Context, id string) (*json.RawMessage, error)
	DrainInspections(ctx context.Context, params *ListInspectionsParams, callback func(*ListInspectionsResponse) error) error
	InitiateInspectionReportExport(ctx context.Context, auditID string, format string, preferenceID string) (string, error)
	CheckInspectionReportExportCompletion(ctx context.Context, auditID string, messageID string) (*InspectionReportExportCompletionResponse, error)
	DownloadInspectionReportFile(ctx context.Context, url string) (io.ReadCloser, error)
}

type apiClient struct {
	accessToken   string
	sling         *sling.Sling
	httpClient    *http.Client
	httpTransport *http.Transport
}

// Opt is an option to configure the APIClient
type Opt func(*apiClient)

// NewAPIClient crates a new instance of the APIClient
func NewAPIClient(addr string, accessToken string, opts ...Opt) Client {
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
		ExpectContinueTimeout: 1 * time.Second,
	}

	httpClient := &http.Client{
		Timeout:   120 * time.Second,
		Transport: newrelic.NewRoundTripper(httpTransport),
	}

	a := apiClient{
		httpClient:    httpClient,
		httpTransport: httpTransport,
		sling:         sling.New().Client(httpClient).Base(addr),
		accessToken:   accessToken,
	}

	for _, opt := range opts {
		opt(&a)
	}

	return &a
}

// HTTPClient returns the http Client used by APIClient
func (a *apiClient) HTTPClient() *http.Client {
	return a.httpClient
}

// OptSetTimeout sets the timeout for the request
func OptSetTimeout(t time.Duration) Opt {
	return func(a *apiClient) {
		a.httpClient.Timeout = t
	}
}

// OptSetProxy sets the proxy URL to use for API requests
func OptSetProxy(proxyURL *url.URL) Opt {
	return func(a *apiClient) {
		a.httpTransport.Proxy = http.ProxyURL(proxyURL)
	}
}

// OptSetInsecureTLS sets whether TLS certs should be verified
func OptSetInsecureTLS(insecureSkipVerify bool) Opt {
	return func(a *apiClient) {
		if a.httpTransport.TLSClientConfig == nil {
			a.httpTransport.TLSClientConfig = &tls.Config{}
		}

		a.httpTransport.TLSClientConfig.InsecureSkipVerify = insecureSkipVerify
	}
}

// OptAddTLSCert adds a certificate at the supplied path to the cert pool
func OptAddTLSCert(certPath string) Opt {
	return func(a *apiClient) {
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
	ModifiedAfter   string   `url:"modified_after,omitempty"`
	TemplateIDs     []string `url:"template,omitempty"`
	Archived        string   `url:"archived,omitempty"`
	Completed       string   `url:"completed,omitempty"`
	IncludeInactive bool     `url:"include_inactive,omitempty"`
	Limit           int      `url:"limit,omitempty"`
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

func (a *apiClient) GetFeed(ctx context.Context, request *GetFeedRequest) (*GetFeedResponse, error) {
	logger := util.GetLogger()

	var (
		result *GetFeedResponse
		errMsg json.RawMessage
		res    *http.Response
		err    error
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

	res, err = sl.Do(req, &result, &errMsg)
	if err != nil {
		return nil, errors.Wrap(err, "Failed request to API")
	}
	if errMsg != nil {
		return result, errors.Errorf("%s", errMsg)
	}
	logger.Debugw("http request",
		"url", req.URL.String(),
		"status", res.Status,
	)

	return result, nil
}

func (a *apiClient) DrainFeed(ctx context.Context, request *GetFeedRequest, feedFn func(*GetFeedResponse) error) error {
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

func (a *apiClient) ListInspections(ctx context.Context, params *ListInspectionsParams) (*ListInspectionsResponse, error) {
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

	if _, err := sl.Do(req, &result, &errMsg); err != nil {
		return nil, errors.Wrap(err, "Failed request to API")
	}

	if errMsg != nil {
		return result, errors.Errorf("%s", errMsg)
	}

	return result, nil
}

func (a *apiClient) GetInspection(ctx context.Context, id string) (*json.RawMessage, error) {
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

	if _, err := sl.Do(req, &result, &errMsg); err != nil {
		return nil, errors.Wrap(err, "Failed request to API")
	}

	if errMsg != nil {
		return result, errors.Errorf("%s", errMsg)
	}

	return result, nil
}

func (a *apiClient) DrainInspections(
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

func (a *apiClient) InitiateInspectionReportExport(ctx context.Context, auditID string, format string, preferenceID string) (string, error) {
	logger := util.GetLogger()

	var (
		result *initiateInspectionReportExportResponse
		errMsg json.RawMessage
		res    *http.Response
		err    error
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

	res, err = sl.Do(req, &result, &errMsg)
	if err != nil {
		return "", errors.Wrap(err, "Failed request to API")
	}
	if errMsg != nil {
		return "", errors.Errorf("%s", errMsg)
	}
	logger.Debugw("http request",
		"url", req.URL.String(),
		"status", res.Status,
	)

	return result.MessageID, nil
}

// InspectionReportExportCompletionResponse represents the response of report export completion status
type InspectionReportExportCompletionResponse struct {
	Status string `json:"status"`
	URL    string `json:"url,omitempty"`
}

func (a *apiClient) CheckInspectionReportExportCompletion(ctx context.Context, auditID string, messageID string) (*InspectionReportExportCompletionResponse, error) {
	logger := util.GetLogger()

	var (
		result *InspectionReportExportCompletionResponse
		errMsg json.RawMessage
		res    *http.Response
		err    error
	)

	url := fmt.Sprintf("audits/%s/report/%s", auditID, messageID)

	sl := a.sling.New().Get(url).
		Set(string(Authorization), "Bearer "+a.accessToken).
		Set(string(IntegrationID), "iauditor-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	res, err = sl.Do(req, &result, &errMsg)
	if err != nil {
		return nil, errors.Wrap(err, "Failed request to API")
	}
	if errMsg != nil {
		return nil, errors.Errorf("%s", errMsg)
	}
	logger.Debugw("http request",
		"url", req.URL.String(),
		"status", res.Status,
	)

	return result, nil
}

func (a *apiClient) DownloadInspectionReportFile(ctx context.Context, url string) (io.ReadCloser, error) {
	logger := util.GetLogger()

	var (
		res *http.Response
		err error
	)

	sl := a.sling.New().Get(url).
		Set(string(Authorization), "Bearer "+a.accessToken).
		Set(string(IntegrationID), "iauditor-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	res, err = a.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Failed request to API")
	}

	logger.Debugw("http request",
		"url", req.URL.String(),
		"status", res.Status,
	)

	return res.Body, err
}
