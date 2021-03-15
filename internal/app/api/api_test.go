package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestAPIClientDrainFeed_should_return_for_as_long_next_page_set(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/feed/inspections").
		Reply(200).
		BodyString(`{
			"metadata": {
				"next_page": "/feed/inspections/next",
				"remaining_records": 0
			},
			"data": [
				{
					"id": "audit_8E2B1F3CB9C94D8792957F9F99E2E4BD"
				},
				{
					"id": "audit_1743ae1aaa8741e6a23db83300e56efe"
				}
			]
		}`)

	gock.New("http://localhost:9999").
		Get("/feed/inspections/next").
		Reply(200).
		BodyString(`{
			"metadata": {
				"next_page": null,
				"remaining_records": 0
			},
			"data": [
				{
					"id": "audit_8E2B1F3CB9C94D8792957F9F99E2E4BD"
				},
				{
					"id": "audit_abc"
				}
			]
		}`)

	apiClient := api.NewClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	calls := 0
	auditIDs := []string{}
	err := apiClient.DrainFeed(context.Background(), &api.GetFeedRequest{
		InitialURL: "/feed/inspections",
	}, func(data *api.GetFeedResponse) error {
		calls += 1

		rows := []map[string]string{}
		err := json.Unmarshal(data.Data, &rows)
		assert.Nil(t, err)

		for _, row := range rows {
			auditIDs = append(auditIDs, row["id"])
		}

		return nil
	})
	assert.Nil(t, err)

	assert.Equal(t, 2, calls)
	assert.Equal(t, []string{
		"audit_8E2B1F3CB9C94D8792957F9F99E2E4BD",
		"audit_1743ae1aaa8741e6a23db83300e56efe",
		"audit_8E2B1F3CB9C94D8792957F9F99E2E4BD",
		"audit_abc",
	}, auditIDs)
}

func TestAPIClientDrainFeed_should_bubble_up_errors_from_callback(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/feed/inspections").
		Reply(200).
		BodyString(`{
			"metadata": {
				"next_page": "/feed/inspections/next",
				"remaining_records": 0
			},
			"data": []
		}`)

	apiClient := api.NewClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	expectedErr := errors.New("test error")
	err := apiClient.DrainFeed(context.Background(), &api.GetFeedRequest{
		InitialURL: "/feed/inspections",
	}, func(data *api.GetFeedResponse) error {
		return expectedErr
	})
	assert.Equal(t, expectedErr, err)
}

func TestAPIClientDrainFeed_should_return_api_errors(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/feed/inspections").
		Reply(500).
		JSON(`{"error": "something bad happened"}`)

	tests := []struct {
		name string
		cr   api.CheckForRetry
		err  string
	}{
		{
			name: "default_retry_policy",
			cr:   api.DefaultRetryPolicy,
			err:  "giving up after 3 attempt(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiClient := api.NewClient("http://localhost:9999", "abc123")
			gock.InterceptClient(apiClient.HTTPClient())

			apiClient.RetryWaitMin = 10 * time.Millisecond
			apiClient.RetryWaitMax = 10 * time.Millisecond
			apiClient.CheckForRetry = tt.cr
			apiClient.RetryMax = 2

			err := apiClient.DrainFeed(context.Background(), &api.GetFeedRequest{
				InitialURL: "/feed/inspections",
			}, func(data *api.GetFeedResponse) error {
				return nil
			})
			if err == nil || !strings.HasSuffix(err.Error(), tt.err) {
				t.Fatalf("expected giving up error, got: %#v", err)
			}
		})
	}
}

func TestApiOptSetTimeout_should_set_timeout(t *testing.T) {
	apiClient := api.NewClient("http://localhost:9999", "abc123", api.OptSetTimeout(time.Second*21))

	assert.Equal(t, time.Second*21, apiClient.HTTPClient().Timeout)
}

func TestAPIClientDrainInspections_should_return_for_as_long_next_page_set(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/audits/search").
		Reply(200).
		BodyString(`{
			"count": 2,
			"total": 2,
			"audits": [
				{
					"audit_id": "audit_8E2B1F3CB9C94D8792957F9F99E2E4BD"
				},
				{
					"audit_id": "audit_1743ae1aaa8741e6a23db83300e56efe"
				}
			]
		}`)

	apiClient := api.NewClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	auditIDs := []string{}
	err := apiClient.DrainInspections(
		context.Background(),
		&api.ListInspectionsParams{},
		func(data *api.ListInspectionsResponse) error {
			for _, inspection := range data.Inspections {
				auditIDs = append(auditIDs, inspection.ID)
			}
			return nil
		})
	assert.Nil(t, err)

	assert.Equal(t, []string{
		"audit_8E2B1F3CB9C94D8792957F9F99E2E4BD",
		"audit_1743ae1aaa8741e6a23db83300e56efe",
	}, auditIDs)
}

