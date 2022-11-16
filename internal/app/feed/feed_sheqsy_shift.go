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

// SheqsyShift represents a user in sheqsy
type SheqsyShift struct {
	ShiftID               int       `json:"shiftId" csv:"shift_id" gorm:"primarykey;column:shift_id;"`
	EmployeeName          string    `json:"employeeName" csv:"employee_name" gorm:"column:employee_name"`
	EmployeeSurname       string    `json:"employeeSurname" csv:"employee_surname" gorm:"column:employee_surname"`
	Email                 string    `json:"email" csv:"email" gorm:"column:email"`
	PhoneNumber           string    `json:"phoneNumber" csv:"phone_number" gorm:"column:phone_number"`
	StartDateTimeUTC      time.Time `json:"startDateTimeUTC" csv:"start_date_time_utc" gorm:"column:start_date_time_utc"`
	FinishDateTimeUTC     time.Time `json:"finishDateTimeUTC" csv:"finish_date_time_utc" gorm:"column:finish_date_time_utc"`
	LastReportedAddress   string    `json:"lastReportedAddress" csv:"last_reported_address" gorm:"column:last_reported_address"`
	LastReportedLatitude  float64   `json:"lastReportedLatitude" csv:"last_reported_latitude" gorm:"column:last_reported_latitude"`
	LastReportedLongitude float64   `json:"lastReportedLongitude" csv:"last_reported_longitude" gorm:"column:last_reported_longitude"`
	Version               int       `json:"version" csv:"version" gorm:"column:version"`
	Departments           string    `json:"departments" csv:"departments" gorm:"column:departments"`
}

// SheqsyShiftFeed is a representation of the users feed
type SheqsyShiftFeed struct{}

// Name is the name of the feed
func (f *SheqsyShiftFeed) Name() string {
	return "sheqsy_shifts"
}

// Model returns the model of the feed row
func (f *SheqsyShiftFeed) Model() interface{} {
	return SheqsyShift{}
}

// RowsModel returns the model of feed rows
func (f *SheqsyShiftFeed) RowsModel() interface{} {
	return &[]*SheqsyShift{}
}

// PrimaryKey returns the primary key(s)
func (f *SheqsyShiftFeed) PrimaryKey() []string {
	return []string{"shift_id"}
}

// Columns returns the columns of the row
func (f *SheqsyShiftFeed) Columns() []string {
	return []string{
		"employee_name",
		"employee_surname",
		"email",
		"phone_number",
		"start_date_time_utc",
		"finish_date_time_utc",
		"last_reported_address",
		"last_reported_latitude",
		"last_reported_longitude",
		"version",
		"departments",
	}
}

// Order returns the ordering when retrieving an export
func (f *SheqsyShiftFeed) Order() string {
	return "shift_id"
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *SheqsyShiftFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*User{})
}

// Export exports the feed to the supplied exporter
func (f *SheqsyShiftFeed) Export(ctx context.Context, apiClient *api.Client, exporter Exporter, companyID string) error {
	logger := util.GetLogger().With(
		"feed", f.Name(),
		"org_id", companyID,
	)

	exporter.InitFeed(f, &InitFeedOptions{
		// Truncate files if upserts aren't supported.
		// This ensure that the export does not contain duplicate rows
		Truncate: !exporter.SupportsUpsert(),
	})

	type apiResp struct {
		Data         []*SheqsyShift `json:"data"`
		LastVersion  int            `json:"lastVersion"`
		HasMoreItems bool           `json:"hasMoreItems"`
		ItemsLeft    int            `json:"itemsLeft"`
	}

	data := apiResp{}

	version := 1
	for version != 0 {
		resp, err := apiClient.Get(ctx, fmt.Sprintf("/SheqsyIntegrationApi/api/v3/companies/%s/shifts/history?ver=%d", companyID, version))
		util.Check(err, "failed fetch data")

		var respBytes []byte
		respBytes = *resp

		gjson.GetBytes(respBytes, "data").ForEach(func(key, value gjson.Result) bool {
			// These timestamps aren't parsable as RFC 3339 strings, so we have to munge them to that format
			respBytes, err = sjson.SetBytes(
				respBytes,
				fmt.Sprintf("data.%d.startDateTimeUTC", key.Int()),
				value.Get("startDateTimeUTC").String()+"Z",
			)
			util.Check(err, "failed to fix timestamp")
			respBytes, err = sjson.SetBytes(
				respBytes,
				fmt.Sprintf("data.%d.finishDateTimeUTC", key.Int()),
				value.Get("finishDateTimeUTC").String()+"Z",
			)
			util.Check(err, "failed to fix timestamp")

			// Departments needs to be a flat string
			departments := []string{}
			value.Get("departments").ForEach(func(key, value gjson.Result) bool {
				departments = append(departments, value.String())
				return true
			})

			respBytes, err = sjson.SetBytes(
				respBytes,
				fmt.Sprintf("data.%d.departments", key.Int()),
				strings.Join(departments, ","),
			)
			util.Check(err, "failed to join departments")
			return true
		})

		err = json.Unmarshal(respBytes, &data)
		util.Check(err, "failed to parse API response")

		if len(data.Data) != 0 {
			// Calculate the size of the batch we can insert into the DB at once. Column count + buffer to account for primary keys
			batchSize := exporter.ParameterLimit() / (len(f.Columns()) + 4)

			for i := 0; i < len(data.Data); i += batchSize {
				j := i + batchSize
				if j > len(data.Data) {
					j = len(data.Data)
				}

				err = exporter.WriteRows(f, data.Data[i:j])
				util.Check(err, "Failed to write data to exporter")
			}
		}

		version = data.LastVersion
		if !data.HasMoreItems {
			version = 0
		}

		logger.With(
			"estimated_remaining", 0,
			"duration_ms", apiClient.Duration.Milliseconds(),
			"export_duration_ms", exporter.GetDuration().Milliseconds(),
		).Info("export batch complete")
	}

	return exporter.FinaliseExport(f, &[]*SheqsyShift{})
}