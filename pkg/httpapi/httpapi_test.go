package httpapi_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

// GetTestClient creates a new test apiClient
func GetTestClient(opts ...httpapi.Opt) *httpapi.Client {
	cfg := httpapi.ClientCfg{
		Addr:                "http://localhost:9999",
		AuthorizationHeader: "abc123",
		IntegrationID:       "safetyculture-exporter",
		IntegrationVersion:  "dev",
	}

	apiClient := httpapi.NewClient(&cfg, opts...)
	apiClient.RetryWaitMin = 10 * time.Millisecond
	apiClient.RetryWaitMax = 10 * time.Millisecond
	apiClient.CheckForRetry = httpapi.DefaultRetryPolicy
	apiClient.RetryMax = 1
	return apiClient
}

func TestApiOptSetTimeout_should_set_timeout(t *testing.T) {
	apiClient := GetTestClient(httpapi.OptSetTimeout(time.Second * 21))

	assert.Equal(t, time.Second*21, apiClient.HTTPClient().Timeout)
}

func TestClient_OptSetTimeout(t *testing.T) {
	cfg := httpapi.ClientCfg{
		Addr:                "fake_addr",
		AuthorizationHeader: "fake_token",
		IntegrationID:       "test",
		IntegrationVersion:  "dev",
	}

	client := httpapi.NewClient(&cfg)
	require.NotNil(t, client)

	opt := httpapi.OptSetTimeout(time.Second * 10)
	opt(client)
	require.NotNil(t, opt)
	assert.EqualValues(t, time.Second*10, client.HTTPClient().Timeout)
}

func TestClient_OptAddTLSCert_WhenEmptyPath(t *testing.T) {
	cfg := httpapi.ClientCfg{
		Addr:                "fake_addr",
		AuthorizationHeader: "fake_token",
		IntegrationID:       "test",
		IntegrationVersion:  "dev",
	}

	client := httpapi.NewClient(&cfg)
	require.NotNil(t, client)

	opt := httpapi.OptAddTLSCert("")
	opt(client)
	require.NotNil(t, opt)
	assert.Nil(t, client.HTTPTransport().TLSClientConfig)
}

func TestClient_OptSetProxy(t *testing.T) {
	cfg := httpapi.ClientCfg{
		Addr:                "fake_addr",
		AuthorizationHeader: "fake_token",
		IntegrationID:       "test",
		IntegrationVersion:  "dev",
	}

	client := httpapi.NewClient(&cfg)
	require.NotNil(t, client)

	u := url.URL{
		Scheme: "https",
		Host:   "fake.com",
	}
	opt := httpapi.OptSetProxy(&u)
	opt(client)

	require.NotNil(t, opt)
}

func TestClient_OptSetInsecureTLS_WhenTrue(t *testing.T) {
	cfg := httpapi.ClientCfg{
		Addr:                "fake_addr",
		AuthorizationHeader: "fake_token",
		IntegrationID:       "test",
		IntegrationVersion:  "dev",
	}

	client := httpapi.NewClient(&cfg)
	require.NotNil(t, client)

	opt := httpapi.OptSetInsecureTLS(true)
	opt(client)
	require.NotNil(t, opt)
	assert.True(t, client.HTTPTransport().TLSClientConfig.InsecureSkipVerify)
}

func TestClient_WhoAmI_WhenOK(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("accounts/user/v1/user:WhoAmI").
		Reply(200).
		BodyString(`{}`)

	apiClient := GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	r, err := httpapi.WhoAmI(context.Background(), apiClient)
	require.Nil(t, err)
	require.NotNil(t, r)
}

func TestClient_WhoAmI_WhenNotOK(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	gock.New("http://localhost:9999").
		Persist().
		Get("accounts/user/v1/user:WhoAmI").
		Reply(500).
		BodyString(`{}`)

	apiClient := GetTestClient()
	apiClient.RetryMax = 3
	gock.InterceptClient(apiClient.HTTPClient())

	r, err := httpapi.WhoAmI(context.Background(), apiClient)
	require.NotNil(t, err)
	require.Nil(t, r)
	assert.EqualValues(t, "api request: http://localhost:9999/accounts/user/v1/user:WhoAmI giving up after 4 attempt(s)", err.Error())
}

func TestClient_WhoAmI_WhenNotOK_ContextCancelled(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	gock.New("http://localhost:9999").
		Persist().
		Get("accounts/user/v1/user:WhoAmI").
		Reply(200).
		BodyString(`{}`)

	apiClient := GetTestClient()
	apiClient.RetryMax = 3
	gock.InterceptClient(apiClient.HTTPClient())

	ctx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()
	r, err := httpapi.WhoAmI(ctx, apiClient)
	require.NotNil(t, err)
	require.Nil(t, r)
	assert.EqualValues(t, "api request: context canceled", err.Error())
}