func TestAPIClientGetInspection(t *testing.T) {
	defer gock.Off()

	auditID := "audit_8E2B1F3CB9C94D8792957F9F99E2E4BD"
	gock.New("http://localhost:9999").
		Get(fmt.Sprintf("/audits/%s", auditID)).
		Reply(200).
		BodyString(`{
			"audit_id": "audit_8E2B1F3CB9C94D8792957F9F99E2E4BD"
		}`)

	apiClient := api.NewClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	resp, err := apiClient.GetInspection(context.Background(), auditID)
	assert.Nil(t, err)

	rows := map[string]string{}
	err = json.Unmarshal(*resp, &rows)
	assert.Nil(t, err)

	expected, ok := rows["audit_id"]
	assert.Equal(t, true, ok)
	assert.Equal(t, expected, auditID)
}

func TestAPIClientGetInspectionWithError(t *testing.T) {
	defer gock.Off()

	auditID := "audit_8E2B1F3CB9C94D8792957F9F99E2E4BD"
	gock.New("http://localhost:9999").
		Get(fmt.Sprintf("/audits/%s", auditID)).
		ReplyError(fmt.Errorf("test error"))

	apiClient := api.NewClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	_, err := apiClient.GetInspection(context.Background(), auditID)
	assert.NotNil(t, err)
}

func TestAPIClientListInspectionWithError(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/audits/search").
		ReplyError(fmt.Errorf("test error"))

	apiClient := api.NewClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	_, err := apiClient.ListInspections(context.Background(), nil)
	assert.NotNil(t, err)
}

func TestDrainInspectionsWithAPIError(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/audits/search").
		ReplyError(fmt.Errorf("test error"))

	apiClient := api.NewClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	err := apiClient.DrainInspections(
		context.Background(),
		&api.ListInspectionsParams{},
		func(data *api.ListInspectionsResponse) error {
			return nil
		})
	assert.NotNil(t, err)
}

func TestDrainInspectionsWithCallbackError(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/audits/search").
		Reply(200).
		BodyString(`{
			"count": 2,
			"total": 2,
			"audits": [
				{
					"audit_id": "audit_8E2B1F3CB9C94D8792957F9F99E2E4BD"
				},
				{
					"audit_id": "audit_1743ae1aaa8741e6a23db83300e56efe"
				}
			]
		}`)

	apiClient := api.NewClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	err := apiClient.DrainInspections(
		context.Background(),
		&api.ListInspectionsParams{},
		func(data *api.ListInspectionsResponse) error {
			return fmt.Errorf("test error")
		})
	assert.NotNil(t, err)
}

func TestGetMediaWithAPIError(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/audits/1234/media/12345").
		ReplyError(fmt.Errorf("test error"))

	apiClient := api.NewClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	_, err := apiClient.GetMedia(
		context.Background(),
		&api.GetMediaRequest{
			URL:     "http://localhost:9999/audits/1234/media/12345",
			AuditID: "1234",
		},
	)
	assert.NotNil(t, err)
}

func TestGetMediaWith204Error(t *testing.T) {
	defer gock.Off()

	result := `{id:"test-id"}`
	gock.New("http://localhost:9999").
		Get("/audits/1234/media/12345").
		Reply(204).
		BodyString(result)

	apiClient := api.NewClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	resp, err := apiClient.GetMedia(
		context.Background(),
		&api.GetMediaRequest{
			URL:     "http://localhost:9999/audits/1234/media/12345",
			AuditID: "1234",
		},
	)
	assert.Nil(t, err)
	assert.Nil(t, resp)
}

func TestGetMediaWithNoContentType(t *testing.T) {
	defer gock.Off()

	result := `{id:"test-id"}`
	gock.New("http://localhost:9999").
		Get("/audits/1234/media/12345").
		Reply(200).
		BodyString(result)

	apiClient := api.NewClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	_, err := apiClient.GetMedia(
		context.Background(),
		&api.GetMediaRequest{
			URL:     "http://localhost:9999/audits/1234/media/12345",
			AuditID: "1234",
		},
	)
	assert.NotNil(t, err)
}

func TestGetMedia(t *testing.T) {
	defer gock.Off()

	result := `{id:"test-id"}`
	header := make(http.Header)
	header["Content-Type"] = []string{"test-content"}
	req := gock.New("http://localhost:9999").
		Get("/audits/1234/media/12345").
		Reply(200).
		BodyString(result)
	req.SetHeader("Content-Type", "test-content")

	apiClient := api.NewClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	expected := &api.GetMediaResponse{
		ContentType: "test-content",
		Body:        []byte(result),
		MediaID:     "12345",
	}
	resp, err := apiClient.GetMedia(
		context.Background(),
		&api.GetMediaRequest{
			URL:     "http://localhost:9999/audits/1234/media/12345",
			AuditID: "1234",
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, resp, expected)
}

func TestAPIClientInitiateInspectionReportExport_should_return_messageID(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Post("/audits/audit_123/report").
		JSON(`{
			"format": "PDF",
			"preference_id": "p123"
		}`).
		Reply(200).
		JSON(`{
			"messageId": "abc"
		}`)

	apiClient := api.NewClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	mId, err := apiClient.InitiateInspectionReportExport(context.Background(), "audit_123", "PDF", "p123")

	assert.Nil(t, err)
	assert.Equal(t, "abc", mId)
}

func TestAPIClientInitiateInspectionReportExport_should_return_error_on_failure(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Post("/audits/audit_123/report").
		JSON(`{"format": "PDF"}`).
		Reply(500).
		JSON(`{"error": "something bad happened"}`)

	tests := []struct {
		name string
		cr   api.CheckForRetry
		err  string
	}{
		{
			name: "default_retry_policy",
			cr:   api.DefaultRetryPolicy,
			err:  "giving up after 2 attempt(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiClient := api.NewClient("http://localhost:9999", "abc123")
			gock.InterceptClient(apiClient.HTTPClient())

			apiClient.RetryWaitMin = 10 * time.Millisecond
			apiClient.RetryWaitMax = 10 * time.Millisecond
			apiClient.CheckForRetry = tt.cr
			apiClient.RetryMax = 1

			_, err := apiClient.InitiateInspectionReportExport(context.Background(), "audit_123", "PDF", "")
			if err == nil || !strings.HasSuffix(err.Error(), tt.err) {
				t.Fatalf("expected giving up error, got: %#v", err)
			}
		})
	}
}

