package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
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

// APIClient is an interface to the iAuditor API
type APIClient interface {
	HTTPClient() *http.Client
	GetFeed(ctx context.Context, request *GetFeedRequest) (*GetFeedResponse, error)
	DrainFeed(ctx context.Context, request *GetFeedRequest, feedFn func(*GetFeedResponse) error) error
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
func NewAPIClient(addr string, accessToken string, opts ...Opt) APIClient {
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

type GetFeedParams struct {
	ModifiedAfter   string   `url:"modified_after,omitempty"`
	TemplateIDs     []string `url:"template,omitempty"`
	Archived        string   `url:"archived,omitempty"`
	Completed       string   `url:"completed,omitempty"`
	IncludeInactive bool     `url:"include_inactive,omitempty"`
	Limit           int      `url:"limit,omitempty"`
}

type GetFeedRequest struct {
	URL        string
	InitialURL string
	Params     GetFeedParams
}

type GetFeedResponse struct {
	Metadata FeedMetadata `json:"metadata"`

	Data json.RawMessage `json:"data"`
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
