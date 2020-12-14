package feed

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/api"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReportExporter is an interface to export data feeds to CSV files
type ReportExporter struct {
	*SQLExporter
	Logger     *zap.SugaredLogger
	ExportPath string
	Mu         sync.Mutex
}

type Report struct {
	AuditID         string    `gorm:"primarykey;column:audit_id"`
	AuditModifiedAt time.Time `gorm:"column:modified_at"`
	PDF             int       `gorm:"column:pdf"`
	WORD            int       `gorm:"column:word"`
}

type SaveReportsResult struct {
	NoChange    int
	PDFReports  int
	PDFErrors   int
	WORDReports int
	WORDErrors  int
}

func (e *ReportExporter) SaveReports(ctx context.Context, apiClient api.APIClient, feed *InspectionFeed, formats []string) error {
	e.Logger.Info("Generating inspection reports")

	exportPDF, exportWORD := false, false
	for _, f := range formats {
		switch f {
		case "PDF":
			exportPDF = true
		case "WORD":
			exportWORD = true
		default:
			e.Logger.Infof("%s is not a valid report format", f)
		}
	}

	if !exportPDF && !exportWORD {
		return fmt.Errorf("No valid export format specified")
	}

	report := &Report{}
	err := e.DB.AutoMigrate(&Report{})
	if err != nil {
		return err
	}

	if !feed.Incremental {
		result := e.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(report)
		if result.Error != nil {
			return errors.Wrap(result.Error, "Unable to truncate table")
		}
	}

	res := &SaveReportsResult{}

	// you can specify level of concurrency by increasing channel size
	buffers := make(chan bool, 3)
	var wg sync.WaitGroup

	limit := 1
	offset := 0
	for {
		rows := &[]*Inspection{}
		resp := e.DB.
			Order(feed.Order()).
			Limit(limit).
			Offset(offset).
			Find(rows)

		if resp.Error != nil {
			err = resp.Error
			break
		}

		if resp.RowsAffected == 0 || resp.RowsAffected == -1 {
			break
		}

		offset = offset + limit

		for _, r := range *rows {
			wg.Add(1)

			go func(inspection *Inspection) {
				defer wg.Done()
				buffers <- true

				rep := e.saveReport(ctx, apiClient, inspection, exportPDF, exportWORD)
				updateReportResult(rep, res)

				<-buffers
			}(r)
		}
	}

	wg.Wait()

	e.Logger.Infof("There were no changes made to %d inspections and no reports downloaded", res.NoChange)

	if res.PDFReports > 0 || res.WORDReports > 0 {
		e.Logger.Infof("Successfully generate %d PDF reports and %d WORD reports", res.PDFReports, res.WORDReports)
	}

	if res.PDFErrors > 0 || res.WORDErrors > 0 {
		return fmt.Errorf("Failed to generate %d PDF reports and %d WORD reports", res.PDFErrors, res.WORDErrors)
	}

	return err
}

func (e *ReportExporter) saveReport(ctx context.Context, apiClient api.APIClient, inspection *Inspection, pdf bool, word bool) *Report {
	auditID := inspection.ID
	exportPDF := pdf
	exportWORD := word

	report := &Report{}
	err := e.DB.First(&report, "audit_id = ?", auditID).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		e.Logger.Errorf("Error during loading report from reports db: %s", err)
		return report
	}

	if report.AuditModifiedAt != inspection.ModifiedAt {
		report.AuditID = auditID
		report.AuditModifiedAt = inspection.ModifiedAt
	} else {
		exportPDF = exportPDF && report.PDF == 0
		exportWORD = exportWORD && report.WORD == 0
		if !exportPDF && !exportWORD {
			return nil
		}
	}

	var wg sync.WaitGroup

	if exportPDF {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err = e.exportInspection(ctx, apiClient, inspection, "PDF")
			if err != nil {
				e.Logger.Errorf("PDF export failed for '%s'. Error: %s", inspection.Name, err)
				report.PDF = -1
			} else {
				report.PDF = 1
			}
		}()
	}

	if exportWORD {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err = e.exportInspection(ctx, apiClient, inspection, "WORD")
			if err != nil {
				e.Logger.Errorf("WORD export failed for '%s'. Error: %s", inspection.Name, err)
				report.WORD = -1
			} else {
				report.WORD = 1
			}
		}()
	}

	wg.Wait()

	result := e.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "audit_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"modified_at", "pdf", "word"}),
	}).Create(&report)

	if result.Error != nil {
		e.Logger.Errorf("Failed to update save report status to local db for %s", inspection.Name)
	}

	return report
}

