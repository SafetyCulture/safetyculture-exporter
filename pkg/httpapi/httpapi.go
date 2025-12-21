package httpapi

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
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
	logger        *zap.SugaredLogger
	BaseURL       string
	sling         *sling.Sling
	httpClient    *http.Client
	httpTransport *http.Transport

	Duration      time.Duration
	CheckForRetry CheckForRetry
	backoff       Backoff
	RetryMax      int
	RetryWaitMin  time.Duration
	RetryWaitMax  time.Duration
	rateLimiter   *RateLimiter
}

type ClientCfg struct {
	Addr                string
	AuthorizationHeader string
	IntegrationID       string
	IntegrationVersion  string
}

// NewClient creates a new instance of the Client
func NewClient(cfg *ClientCfg, opts ...Opt) *Client {

	dialFn := func(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
		return dialer.DialContext
	}

	httpTransport := &http.Transport{
		DialContext: dialFn(&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}),
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

	s := sling.New().Client(httpClient).Base(cfg.Addr).
		Set(string(Authorization), cfg.AuthorizationHeader).
		Set(string(IntegrationID), cfg.IntegrationID).
		Set(string(IntegrationVersion), cfg.IntegrationVersion)

	a := &Client{
		logger:        logger.GetLogger(),
		httpClient:    httpClient,
		BaseURL:       cfg.Addr,
		httpTransport: httpTransport,
		sling:         s,
		Duration:      0,
		CheckForRetry: DefaultRetryPolicy,
		backoff:       DefaultBackoff,
		RetryMax:      defaultRetryMax,
		RetryWaitMin:  defaultRetryWaitMin,
		RetryWaitMax:  defaultRetryWaitMax,
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
	return a.httpClient
}

// HTTPTransport returns the http Transport used by APIClient
func (a *Client) HTTPTransport() *http.Transport {
	return a.httpTransport
}

func (a *Client) NewSling() *sling.Sling {
	return a.sling.New()
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

// OptSetRateLimiter configures rate limiting for the client
func OptSetRateLimiter(limiter *RateLimiter) Opt {
	return func(a *Client) {
		a.rateLimiter = limiter
	}
}

// WithRateLimiter returns a shallow copy of the client with the specified rate limiter.
// This allows different feeds to have different rate limits while sharing the underlying HTTP client.
// The returned client shares all HTTP resources (connection pool, etc.) with the original.
func (a *Client) WithRateLimiter(limiter *RateLimiter) *Client {
	return &Client{
		logger:        a.logger,
		BaseURL:       a.BaseURL,
		sling:         a.sling,
		httpClient:    a.httpClient,
		httpTransport: a.httpTransport,
		Duration:      a.Duration,
		CheckForRetry: a.CheckForRetry,
		backoff:       a.backoff,
		RetryMax:      a.RetryMax,
		RetryWaitMin:  a.RetryWaitMin,
		RetryWaitMax:  a.RetryWaitMax,
		rateLimiter:   limiter,
	}
}

// Wait blocks until the rate limiter allows a request or the context is cancelled.
// This is useful for testing and for manual rate limiting in application code.
func (a *Client) Wait(ctx context.Context) error {
	if a.rateLimiter == nil {
		return nil
	}
	return a.rateLimiter.Wait(ctx)
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

func (a *Client) Do(ctx context.Context, doer HTTPDoer) (*http.Response, error) {
	u := doer.URL()
	iter := 0
	for {
		if ctx.Err() != nil && ctx.Err().Error() == "context canceled" {
			return nil, ctx.Err()
		}

		iter++
		a.logger.Debugw("http request", "url", u)

		// Apply rate limiting before making the request
		if a.rateLimiter != nil {
			if err := a.rateLimiter.Wait(ctx); err != nil {
				return nil, fmt.Errorf("rate limit wait: %w", err)
			}
		}

		start := time.Now()
		resp, err := doer.Do()
		a.Duration = time.Since(start)

		status := ""
		if resp != nil {
			status = resp.Status
		}

		if err != nil {
			a.logger.Errorw("http request error", "url", u, "status", status, "err", err)
		}

		a.logger.Debugw("http response", "url", u, "status", status)

		// Check if we should continue with the retries
		shouldRetry, _ := a.CheckForRetry(resp, err)
		if !shouldRetry {
			if resp != nil {
				switch status := resp.StatusCode; {
				case status >= 200 && status <= 299:
					return resp, nil

				case status == http.StatusNotFound:
					a.logger.Errorw("http request error status",
						"url", u,
						"status", status,
					)
					return resp, nil

				case status == http.StatusForbidden:
					a.logger.Errorw("no access to this resource", "url", u, "status", status)
					return resp, nil

				default:
					a.logger.Errorw("http request error status",
						"url", u,
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

		wait := a.backoff(a.RetryWaitMin, a.RetryWaitMax, iter, resp)
		a.logger.Infof("retrying URL %s after %v", u, wait)

		time.Sleep(wait)
	}

	return nil, fmt.Errorf("%s giving up after %d attempt(s)", u, iter+1)
}
