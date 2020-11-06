package feed

import (
	"context"
	"sync"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/SafetyCulture/iauditor-exporter/internal/app/util"
	"github.com/spf13/viper"
)

func ExportFeeds(v *viper.Viper, apiClient api.APIClient, exporter Exporter) error {
	logger := util.GetLogger()
	ctx := context.Background()

	var wg sync.WaitGroup
	tables := v.GetStringSlice("export.tables")
	tablesMap := map[string]bool{}
	for _, table := range tables {
		tablesMap[table] = true
	}

	// TODO. Should validate auth before doing anything

	if tablesMap["inspections"] || len(tables) == 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := (&InspectionFeed{
				SkipIDs:       v.GetStringSlice("export.inspection.skip_ids"),
				ModifiedAfter: v.GetString("export.inspection.modified_after"),
				TemplateIDs:   v.GetStringSlice("export.template_ids"),
				Archived:      v.GetString("export.inspection.archived"),
				Completed:     v.GetString("export.inspection.completed"),
				Incremental:   v.GetBool("export.inspection.incremental"),
			}).Export(ctx, apiClient, exporter)
			util.Check(err, "failed to export")
		}()
	}

	if tablesMap["inspection_items"] || len(tables) == 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := (&InspectionItemFeed{
				SkipIDs:         v.GetStringSlice("export.inspection.skip_ids"),
				ModifiedAfter:   v.GetString("export.inspection.modified_after"),
				TemplateIDs:     v.GetStringSlice("export.template_ids"),
				Archived:        v.GetString("export.inspection.archived"),
				Completed:       v.GetString("export.inspection.completed"),
				IncludeInactive: v.GetBool("export.inspection.included_inactive_items"),
				Incremental:     v.GetBool("export.inspection.incremental"),
			}).Export(ctx, apiClient, exporter)
			util.Check(err, "failed to export")
		}()
	}

	if tablesMap["templates"] || len(tables) == 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := (&TemplateFeed{
				Incremental: v.GetBool("export.inspection.incremental"),
			}).Export(ctx, apiClient, exporter)
			util.Check(err, "failed to export")
		}()
	}

	if tablesMap["sites"] || len(tables) == 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := (&SiteFeed{}).Export(ctx, apiClient, exporter)
			util.Check(err, "failed to export")
		}()
	}

	if tablesMap["users"] || len(tables) == 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := (&UserFeed{}).Export(ctx, apiClient, exporter)
			util.Check(err, "failed to export")
		}()
	}

	if tablesMap["groups"] || len(tables) == 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := (&GroupFeed{}).Export(ctx, apiClient, exporter)
			util.Check(err, "failed to export")
		}()
	}

	if tablesMap["group_users"] || len(tables) == 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := (&GroupUserFeed{}).Export(ctx, apiClient, exporter)
			util.Check(err, "failed to export")
		}()
	}

	if tablesMap["schedules"] || len(tables) == 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := (&ScheduleFeed{
				TemplateIDs: v.GetStringSlice("export.template_ids"),
			}).Export(ctx, apiClient, exporter)
			util.Check(err, "failed to export")
		}()
	}

	if tablesMap["schedule_assignees"] || len(tables) == 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := (&ScheduleAssigneeFeed{
				TemplateIDs: v.GetStringSlice("export.template_ids"),
			}).Export(ctx, apiClient, exporter)
			util.Check(err, "failed to export")
		}()
	}

	if tablesMap["schedule_occurrence"] || len(tables) == 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := (&ScheduleOccurrenceFeed{
				TemplateIDs: v.GetStringSlice("export.template_ids"),
			}).Export(ctx, apiClient, exporter)
			util.Check(err, "failed to export")
		}()
	}

	wg.Wait()

	logger.Info("Export finished")

	return nil
}
