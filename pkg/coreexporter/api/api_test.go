package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/coreexporter/api"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestClient_DrainDeletedInspections(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Post("/accounts/history/v1/activity_log/list").
		BodyString(`{"org_id":"","page_size":4,"page_token":"","filters":{"timeframe":{"from":"2022-06-30T10:43:17Z"},"event_types":["inspection.deleted"],"limit":4}}`).
		Reply(http.StatusOK).
		File(path.Join("fixtures", "inspections_deleted_page_1.json"))

	gock.New("http://localhost:9999").
		Post("/accounts/history/v1/activity_log/list").
		BodyString(`{"org_id":"","page_size":4,"page_token":"eyJldmVudF90eXBlcyI6WyJpbnNwZWN0aW9uLmFyY2hpdmVkIl0sImxpbWl0Ijo0LCJvZmZzZXQiOjR9","filters":{"timeframe":{"from":"2022-06-30T10:43:17Z"},"event_types":["inspection.deleted"],"limit":4}}`).
		Reply(http.StatusOK).
		File(path.Join("fixtures", "inspections_deleted_page_2.json"))

	gock.New("http://localhost:9999").
		Post("/accounts/history/v1/activity_log/list").
		BodyString(`{"org_id":"","page_size":4,"page_token":"eyJldmVudF90eXBlcyI6WyJpbnNwZWN0aW9uLmFyY2hpdmVkIl0sImxpbWl0Ijo0LCJvZmZzZXQiOjh9","filters":{"timeframe":{"from":"2022-06-30T10:43:17Z"},"event_types":["inspection.deleted"],"limit":4}}`).
		Reply(http.StatusOK).
		File(path.Join("fixtures", "inspections_deleted_page_3.json"))

	gock.New("http://localhost:9999").
		Post("/accounts/history/v1/activity_log/list").
		BodyString(`{"org_id":"","page_size":4,"page_token":"eyJldmVudF90eXBlcyI6WyJpbnNwZWN0aW9uLmFyY2hpdmVkIl0sImxpbWl0Ijo0LCJvZmZzZXQiOjEyfQ==","filters":{"timeframe":{"from":"2022-06-30T10:43:17Z"},"event_types":["inspection.deleted"],"limit":4}}`).
		Reply(http.StatusOK).
		File(path.Join("fixtures", "inspections_deleted_page_4.json"))

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	fakeTime, err := time.Parse(time.RFC3339, "2022-06-30T10:43:17Z")
	require.Nil(t, err)
	req := api.NewGetAccountsActivityLogRequest(4, fakeTime)

	calls := 0
	var deletedIds = make([]string, 0, 15)
	fn := func(res *api.GetAccountsActivityLogResponse) error {
		calls++
		for _, a := range res.Activities {
			deletedIds = append(deletedIds, a.Metadata["inspection_id"])
		}
		return nil
	}
	err = apiClient.DrainAccountActivityHistoryLog(context.TODO(), req, fn)
	require.Nil(t, err)
	assert.EqualValues(t, 4, calls)
	require.EqualValues(t, 15, len(deletedIds))
	assert.EqualValues(t, "3b8ac4f4-e904-453e-b5a0-b5cceedb0ee1", deletedIds[0])
	assert.EqualValues(t, "4b3bc1d5-3011-4f81-94d4-125d2bce7ca8", deletedIds[1])
	assert.EqualValues(t, "6bd628a6-5188-425f-89ef-81f9dfcdf5cd", deletedIds[2])
	assert.EqualValues(t, "d722fc86-defa-4de2-b8d7-c0a3e0ec6ce4", deletedIds[3])
	assert.EqualValues(t, "ed8b3911-4141-41c4-946c-167bb6f61109", deletedIds[4])
	assert.EqualValues(t, "fd95cb4b-e1e7-488b-ba58-93fecd2379dc", deletedIds[5])
	assert.EqualValues(t, "1878c1e2-8a42-4f63-9e07-2e605f76762b", deletedIds[6])
	assert.EqualValues(t, "9e28ab2c-ce8c-44a7-81d3-76d0ac47dc91", deletedIds[7])
	assert.EqualValues(t, "48d61915-98c8-4d05-b786-4948dad199be", deletedIds[8])
	assert.EqualValues(t, "331727d2-4976-45da-857a-6d080dc645a9", deletedIds[9])
	assert.EqualValues(t, "1f2c9c1b-6f35-4bae-9b38-4094b40e13c1", deletedIds[10])
	assert.EqualValues(t, "35583d49-6421-40a8-a6f5-591c718c6025", deletedIds[11])
	assert.EqualValues(t, "eb49e9f8-4a3c-4b8f-a180-7ba0d171e93d", deletedIds[12])
	assert.EqualValues(t, "47ac0dce-16f9-4d73-b517-8372368af162", deletedIds[13])
	assert.EqualValues(t, "6d2f8bd5-a965-4046-b2b4-ccdf8341c9f0", deletedIds[14])
}

