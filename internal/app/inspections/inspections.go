package inspections

import (
	"context"
	"sync"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/config"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/exporter"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const maxGoRoutines = 10

type tFunc func(api.Inspection)

type inspectionClient struct {
	*zap.SugaredLogger

	apiClient     api.APIClient
	exporter      exporter.Exporter
	SkipIDs       []string
	ModifiedAfter string
	TemplateIDs   []string
	Archived      string
	Completed     string
	Incremental   bool
}

type InspectionClient interface {
	Export(ctx context.Context) error
}

func NewInspectionClient(v *viper.Viper, apiClient api.APIClient, exporter exporter.Exporter) InspectionClient {
	inspectionConfig := config.GetInspectionConfig(v)
	templateIDs := v.GetStringSlice("export.template_ids")

	return &inspectionClient{
		apiClient:     apiClient,
		exporter:      exporter,
		SkipIDs:       inspectionConfig.SkipIDs,
		ModifiedAfter: inspectionConfig.ModifiedAfter,
		TemplateIDs:   templateIDs,
		Archived:      inspectionConfig.Archived,
		Completed:     inspectionConfig.Completed,
		Incremental:   inspectionConfig.Incremental,
		SugaredLogger: util.GetLogger(),
	}
}

func (client *inspectionClient) Name() string {
	return "inspections"
}

// throttleFunc throttles the given function
func throttleFunc(tfunc tFunc, inspection api.Inspection, guard chan struct{}) {
	// If the channel is full then this go-routine will be in waiting state.
	// Multiple go routines can access this channel without any lock as the
	// the channel internally maintains the lock for send and receive operations
	guard <- struct{}{}
	tfunc(inspection)
	<-guard

	return
}

func (client *inspectionClient) Export(ctx context.Context) error {
	var wg sync.WaitGroup

	client.Infof("%s: exporting", client.Name())

	skipIDs := map[string]bool{}
	for _, id := range client.SkipIDs {
		skipIDs[id] = true
	}

	// Create a buffered channel of length maxGoroutines
	guard := make(chan struct{}, maxGoRoutines)

	operation := func(row api.Inspection) {
		inspection, err := client.apiClient.GetInspection(ctx, row.ID)
		util.Check(err, "Failed to get inspection")

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

	modifiedAt := client.exporter.GetLastModifiedAt()
	if modifiedAt != nil {
		params.ModifiedAfter = *modifiedAt
	}

	err := client.apiClient.DrainInspections(
		ctx,
		params,
		callback,
	)
	util.Check(err, "Failed to list inspections")

	client.Info("Export finished")

	return nil
}
