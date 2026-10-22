package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/SafetyCulture/mitti-exporter/pkg/httpapi"
	"github.com/SafetyCulture/mitti-exporter/pkg/logger"
	"github.com/gofrs/uuid"
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

	// The fields below are derived from the API fields: the legacy schedules columns, then the targets
	Description     string     `json:"-" csv:"description"`
	ExportedAt      time.Time  `json:"-" csv:"exported_at" gorm:"autoUpdateTime"`
	FromDate        *time.Time `json:"-" csv:"from_date"`
	ToDate          *time.Time `json:"-" csv:"to_date"`
	StartTimeHour   int        `json:"-" csv:"start_time_hour"`
	StartTimeMinute int        `json:"-" csv:"start_time_minute"`
	AllMustComplete bool       `json:"-" csv:"all_must_complete"`
	CanLateSubmit   bool       `json:"-" csv:"can_late_submit"`
	SiteID          string     `json:"-" csv:"site_id" gorm:"size:41"`
	CreatorUserID   string     `json:"-" csv:"creator_user_id" gorm:"size:37"`
	AssetID         string     `json:"-" csv:"asset_id" gorm:"size:37"`
	TargetType      string     `json:"-" csv:"target_type" gorm:"size:10"`
	// TargetIDs is the comma-joined list of the site or asset UUIDs of the schedule
	TargetIDs string `json:"-" csv:"target_ids"`
}

// UnmarshalJSON decodes the API fields and fills the derived fields. A derived field that can't be parsed is left empty
func (s *ScheduleV2) UnmarshalJSON(b []byte) error {
	type apiSchedule ScheduleV2
	if err := json.Unmarshal(b, (*apiSchedule)(s)); err != nil {
		return err
	}

	var ids struct {
		SiteIDs  []string `json:"site_ids"`
		AssetIDs []string `json:"asset_ids"`
	}
	if err := json.Unmarshal(b, &ids); err != nil {
		return err
	}

	s.Description = s.Title
	s.AllMustComplete = s.CompletionRule == "all"
	s.CanLateSubmit = s.IsLateSubmitAllowed
	s.CreatorUserID = s.CreatedBy
	if len(ids.SiteIDs) == 1 {
		s.SiteID = ids.SiteIDs[0]
	}
	if len(ids.AssetIDs) == 1 {
		s.AssetID = ids.AssetIDs[0]
	}

	switch s.ScheduleType {
	case "SCHEDULE_TYPE_SITES":
		s.TargetType = "SITE"
		s.TargetIDs = strings.Join(siteUUIDs(ids.SiteIDs), ",")
	case "SCHEDULE_TYPE_ASSETS":
		s.TargetType = "ASSET"
		s.TargetIDs = strings.Join(ids.AssetIDs, ",")
	}

	if toDate, err := time.Parse(time.RFC3339, s.EndingRuleEndDate); err == nil {
		s.ToDate = &toDate
	}
	if fromDate := parseDTStart(s.Recurrence, s.Timezone); fromDate != nil {
		s.FromDate = fromDate
		s.StartTimeHour = fromDate.Hour()
		s.StartTimeMinute = fromDate.Minute()
	}
	return nil
}

// siteUUIDs converts site IDs from the location_<hex> form to UUIDs, the form of target_id.
// An ID that isn't in that form is kept as it is.
func siteUUIDs(siteIDs []string) []string {
	result := make([]string, len(siteIDs))
	for i, siteID := range siteIDs {
		result[i] = siteID
		hex, found := strings.CutPrefix(siteID, "location_")
		if !found || len(hex) != 32 {
			continue
		}
		if u, err := uuid.FromString(hex); err == nil {
			result[i] = u.String()
		}
	}
	return result
}

