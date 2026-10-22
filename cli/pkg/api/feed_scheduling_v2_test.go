package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/SafetyCulture/mitti-exporter/pkg/internal/feed"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestScheduleV2Feed_should_follow_pagination_and_keep_arrays_as_json(t *testing.T) {
	defer gock.Off()

	exporter, err := getInmemorySQLExporter("")
	assert.NoError(t, err)

	apiClient := GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())
	gock.New("http://localhost:9999").Get("/scheduling/v1/feed/schedules").
		MatchParam("limit", "1000").
		Reply(200).
		BodyString(`{"metadata":{"next_page":"/scheduling/v1/feed/schedules/page2","next_page_token":"t"},"data":[{"id":"scheduleitem_1","title":"One","modified_at":"2026-01-01T00:00:00Z","created_at":"2026-01-01T00:00:00Z","site_ids":["site_1","site_2"],"assignees":["user_1"],"ending_rule_end_after_repeats":5}]}`)
	gock.New("http://localhost:9999").Get("/scheduling/v1/feed/schedules/page2").
		Reply(200).
		BodyString(`{"metadata":{},"data":[{"id":"scheduleitem_2","title":"Two","modified_at":"2026-01-01T00:00:00Z","created_at":"2026-01-01T00:00:00Z"}]}`)

	assert.NoError(t, feed.NewScheduleV2Feed(nil).Export(context.Background(), apiClient, exporter, ""))

	var rows []feed.ScheduleV2
	assert.NoError(t, exporter.DB.Table("schedules_v2").Order("schedule_id").Scan(&rows).Error)
	assert.Len(t, rows, 2)
	assert.Equal(t, feed.JSONArray(`["site_1","site_2"]`), rows[0].SiteIDs)
	assert.Equal(t, feed.JSONArray(`["user_1"]`), rows[0].Assignees)
	assert.Equal(t, int32(5), *rows[0].EndingRuleEndAfterRepeats)
	assert.Nil(t, rows[1].EndingRuleEndAfterRepeats)
	assert.Equal(t, feed.JSONArray(""), rows[1].SiteIDs)
}

