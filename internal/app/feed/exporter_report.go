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
	Logger       *zap.SugaredLogger
	ExportPath   string
	PreferenceID string
	Format       []string
	Mu           sync.Mutex
}

type reportExportFormat struct {
	PDF  bool
	WORD bool
}

type reportExport struct {
	AuditID         string    `gorm:"primarykey;column:audit_id"`
	AuditModifiedAt time.Time `gorm:"column:modified_at"`
	PDF             int       `gorm:"column:pdf"`
	WORD            int       `gorm:"column:word"`
}

type reportExportResult struct {
	NoChange    int
	PDFReports  int
	PDFErrors   int
	WORDReports int
	WORDErrors  int
}

// SaveReports downloads and stores inspection reports on disk
func (e *ReportExporter) SaveReports(ctx context.Context, apiClient api.Client, feed *InspectionFeed) error {
	e.Logger.Info("Generating inspection reports")

	format, err := e.getFormats()
	if err != nil {
		return fmt.Errorf("No valid export format specified")
	}

	report := &reportExport{}
	err = e.DB.AutoMigrate(&reportExport{})
	if err != nil {
		return err
	}

	if !feed.Incremental {
		result := e.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(report)
		if result.Error != nil {
			return errors.Wrap(result.Error, "Unable to truncate table")
		}
	}

	res := &reportExportResult{}

	// you can specify level of concurrency by increasing channel size
	buffers := make(chan bool, 3)
	var wg sync.WaitGroup

	var totalInspections int64 = 0
	cntRsp := e.DB.Model(&Inspection{}).Count(&totalInspections)
	if cntRsp.Error != nil {
		return cntRsp.Error
	}

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

			go func(inspection *Inspection, remaining int64) {
				defer wg.Done()
				buffers <- true

				rep := e.saveReport(ctx, apiClient, inspection, format)
				e.updateReportResult(rep, res, inspection, remaining)

				<-buffers
			}(r, totalInspections-int64(offset))
		}
	}

	wg.Wait()

	if res.NoChange > 0 {
		e.Logger.Infof("There were no changes made to %d inspections and no reports downloaded", res.NoChange)
	}

	if res.PDFReports > 0 || res.WORDReports > 0 {
		e.Logger.Infof("Successfully generate %d PDF reports and %d WORD reports", res.PDFReports, res.WORDReports)
	}

	if res.PDFErrors > 0 || res.WORDErrors > 0 {
		err = fmt.Errorf("Failed to generate %d PDF reports and %d WORD reports", res.PDFErrors, res.WORDErrors)
	}

	return err
}

func (e *ReportExporter) getFormats() (*reportExportFormat, error) {
	format := &reportExportFormat{}
	for _, f := range e.Format {
		switch f {
		case "PDF":
			format.PDF = true
		case "WORD":
			format.WORD = true
		default:
			e.Logger.Infof("%s is not a valid report format", f)
		}
	}

	if !format.PDF && !format.WORD {
		return nil, fmt.Errorf("No valid export format specified")
	}

	return format, nil
}

func (e *ReportExporter) saveReport(ctx context.Context, apiClient api.Client, inspection *Inspection, format *reportExportFormat) *reportExport {
	exportPDF, exportWORD := format.PDF, format.WORD

	report := &reportExport{}
	err := e.DB.First(&report, "audit_id = ?", inspection.ID).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		e.Logger.Errorf("Error during loading report from reports db: %s", err)
		return report
	}

	if report.AuditModifiedAt != inspection.ModifiedAt {
		report.AuditID = inspection.ID
		report.AuditModifiedAt = inspection.ModifiedAt
	} else {
		exportPDF = exportPDF && report.PDF != 1
		exportWORD = exportWORD && report.WORD != 1
		if !exportPDF && !exportWORD {
			return nil
		}
	}

	if exportPDF {
		err = e.exportInspection(ctx, apiClient, inspection, "PDF")
		if err != nil {
			e.Logger.Errorf("PDF export failed for '%s'. Error: %s", inspection.Name, err)
			report.PDF = -1
		} else {
			report.PDF = 1
		}
	}

	if exportWORD {
		err = e.exportInspection(ctx, apiClient, inspection, "WORD")
		if err != nil {
			e.Logger.Errorf("WORD export failed for '%s'. Error: %s", inspection.Name, err)
			report.WORD = -1
		} else {
			report.WORD = 1
		}
	}

	result := e.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "audit_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"modified_at", "pdf", "word"}),
	}).Create(&report)

	if result.Error != nil {
		e.Logger.Errorf("Failed to update save report status to local db for %s", inspection.Name)
	}

	return report
}

