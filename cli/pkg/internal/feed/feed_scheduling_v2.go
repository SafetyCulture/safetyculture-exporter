package feed

import (
	"time"
)

// The feeds in this file use the Scheduling feed service (/scheduling/v1/feed/*) which replaces the legacy
// /feed/schedules* endpoints for organisations on the new Schedules experience.
// They export to separate *_v2 tables so organisations that are still on the legacy experience keep working.

// maxSchedulingFeedLimit is the largest page size the Scheduling feed service accepts (the default is 50)
const maxSchedulingFeedLimit = 1000

// ScheduleV2 represents a row from the Scheduling feed service schedules feed
type ScheduleV2 struct {
	ID                         string    `json:"id" csv:"schedule_id" gorm:"primarykey;column:schedule_id;size:128"`
	Title                      string    `json:"title" csv:"title"`
	Recurrence                 string    `json:"recurrence" csv:"recurrence"`
	Duration                   string    `json:"duration" csv:"duration" gorm:"size:50"`
	ModifiedAt                 time.Time `json:"modified_at" csv:"modified_at"`
	OrganisationID             string    `json:"organisation_id" csv:"organisation_id" gorm:"size:37"`
	CreatedAt                  time.Time `json:"created_at" csv:"created_at"`
	CompletionRule             string    `json:"completion_rule" csv:"completion_rule" gorm:"size:10"`
	Status                     string    `json:"status" csv:"status" gorm:"size:20"`
	Timezone                   string    `json:"timezone" csv:"timezone"`
	TimezoneType               string    `json:"timezone_type" csv:"timezone_type" gorm:"size:30"`
	IsLateSubmitAllowed        bool      `json:"is_late_submit_allowed" csv:"is_late_submit_allowed"`
	SiteIDs                    JSONArray `json:"site_ids" csv:"site_ids"`
	TemplateID                 string    `json:"template_id" csv:"template_id" gorm:"size:100"`
	CreatedBy                  string    `json:"created_by" csv:"created_by" gorm:"size:37"`
	GracePeriodDuration        string    `json:"grace_period_duration" csv:"grace_period_duration" gorm:"size:50"`
	SiteBasedAssignmentEnabled bool      `json:"site_based_assignment_enabled" csv:"site_based_assignment_enabled"`
	EndingRuleType             string    `json:"ending_rule_type" csv:"ending_rule_type" gorm:"size:20"`
	EndingRuleEndDate          string    `json:"ending_rule_end_date" csv:"ending_rule_end_date"`
	EndingRuleEndAfterRepeats  *int32    `json:"ending_rule_end_after_repeats" csv:"ending_rule_end_after_repeats"`
	Assignees                  JSONArray `json:"assignees" csv:"assignees"`
	Intersections              JSONArray `json:"intersections" csv:"intersections"`
	ScheduleType               string    `json:"schedule_type" csv:"schedule_type" gorm:"size:30"`
	AssetIDs                   JSONArray `json:"asset_ids" csv:"asset_ids"`
}

// TableName makes the table match the name of the feed. By default gorm derives it from the type name
func (ScheduleV2) TableName() string {
	return "schedules_v2"
}

// ScheduleAssigneeV2 represents a row from the Scheduling feed service schedule_assignees feed
type ScheduleAssigneeV2 struct {
	ID             string `json:"id" csv:"id" gorm:"primarykey;size:375"`
	ScheduleID     string `json:"schedule_id" csv:"schedule_id" gorm:"size:128"`
	AssigneeID     string `json:"assignee_id" csv:"assignee_id" gorm:"size:256"`
	Type           string `json:"type" csv:"type" gorm:"size:30"`
	Name           string `json:"name" csv:"name"`
	OrganisationID string `json:"organisation_id" csv:"organisation_id" gorm:"size:37"`
	CompletionRule string `json:"completion_rule" csv:"completion_rule" gorm:"size:10"`
}

// TableName makes the table match the name of the feed. By default gorm derives it from the type name
func (ScheduleAssigneeV2) TableName() string {
	return "schedule_assignees_v2"
}