// parseDTStart returns the DTSTART of a recurrence in its local time zone, or nil when there is none or it can't be parsed.
// A DTSTART without TZID or a Z suffix is in the time zone of the schedule.
func parseDTStart(recurrence, timezone string) *time.Time {
	for _, line := range strings.Split(recurrence, "\n") {
		rest, found := strings.CutPrefix(strings.TrimSpace(line), "DTSTART")
		if !found {
			continue
		}
		params, value, ok := strings.Cut(rest, ":")
		if !ok {
			return nil
		}

		tzName := timezone
		if tzid, found := strings.CutPrefix(params, ";TZID="); found {
			tzName = tzid
		}
		loc, err := time.LoadLocation(tzName)
		if err != nil {
			return nil
		}

		layout := "20060102T150405"
		if strings.HasSuffix(value, "Z") {
			layout += "Z"
			loc = time.UTC
		}
		t, err := time.ParseInLocation(layout, value, loc)
		if err != nil {
			return nil
		}
		return &t
	}
	return nil
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
	// TargetID is the site or asset UUID of the row, empty for schedules without sites or assets
	TargetID   string `json:"target_id" csv:"target_id" gorm:"primarykey;size:36"`
	TargetType string `json:"target_type" csv:"target_type" gorm:"size:10"`
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
	// TargetID is the site or asset UUID of the row, empty for schedules without sites or assets
	TargetID          string `json:"target_id" csv:"target_id" gorm:"primarykey;size:36"`
	TargetType        string `json:"target_type" csv:"target_type" gorm:"size:10"`
	AssigneeStatus    string `json:"assignee_status" csv:"assignee_status" gorm:"size:20"`
	CompletedByUserID string `json:"completed_by_user_id" csv:"completed_by_user_id" gorm:"size:37"`

	// The fields below are derived from the API fields: the legacy schedule_occurrences columns, then the assignment
	UserID string `json:"-" csv:"user_id" gorm:"size:256"`
	Note   string `json:"-" csv:"note"`
	// AssigneeType is how the user was assigned, with the type values of the schedule assignees feed:
	// user, role, location or location_intersection
	AssigneeType    string `json:"-" csv:"assignee_type" gorm:"size:30"`
	AssigneeUserID  string `json:"-" csv:"assignee_user_id" gorm:"size:37"`
	AssigneeSiteID  string `json:"-" csv:"assignee_site_id" gorm:"size:41"`
	AssigneeGroupID string `json:"-" csv:"assignee_group_id" gorm:"size:37"`

	// The names are looked up when exporting. A name is empty when it can't be found
	UserName            string `json:"-" csv:"user_name"`
	AssigneeUserName    string `json:"-" csv:"assignee_user_name"`
	AssigneeSiteName    string `json:"-" csv:"assignee_site_name"`
	AssigneeGroupName   string `json:"-" csv:"assignee_group_name"`
	CompletedByUserName string `json:"-" csv:"completed_by_user_name"`
	// TargetName is the name of the site, or the code of the asset
	TargetName string `json:"-" csv:"target_name"`
}

// UnmarshalJSON decodes the API fields and fills the derived fields
func (o *ScheduleOccurrenceV2) UnmarshalJSON(b []byte) error {
	type apiOccurrence ScheduleOccurrenceV2
	if err := json.Unmarshal(b, (*apiOccurrence)(o)); err != nil {
		return err
	}

	o.UserID = o.AssigneeID
	o.Note = o.WontDoNote

	// assignee_from is user_<id>, role_<id>, location_<id> or location_<id>+role_<id> for a site and group intersection
	switch from := o.AssigneeFrom; {
	case strings.HasPrefix(from, "user_"):
		o.AssigneeType = "user"
		o.AssigneeUserID = from
	case strings.HasPrefix(from, "role_"):
		o.AssigneeType = "role"
		o.AssigneeGroupID = from
	case strings.HasPrefix(from, "location_"):
		site, group, found := strings.Cut(from, "+")
		o.AssigneeSiteID = site
		o.AssigneeType = "location"
		if found {
			o.AssigneeType = "location_intersection"
			o.AssigneeGroupID = group
		}
	}
	return nil
}

// TableName makes the table match the name of the feed. By default gorm derives it from the type name
func (ScheduleOccurrenceV2) TableName() string {
	return "schedule_occurrences_v2"
}

// nameLookup has the names of the users, groups, sites and assets of the organisation by ID.
// Sites are found by both their location_<hex> ID and their UUID
type nameLookup struct {
	users  map[string]string
	groups map[string]string
	sites  map[string]string
	assets map[string]string
}

// loadNames reads the names from the users, groups, sites and assets feeds
func loadNames(ctx context.Context, apiClient *httpapi.Client) (*nameLookup, error) {
	names := &nameLookup{
		users:  map[string]string{},
		groups: map[string]string{},
		sites:  map[string]string{},
		assets: map[string]string{},
	}

	err := drainRows(ctx, apiClient, &GetFeedRequest{InitialURL: "/feed/users"}, func(u *User) {
		names.users[u.ID] = strings.TrimSpace(u.Firstname + " " + u.Lastname)
	})
	if err != nil {
		return nil, fmt.Errorf("load user names: %w", err)
	}

	err = drainRows(ctx, apiClient, &GetFeedRequest{InitialURL: "/feed/groups"}, func(g *Group) {
		names.groups[g.ID] = g.Name
	})
	if err != nil {
		return nil, fmt.Errorf("load group names: %w", err)
	}

	showOnlyLeafNodes := false
	sitesRequest := &GetFeedRequest{
		InitialURL: "/feed/sites",
		Params:     GetFeedParams{IncludeDeleted: true, ShowOnlyLeafNodes: &showOnlyLeafNodes},
	}
	err = drainRows(ctx, apiClient, sitesRequest, func(site *Site) {
		names.sites[site.ID] = site.Name
		if site.SiteUUID != "" {
			names.sites[site.SiteUUID] = site.Name
		}
	})
	if err != nil {
		return nil, fmt.Errorf("load site names: %w", err)
	}

	err = drainRows(ctx, apiClient, &GetFeedRequest{InitialURL: "/feed/assets"}, func(a *Asset) {
		names.assets[a.ID] = a.Code
	})
	if err != nil {
		return nil, fmt.Errorf("load asset names: %w", err)
	}
	return names, nil
}