func (e *ReportExporter) exportInspection(ctx context.Context, apiClient api.Client, inspection *Inspection, format string) error {
	messageID, err := apiClient.InitiateInspectionReportExport(ctx, inspection.ID, format, e.PreferenceID)
	if err != nil {
		return err
	}

	tries := 0

	for {
		// wait for a second before checking for report completion
		time.Sleep(1 * time.Second)
		rec, cErr := apiClient.CheckInspectionReportExportCompletion(ctx, inspection.ID, messageID)
		if cErr != nil {
			err = cErr
			break
		} else if rec.Status == "SUCCESS" {
			resp, dErr := apiClient.DownloadInspectionReportFile(ctx, rec.URL)
			if dErr != nil {
				err = dErr
				break
			}

			// only allow one process to access disk at the same time
			// this way we won't allow process to overwrite reports with the same name
			e.Mu.Lock()
			err = saveReportResponse(resp, inspection, e.ExportPath, format)
			e.Mu.Unlock()
			break
		} else if rec.Status == "FAILED" {
			err = fmt.Errorf("%s report generation failed on server for %s", format, fmt.Sprintf("%s (%s)", inspection.Name, inspection.ID))
			break
		}

		// make sure we stop checking after a while
		tries++
		if tries == 15 {
			err = fmt.Errorf("%s report generation for %s terminated after %d tries", format, inspection.Name, tries)
			break
		}
	}

	return err
}

func (e *ReportExporter) updateReportResult(rep *reportExport, res *reportExportResult, inspection *Inspection, remaining int64) {
	fn := fmt.Sprintf("%s (%s)", inspection.Name, inspection.ID)
	if rep == nil {
		res.NoChange++
		e.Logger.Infof("No changes were made to %s", fn)
	} else {
		if rep.PDF == 1 {
			res.PDFReports++
			e.Logger.Infof("Saved PDF report for %s", fn)
		} else if rep.PDF == -1 {
			res.PDFErrors++
		}

		if rep.WORD == 1 {
			res.WORDReports++
			e.Logger.Infof("Saved Word report for %s", fn)
		} else if rep.WORD == -1 {
			res.WORDErrors++
		}

		e.Logger.Infof("%d inspections remaining", remaining)
	}
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
	res = strings.ReplaceAll(res, " // ", "-")
	res = strings.ReplaceAll(res, "/", "-")
	res = strings.ReplaceAll(res, " ", "-")
	return res
}

func getFileExtension(format string) string {
	switch format {
	case "PDF":
		return "pdf"
	case "WORD":
		return "docx"
	default:
		return ""
	}
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
		}

		dupIndex++
	}

	return ""
}

// NewReportExporter returns a new instance of ReportExporter
func NewReportExporter(exportPath string, format []string, preferenceID string) (*ReportExporter, error) {
	sqlExporter, err := NewSQLExporter("sqlite", filepath.Join(exportPath, "reports.db"), true, "")
	if err != nil {
		return nil, err
	}

	return &ReportExporter{
		SQLExporter:  sqlExporter,
		Logger:       sqlExporter.Logger,
		ExportPath:   exportPath,
		Format:       format,
		PreferenceID: preferenceID,
	}, nil
}