// ScheduleOccurrenceV2 represents a row from the Scheduling feed service schedule_occurrences feed
type ScheduleOccurrenceV2 struct {
	ID               string `json:"id" csv:"id" gorm:"primarykey;size:375"`
	ScheduleID       string `json:"schedule_id" csv:"schedule_id" gorm:"size:128"`
	OccurrenceID     string `json:"occurrence_id" csv:"occurrence_id" gorm:"size:128"`
	TemplateID       string `json:"template_id" csv:"template_id" gorm:"size:100"`
	StartTime        string `json:"start_time" csv:"start_time" gorm:"size:40"`
	DueTime          string `json:"due_time" csv:"due_time" gorm:"size:40"`
	MissTime         string `json:"miss_time" csv:"miss_time" gorm:"size:40"`
	OccurrenceStatus string `json:"occurrence_status" csv:"occurrence_status" gorm:"size:20"`
	AuditID          string `json:"audit_id" csv:"audit_id" gorm:"size:100"`
	CompletedAt      string `json:"completed_at" csv:"completed_at" gorm:"size:40"`
	CompletionRule   string `json:"completion_rule" csv:"completion_rule" gorm:"size:10"`
	OrganisationID   string `json:"organisation_id" csv:"organisation_id" gorm:"size:37"`
	WontDoNote       string `json:"wont_do_note" csv:"wont_do_note"`
	AssigneeID       string `json:"assignee_id" csv:"assignee_id" gorm:"size:256"`
	AssigneeFrom     string `json:"assignee_from" csv:"assignee_from" gorm:"size:300"`
}

// TableName makes the table match the name of the feed. By default gorm derives it from the type name
func (ScheduleOccurrenceV2) TableName() string {
	return "schedule_occurrences_v2"
}

// NewScheduleV2Feed returns the feed for /scheduling/v1/feed/schedules
func NewScheduleV2Feed(templateIDs []string) Feed {
	return &rowFeed[ScheduleV2]{
		name:         "schedules_v2",
		initialURL:   "/scheduling/v1/feed/schedules",
		primaryKey:   []string{"schedule_id"},
		order:        "schedule_id",
		hasRemaining: false, // FeedMetadataSimple has no remaining_records
		params: GetFeedParams{
			TemplateIDs: templateIDs,
			Limit:       maxSchedulingFeedLimit,
		},
		key: func(r *ScheduleV2) string { return r.ID },
		columns: []string{
			"title", "recurrence", "duration", "modified_at", "organisation_id", "created_at", "completion_rule",
			"status", "timezone", "timezone_type", "is_late_submit_allowed", "site_ids", "template_id", "created_by",
			"grace_period_duration", "site_based_assignment_enabled", "ending_rule_type", "ending_rule_end_date",
			"ending_rule_end_after_repeats", "assignees", "intersections", "schedule_type", "asset_ids",
		},
	}
}

// NewScheduleAssigneeV2Feed returns the feed for /scheduling/v1/feed/schedule_assignees
func NewScheduleAssigneeV2Feed(templateIDs []string) Feed {
	return &rowFeed[ScheduleAssigneeV2]{
		name:       "schedule_assignees_v2",
		initialURL: "/scheduling/v1/feed/schedule_assignees",
		primaryKey: []string{"id"},
		order:      "schedule_id, assignee_id",
		params: GetFeedParams{
			TemplateIDs: templateIDs,
			Limit:       maxSchedulingFeedLimit,
		},
		key:     func(r *ScheduleAssigneeV2) string { return r.ID },
		columns: []string{"schedule_id", "assignee_id", "type", "name", "organisation_id", "completion_rule"},
	}
}

// NewScheduleOccurrenceV2Feed returns the feed for /scheduling/v1/feed/schedule_occurrences.
// startDate filters on the due time of the occurrences. When it is zero the API default window applies (5 months back, 1 month ahead).
func NewScheduleOccurrenceV2Feed(templateIDs []string, startDate time.Time) Feed {
	return &rowFeed[ScheduleOccurrenceV2]{
		name:       "schedule_occurrences_v2",
		initialURL: "/scheduling/v1/feed/schedule_occurrences",
		primaryKey: []string{"id"},
		order:      "schedule_id, due_time",
		params: GetFeedParams{
			TemplateIDs: templateIDs,
			StartDate:   startDate,
			Limit:       maxSchedulingFeedLimit,
		},
		key: func(r *ScheduleOccurrenceV2) string { return r.ID },
		columns: []string{
			"schedule_id", "occurrence_id", "template_id", "start_time", "due_time", "miss_time", "occurrence_status",
			"audit_id", "completed_at", "completion_rule", "organisation_id", "wont_do_note", "assignee_id", "assignee_from",
		},
	}
}