func TestClient_HeadersShouldMatch(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	gock.New("http://localhost:9999").
		Get("accounts/user/v1/user:WhoAmI").
		MatchHeader("sc-integration-id", "safetyculture-exporter").
		MatchHeader("sc-integration-version", "dev").
		Reply(200).
		BodyString(`{}`)

	apiClient := GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	r, err := httpapi.WhoAmI(context.Background(), apiClient)
	require.Nil(t, err)
	require.NotNil(t, r)
}

func TestClient_WithRateLimiter_CreatesShallowCopy(t *testing.T) {
	// Create original client
	cfg := httpapi.ClientCfg{
		Addr:                "http://localhost:9999",
		AuthorizationHeader: "test-token",
		IntegrationID:       "test-id",
		IntegrationVersion:  "v1.0",
	}
	originalClient := httpapi.NewClient(&cfg)

	// Create rate limiter
	rateLimiter := httpapi.NewRateLimiter(httpapi.RateLimiterConfig{
		RequestsPerMinute: 60,
		BurstSize:         60,
		Enabled:           true,
		Name:              "test",
	})

	// Create copy with rate limiter
	rateLimitedClient := originalClient.WithRateLimiter(rateLimiter)

	// Verify they are different instances
	assert.NotSame(t, originalClient, rateLimitedClient, "should create a new client instance")

	// Verify they share the same underlying HTTP resources
	assert.Same(t, originalClient.HTTPClient(), rateLimitedClient.HTTPClient(), "should share HTTP client")
	assert.Same(t, originalClient.HTTPTransport(), rateLimitedClient.HTTPTransport(), "should share HTTP transport")

	// Verify base URL is copied
	assert.Equal(t, originalClient.BaseURL, rateLimitedClient.BaseURL)
}

func TestClient_WithRateLimiter_AppliesRateLimiting(t *testing.T) {
	defer gock.Off()

	// Mock endpoint that always succeeds
	gock.New("http://localhost:9999").
		Get("/test").
		Persist().
		Reply(200).
		BodyString(`{"status": "ok"}`)

	// Create client with rate limiter (60 requests per minute = 1 per second)
	cfg := httpapi.ClientCfg{
		Addr:                "http://localhost:9999",
		AuthorizationHeader: "test-token",
		IntegrationID:       "test-id",
		IntegrationVersion:  "v1.0",
	}
	client := httpapi.NewClient(&cfg)

	rateLimiter := httpapi.NewRateLimiter(httpapi.RateLimiterConfig{
		RequestsPerMinute: 60, // 1 request per second
		BurstSize:         2,  // Allow initial burst of 2
		Enabled:           true,
		Name:              "test",
		Algorithm:         httpapi.AlgorithmTokenBucket, // Use token bucket for this test
	})
	rateLimitedClient := client.WithRateLimiter(rateLimiter)
	gock.InterceptClient(rateLimitedClient.HTTPClient())

	// Make 3 requests - first 2 should be fast (burst), 3rd should be rate-limited
	start := time.Now()

	for i := 0; i < 3; i++ {
		_, err := httpapi.ExecuteGet[map[string]string](context.Background(), rateLimitedClient, "/test", nil)
		require.NoError(t, err)
	}

	elapsed := time.Since(start)

	// The 3rd request should have been delayed by ~1 second due to rate limiting
	// (burst allows 2, then we wait for the next token)
	assert.GreaterOrEqual(t, elapsed, 900*time.Millisecond, "should enforce rate limiting")
}

func TestClient_WithRateLimiter_NoRateLimitingWhenNil(t *testing.T) {
	defer gock.Off()

	// Mock endpoint
	gock.New("http://localhost:9999").
		Get("/test").
		Times(3).
		Reply(200).
		BodyString(`{"status": "ok"}`)

	// Create client without rate limiter
	cfg := httpapi.ClientCfg{
		Addr:                "http://localhost:9999",
		AuthorizationHeader: "test-token",
		IntegrationID:       "test-id",
		IntegrationVersion:  "v1.0",
	}
	client := httpapi.NewClient(&cfg)
	gock.InterceptClient(client.HTTPClient())

	// Make 3 requests quickly
	start := time.Now()

	for i := 0; i < 3; i++ {
		_, err := httpapi.ExecuteGet[map[string]string](context.Background(), client, "/test", nil)
		require.NoError(t, err)
	}

	elapsed := time.Since(start)

	// Should complete quickly without rate limiting
	assert.Less(t, elapsed, 500*time.Millisecond, "should not apply rate limiting")
}
