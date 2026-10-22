package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/SafetyCulture/mitti-exporter/pkg/httpapi"
	"github.com/SafetyCulture/mitti-exporter/pkg/logger"
	"github.com/gofrs/uuid"
)

// The feeds in this file use the Scheduling feed service (/scheduling/v1/feed/*) which replaces the legacy
// /feed/schedules* endpoints for organisations on the new Schedules experience.
// They export to separate new_* tables so organisations that are still on the legacy experience keep working.
// Each table has the columns of the legacy table, plus the columns the new Schedules exporter PRD adds.

// maxSchedulingFeedLimit is the largest page size the Scheduling feed service accepts (the default is 50)
const maxSchedulingFeedLimit = 1000

// ScheduleV2 represents a row from the Scheduling feed service schedules feed.
// The legacy schedules columns come first, then the columns the PRD adds
type ScheduleV2 struct {
	ID              string     `json:"id" csv:"schedule_id" gorm:"primarykey;column:schedule_id;size:128"`
	Description     string     `json:"-" csv:"description"`
	Recurrence      string     `json:"recurrence" csv:"recurrence"`
	Duration        string     `json:"duration" csv:"duration" gorm:"size:50"`
	ModifiedAt      time.Time  `json:"modified_at" csv:"modified_at"`
	ExportedAt      time.Time  `json:"-" csv:"exported_at" gorm:"autoUpdateTime"`
	FromDate        *time.Time `json:"-" csv:"from_date"`
	ToDate          *time.Time `json:"-" csv:"to_date"`
	StartTimeHour   int        `json:"-" csv:"start_time_hour"`
	StartTimeMinute int        `json:"-" csv:"start_time_minute"`
	AllMustComplete bool       `json:"-" csv:"all_must_complete"`
	Status          string     `json:"status" csv:"status" gorm:"size:20"`
	OrganisationID  string     `json:"organisation_id" csv:"organisation_id" gorm:"size:37"`
	Timezone        string     `json:"timezone" csv:"timezone"`
	CanLateSubmit   bool       `json:"-" csv:"can_late_submit"`
	SiteID          string     `json:"-" csv:"site_id" gorm:"size:41"`
	TemplateID      string     `json:"template_id" csv:"template_id" gorm:"size:100"`
	CreatorUserID   string     `json:"-" csv:"creator_user_id" gorm:"size:37"`
	AssetID         string     `json:"-" csv:"asset_id" gorm:"size:37"`
	TargetType      string     `json:"-" csv:"target_type" gorm:"size:10"`
	// TargetIDs is the JSON array of the site or asset UUIDs of the schedule, empty for schedules without sites or assets
	TargetIDs JSONArray `json:"-" csv:"target_ids"`
	// TimezoneType is CUSTOM_TIMEZONE or SITE_TIMEZONE. With SITE_TIMEZONE, recurrence has the site wall-clock time
	// as UTC and each site uses its own time zone
	TimezoneType string `json:"timezone_type" csv:"timezone_type" gorm:"size:30"`
}