func (e *ReportExporter) exportInspection(ctx context.Context, apiClient api.APIClient, inspection *Inspection, format string) error {
	fn := fmt.Sprintf("%s (%s)", inspection.Name, inspection.ID)
	e.Logger.Infof("Exporting %s report for %s", format, fn)

	mId, err := apiClient.InitiateInspectionReportExport(ctx, inspection.ID, format)

	if err != nil {
		return err
	}

	tries := 0

	for {
		// wait for a second before checking for report completion
		time.Sleep(1 * time.Second)
		du, err := apiClient.CheckInspectionReportExportCompletion(ctx, inspection.ID, mId)
		if err != nil {
			break
		} else if du.Status == "SUCCESS" {
			resp, err := apiClient.DownloadInspectionReportFile(ctx, du.URL)
			if err != nil {
				break
			}

			// only allow one process to access disk at the same time
			// this way we won't allow process to overwrite reports with the same name
			e.Mu.Lock()
			err = saveReportResponse(resp, inspection, e.ExportPath, format)
			e.Mu.Unlock()
			break
		} else if du.Status == "FAILED" {
			err = fmt.Errorf("%s report generation failed on server for %s", format, fn)
			break
		}

		// make sure we stop checking after a while
		tries += 1
		if tries == 20 {
			err = fmt.Errorf("%s report generation for %s terminated after %d tries", format, inspection.Name, tries)
			break
		}
	}

	return err
}

func saveReportResponse(resp io.ReadCloser, inspection *Inspection, path string, format string) error {
	filePath := getFilePath(path, inspection, format)
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, resp)
	resp.Close()
	out.Close()
	return err
}

func sanitizeName(name string) string {
	res := strings.ReplaceAll(name, " / ", "-")
	res = strings.ReplaceAll(res, " ", "-")
	return res
}

func getFileExtension(format string) string {
	switch format {
	case "PDF":
		return "pdf"
	case "WORD":
		return "docx"
	}
	return ""
}

func getFilePath(exportPath string, inspection *Inspection, format string) string {
	dupIndex := 0
	for true {
		fileName := sanitizeName(inspection.Name)
		if strings.TrimSpace(fileName) == "" {
			fileName = inspection.ID
		}
		if dupIndex > 0 {
			fileName = fmt.Sprintf("%s (%d)", fileName, dupIndex)
		}
		exportFilePath := filepath.Join(exportPath, fmt.Sprintf("%s.%s", fileName, getFileExtension(format)))
		if _, err := os.Stat(exportFilePath); os.IsNotExist(err) {
			return exportFilePath
		} else {
			dupIndex += 1
		}
	}
	return ""
}

func updateReportResult(rep *Report, res *SaveReportsResult) {
	if rep == nil {
		res.NoChange += 1
	} else {
		if rep.PDF == 1 {
			res.PDFReports += 1
		} else if rep.PDF == -1 {
			res.PDFErrors += 1
		}

		if rep.WORD == 1 {
			res.WORDReports += 1
		} else if rep.WORD == -1 {
			res.WORDErrors += 1
		}
	}
}

func NewReportExporter(exportPath string) (*ReportExporter, error) {
	sqlExporter, err := NewSQLExporter("sqlite", filepath.Join(exportPath, "reports.db"), true)
	if err != nil {
		return nil, err
	}

	return &ReportExporter{
		SQLExporter: sqlExporter,
		Logger:      sqlExporter.Logger,
		ExportPath:  exportPath,
	}, nil
}
