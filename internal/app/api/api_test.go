package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	apiClient := api.NewAPIClient("http://localhost:9999", "abc123")
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

	apiClient := api.NewAPIClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	expectedErr := errors.New("test error")
	err := apiClient.DrainFeed(context.Background(), &api.GetFeedRequest{
		InitialURL: "/feed/inspections",
	}, func(data *api.GetFeedResponse) error {
		return expectedErr
	})
	assert.Equal(t, expectedErr, err)
}

func TestApiOptSetTimeout_should_set_timeout(t *testing.T) {
	apiClient := api.NewAPIClient("http://localhost:9999", "abc123", api.OptSetTimeout(time.Second*21))

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

	apiClient := api.NewAPIClient("http://localhost:9999", "abc123")
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

	apiClient := api.NewAPIClient("http://localhost:9999", "abc123")
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

	apiClient := api.NewAPIClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	_, err := apiClient.GetInspection(context.Background(), auditID)
	assert.NotNil(t, err)
}

func TestAPIClientListInspectionWithError(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/audits/search").
		ReplyError(fmt.Errorf("test error"))

	apiClient := api.NewAPIClient("http://localhost:9999", "abc123")
	gock.InterceptClient(apiClient.HTTPClient())

	_, err := apiClient.ListInspections(context.Background(), nil)
	assert.NotNil(t, err)
}

func TestDrainInspectionsWithAPIError(t *testing.T) {
	defer gock.Off()

	gock.New("http://localhost:9999").
		Get("/audits/search").
		ReplyError(fmt.Errorf("test error"))

	apiClient := api.NewAPIClient("http://localhost:9999", "abc123")
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

	apiClient := api.NewAPIClient("http://localhost:9999", "abc123")
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

	apiClient := api.NewAPIClient("http://localhost:9999", "abc123")
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
