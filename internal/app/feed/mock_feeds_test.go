package feed_test

import (
	"net/http"

	"gopkg.in/h2non/gock.v1"
)

func initMockFeedsSet1(httpClient *http.Client) {
	gock.InterceptClient(httpClient)

	gock.New("http://localhost:9999").
		Get("/feed/inspections").
		Reply(200).
		File("mocks/set_1/feed_inspections_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/inspections/2").
		Reply(200).
		File("mocks/set_1/feed_inspections_2.json")

	gock.New("http://localhost:9999").
		Get("/feed/inspection_items").
		Reply(200).
		File("mocks/set_1/feed_inspection_items_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/users").
		Reply(200).
		File("mocks/set_1/feed_users_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/groups").
		Reply(200).
		File("mocks/set_1/feed_groups_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/group_users").
		Reply(200).
		File("mocks/set_1/feed_group_users_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/sites").
		Reply(200).
		File("mocks/set_1/feed_sites_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/templates").
		Reply(200).
		File("mocks/set_1/feed_templates_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/template_permissions").
		Reply(200).
		File("mocks/set_1/feed_template_permissions_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/schedules").
		Reply(200).
		File("mocks/set_1/feed_schedules_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/schedule_assignees").
		Reply(200).
		File("mocks/set_1/feed_schedule_assignees_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/schedule_occurrences").
		Reply(200).
		File("mocks/set_1/feed_schedule_occurrences_1.json")
}

func initMockFeedsSet2(httpClient *http.Client) {
	gock.InterceptClient(httpClient)

	gock.New("http://localhost:9999").
		Get("/feed/inspections").
		Reply(200).
		File("mocks/set_1/feed_inspections_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/inspections/2").
		Reply(200).
		File("mocks/set_1/feed_inspections_2.json")

	gock.New("http://localhost:9999").
		Get("/feed/inspection_items").
		Reply(200).
		File("mocks/set_1/feed_inspection_items_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/users").
		Reply(200).
		File("mocks/set_1/feed_users_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/groups").
		Reply(200).
		File("mocks/set_1/feed_groups_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/sites").
		Reply(200).
		File("mocks/set_1/feed_sites_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/templates").
		Reply(200).
		File("mocks/set_1/feed_templates_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/template_permissions").
		Reply(200).
		File("mocks/set_1/feed_template_permissions_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/schedules").
		Reply(200).
		File("mocks/set_1/feed_schedules_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/schedule_occurrences").
		Reply(200).
		File("mocks/set_1/feed_schedule_occurrences_1.json")

	// Set 2 is just set 1 with different group users and schedule assignees
	gock.New("http://localhost:9999").
		Get("/feed/group_users").
		Reply(200).
		File("mocks/set_2/feed_group_users_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/schedule_assignees").
		Reply(200).
		File("mocks/set_2/feed_schedule_assignees_1.json")
}

func initMockFeedsSet3(httpClient *http.Client) {
	gock.InterceptClient(httpClient)

	gock.New("http://localhost:9999").
		Get("/feed/inspections").
		Reply(200).
		File("mocks/set_3/feed_inspections_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/inspection_items").
		Reply(200).
		File("mocks/set_3/feed_inspection_items_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/users").
		Reply(200).
		File("mocks/set_1/feed_users_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/groups").
		Reply(200).
		File("mocks/set_1/feed_groups_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/group_users").
		Reply(200).
		File("mocks/set_1/feed_group_users_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/sites").
		Reply(200).
		File("mocks/set_1/feed_sites_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/templates").
		Reply(200).
		File("mocks/set_1/feed_templates_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/template_permissions").
		Reply(200).
		File("mocks/set_1/feed_template_permissions_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/schedules").
		Reply(200).
		File("mocks/set_1/feed_schedules_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/schedule_assignees").
		Reply(200).
		File("mocks/set_1/feed_schedule_assignees_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/schedule_occurrences").
		Reply(200).
		File("mocks/set_1/feed_schedule_occurrences_1.json")
}

func resetMocks(httpClient *http.Client) {
	gock.Off()
	gock.Clean()
	gock.RestoreClient(httpClient)
}
