package httpapi

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/version"
	"github.com/dghubble/sling"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

/* THIS CODE WAS MOVED FROM API INTERNAL */

var (
	// Default retry configuration
	defaultRetryWaitMin = 1 * time.Second
	defaultRetryWaitMax = 30 * time.Second
	defaultRetryMax     = 4
)

type Client struct {
	logger              *zap.SugaredLogger
	AuthorizationHeader string
	BaseURL             string
	Sling               *sling.Sling
	HttpClient          *http.Client
	httpTransport       *http.Transport

	Duration      time.Duration
	CheckForRetry CheckForRetry
	Backoff       Backoff
	RetryMax      int
	RetryWaitMin  time.Duration
	RetryWaitMax  time.Duration
}

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
		logger:              logger.GetLogger(),
		HttpClient:          httpClient,
		BaseURL:             addr,
		httpTransport:       httpTransport,
		Sling:               sling.New().Client(httpClient).Base(addr),
		AuthorizationHeader: authorizationHeader,
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

// Opt is an option to configure the Client
type Opt func(*Client)

// HTTPClient returns the http Client used by APIClient
func (a *Client) HTTPClient() *http.Client {
	return a.HttpClient
}

// HTTPTransport returns the http Transport used by APIClient
func (a *Client) HTTPTransport() *http.Transport {
	return a.httpTransport
}

// OptSetTimeout sets the timeout for the request
func OptSetTimeout(t time.Duration) Opt {
	return func(a *Client) {
		a.HttpClient.Timeout = t
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

func (a *Client) Do(doer HTTPDoer) (*http.Response, error) {
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

// Get makes a get request
func (a *Client) Get(ctx context.Context, url string) (*json.RawMessage, error) {
	var (
		result *json.RawMessage
		errMsg json.RawMessage
	)

	sl := a.Sling.New().Get(url).
		Set(string(Authorization), a.AuthorizationHeader).
		Set(string(IntegrationID), "safetyculture-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	_, err := a.Do(&util.SlingHTTPDoer{
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

// WhoAmI returns the details for the user who is making the request
func (a *Client) WhoAmI(ctx context.Context) (*WhoAmIResponse, error) {
	var (
		result *WhoAmIResponse
		errMsg json.RawMessage
	)

	sl := a.Sling.New().Get("accounts/user/v1/user:WhoAmI").
		Set(string(Authorization), a.AuthorizationHeader).
		Set(string(IntegrationID), "safetyculture-exporter").
		Set(string(IntegrationVersion), version.GetVersion()).
		Set(string(XRequestID), util.RequestIDFromContext(ctx))

	req, _ := sl.Request()
	req = req.WithContext(ctx)

	_, err := a.Do(&util.SlingHTTPDoer{
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
