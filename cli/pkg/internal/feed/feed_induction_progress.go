package feed

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/events"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
)

// InductionProgress represents a row for the feed
type InductionProgress struct {
	OpenedAt                     string `json:"opened_at" csv:"opened_at"`
	CompletedAt                  string `json:"completed_at" csv:"completed_at"`
	IsWaitingForCompletion       bool   `json:"is_waiting_for_completion" csv:"is_waiting_for_completion"`
	StepProgressesCount          int32  `json:"step_progresses_count" csv:"step_progresses_count"`
	StepProgressesCompletedCount int32  `json:"step_progresses_completed_count" csv:"step_progresses_completed_count"`
	InductionID                  string `json:"induction_id" csv:"induction_id" gorm:"primarykey;column:induction_id;size:64"`
	InductionVersionID           string `json:"induction_version_id" csv:"induction_version_id" gorm:"size:64"`
	InductionTitle               string `json:"induction_title" csv:"induction_title"`
	OrgID                        string `json:"org_id" csv:"org_id" gorm:"size:37"`
	UserEmail                    string `json:"user_email" csv:"user_email" gorm:"size:256"`
	UserFirstName                string `json:"user_first_name" csv:"user_first_name"`
	UserLastName                 string `json:"user_last_name" csv:"user_last_name"`
	UserID                       string `json:"user_id" csv:"user_id" gorm:"primarykey;column:user_id;size:37"`
	UserExternalID               string `json:"user_external_id" csv:"user_external_id"`
}

// InductionProgressFeed is a representation of the feed
type InductionProgressFeed struct {
	Incremental bool
	Limit       int
}

// Name is the name of the feed
func (f *InductionProgressFeed) Name() string {
	return "induction_progresses"
}

// HasRemainingInformation returns true if the feed returns remaining items information
func (f *InductionProgressFeed) HasRemainingInformation() bool {
	return false
}

// Model returns the model of the feed row
func (f *InductionProgressFeed) Model() interface{} {
	return InductionProgress{}
}

// RowsModel returns the model of feed rows
func (f *InductionProgressFeed) RowsModel() interface{} {
	return &[]*InductionProgress{}
}

// PrimaryKey returns the primary key(s)
func (f *InductionProgressFeed) PrimaryKey() []string {
	return []string{"induction_id", "user_id"}
}

func (f *InductionProgressFeed) Columns() []string {
	return []string{
		"opened_at",
		"completed_at",
		"is_waiting_for_completion",
		"step_progresses_count",
		"step_progresses_completed_count",
		"induction_id",
		"induction_version_id",
		"induction_title",
		"org_id",
		"user_email",
		"user_first_name",
		"user_last_name",
		"user_id",
		"user_external_id",
	}
}

// Order returns the ordering when retrieving an export
func (f *InductionProgressFeed) Order() string {
	return "opened_at"
}

func (f *InductionProgressFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*InductionProgress{})
}

func (f *InductionProgressFeed) Export(ctx context.Context, apiClient *httpapi.Client, exporter Exporter, orgID string) error {
	l := logger.GetLogger().With("feed", f.Name(), "org_id", orgID)
	status := GetExporterStatus()

	if err := exporter.InitFeed(f, &InitFeedOptions{
		// Delete data if incremental refresh is disabled so there is no duplicates
		Truncate: !f.Incremental,
	}); err != nil {
		return events.WrapEventError(err, "init feed")
	}

	drainFn := func(resp *GetFeedResponse) error {
		var rows []*InductionProgress

		if err := json.Unmarshal(resp.Data, &rows); err != nil {
			return events.NewEventErrorWithMessage(err, events.ErrorSeverityError, events.ErrorSubSystemDataIntegrity, false, "map data")
		}

		numRows := len(rows)
		if numRows != 0 {
			// Calculate the size of the batch we can insert into the DB at once. Column count + buffer to account for primary keys
			batchSize := exporter.ParameterLimit() / (len(f.Columns()) + 4)
			err := util.SplitSliceInBatch(batchSize, rows, func(batch []*InductionProgress) error {
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
		status.IncrementStatus(f.Name(), int64(numRows), apiClient.Duration.Milliseconds())

		l.With(
			"downloaded", status.ReadCounter(f.Name()),
			"duration_ms", apiClient.Duration.Milliseconds(),
			"export_duration_ms", exporter.GetDuration().Milliseconds(),
		).Info("export batch complete")
		return nil
	}

	req := &GetFeedRequest{
		InitialURL: "/inductions/v1/feed/onboarding-progress",
		Params: GetFeedParams{
			Limit: f.Limit,
		},
	}

	if err := DrainFeed(ctx, apiClient, req, drainFn); err != nil {
		return events.WrapEventError(err, fmt.Sprintf("feed %q", f.Name()))
	}
	return exporter.FinaliseExport(f, &[]*InductionProgress{})
}
