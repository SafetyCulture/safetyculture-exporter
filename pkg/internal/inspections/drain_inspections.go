package inspections

import (
	"context"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/httpapi"
)

// DrainInspections fetches the inspections in batches and triggers the callback
// for each batch.
func DrainInspections(ctx context.Context, apiClient *httpapi.Client, params *ListInspectionsParams, callback func(*ListInspectionsResponse) error) error {
	modifiedAfter := params.ModifiedAfter

	for {
		resp, err := ListInspections(
			ctx,
			apiClient,
			&ListInspectionsParams{
				ModifiedAfter: modifiedAfter,
				TemplateIDs:   params.TemplateIDs,
				Archived:      params.Archived,
				Completed:     params.Completed,
			},
		)
		if err != nil {
			return err
		}

		if err := callback(resp); err != nil {
			return err
		}

		if (resp.Total - resp.Count) == 0 {
			break
		}
		modifiedAfter = resp.Inspections[resp.Count-1].ModifiedAt
	}

	return nil
}