// UnmarshalJSON decodes the API fields and fills the legacy and target columns. A column that can't be parsed is left empty
func (s *ScheduleV2) UnmarshalJSON(b []byte) error {
	type apiSchedule ScheduleV2
	if err := json.Unmarshal(b, (*apiSchedule)(s)); err != nil {
		return err
	}

	var api struct {
		Title               string   `json:"title"`
		CompletionRule      string   `json:"completion_rule"`
		IsLateSubmitAllowed bool     `json:"is_late_submit_allowed"`
		CreatedBy           string   `json:"created_by"`
		EndingRuleEndDate   string   `json:"ending_rule_end_date"`
		ScheduleType        string   `json:"schedule_type"`
		SiteIDs             []string `json:"site_ids"`
		AssetIDs            []string `json:"asset_ids"`
	}
	if err := json.Unmarshal(b, &api); err != nil {
		return err
	}

	s.Description = api.Title
	s.AllMustComplete = api.CompletionRule == "all"
	s.CanLateSubmit = api.IsLateSubmitAllowed
	s.CreatorUserID = api.CreatedBy
	if len(api.SiteIDs) == 1 {
		s.SiteID = api.SiteIDs[0]
	}
	if len(api.AssetIDs) == 1 {
		s.AssetID = api.AssetIDs[0]
	}

	var targetIDs []string
	switch api.ScheduleType {
	case "SCHEDULE_TYPE_SITES":
		s.TargetType = "SITE"
		targetIDs = siteUUIDs(api.SiteIDs)
	case "SCHEDULE_TYPE_ASSETS":
		s.TargetType = "ASSET"
		targetIDs = api.AssetIDs
	}
	if s.TargetType != "" {
		ids, err := json.Marshal(targetIDs)
		if err != nil {
			return err
		}
		s.TargetIDs = JSONArray(ids)
	}

	if toDate, err := time.Parse(time.RFC3339, api.EndingRuleEndDate); err == nil {
		s.ToDate = &toDate
	}
	if fromDate := parseDTStart(s.Recurrence, s.Timezone); fromDate != nil {
		// The start time is the local time of the schedule, and the date is in UTC like the other times
		s.StartTimeHour = fromDate.Hour()
		s.StartTimeMinute = fromDate.Minute()
		utc := fromDate.UTC()
		s.FromDate = &utc
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
	return "new_schedules"
}

// ScheduleAssigneeV2 represents a row from the Scheduling feed service schedule_assignees feed.
// The legacy schedule_assignees columns come first, then the target
type ScheduleAssigneeV2 struct {
	ID             string    `json:"id" csv:"id" gorm:"primarykey;size:375"`
	ScheduleID     string    `json:"schedule_id" csv:"schedule_id" gorm:"size:128"`
	AssigneeID     string    `json:"assignee_id" csv:"assignee_id" gorm:"size:256"`
	OrganisationID string    `json:"organisation_id" csv:"organisation_id" gorm:"size:37"`
	Type           string    `json:"type" csv:"type" gorm:"size:30"`
	Name           string    `json:"name" csv:"name"`
	ExportedAt     time.Time `json:"-" csv:"exported_at" gorm:"autoUpdateTime"`
	// TargetID is the site or asset UUID of the row, empty for schedules without sites or assets.
	// It is part of the primary key, as the API returns the same id once per site or asset
	TargetID   string `json:"target_id" csv:"target_id" gorm:"primarykey;size:36"`
	TargetType string `json:"target_type" csv:"target_type" gorm:"size:10"`
}

// TableName makes the table match the name of the feed. By default gorm derives it from the type name
func (ScheduleAssigneeV2) TableName() string {
	return "new_schedule_assignees"
}

// ScheduleOccurrenceV2 represents a row from the Scheduling feed service schedule_occurrences feed.
// The legacy schedule_occurrences columns come first, then the columns the PRD adds
type ScheduleOccurrenceV2 struct {
	ID               string    `json:"id" csv:"id" gorm:"primarykey;size:375"`
	ScheduleID       string    `json:"schedule_id" csv:"schedule_id" gorm:"size:128"`
	OccurrenceID     string    `json:"occurrence_id" csv:"occurrence_id" gorm:"size:128"`
	TemplateID       string    `json:"template_id" csv:"template_id" gorm:"size:100"`
	OrganisationID   string    `json:"organisation_id" csv:"organisation_id" gorm:"size:37"`
	StartTime        string    `json:"start_time" csv:"start_time" gorm:"size:40"`
	DueTime          string    `json:"due_time" csv:"due_time" gorm:"size:40"`
	MissTime         string    `json:"miss_time" csv:"miss_time" gorm:"size:40"`
	OccurrenceStatus string    `json:"occurrence_status" csv:"occurrence_status" gorm:"size:20"`
	AuditID          string    `json:"audit_id" csv:"audit_id" gorm:"size:100"`
	CompletedAt      string    `json:"completed_at" csv:"completed_at" gorm:"size:40"`
	ExportedAt       time.Time `json:"-" csv:"exported_at" gorm:"autoUpdateTime"`
	UserID           string    `json:"assignee_id" csv:"user_id" gorm:"size:256"`
	AssigneeStatus   string    `json:"assignee_status" csv:"assignee_status" gorm:"size:20"`
	Note             string    `json:"-" csv:"note"`

	// AssigneeType is how the user was assigned: user, site or group. A group within a site is a site with the group set
	AssigneeType        string `json:"-" csv:"assignee_type" gorm:"size:10"`
	AssigneeUserID      string `json:"-" csv:"assignee_user_id" gorm:"size:37"`
	AssigneeUserName    string `json:"-" csv:"assignee_user_name"`
	AssigneeSiteID      string `json:"-" csv:"assignee_site_id" gorm:"size:41"`
	AssigneeSiteName    string `json:"-" csv:"assignee_site_name"`
	AssigneeGroupID     string `json:"-" csv:"assignee_group_id" gorm:"size:37"`
	AssigneeGroupName   string `json:"-" csv:"assignee_group_name"`
	CompletedByUserID   string `json:"completed_by_user_id" csv:"completed_by_user_id" gorm:"size:37"`
	CompletedByUserName string `json:"-" csv:"completed_by_user_name"`
	// TargetID is the site or asset UUID of the row, empty for schedules without sites or assets.
	// It is part of the primary key, as the API returns the same id once per site or asset
	TargetID   string `json:"target_id" csv:"target_id" gorm:"primarykey;size:36"`
	TargetType string `json:"target_type" csv:"target_type" gorm:"size:10"`
	// TargetName is the name of the site, or the code of the asset
	TargetName string `json:"-" csv:"target_name"`
	WontDoNote string `json:"wont_do_note" csv:"wont_do_note"`
}

// UnmarshalJSON decodes the API fields and fills the assignment and legacy columns.
// The names are filled when exporting
func (o *ScheduleOccurrenceV2) UnmarshalJSON(b []byte) error {
	type apiOccurrence ScheduleOccurrenceV2
	if err := json.Unmarshal(b, (*apiOccurrence)(o)); err != nil {
		return err
	}

	var api struct {
		AssigneeFrom string `json:"assignee_from"`
	}
	if err := json.Unmarshal(b, &api); err != nil {
		return err
	}

	o.Note = o.WontDoNote

	// assignee_from is user_<id>, role_<id>, location_<id> or location_<id>+role_<id> for a group within a site
	switch from := api.AssigneeFrom; {
	case strings.HasPrefix(from, "user_"):
		o.AssigneeType = "user"
		o.AssigneeUserID = from
	case strings.HasPrefix(from, "role_"):
		o.AssigneeType = "group"
		o.AssigneeGroupID = from
	case strings.HasPrefix(from, "location_"):
		site, group, found := strings.Cut(from, "+")
		o.AssigneeType = "site"
		o.AssigneeSiteID = site
		if found {
			o.AssigneeGroupID = group
		}
	}
	return nil
}

// TableName makes the table match the name of the feed. By default gorm derives it from the type name
func (ScheduleOccurrenceV2) TableName() string {
	return "new_schedule_occurrences"
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
		logger.GetLogger().With("feed", "new_schedule_occurrences", "error", err).Warn("names are not exported")
		return nil
	}
	return names.fillNames
}

