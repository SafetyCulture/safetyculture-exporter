package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/internal/app/api"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/util"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// SheqsyEmployee represents a user in sheqsy
type SheqsyEmployee struct {
	EmployeeUID             string    `json:"employeeUId" csv:"employee_uid" gorm:"primarykey;column:employee_uid;size:36"`
	EmployeeID              int       `json:"employeeId" csv:"employee_id" gorm:"column:employee_id"`
	ExternalID              *string   `json:"externalId" csv:"external_id" gorm:"column:external_id"`
	FirstName               string    `json:"firstName" csv:"first_name" gorm:"column:first_name"`
	LastName                string    `json:"lastName" csv:"last_name" gorm:"column:last_name"`
	Email                   string    `json:"email" csv:"email" gorm:"column:email"`
	AcceptedActivitiesCount int       `json:"acceptedActivitiesCount" csv:"accepted_activities_count" gorm:"column:accepted_activities_count"`
	PendingActivitiesCount  int       `json:"pendingActivitiesCount" csv:"pending_activities_count" gorm:"column:pending_activities_count"`
	IsInPanic               bool      `json:"isInPanic" csv:"is_in_panic" gorm:"column:is_in_panic"`
	Status                  string    `json:"status" csv:"status" gorm:"column:status"`
	LastActivityDateTimeUTC string    `json:"lastActivityDateTimeUTC" csv:"last_activity_date_time_utc" gorm:"column:last_activity_date_time_utc"`
	Departments             string    `json:"departments" csv:"departments" gorm:"column:departments"`
	ExportedAt              time.Time `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
}

// SheqsyEmployeeFeed is a representation of the users feed
type SheqsyEmployeeFeed struct{}

// Name is the name of the feed
func (f *SheqsyEmployeeFeed) Name() string {
	return "sheqsy_employees"
}

// Model returns the model of the feed row
func (f *SheqsyEmployeeFeed) Model() interface{} {
	return SheqsyEmployee{}
}

// RowsModel returns the model of feed rows
func (f *SheqsyEmployeeFeed) RowsModel() interface{} {
	return &[]*SheqsyEmployee{}
}

// PrimaryKey returns the primary key(s)
func (f *SheqsyEmployeeFeed) PrimaryKey() []string {
	return []string{"employee_uid"}
}

// Columns returns the columns of the row
func (f *SheqsyEmployeeFeed) Columns() []string {
	return []string{
		"employee_id",
		"external_id",
		"first_name",
		"last_name",
		"email",
		"accepted_activities_count",
		"pending_activities_count",
		"is_in_panic",
		"status",
		"last_activity_date_time_utc",
		"exported_at",
	}
}

// Order returns the ordering when retrieving an export
func (f *SheqsyEmployeeFeed) Order() string {
	return "employee_uid"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *SheqsyEmployeeFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*SheqsyEmployee{})
}

// Export exports the feed to the supplied exporter
func (f *SheqsyEmployeeFeed) Export(ctx context.Context, apiClient *api.Client, exporter Exporter, companyID string) error {
	logger := util.GetLogger().With(
		"feed", f.Name(),
		"org_id", companyID,
	)

	exporter.InitFeed(f, &InitFeedOptions{
		// Truncate files if upserts aren't supported.
		// This ensure that the export does not contain duplicate rows
		Truncate: !exporter.SupportsUpsert(),
	})

	var rows []*SheqsyEmployee

	resp, err := apiClient.Get(ctx, fmt.Sprintf("/SheqsyIntegrationApi/api/v3/companies/%s/employees", companyID))
	util.Check(err, "failed fetch data")

	var respBytes []byte
	respBytes = *resp

	gjson.ParseBytes(respBytes).ForEach(func(key, value gjson.Result) bool {
		// These timestamps aren't parsable as RFC 3339 strings, so we have to munge them to that format
		lastSeen := value.Get("lastActivityDateTimeUTC").String()
		if len(lastSeen) != 0 {
			respBytes, err = sjson.SetBytes(
				respBytes,
				fmt.Sprintf("%d.lastActivityDateTimeUTC", key.Int()),
				lastSeen+"Z",
			)
			util.Check(err, "failed to update lastActivityDateTimeUTC")
		}

		var departments []string
		value.Get("departments.#.name").ForEach(func(key, value gjson.Result) bool {
			departments = append(departments, value.String())
			return true
		})
		util.Check(err, "failed to set departments")
		respBytes, err = sjson.SetBytes(
			respBytes,
			fmt.Sprintf("%d.departments", key.Int()),
			strings.Join(departments, ","),
		)
		util.Check(err, "failed to set departments")
		return true
	})

	err = json.Unmarshal(respBytes, &rows)
	util.Check(err, "failed to parse API response")

	if len(rows) != 0 {
		// Calculate the size of the batch we can insert into the DB at once. Column count + buffer to account for primary keys
		batchSize := exporter.ParameterLimit() / (len(f.Columns()) + 4)

		for i := 0; i < len(rows); i += batchSize {
			j := i + batchSize
			if j > len(rows) {
				j = len(rows)
			}

			err = exporter.WriteRows(f, rows[i:j])
			util.Check(err, "Failed to write data to exporter")
		}
	}

	logger.With(
		"estimated_remaining", 0,
		"duration_ms", apiClient.Duration.Milliseconds(),
		"export_duration_ms", exporter.GetDuration().Milliseconds(),
	).Info("export batch complete")

	return exporter.FinaliseExport(f, &[]*SheqsyEmployee{})
}
