// +build mitmproxy

package api_test

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/stretchr/testify/assert"
)

func TestAPIClient_should_proxy_requests_when_proxy_supplied(t *testing.T) {
	proxyURL, err := url.Parse(os.Getenv("TEST_PROXY_URL"))
	assert.Nil(t, err)

	apiClient := api.NewAPIClient(
		"https://api.safetyculture.io",
		"invalid_token",
		api.OptAddTLSCert(os.Getenv("TEST_TLS_CERT_PATH")),
		api.OptSetProxy(proxyURL),
	)

	_, err = apiClient.GetFeed(context.Background(), &api.GetFeedRequest{
		InitialURL: "/feed/inspections",
	})
	// Expect the requests makes it through, but is rejected for bad token
	assert.Contains(t, err.Error(), `{"statusCode":401,"error":"Unauthorized","message":"Bad token or token expired"}`)
}

func TestAPIClient_should_proxy_also_in_insecure_mode(t *testing.T) {
	proxyURL, err := url.Parse(os.Getenv("TEST_PROXY_URL"))
	assert.Nil(t, err)

	apiClient := api.NewAPIClient(
		"https://api.safetyculture.io",
		"invalid_token",
		api.OptSetInsecureTLS(true),
		api.OptSetProxy(proxyURL),
	)

	_, err = apiClient.GetFeed(context.Background(), &api.GetFeedRequest{
		InitialURL: "/feed/inspections",
	})
	// Expect the requests makes it through, but is rejected for bad token
	assert.Contains(t, err.Error(), `{"statusCode":401,"error":"Unauthorized","message":"Bad token or token expired"}`)
}

func TestAPIClient_should_fail_without_valid_ca(t *testing.T) {
	proxyURL, err := url.Parse(os.Getenv("TEST_PROXY_URL"))
	assert.Nil(t, err)

	apiClient := api.NewAPIClient(
		"https://api.safetyculture.io",
		"invalid_token",
		api.OptSetProxy(proxyURL),
	)

	_, err = apiClient.GetFeed(context.Background(), &api.GetFeedRequest{
		InitialURL: "/feed/inspections",
	})
	// Expect the requests makes it through, but is rejected for bad token
	assert.Contains(t, err.Error(), "x509: certificate signed by unknown authority")
}