// withScheduleStatuses adds the status filter of the schedules and schedule assignees feeds to their path.
// No statuses, or none that are known, keeps the path so schedules of every status are returned
func withScheduleStatuses(path string, statuses []string) string {
	show := map[string]bool{}
	for _, status := range statuses {
		show[strings.ToUpper(strings.TrimSpace(status))] = true
	}
	if !show["ACTIVE"] && !show["PAUSED"] && !show["FINISHED"] {
		return path
	}

	query := url.Values{}
	query.Set("show_active", strconv.FormatBool(show["ACTIVE"]))
	query.Set("show_paused", strconv.FormatBool(show["PAUSED"]))
	query.Set("show_finished", strconv.FormatBool(show["FINISHED"]))
	return path + "?" + query.Encode()
}

// NewScheduleV2Feed returns the feed for /scheduling/v1/feed/schedules.
// statuses filters on the schedule status (ACTIVE, PAUSED, FINISHED); it is empty for every status
func NewScheduleV2Feed(templateIDs []string, statuses []string) Feed {
	return &rowFeed[ScheduleV2]{
		name:         "new_schedules",
		initialURL:   withScheduleStatuses("/scheduling/v1/feed/schedules", statuses),
		primaryKey:   []string{"schedule_id"},
		order:        "schedule_id",
		hasRemaining: false, // FeedMetadataSimple has no remaining_records
		params: GetFeedParams{
			TemplateIDs: templateIDs,
			Limit:       maxSchedulingFeedLimit,
		},
		key: func(r *ScheduleV2) string { return r.ID },
		columns: []string{
			"description", "recurrence", "duration", "modified_at", "exported_at", "from_date", "to_date",
			"start_time_hour", "start_time_minute", "all_must_complete", "status", "organisation_id", "timezone",
			"can_late_submit", "site_id", "template_id", "creator_user_id", "asset_id", "target_type", "target_ids",
			"timezone_type",
		},
	}
}