func TestClient_DrainDeletedInspections_WhenApiReturnsError(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Persist().
		Post("/accounts/history/v1/activity_log/list").
		Reply(http.StatusInternalServerError).
		JSON(`{"error": "something bad happened"}`)

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	fakeTime, err := time.Parse(time.RFC3339, "2022-06-30T10:43:17Z")
	require.Nil(t, err)
	req := api.NewGetAccountsActivityLogRequest(14, fakeTime)
	fn := func(res *api.GetAccountsActivityLogResponse) error {
		return nil
	}
	err = apiClient.DrainAccountActivityHistoryLog(context.TODO(), req, fn)
	require.NotNil(t, err)
	assert.EqualValues(t, "api request: http://localhost:9999/accounts/history/v1/activity_log/list giving up after 2 attempt(s)", err.Error())
}

func TestClient_DrainDeletedInspections_WhenFeedFnReturnsError(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Post("/accounts/history/v1/activity_log/list").
		BodyString(`{"org_id":"","page_size":4,"page_token":"","filters":{"timeframe":{"from":"2022-06-30T10:43:17Z"},"event_types":["inspection.deleted"],"limit":4}}`).
		Reply(http.StatusOK).
		File(path.Join("fixtures", "inspections_deleted_page_1.json"))

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	fakeTime, err := time.Parse(time.RFC3339, "2022-06-30T10:43:17Z")
	require.Nil(t, err)
	req := api.NewGetAccountsActivityLogRequest(4, fakeTime)

	fn := func(res *api.GetAccountsActivityLogResponse) error {
		return fmt.Errorf("ERROR_GetAccountsActivityLogResponse")
	}
	err = apiClient.DrainAccountActivityHistoryLog(context.TODO(), req, fn)
	require.NotNil(t, err)
	assert.EqualValues(t, "ERROR_GetAccountsActivityLogResponse", err.Error())
}

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

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	calls := 0
	var auditIDs []string
	err := apiClient.DrainFeed(context.Background(), &api.GetFeedRequest{
		InitialURL: "/feed/inspections",
	}, func(data *api.GetFeedResponse) error {
		calls += 1

		var rows []map[string]string
		err := json.Unmarshal(data.Data, &rows)
		assert.NoError(t, err)

		for _, row := range rows {
			auditIDs = append(auditIDs, row["id"])
		}

		return nil
	})
	assert.NoError(t, err)

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

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	expectedErr := errors.New("test error")
	err := apiClient.DrainFeed(context.Background(), &api.GetFeedRequest{
		InitialURL: "/feed/inspections",
	}, func(data *api.GetFeedResponse) error {
		return expectedErr
	})
	assert.EqualValues(t, expectedErr.Error(), err.Error())
}

func TestClient_DrainFeed_WhenApiReturns403Error(t *testing.T) {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	gock.New("http://localhost:9999").
		Get("/feed/inspections").
		Reply(403).
		BodyString(`{"statusCode":403,"error":"Forbidden","message":"The caller does not have permission to execute the specified operation"}`)

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	err := apiClient.DrainFeed(context.Background(), &api.GetFeedRequest{
		InitialURL: "/feed/inspections",
	}, func(data *api.GetFeedResponse) error {
		return nil
	})
	assert.EqualValues(t, `{"status_code":403,"resource":"/feed/inspections","message":"{\"statusCode\":403,\"error\":\"Forbidden\",\"message\":\"The caller does not have permission to execute the specified operation\"}"}`, err.Error())
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
			err:  "giving up after 2 attempt(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiClient := api.GetTestClient()
			gock.InterceptClient(apiClient.HTTPClient())

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
	apiClient := api.GetTestClient(api.OptSetTimeout(time.Second * 21))

	assert.Equal(t, time.Second*21, apiClient.HTTPClient().Timeout)
}

func TestClient_OptSetTimeout(t *testing.T) {
	client := api.NewClient("fake_addr", "fake_token")
	require.NotNil(t, client)

	opt := api.OptSetTimeout(time.Second * 10)
	opt(client)
	require.NotNil(t, opt)
	assert.EqualValues(t, time.Second*10, client.HTTPClient().Timeout)
}

func TestClient_OptAddTLSCert_WhenEmptyPath(t *testing.T) {
	client := api.NewClient("fake_addr", "fake_token")
	require.NotNil(t, client)

	opt := api.OptAddTLSCert("")
	opt(client)
	require.NotNil(t, opt)
	assert.Nil(t, client.HTTPTransport().TLSClientConfig)
}

func TestClient_OptSetProxy(t *testing.T) {
	client := api.NewClient("fake_addr", "fake_token")
	require.NotNil(t, client)

	u := url.URL{
		Scheme: "https",
		Host:   "fake.com",
	}
	opt := api.OptSetProxy(&u)
	opt(client)

	require.NotNil(t, opt)
}

