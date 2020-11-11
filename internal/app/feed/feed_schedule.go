package feed

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
)

// Schedule represents a row from the schedules feed
type Schedule struct {
	ID              string     `json:"id" csv:"schedule_id" gorm:"primarykey;column:schedule_id"`
	Description     string     `json:"description" csv:"description"`
	Recurrence      string     `json:"recurrence" csv:"recurrence"`
	Duration        string     `json:"duration" csv:"duration"`
	ModifiedAt      time.Time  `json:"modified_at" csv:"modified_at"`
	ExportedAt      time.Time  `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
	FromDate        time.Time  `json:"from_date" csv:"from_date"`
	ToDate          *time.Time `json:"to_date" csv:"to_date"`
	StartTimeHour   int        `json:"start_time_hour" csv:"start_time_hour"`
	StartTimeMinute int        `json:"start_time_minute" csv:"start_time_minute"`
	AllMustComplete bool       `json:"all_must_complete" csv:"all_must_complete"`
	Status          string     `json:"status" csv:"status"`
	Timezone        string     `json:"timezone" csv:"timezone"`
	CanLateSubmit   bool       `json:"can_late_submit" csv:"can_late_submit"`
	SiteID          string     `json:"site_id" csv:"site_id"`
	TemplateID      string     `json:"template_id" csv:"template_id"`
	CreatorUserID   string     `json:"creator_user_id" csv:"creator_user_id"`
}

// ScheduleFeed is a representation of the schedules feed
type ScheduleFeed struct {
	TemplateIDs []string
}

// Name is the name of the feed
func (f *ScheduleFeed) Name() string {
	return "schedules"
}

// Model returns the model of the feed row
func (f *ScheduleFeed) Model() interface{} {
	return Schedule{}
}

// PrimaryKey returns the primary key(s)
func (f *ScheduleFeed) PrimaryKey() []string {
	return []string{"schedule_id"}
}

// Columns returns the columns of the row
func (f *ScheduleFeed) Columns() []string {
	return []string{
		"description",
		"recurrence",
		"duration",
		"modified_at",
		"exported_at",
		"from_date",
		"to_date",
		"start_time_hour",
		"start_time_minute",
		"all_must_complete",
		"status",
		"timezone",
		"can_late_submit",
		"site_id",
		"template_id",
		"creator_user_id",
	}
}

// Order returns the ordering when retrieving an export
func (f *ScheduleFeed) Order() string {
	return "schedule_id"
}

// Export exports the feed to the supplied exporter
func (f *ScheduleFeed) Export(ctx context.Context, apiClient api.APIClient, exporter Exporter) error {
	logger := util.GetLogger()
	feedName := f.Name()

	logger.Infof("%s: exporting", feedName)

	exporter.InitFeed(f, &InitFeedOptions{
		// Truncate files if upserts aren't supported.
		// This ensure that the export does not contain duplicate rows
		Truncate: exporter.SupportsUpsert() == false,
	})

	err := apiClient.DrainFeed(ctx, &api.GetFeedRequest{
		InitialURL: "/feed/schedules",
		Params: api.GetFeedParams{
			TemplateIDs: f.TemplateIDs,
		},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*Schedule{}

		err := json.Unmarshal(resp.Data, &rows)
		util.Check(err, "Failed to unmarshal data to struct")

		if len(rows) != 0 {
			err = exporter.WriteRows(f, rows)
			util.Check(err, "Failed to write data to exporter")
		}

		logger.Infof("%s: %d remaining", feedName, resp.Metadata.RemainingRecords)
		return nil
	})
	util.Check(err, "Failed to export feed")

	return exporter.FinaliseExport(f, &[]*Schedule{})
}
