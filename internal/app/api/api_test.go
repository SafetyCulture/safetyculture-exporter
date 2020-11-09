package api_test

import (
	"context"
	"encoding/json"
	"errors"
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