func TestClient_OptSetInsecureTLS_WhenTrue(t *testing.T) {
	client := api.NewClient("fake_addr", "fake_token")
	require.NotNil(t, client)

	opt := api.OptSetInsecureTLS(true)
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

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	r, err := apiClient.WhoAmI(context.Background())
	require.Nil(t, err)
	require.NotNil(t, r)
}

func TestClient_WhoAmI_WhenNotOK(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("accounts/user/v1/user:WhoAmI").
		Reply(500).
		BodyString(`{}`)

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	r, err := apiClient.WhoAmI(context.Background())
	require.NotNil(t, err)
	require.Nil(t, r)
	assert.EqualValues(t, "api request: http://localhost:9999/accounts/user/v1/user:WhoAmI giving up after 2 attempt(s)", err.Error())
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

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	var auditIDs []string
	err := apiClient.DrainInspections(
		context.Background(),
		&api.ListInspectionsParams{},
		func(data *api.ListInspectionsResponse) error {
			for _, inspection := range data.Inspections {
				auditIDs = append(auditIDs, inspection.ID)
			}
			return nil
		})
	assert.NoError(t, err)

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

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	resp, err := apiClient.GetInspection(context.Background(), auditID)
	assert.NoError(t, err)

	rows := map[string]string{}
	err = json.Unmarshal(*resp, &rows)
	assert.NoError(t, err)

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

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	_, err := apiClient.GetInspection(context.Background(), auditID)
	assert.NotNil(t, err)
}

func TestAPIClientListInspectionWithError(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/audits/search").
		ReplyError(fmt.Errorf("test error"))

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	_, err := apiClient.ListInspections(context.Background(), nil)
	assert.NotNil(t, err)
}

func TestDrainInspectionsWithAPIError(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/audits/search").
		ReplyError(fmt.Errorf("test error"))

	apiClient := api.GetTestClient()
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

	apiClient := api.GetTestClient()
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

	apiClient := api.GetTestClient()
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

func TestGetMediaWith403Error(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/audits/1234/media/12345").
		Reply(403).
		JSON(`{"error": "something bad happened"}`)

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	_, err := apiClient.GetMedia(
		context.Background(),
		&api.GetMediaRequest{
			URL:     "http://localhost:9999/audits/1234/media/12345",
			AuditID: "1234",
		},
	)
	assert.NoError(t, err)
}

func TestGetMediaWith404Error(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/audits/1234/media/12345").
		Reply(404).
		JSON(`{"error": "something bad happened"}`)

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	_, err := apiClient.GetMedia(
		context.Background(),
		&api.GetMediaRequest{
			URL:     "http://localhost:9999/audits/1234/media/12345",
			AuditID: "1234",
		},
	)
	assert.NoError(t, err)
}

func TestGetMediaWith405Error(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/audits/1234/media/12345").
		Reply(405).
		JSON(`{"error": "something bad happened"}`)

	apiClient := api.GetTestClient()
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

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	resp, err := apiClient.GetMedia(
		context.Background(),
		&api.GetMediaRequest{
			URL:     "http://localhost:9999/audits/1234/media/12345",
			AuditID: "1234",
		},
	)
	assert.NoError(t, err)
	assert.Nil(t, resp)
}

func TestGetMediaWithNoContentType(t *testing.T) {
	defer gock.Off()

	result := `{id:"test-id"}`
	gock.New("http://localhost:9999").
		Get("/audits/1234/media/12345").
		Reply(200).
		BodyString(result)

	apiClient := api.GetTestClient()
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
	gock.New("http://localhost:9999").
		Get("/audits/1234/media/12345").
		Reply(200).
		BodyString(result).
		SetHeader("Content-Type", "test-content")

	apiClient := api.GetTestClient()
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
	assert.NoError(t, err)
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

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	mId, err := apiClient.InitiateInspectionReportExport(context.Background(), "audit_123", "PDF", "p123")

	assert.NoError(t, err)
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
			apiClient := api.GetTestClient()
			gock.InterceptClient(apiClient.HTTPClient())

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

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	res, err := apiClient.CheckInspectionReportExportCompletion(context.Background(), "audit_123", "abc")

	assert.NoError(t, err)
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
			apiClient := api.GetTestClient()
			gock.InterceptClient(apiClient.HTTPClient())

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

	apiClient := api.GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())

	res, err := apiClient.DownloadInspectionReportFile(context.Background(), "http://localhost:9999/report-exports/abc")

	assert.NoError(t, err)

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
			apiClient := api.GetTestClient()
			gock.InterceptClient(apiClient.HTTPClient())

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
	req.SetHeader("X-RateLimit-Reset", "1")

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
			apiClient := api.GetTestClient()
			gock.InterceptClient(apiClient.HTTPClient())
			apiClient.RetryMax = 1

			_, err := apiClient.GetInspection(context.Background(), "1234")
			if err == nil || !strings.HasSuffix(err.Error(), tt.err) {
				t.Fatalf("expected giving up error, got: %#v", err)
			}
		})
	}
}
