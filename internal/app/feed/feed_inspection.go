package feed

import (
	"context"
	"encoding/json"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
)

// Inspection represents a row from the inspections feed
type Inspection struct {
	ID              string     `json:"id" csv:"audit_id" gorm:"primarykey;column:audit_id"`
	Name            string     `json:"name" csv:"name" gorm:"size:16383" gorm:"size:16383"`
	Archived        bool       `json:"archived" csv:"archived"`
	OwnerName       string     `json:"owner_name" csv:"owner_name" gorm:"size:16383"`
	OwnerID         string     `json:"owner_id" csv:"owner_id"`
	AuthorName      string     `json:"author_name" csv:"author_name" gorm:"size:16383"`
	AuthorID        string     `json:"author_id" csv:"author_id"`
	Score           float32    `json:"score" csv:"score"`
	MaxScore        float32    `json:"max_score" csv:"max_score"`
	ScorePercentage float32    `json:"score_percentage" csv:"score_percentage"`
	Duration        int64      `json:"duration" csv:"duration"`
	TemplateID      string     `json:"template_id" csv:"template_id"`
	TemplateName    string     `json:"template_name" csv:"template_name" gorm:"size:16383"`
	TemplateAuthor  string     `json:"template_author" csv:"template_author" gorm:"size:16383"`
	SiteID          string     `json:"site_id" csv:"site_id"`
	DateStarted     time.Time  `json:"date_started" csv:"date_started"`
	DateCompleted   *time.Time `json:"date_completed" csv:"date_completed"`
	DateModified    time.Time  `json:"date_modified" csv:"date_modified"`
	CreatedAt       time.Time  `json:"created_at" csv:"created_at"`
	ModifiedAt      time.Time  `json:"modified_at" csv:"modified_at"`
	ExportedAt      time.Time  `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
	DocumentNo      string     `json:"document_no" csv:"document_no" gorm:"size:16383"`
	PreparedBy      string     `json:"prepared_by" csv:"prepared_by" gorm:"size:16383"`
	Location        string     `json:"location" csv:"location" gorm:"size:16383"`
	ConductedOn     *time.Time `json:"conducted_on" csv:"conducted_on"`
	Personnel       string     `json:"personnel" csv:"personnel" gorm:"size:16383"`
	ClientSite      string     `json:"client_site" csv:"client_site" gorm:"size:16383"`
	Latitude        *float64   `json:"latitude" csv:"latitude"`
	Longitude       *float64   `json:"longitude" csv:"longitude"`
	WebReportLink   string     `json:"web_report_link" csv:"web_report_link" gorm:"size:16383"`
}

// InspectionFeed is a representation of the inspections feed
type InspectionFeed struct {
	SkipIDs       []string
	ModifiedAfter time.Time
	TemplateIDs   []string
	Archived      string
	Completed     string
	Incremental   bool
	Limit         int
}

// Name is the name of the feed
func (f *InspectionFeed) Name() string {
	return "inspections"
}

// Model returns the model of the feed row
func (f *InspectionFeed) Model() interface{} {
	return Inspection{}
}

// RowsModel returns the model of feed rows
func (f *InspectionFeed) RowsModel() interface{} {
	return &[]*Inspection{}
}

// PrimaryKey returns the primary key(s)
func (f *InspectionFeed) PrimaryKey() []string {
	return []string{"audit_id"}
}

// Columns returns the columns of the row
func (f *InspectionFeed) Columns() []string {
	return []string{
		"name",
		"archived",
		"owner_name",
		"owner_id",
		"author_name",
		"author_id",
		"score",
		"max_score",
		"score_percentage",
		"duration",
		"template_id",
		"template_name",
		"template_author",
		"date_started",
		"date_completed",
		"date_modified",
		"created_at",
		"modified_at",
		"exported_at",
		"document_no",
		"prepared_by",
		"location",
		"conducted_on",
		"personnel",
		"client_site",
		"latitude",
		"longitude",
		"web_report_link",
	}
}

// Order returns the ordering when retrieving an export
func (f *InspectionFeed) Order() string {
	return "modified_at ASC, audit_id"
}

func (f *InspectionFeed) writeRows(exporter Exporter, rows []*Inspection) error {
	skipIDs := map[string]bool{}
	for _, id := range f.SkipIDs {
		skipIDs[id] = true
	}

	// Calculate the size of the batch we can insert into the DB at once. Column count + buffer to account for primary keys
	batchSize := exporter.ParameterLimit() / (len(f.Columns()) + 4)
	for i := 0; i < len(rows); i += batchSize {
		j := i + batchSize
		if j > len(rows) {
			j = len(rows)
		}

		// Some audits in production have the same item ID multiple times
		// We can't insert them simultaneously. This means we are dropping data, which sucks.
		rowsToInsert := []*Inspection{}
		for _, row := range rows[i:j] {
			skip := skipIDs[row.ID]
			if !skip {
				rowsToInsert = append(rowsToInsert, row)
			}
		}

		err := exporter.WriteRows(f, rowsToInsert)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateSchema creates the schema of the feed for the supplied exporter
func (f *InspectionFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*Inspection{})
}

// Export exports the feed to the supplied exporter
func (f *InspectionFeed) Export(ctx context.Context, apiClient *api.Client, exporter Exporter) error {
	logger := util.GetLogger()
	feedName := f.Name()

	exporter.InitFeed(f, &InitFeedOptions{
		// Delete data if incremental refresh is disabled so there is no duplicates
		Truncate: f.Incremental == false,
	})

	var err error
	f.ModifiedAfter, err = exporter.LastModifiedAt(f, f.ModifiedAfter)
	util.Check(err, "unable to load modified after")

	logger.Infof("%s: exporting since %s", feedName, f.ModifiedAfter.Format(time.RFC1123))

	err = apiClient.DrainFeed(ctx, &api.GetFeedRequest{
		InitialURL: "/feed/inspections",
		Params: api.GetFeedParams{
			ModifiedAfter: f.ModifiedAfter,
			TemplateIDs:   f.TemplateIDs,
			Archived:      f.Archived,
			Completed:     f.Completed,
			Limit:         f.Limit,
		},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*Inspection{}

		err := json.Unmarshal(resp.Data, &rows)
		util.Check(err, "Failed to unmarshal inspections data to struct")

		if len(rows) != 0 {
			err = f.writeRows(exporter, rows)
			util.Check(err, "Failed to write data to exporter")
		}

		logger.Infof("%s: %d remaining", feedName, resp.Metadata.RemainingRecords)
		return nil
	})
	util.Check(err, "Failed to export feed")

	return exporter.FinaliseExport(f, &[]*Inspection{})
}
