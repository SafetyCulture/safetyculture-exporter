package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/events"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
)

// TrainingCourseProgress represents a row for the feed
type TrainingCourseProgress struct {
	ID               string     `json:"id" csv:"course_progress_id" gorm:"primarykey;column:course_progress_id;size:512"`
	OpenedAt         *time.Time `json:"opened_at" csv:"opened_at"`
	CompletedAt      *time.Time `json:"completed_at" csv:"completed_at"`
	TotalLessons     int32      `json:"total_lessons" csv:"total_lessons"`
	CompletedLessons int32      `json:"completed_lessons" csv:"completed_lessons"`
	CourseID         string     `json:"course_id" csv:"course_id" gorm:"size:64"`
	CourseExternalID string     `json:"course_external_id" csv:"course_external_id" gorm:"size:256"`
	CourseTitle      string     `json:"course_title" csv:"course_title"`
	UserEmail        string     `json:"user_email" csv:"user_email" gorm:"size:256"`
	UserFirstName    string     `json:"user_first_name" csv:"user_first_name"`
	UserLastName     string     `json:"user_last_name" csv:"user_last_name"`
	UserID           string     `json:"user_id" csv:"user_id" gorm:"size:37"`
	UserExternalID   string     `json:"user_external_id" csv:"user_external_id"`
	ProgressPercent  float32    `json:"progress_percent" csv:"progress_percent"`
	Score            int32      `json:"score" csv:"score"`
	DueAt            *time.Time `json:"due_at" csv:"due_at"`
}

// TrainingCourseProgressFeed is a representation of the feed
type TrainingCourseProgressFeed struct {
	Incremental bool
	Offset      int
}

// Name is the name of the feed
func (f *TrainingCourseProgressFeed) Name() string {
	return "training_course_progress"
}

// Model returns the model of the feed row
func (f *TrainingCourseProgressFeed) Model() any {
	return TrainingCourseProgress{}
}

// RowsModel returns the model of feed rows
func (f *TrainingCourseProgressFeed) RowsModel() any {
	return &[]*TrainingCourseProgress{}
}

// PrimaryKey returns the primary key(s)
func (f *TrainingCourseProgressFeed) PrimaryKey() []string {
	return []string{"course_id", "user_id"}
}

func (f *TrainingCourseProgressFeed) Columns() []string {
	return []string{
		"opened_at",
		"completed_at",
		"total_lessons",
		"completed_lessons",
		"course_id",
		"course_external_id",
		"course_title",
		"user_email",
		"user_first_name",
		"user_last_name",
		"user_id",
		"user_external_id",
		"progress_percent",
		"score",
		"due_at",
	}
}

// Order returns the ordering when retrieving an export
func (f *TrainingCourseProgressFeed) Order() string {
	return "opened_at"
}

func (f *TrainingCourseProgressFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*TrainingCourseProgress{})
}

func (f *TrainingCourseProgressFeed) Export(ctx context.Context, apiClient *httpapi.Client, exporter Exporter, orgID string) error {
	l := logger.GetLogger().With("feed", f.Name(), "org_id", orgID)
	status := GetExporterStatus()

	if err := exporter.InitFeed(f, &InitFeedOptions{
		// Delete data if incremental refresh is disabled so there is no duplicates
		Truncate: !f.Incremental,
	}); err != nil {
		return events.WrapEventError(err, "init feed")
	}

	drainFn := func(resp *GetFeedResponse) error {
		var rows []*TrainingCourseProgress

		if err := json.Unmarshal(resp.Data, &rows); err != nil {
			return events.NewEventErrorWithMessage(err, events.ErrorSeverityError, events.ErrorSubSystemDataIntegrity, false, "map data")
		}

		if len(rows) != 0 {
			// Calculate the size of the batch we can insert into the DB at once. Column count + buffer to account for primary keys
			batchSize := exporter.ParameterLimit() / (len(f.Columns()) + 4)
			err := util.SplitSliceInBatch(batchSize, rows, func(batch []*TrainingCourseProgress) error {
				if err := exporter.WriteRows(f, batch); err != nil {
					return events.WrapEventError(err, "write rows")
				}
				return nil
			})

			if err != nil {
				return err
			}
		}

		// note: this feed api doesn't return remaining items
		status.UpdateStatus(f.Name(), -1, apiClient.Duration.Milliseconds())

		l.With(
			"estimated_remaining", -1,
			"duration_ms", apiClient.Duration.Milliseconds(),
			"export_duration_ms", exporter.GetDuration().Milliseconds(),
		).Info("export batch complete")
		return nil
	}

	req := &GetFeedRequest{
		InitialURL: "/feed/training-course-progress",
		Params: GetFeedParams{
			Offset: f.Offset,
		},
	}

	if err := DrainFeed(ctx, apiClient, req, drainFn); err != nil {
		return events.WrapEventError(err, fmt.Sprintf("feed %q", f.Name()))
	}
	return exporter.FinaliseExport(f, &[]*TrainingCourseProgress{})
}
