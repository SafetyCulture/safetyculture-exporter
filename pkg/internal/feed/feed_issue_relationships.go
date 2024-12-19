package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/events"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/util"
	"github.com/SafetyCulture/safetyculture-exporter/pkg/logger"
	"time"
)

const issueRelationshipFeedPath = "/tasks/v1/feed/issue_relations"
const issueRelationshipSortingColumn = "rel_created_at"

// IssueRelationship represents a row from the issue relationship feed
type IssueRelationship struct {
	Id             string    `json:"id" csv:"id" gorm:"primaryKey;column:id;size:73"`
	FromId         string    `json:"from_id" csv:"from_id" gorm:"size:36"`
	FromLabel      string    `json:"from_label" csv:"from_label"`
	RelType        string    `json:"rel_type" csv:"rel_type"`
	RelCreatedAt   time.Time `json:"rel_created_at" csv:"rel_created_at"`
	ToId           string    `json:"to_id" csv:"to_id" gorm:"size:36"`
	ToLabel        string    `json:"to_label" csv:"to_label"`
	OrganisationID string    `json:"organisation_id" csv:"organisation_id" gorm:"size:36"`
}

// IssueRelationshipFeed is a representation of the issue relationship feed
type IssueRelationshipFeed struct {
	RelCreatedAt time.Time
	Incremental  bool
	Limit        int
}

func (f *IssueRelationshipFeed) Name() string {
	return "issue_relationships"
}

func (f *IssueRelationshipFeed) HasRemainingInformation() bool {
	return false
}

func (f *IssueRelationshipFeed) Model() interface{} {
	return IssueRelationship{}
}

func (f *IssueRelationshipFeed) RowsModel() interface{} {
	return &[]*IssueRelationship{}
}

func (f *IssueRelationshipFeed) PrimaryKey() []string {
	return []string{"id"}
}

func (f *IssueRelationshipFeed) Columns() []string {
	return []string{"id", "from_id", "from_label", "rel_type", "rel_created_at", "to_id", "to_label"}
}

func (f *IssueRelationshipFeed) Order() string {
	return "id"
}

func (f *IssueRelationshipFeed) CreateSchema(exporter Exporter) error {
	return exporter.CreateSchema(f, &[]*IssueRelationship{})
}

func (f *IssueRelationshipFeed) Export(ctx context.Context, apiClient *httpapi.Client, exporter Exporter, s12OrgID string) error {
	l := logger.GetLogger().With("feed", f.Name(), "org_id", s12OrgID)
	orgID := util.ConvertS12ToUUID(s12OrgID)
	if orgID.IsNil() {
		return fmt.Errorf("cannot convert given %q organisation ID to UUID", s12OrgID)
	}

	status := GetExporterStatus()

	if err := exporter.InitFeed(f, &InitFeedOptions{
		// Truncate files if upserts aren't supported.
		// This ensures that the export does not contain duplicate rows
		Truncate: !f.Incremental,
	}); err != nil {
		return events.WrapEventError(err, "init feed")
	}

	if f.Incremental {
		// note, this table doesn't have an orgID
		f.RelCreatedAt = exporter.LastRecord(f, f.RelCreatedAt, orgID.String(), issueRelationshipSortingColumn)
		l.Info("resuming issue relationship feed from ", f.RelCreatedAt.String())
	}

	drainFn := func(resp *GetFeedResponse) error {
		var rows []*IssueRelationship

		if err := json.Unmarshal(resp.Data, &rows); err != nil {
			return events.NewEventErrorWithMessage(err, events.ErrorSeverityError, events.ErrorSubSystemDataIntegrity, false, "map data")
		}

		numRows := len(rows)
		if numRows != 0 {
			// Calculate the size of the batch we can insert into the DB at once. Column count + buffer to account for primary keys
			batchSize := exporter.ParameterLimit() / (len(f.Columns()) + 4)
			err := util.SplitSliceInBatch(batchSize, rows, func(batch []*IssueRelationship) error {
				// hydrate org_id
				for _, r := range rows {
					r.OrganisationID = orgID.String()
				}

				if err := exporter.WriteRows(f, batch); err != nil {
					return events.WrapEventError(err, "write rows")
				}
				return nil
			})

			if err != nil {
				return err
			}
		}

		status.IncrementStatus(f.Name(), int64(numRows), apiClient.Duration.Milliseconds())

		l.With(
			"downloaded", status.ReadCounter(f.Name()),
			"duration_ms", apiClient.Duration.Milliseconds(),
			"export_duration_ms", exporter.GetDuration().Milliseconds(),
		).Info("export batch complete")
		return nil
	}

	req := &GetFeedRequest{
		InitialURL: issueRelationshipFeedPath,
		Params: GetFeedParams{
			Limit:        100,
			CreatedAfter: f.RelCreatedAt,
		},
	}

	if err := DrainFeed(ctx, apiClient, req, drainFn); err != nil {
		return events.WrapEventError(err, fmt.Sprintf("feed %q", f.Name()))
	}
	return exporter.FinaliseExport(f, &[]*IssueRelationship{})
}