// NewScheduleAssigneeV2Feed returns the feed for /scheduling/v1/feed/schedule_assignees.
// statuses filters on the schedule status (ACTIVE, PAUSED, FINISHED); it is empty for every status
func NewScheduleAssigneeV2Feed(templateIDs []string, statuses []string) Feed {
	return &rowFeed[ScheduleAssigneeV2]{
		name:       "new_schedule_assignees",
		initialURL: withScheduleStatuses("/scheduling/v1/feed/schedule_assignees", statuses),
		primaryKey: []string{"id", "target_id"},
		order:      "schedule_id, assignee_id, target_id",
		params: GetFeedParams{
			TemplateIDs: templateIDs,
			Limit:       maxSchedulingFeedLimit,
		},
		key:     func(r *ScheduleAssigneeV2) string { return r.ID + "|" + r.TargetID },
		columns: []string{"schedule_id", "assignee_id", "organisation_id", "type", "name", "exported_at", "target_type"},
	}
}

// NewScheduleOccurrenceV2Feed returns the feed for /scheduling/v1/feed/schedule_occurrences.
// startDate and endDate filter on the due time of the occurrences. When one is zero, the API default applies to it
// (5 months back for the start, 1 month ahead for the end).
func NewScheduleOccurrenceV2Feed(templateIDs []string, startDate, endDate time.Time) Feed {
	return &rowFeed[ScheduleOccurrenceV2]{
		name:       "new_schedule_occurrences",
		initialURL: "/scheduling/v1/feed/schedule_occurrences",
		primaryKey: []string{"id", "target_id"},
		order:      "schedule_id, due_time, target_id",
		params: GetFeedParams{
			TemplateIDs: templateIDs,
			StartDate:   startDate,
			EndDate:     endDate,
			Limit:       maxSchedulingFeedLimit,
		},
		key:     func(r *ScheduleOccurrenceV2) string { return r.ID + "|" + r.TargetID },
		prepare: prepareOccurrenceNames,
		columns: []string{
			"schedule_id", "occurrence_id", "template_id", "organisation_id", "start_time", "due_time", "miss_time",
			"occurrence_status", "audit_id", "completed_at", "exported_at", "user_id", "assignee_status", "note",
			"assignee_type", "assignee_user_id", "assignee_user_name", "assignee_site_id", "assignee_site_name",
			"assignee_group_id", "assignee_group_name", "completed_by_user_id", "completed_by_user_name", "target_type",
			"target_name", "wont_do_note",
		},
	}
}