func TestSchedulingV2Feeds_should_export_every_column_and_a_row_per_target(t *testing.T) {
	defer gock.Off()

	exporter, err := getInmemorySQLExporter("")
	assert.NoError(t, err)

	apiClient := GetTestClient()
	gock.InterceptClient(apiClient.HTTPClient())
	gock.New("http://localhost:9999").Get("/scheduling/v1/feed/schedules").
		Reply(200).
		BodyString(`{"metadata":{},"data":[{"id":"scheduleitem_1","title":"Daily check","recurrence":"DTSTART;TZID=Australia/Sydney:20260928T093000\nRRULE:FREQ=DAILY","duration":"PT8H","modified_at":"2026-09-28T05:59:08Z","organisation_id":"role_org","created_at":"2026-09-28T05:59:08Z","completion_rule":"all","status":"ACTIVE","timezone":"Australia/Sydney","timezone_type":"CUSTOM_TIMEZONE","is_late_submit_allowed":true,"site_ids":["location_75fc77d9a3b14c5e9f0e1a2b3c4d5e6f"],"template_id":"template_1","created_by":"user_1","grace_period_duration":"until_next","site_based_assignment_enabled":true,"ending_rule_type":"end_date","ending_rule_end_date":"2026-12-31T00:00:00Z","assignees":["role_1"],"intersections":[],"schedule_type":"SCHEDULE_TYPE_SITES","asset_ids":[]}]}`)
	gock.New("http://localhost:9999").Get("/scheduling/v1/feed/schedule_assignees").
		Reply(200).
		BodyString(`{"metadata":{},"data":[
			{"id":"scheduleitem_1_role_1","schedule_id":"scheduleitem_1","assignee_id":"role_1","type":"role","name":"Cleaners","organisation_id":"role_org","completion_rule":"all","target_id":"site_1","target_type":"SITE"},
			{"id":"scheduleitem_1_role_1","schedule_id":"scheduleitem_1","assignee_id":"role_1","type":"role","name":"Cleaners","organisation_id":"role_org","completion_rule":"all","target_id":"site_2","target_type":"SITE"}
		]}`)
	gock.New("http://localhost:9999").Get("/feed/users").
		Reply(200).
		BodyString(`{"metadata":{},"data":[{"id":"user_1","firstname":"Ada","lastname":"Lovelace"},{"id":"user_2","firstname":"Alan","lastname":"Turing"}]}`)
	gock.New("http://localhost:9999").Get("/feed/groups").
		Reply(200).
		BodyString(`{"metadata":{},"data":[{"id":"role_1","name":"Cleaners"}]}`)
	gock.New("http://localhost:9999").Get("/feed/sites").
		Reply(200).
		BodyString(`{"metadata":{},"data":[{"id":"location_1","site_uuid":"site_1","name":"Sydney"},{"id":"location_2","site_uuid":"site_2","name":"Melbourne"}]}`)
	gock.New("http://localhost:9999").Get("/feed/assets").
		Reply(200).
		BodyString(`{"metadata":{},"data":[]}`)
	gock.New("http://localhost:9999").Get("/scheduling/v1/feed/schedule_occurrences").
		Reply(200).
		BodyString(`{"metadata":{},"data":[
			{"id":"scheduleitem_1_occurrence_1_user_1","schedule_id":"scheduleitem_1","occurrence_id":"1","template_id":"template_1","start_time":"2026-09-28T09:30:00Z","due_time":"2026-09-28T17:30:00Z","occurrence_status":"COMPLETED","audit_id":"audit_1","completed_at":"2026-09-28T10:00:00Z","completion_rule":"all","organisation_id":"role_org","assignee_id":"user_1","assignee_from":"location_1+role_1","target_id":"site_1","target_type":"SITE","assignee_status":"COMPLETED","completed_by_user_id":"user_2"},
			{"id":"scheduleitem_1_occurrence_1_user_1","schedule_id":"scheduleitem_1","occurrence_id":"1","template_id":"template_1","start_time":"2026-09-28T09:30:00Z","due_time":"2026-09-28T17:30:00Z","occurrence_status":"WONT_DO","completion_rule":"all","organisation_id":"role_org","wont_do_note":"closed","assignee_id":"user_1","assignee_from":"role_1","target_id":"site_2","target_type":"SITE","assignee_status":"WONT_DO"}
		]}`)

	assert.NoError(t, feed.NewScheduleV2Feed(nil).Export(context.Background(), apiClient, exporter, ""))
	assert.NoError(t, feed.NewScheduleAssigneeV2Feed(nil).Export(context.Background(), apiClient, exporter, ""))
	assert.NoError(t, feed.NewScheduleOccurrenceV2Feed(nil, time.Time{}).Export(context.Background(), apiClient, exporter, ""))

	var schedules []feed.ScheduleV2
	assert.NoError(t, exporter.DB.Table("schedules_v2").Scan(&schedules).Error)
	assert.Len(t, schedules, 1)
	schedule := schedules[0]
	sydney, err := time.LoadLocation("Australia/Sydney")
	assert.NoError(t, err)
	assert.True(t, time.Date(2026, 9, 28, 9, 30, 0, 0, sydney).Equal(*schedule.FromDate))
	assert.True(t, time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC).Equal(*schedule.ToDate))
	assert.False(t, schedule.ExportedAt.IsZero())
	schedule.FromDate, schedule.ToDate, schedule.ExportedAt = nil, nil, time.Time{}
	schedule.ModifiedAt, schedule.CreatedAt = schedule.ModifiedAt.UTC(), schedule.CreatedAt.UTC()
	assert.Equal(t, feed.ScheduleV2{
		ID:                         "scheduleitem_1",
		Title:                      "Daily check",
		Recurrence:                 "DTSTART;TZID=Australia/Sydney:20260928T093000\nRRULE:FREQ=DAILY",
		Duration:                   "PT8H",
		ModifiedAt:                 time.Date(2026, 9, 28, 5, 59, 8, 0, time.UTC),
		OrganisationID:             "role_org",
		CreatedAt:                  time.Date(2026, 9, 28, 5, 59, 8, 0, time.UTC),
		CompletionRule:             "all",
		Status:                     "ACTIVE",
		Timezone:                   "Australia/Sydney",
		TimezoneType:               "CUSTOM_TIMEZONE",
		IsLateSubmitAllowed:        true,
		SiteIDs:                    `["location_75fc77d9a3b14c5e9f0e1a2b3c4d5e6f"]`,
		TemplateID:                 "template_1",
		CreatedBy:                  "user_1",
		GracePeriodDuration:        "until_next",
		SiteBasedAssignmentEnabled: true,
		EndingRuleType:             "end_date",
		EndingRuleEndDate:          "2026-12-31T00:00:00Z",
		Assignees:                  `["role_1"]`,
		Intersections:              `[]`,
		ScheduleType:               "SCHEDULE_TYPE_SITES",
		AssetIDs:                   `[]`,
		Description:                "Daily check",
		StartTimeHour:              9,
		StartTimeMinute:            30,
		AllMustComplete:            true,
		CanLateSubmit:              true,
		SiteID:                     "location_75fc77d9a3b14c5e9f0e1a2b3c4d5e6f",
		CreatorUserID:              "user_1",
		TargetType:                 "SITE",
		TargetIDs:                  "75fc77d9-a3b1-4c5e-9f0e-1a2b3c4d5e6f",
	}, schedule)

	var assignees []feed.ScheduleAssigneeV2
	assert.NoError(t, exporter.DB.Table("schedule_assignees_v2").Order("target_id").Scan(&assignees).Error)
	assigneeOf := func(targetID string) feed.ScheduleAssigneeV2 {
		return feed.ScheduleAssigneeV2{
			ID: "scheduleitem_1_role_1", ScheduleID: "scheduleitem_1", AssigneeID: "role_1", Type: "role", Name: "Cleaners",
			OrganisationID: "role_org", CompletionRule: "all", TargetID: targetID, TargetType: "SITE",
		}
	}
	assert.Equal(t, []feed.ScheduleAssigneeV2{assigneeOf("site_1"), assigneeOf("site_2")}, assignees)

	var occurrences []feed.ScheduleOccurrenceV2
	assert.NoError(t, exporter.DB.Table("schedule_occurrences_v2").Order("target_id").Scan(&occurrences).Error)
	assert.Equal(t, []feed.ScheduleOccurrenceV2{
		{
			ID: "scheduleitem_1_occurrence_1_user_1", ScheduleID: "scheduleitem_1", OccurrenceID: "1", TemplateID: "template_1",
			StartTime: "2026-09-28T09:30:00Z", DueTime: "2026-09-28T17:30:00Z", OccurrenceStatus: "COMPLETED", AuditID: "audit_1",
			CompletedAt: "2026-09-28T10:00:00Z", CompletionRule: "all", OrganisationID: "role_org", AssigneeID: "user_1",
			AssigneeFrom: "location_1+role_1", TargetID: "site_1", TargetType: "SITE", AssigneeStatus: "COMPLETED",
			CompletedByUserID: "user_2", UserID: "user_1", AssigneeType: "location_intersection", AssigneeSiteID: "location_1",
			AssigneeGroupID: "role_1", UserName: "Ada Lovelace", AssigneeSiteName: "Sydney", AssigneeGroupName: "Cleaners",
			CompletedByUserName: "Alan Turing", TargetName: "Sydney",
		},
		{
			ID: "scheduleitem_1_occurrence_1_user_1", ScheduleID: "scheduleitem_1", OccurrenceID: "1", TemplateID: "template_1",
			StartTime: "2026-09-28T09:30:00Z", DueTime: "2026-09-28T17:30:00Z", OccurrenceStatus: "WONT_DO",
			CompletionRule: "all", OrganisationID: "role_org", WontDoNote: "closed", AssigneeID: "user_1", AssigneeFrom: "role_1",
			TargetID: "site_2", TargetType: "SITE", AssigneeStatus: "WONT_DO", UserID: "user_1", Note: "closed",
			AssigneeType: "role", AssigneeGroupID: "role_1", UserName: "Ada Lovelace", AssigneeGroupName: "Cleaners",
			TargetName: "Melbourne",
		},
	}, occurrences)
}