// drainRows calls add for every row of a feed
func drainRows[T any](ctx context.Context, apiClient *httpapi.Client, req *GetFeedRequest, add func(*T)) error {
	return DrainFeed(ctx, apiClient, req, func(resp *GetFeedResponse) error {
		var rows []*T
		if err := json.Unmarshal(resp.Data, &rows); err != nil {
			return fmt.Errorf("map data: %w", err)
		}
		for _, row := range rows {
			add(row)
		}
		return nil
	})
}

// fillNames sets the names of the occurrence
func (n *nameLookup) fillNames(o *ScheduleOccurrenceV2) {
	o.UserName = n.users[o.UserID]
	o.AssigneeUserName = n.users[o.AssigneeUserID]
	o.AssigneeSiteName = n.sites[o.AssigneeSiteID]
	o.AssigneeGroupName = n.groups[o.AssigneeGroupID]
	o.CompletedByUserName = n.users[o.CompletedByUserID]
	switch o.TargetType {
	case "SITE":
		o.TargetName = n.sites[o.TargetID]
	case "ASSET":
		o.TargetName = n.assets[o.TargetID]
	}
}

// prepareOccurrenceNames loads the names for the occurrences. The names are left empty when they can't be loaded
func prepareOccurrenceNames(ctx context.Context, apiClient *httpapi.Client) func(*ScheduleOccurrenceV2) {
	names, err := loadNames(ctx, apiClient)
	if err != nil {
		logger.GetLogger().With("feed", "schedule_occurrences_v2", "error", err).Warn("names are not exported")
		return nil
	}
	return names.fillNames
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
			"description", "exported_at", "from_date", "to_date", "start_time_hour", "start_time_minute",
			"all_must_complete", "can_late_submit", "site_id", "creator_user_id", "asset_id", "target_type",
			"target_ids",
		},
	}
}

// NewScheduleAssigneeV2Feed returns the feed for /scheduling/v1/feed/schedule_assignees
func NewScheduleAssigneeV2Feed(templateIDs []string) Feed {
	return &rowFeed[ScheduleAssigneeV2]{
		name:       "schedule_assignees_v2",
		initialURL: "/scheduling/v1/feed/schedule_assignees",
		primaryKey: []string{"id", "target_id"},
		order:      "schedule_id, assignee_id, target_id",
		params: GetFeedParams{
			TemplateIDs: templateIDs,
			Limit:       maxSchedulingFeedLimit,
		},
		key: func(r *ScheduleAssigneeV2) string { return r.ID + "|" + r.TargetID },
		columns: []string{
			"schedule_id", "assignee_id", "type", "name", "organisation_id", "completion_rule", "target_type",
		},
	}
}

// NewScheduleOccurrenceV2Feed returns the feed for /scheduling/v1/feed/schedule_occurrences.
// startDate filters on the due time of the occurrences. When it is zero the API default window applies (5 months back, 1 month ahead).
func NewScheduleOccurrenceV2Feed(templateIDs []string, startDate time.Time) Feed {
	return &rowFeed[ScheduleOccurrenceV2]{
		name:       "schedule_occurrences_v2",
		initialURL: "/scheduling/v1/feed/schedule_occurrences",
		primaryKey: []string{"id", "target_id"},
		order:      "schedule_id, due_time, target_id",
		params: GetFeedParams{
			TemplateIDs: templateIDs,
			StartDate:   startDate,
			Limit:       maxSchedulingFeedLimit,
		},
		key:     func(r *ScheduleOccurrenceV2) string { return r.ID + "|" + r.TargetID },
		prepare: prepareOccurrenceNames,
		columns: []string{
			"schedule_id", "occurrence_id", "template_id", "start_time", "due_time", "miss_time", "occurrence_status",
			"audit_id", "completed_at", "completion_rule", "organisation_id", "wont_do_note", "assignee_id", "assignee_from",
			"target_type", "assignee_status", "completed_by_user_id", "user_id", "note", "assignee_type",
			"assignee_user_id", "assignee_site_id", "assignee_group_id", "user_name", "assignee_user_name",
			"assignee_site_name", "assignee_group_name", "completed_by_user_name", "target_name",
		},
	}
}