func TestAPIClientCheckInspectionReportExportCompletion_should_return_status(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/audits/audit_123/report/abc").
		Reply(200).
		JSON(`{
			"status": "SUCCESS",
			"url": "http://domain.com/report"
		}`)

	apiClient := api.NewClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	res, err := apiClient.CheckInspectionReportExportCompletion(context.Background(), "audit_123", "abc")

	assert.Nil(t, err)
	assert.Equal(t, res.Status, "SUCCESS")
	assert.Equal(t, res.URL, "http://domain.com/report")
}

func TestAPIClientCheckInspectionReportExportCompletion_should_return_error_on_failure(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/audits/audit_123/report/abc").
		Reply(500).
		JSON(`{"error": "something bad happened"}`)

	tests := []struct {
		name string
		cr   api.CheckForRetry
		err  string
	}{
		{
			name: "default_retry_policy",
			cr:   api.DefaultRetryPolicy,
			err:  "giving up after 2 attempt(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiClient := api.NewClient("http://localhost:9999", "abc123")
			gock.InterceptClient(apiClient.HTTPClient())

			apiClient.RetryWaitMin = 10 * time.Millisecond
			apiClient.RetryWaitMax = 10 * time.Millisecond
			apiClient.CheckForRetry = tt.cr
			apiClient.RetryMax = 1

			_, err := apiClient.CheckInspectionReportExportCompletion(context.Background(), "audit_123", "abc")
			if err == nil || !strings.HasSuffix(err.Error(), tt.err) {
				t.Fatalf("expected giving up error, got: %#v", err)
			}
		})
	}
}

func TestAPIClientDownloadInspectionReportFile_should_return_status(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/report-exports/abc").
		Reply(200).
		Body(bytes.NewBuffer([]byte(`file content`)))

	apiClient := api.NewClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	res, err := apiClient.DownloadInspectionReportFile(context.Background(), "http://localhost:9999/report-exports/abc")

	assert.Nil(t, err)

	buf := new(bytes.Buffer)
	buf.ReadFrom(res)
	assert.Equal(t, buf.String(), "file content")
}

func TestAPIClientDownloadInspectionReportFile_should_return_error_on_failure(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/report-exports/abc").
		Reply(500).
		BodyString("somthing bad happened")

	tests := []struct {
		name string
		cr   api.CheckForRetry
		err  string
	}{
		{
			name: "default_retry_policy",
			cr:   api.DefaultRetryPolicy,
			err:  "giving up after 2 attempt(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiClient := api.NewClient("http://localhost:9999", "abc123")
			gock.InterceptClient(apiClient.HTTPClient())

			apiClient.RetryWaitMin = 10 * time.Millisecond
			apiClient.RetryWaitMax = 10 * time.Millisecond
			apiClient.CheckForRetry = tt.cr
			apiClient.RetryMax = 1

			_, err := apiClient.DownloadInspectionReportFile(context.Background(), "http://localhost:9999/report-exports/abc")
			if err == nil || !strings.HasSuffix(err.Error(), tt.err) {
				t.Fatalf("expected giving up error, got: %#v", err)
			}
		})
	}
}

func TestAPIClientBackoff429TooManyRequest(t *testing.T) {
	defer gock.Off()

	req := gock.New("http://localhost:9999").
		Get(fmt.Sprintf("/audits/%s", "1234")).
		Reply(429)
	now := time.Now().Unix() * 1000
	req.SetHeader("X-Rate-Limit-Reset", strconv.FormatInt(now, 10))

	tests := []struct {
		name string
		bo   api.Backoff
		err  string
	}{
		{
			name: "default_backoff_policy",
			bo:   api.DefaultBackoff,
			err:  "giving up after 2 attempt(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiClient := api.NewClient("http://localhost:9999", "abc123")
			gock.InterceptClient(apiClient.HTTPClient())
			apiClient.RetryMax = 1

			_, err := apiClient.GetInspection(context.Background(), "1234")
			if err == nil || !strings.HasSuffix(err.Error(), tt.err) {
				t.Fatalf("expected giving up error, got: %#v", err)
			}
		})
	}
}
