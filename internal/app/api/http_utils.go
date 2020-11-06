package api

type Header string

const (
	Authorization      Header = "Authorization"
	ContentType        Header = "Content-Type"
	XRequestID         Header = "X-Request-ID"
	IntegrationID      Header = "sc-integration-id"
	IntegrationVersion Header = "sc-integration-version"
)
