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
	Name            string     `json:"name" csv:"name"`
	Archived        bool       `json:"archived" csv:"archived"`
	OwnerName       string     `json:"owner_name" csv:"owner_name"`
	OwnerID         string     `json:"owner_id" csv:"owner_id"`
	AuthorName      string     `json:"author_name" csv:"author_name"`
	AuthorID        string     `json:"author_id" csv:"author_id"`
	Score           float32    `json:"score" csv:"score"`
	MaxScore        float32    `json:"max_score" csv:"max_score"`
	ScorePercentage float32    `json:"score_percentage" csv:"score_percentage"`
	Duration        int64      `json:"duration" csv:"duration"`
	TemplateID      string     `json:"template_id" csv:"template_id"`
	TemplateName    string     `json:"template_name" csv:"template_name"`
	TemplateAuthor  string     `json:"template_author" csv:"template_author"`
	SiteID          string     `json:"site_id" csv:"site_id"`
	DateStarted     time.Time  `json:"date_started" csv:"date_started"`
	DateCompleted   *time.Time `json:"date_completed" csv:"date_completed"`
	DateModified    time.Time  `json:"date_modified" csv:"date_modified"`
	CreatedAt       time.Time  `json:"created_at" csv:"created_at"`
	ModifiedAt      time.Time  `json:"modified_at" csv:"modified_at"`
	ExportedAt      time.Time  `json:"exported_at" csv:"exported_at" gorm:"autoUpdateTime"`
	DocumentNo      string     `json:"document_no" csv:"document_no"`
	PreparedBy      string     `json:"prepared_by" csv:"prepared_by"`
	Location        string     `json:"location" csv:"location"`
	ConductedOn     *time.Time `json:"conducted_on" csv:"conducted_on"`
	Personnel       string     `json:"personnel" csv:"personnel"`
	ClientSite      string     `json:"client_site" csv:"client_site"`
}

// InspectionFeed is a representation of the inspections feed
type InspectionFeed struct {
	rows *[]*Inspection

	SkipIDs       []string
	ModifiedAfter string
	TemplateIDs   []string
	Archived      string
	Completed     string
	Incremental   bool
}

// Name is the name of the feed
func (f *InspectionFeed) Name() string {
	return "inspections"
}

// Model returns the model of the feed row
func (f *InspectionFeed) Model() interface{} {
	return Inspection{}
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

	// Calculate the size of the batch we can insert into the DB at once. Column count + buffer
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
func (f *InspectionFeed) Export(ctx context.Context, apiClient api.APIClient, exporter Exporter) error {
	logger := util.GetLogger()
	feedName := f.Name()

	logger.Infof("%s: exporting", feedName)

	exporter.InitFeed(f, &InitFeedOptions{
		// Delete data if incremental refresh is disabled so there is no duplicates
		Truncate: f.Incremental == false,
	})

	lastModifiedAt, err := exporter.LastModifiedAt(f)
	util.Check(err, "unable to load modified after")
	if lastModifiedAt != nil {
		f.ModifiedAfter = lastModifiedAt.Format(time.RFC3339Nano)
	}

	err = apiClient.DrainFeed(ctx, &api.GetFeedRequest{
		InitialURL: "/feed/inspections",
		Params: api.GetFeedParams{
			ModifiedAfter: f.ModifiedAfter,
			TemplateIDs:   f.TemplateIDs,
			Archived:      f.Archived,
			Completed:     f.Completed,
		},
	}, func(resp *api.GetFeedResponse) error {
		rows := []*Inspection{}

		err := json.Unmarshal(resp.Data, &rows)
		util.Check(err, "Failed to unmarshal data to struct")

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
