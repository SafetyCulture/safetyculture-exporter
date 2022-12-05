package inspections

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SafetyCulture/safetyculture-exporter/internal/app/api"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/config"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/exporter"
	"github.com/SafetyCulture/safetyculture-exporter/internal/app/util"
	"go.uber.org/zap"
)

const maxGoRoutines = 10

// Client to be used with inspections
type Client struct {
	*zap.SugaredLogger

	apiClient     *api.Client
	exporter      exporter.SafetyCultureJSONExporter
	SkipIDs       []string
	ModifiedAfter time.Time
	TemplateIDs   []string
	Archived      string
	Completed     string
	Incremental   bool
}

// InspectionClient is an interface to get inspections from server
type InspectionClient interface {
	Export(ctx context.Context) error
}

// NewInspectionClient returns a new instance of InspectionClient
func NewInspectionClient(cfg *config.ExporterConfiguration, apiClient *api.Client, exporter exporter.SafetyCultureJSONExporter) InspectionClient {
	return &Client{
		apiClient:     apiClient,
		exporter:      exporter,
		SkipIDs:       cfg.Export.Inspection.SkipIds,
		ModifiedAfter: cfg.Export.ModifiedAfter.Time,
		TemplateIDs:   cfg.Export.TemplateIds,
		Archived:      cfg.Export.Inspection.Archived,
		Completed:     cfg.Export.Inspection.Completed,
		Incremental:   cfg.Export.Incremental,
		SugaredLogger: util.GetLogger(),
	}
}

// Name returns the name of the client
func (client *Client) Name() string {
	return "inspections"
}

// throttleFunc throttles the given function
func throttleFunc(tfunc func(api.Inspection), inspection api.Inspection, guard chan struct{}) {
	// If the channel is full then this go-routine will be in waiting state.
	// Multiple go routines can access this channel without any lock as the
	// the channel internally maintains the lock for send and receive operations
	guard <- struct{}{}
	tfunc(inspection)
	<-guard
}

// Export triggers the export
func (client *Client) Export(ctx context.Context) error {
	var wg sync.WaitGroup

	skipIDs := map[string]bool{}
	for _, id := range client.SkipIDs {
		skipIDs[id] = true
	}

	// Create a buffered channel of length maxGoroutines
	guard := make(chan struct{}, maxGoRoutines)

	operation := func(row api.Inspection) {
		inspection, err := client.apiClient.GetInspection(ctx, row.ID)
		util.Check(err, fmt.Sprintf("Failed to get inspection with id: %s", row.ID))

		client.exporter.WriteRow(row.ID, inspection)
	}

	callback := func(resp *api.ListInspectionsResponse) error {
		n := len(resp.Inspections)
		for _, row := range resp.Inspections {
			skip := skipIDs[row.ID]
			if skip {
				continue
			}

			wg.Add(1)
			go func(row api.Inspection) {
				defer wg.Done()
				throttleFunc(operation, row, guard)
			}(row)
		}
		wg.Wait()

		if n > 0 {
			client.exporter.SetLastModifiedAt(resp.Inspections[n-1].ModifiedAt)
		}
		return nil
	}

	params := &api.ListInspectionsParams{
		TemplateIDs: client.TemplateIDs,
		Archived:    client.Archived,
		Completed:   client.Completed,
	}

	modifiedAt := client.exporter.GetLastModifiedAt(client.ModifiedAfter)
	if modifiedAt != nil {
		params.ModifiedAfter = *modifiedAt
	}
	client.Infof("%s: exporting since %s", client.Name(), params.ModifiedAfter.Format(time.RFC1123))

	err := client.apiClient.DrainInspections(
		ctx,
		params,
		callback,
	)
	util.Check(err, "failed to list inspections")

	client.Info("Export finished")

	return nil
}
