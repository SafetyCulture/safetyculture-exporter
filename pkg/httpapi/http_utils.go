package httpapi

import (
	"net/http"
)

// Header is used to represent name of a header
type Header string

// Headers that are sent with each request when making api calls
const (
	Authorization      Header = "Authorization"
	ContentType        Header = "Content-Type"
	XRequestID         Header = "X-Request-ID"
	IntegrationID      Header = "sc-integration-id"
	IntegrationVersion Header = "sc-integration-version"
	XRateLimitReset    Header = "X-RateLimit-Reset"
)

// HTTPDoer executes http requests.
type HTTPDoer interface {
	Do() (*http.Response, error)
	URL() string
	Error() interface{}
}

type defaultHTTPDoer struct {
	req        *http.Request
	httpClient *http.Client
}

func (b *defaultHTTPDoer) Do() (*http.Response, error) {
	return b.httpClient.Do(b.req)
}

func (b *defaultHTTPDoer) URL() string {
	return b.req.URL.String()
}

func (b *defaultHTTPDoer) Error() interface{} {
	return nil
}
