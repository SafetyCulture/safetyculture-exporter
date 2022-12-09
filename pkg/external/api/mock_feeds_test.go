package api_test

import (
	"net/http"
	"path"

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
		Get("/feed/site_members").
		Reply(200).
		File("mocks/set_1/feed_site_members_1.json")

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

	gock.New("http://localhost:9999").
		Get("/feed/actions").
		Reply(200).
		File("mocks/set_1/feed_actions_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/action_assignees").
		Reply(200).
		File("mocks/set_1/feed_action_assignees_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/issues").
		Reply(200).
		File("mocks/set_1/feed_issues_1.json")

	gock.New("http://localhost:9999/feed/issues?limit=20&next_page_token=QGI5Nzk5NzMtZWMxMy00YzVmLTk1NDQtMTI0YjQ1M2I2OWYp").
		Get("/feed/issues").
		Reply(200).
		File("mocks/set_1/feed_issues_2.json")

	gock.New("http://localhost:9999").
		Post("/accounts/history/v1/activity_log/list").
		BodyString(`{"org_id":"","page_size":0,"page_token":"","filters":{"timeframe":{"from":"0001-01-01T00:00:00Z"},"event_types":["inspection.deleted"],"limit":0}}`).
		Reply(http.StatusOK).
		File(path.Join("mocks", "set_1", "inspections_deleted_single_page.json"))

	gock.New("http://localhost:9999").
		Persist().
		Get("/SheqsyIntegrationApi/api/v3/companies/ada3042f-16a4-4249-915d-dc088adef92a/employees").
		Reply(200).
		File("mocks/set_1/sheqsy_employees_1.json")

	gock.New("http://localhost:9999").
		Get("/SheqsyIntegrationApi/api/v3/companies/ada3042f-16a4-4249-915d-dc088adef92a/departments").
		Reply(200).
		File("mocks/set_1/sheqsy_departments_1.json")

	gock.New("http://localhost:9999").
		Get("/SheqsyIntegrationApi/api/v3/companies/ada3042f-16a4-4249-915d-dc088adef92a/shifts/history").
		Reply(200).
		File("mocks/set_1/sheqsy_shifts_1.json")

	gock.New("http://localhost:9999").
		Get("/SheqsyIntegrationApi/api/v3/companies/ada3042f-16a4-4249-915d-dc088adef92a/activities/history").
		Reply(200).
		File("mocks/set_1/sheqsy_activities_1.json")

	gock.New("http://localhost:9999").
		Get("/SheqsyIntegrationApi/api/v3/companies/ada3042f-16a4-4249-915d-dc088adef92a/shifts/history").
		MatchParam("ver", "37381147").
		Reply(200).
		File("mocks/set_1/sheqsy_shifts_2.json")

	gock.New("http://localhost:9999").
		Get("/SheqsyIntegrationApi/api/v3/companies/ada3042f-16a4-4249-915d-dc088adef92a/activities/history").
		MatchParam("ver", "36384957").
		Reply(200).
		File("mocks/set_1/sheqsy_activities_2.json")

	gock.New("http://localhost:9999").
		Get("/feed/assets").
		Reply(200).
		File("mocks/set_1/feed_assets_1.json")
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
		Get("/feed/site_members").
		Reply(200).
		File("mocks/set_1/feed_site_members_1.json")

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

	gock.New("http://localhost:9999").
		Get("/feed/actions").
		Reply(200).
		File("mocks/set_1/feed_actions_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/action_assignees").
		Reply(200).
		File("mocks/set_2/feed_action_assignees_2.json")

	gock.New("http://localhost:9999").
		Get("/feed/issues").
		Reply(200).
		File("mocks/set_1/feed_issues_1.json")

	gock.New("http://localhost:9999/feed/issues?limit=20&next_page_token=QGI5Nzk5NzMtZWMxMy00YzVmLTk1NDQtMTI0YjQ1M2I2OWYp").
		Get("/feed/issues").
		Reply(200).
		File("mocks/set_1/feed_issues_2.json")

	gock.New("http://localhost:9999").
		Get("/feed/assets").
		Reply(200).
		File("mocks/set_1/feed_assets_1.json")
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
		Get("/feed/site_members").
		Reply(200).
		File("mocks/set_1/feed_site_members_1.json")

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

	gock.New("http://localhost:9999").
		Get("/feed/actions").
		Reply(200).
		File("mocks/set_1/feed_actions_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/action_assignees").
		Reply(200).
		File("mocks/set_1/feed_action_assignees_1.json")

	gock.New("http://localhost:9999").
		Get("/feed/issues").
		Reply(200).
		File("mocks/set_1/feed_issues_1.json")

	gock.New("http://localhost:9999/feed/issues?limit=20&next_page_token=QGI5Nzk5NzMtZWMxMy00YzVmLTk1NDQtMTI0YjQ1M2I2OWYp").
		Get("/feed/issues").
		Reply(200).
		File("mocks/set_1/feed_issues_2.json")

	gock.New("http://localhost:9999").
		Get("/feed/assets").
		Reply(200).
		File("mocks/set_1/feed_assets_1.json")
}

func initMockIssuesFeed(httpClient *http.Client) {
	gock.InterceptClient(httpClient)

	gock.New("http://localhost:9999").
		Get("/feed/issues").
		Reply(200).
		File("mocks/set_1/feed_issues_1.json")

	gock.New("http://localhost:9999/feed/issues?limit=20&next_page_token=QGI5Nzk5NzMtZWMxMy00YzVmLTk1NDQtMTI0YjQ1M2I2OWYp").
		Get("/feed/issues").
		Reply(200).
		File("mocks/set_1/feed_issues_2.json")
}

func resetMocks(httpClient *http.Client) {
	gock.Off()
	gock.Clean()
	gock.RestoreClient(httpClient)
}
