package api

// Header is used to represent name of a header
type Header string

// Headers that are sent with each request when making api calls
const (
	Authorization      Header = "Authorization"
	ContentType        Header = "Content-Type"
	XRequestID         Header = "X-Request-ID"
	IntegrationID      Header = "sc-integration-id"
	IntegrationVersion Header = "sc-integration-version"
)
