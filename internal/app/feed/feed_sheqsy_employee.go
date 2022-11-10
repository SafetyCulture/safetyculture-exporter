package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/internal/app/api"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/util"
)

// SheqsyEmployee represents a user in sheqsy
type SheqsyEmployee struct {
	EmployeeUID             string    `json:"employeeUId" csv:"employee_uid" gorm:"primarykey;column:employee_uid;size:32"`
	EmployeeID              int       `json:"employeeId" csv:"employee_id" gorm:"column:employee_id"`
	ExternalID              *string   `json:"externalId" csv:"external_id" gorm:"column:external_id"`
	FirstName               string    `json:"firstName" csv:"first_name" gorm:"column:first_name"`
	LastName                string    `json:"lastName" csv:"last_name" gorm:"column:last_name"`
	Email                   string    `json:"email" csv:"email" gorm:"column:email"`
	AcceptedActivitiesCount int       `json:"acceptedActivitiesCount" csv:"accepted_activities_count" gorm:"column:accepted_activities_count"`
	PendingActivitiesCount  int       `json:"pendingActivitiesCount" csv:"pending_activities_count" gorm:"column:pending_activities_count"`
	IsInPanic               bool      `json:"isInPanic" csv:"is_in_panic" gorm:"column:is_in_panic"`
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
		"exported_at",
	}
}

// Order returns the ordering when retrieving an export
func (f *SheqsyEmployeeFeed) Order() string {
	return "employee_uid"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *SheqsyEmployeeFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*User{})
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

	rows := []*SheqsyEmployee{}

	resp, err := apiClient.Get(ctx, fmt.Sprintf("/SheqsyIntegrationApi/api/v3/companies/%s/employees", companyID))
	util.Check(err, "failed fetch data")

	err = json.Unmarshal(*resp, &rows)
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
