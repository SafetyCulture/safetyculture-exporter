package feed

import (
	"sync"
)

var lock = &sync.Mutex{}
var statusInstance *ExportStatus

// GetExporterStatus will return a singleton instance of the ExporterStatus
func GetExporterStatus() *ExportStatus {
	lock.Lock()
	defer lock.Unlock()
	if statusInstance == nil {
		statusInstance = &ExportStatus{
			status: map[string]*ExportStatusItem{},
		}
	}
	return statusInstance
}

type ExportStatus struct {
	lock     sync.Mutex
	status   map[string]*ExportStatusItem
	finished bool
	started  bool
}

type ExportStatusItemStage string

const StageApi ExportStatusItemStage = "API_DOWNLOAD"
const StageCsv ExportStatusItemStage = "CSV_EXPORT"

type ExportStatusItem struct {
	Name          string
	Stage         ExportStatusItemStage
	Started       bool
	Finished      bool
	HasError      bool
	EstRemaining  int64
	StatusMessage string
	DurationMs    int64
}

func (e *ExportStatus) Reset() {
	e.lock.Lock()
	e.started = false
	e.finished = false
	e.status = map[string]*ExportStatusItem{}
	e.lock.Unlock()
}

func (e *ExportStatus) StartFeedExport(feedName string) {
	e.lock.Lock()
	e.status[feedName] = &ExportStatusItem{
		Name:     feedName,
		Stage:    StageApi,
		Started:  true,
		Finished: false,
	}
	e.lock.Unlock()
}

func (e *ExportStatus) UpdateStatus(feedName string, remaining int64, durationMs int64) {
	e.lock.Lock()
	if _, ok := e.status[feedName]; ok {
		e.status[feedName].EstRemaining = remaining
		e.status[feedName].DurationMs = durationMs
	}
	e.lock.Unlock()
}

func (e *ExportStatus) UpdateStage(feedName string, stage ExportStatusItemStage) {
	e.lock.Lock()
	if _, ok := e.status[feedName]; ok {
		e.status[feedName].Stage = stage
	}
	e.lock.Unlock()
}

func (e *ExportStatus) FinishFeedExport(feedName string, err error) {
	e.lock.Lock()
	if _, ok := e.status[feedName]; ok {
		e.status[feedName].Finished = true
		if err != nil {
			e.status[feedName].HasError = true
			e.status[feedName].StatusMessage = err.Error()
		}
	}
	e.lock.Unlock()
}

func (e *ExportStatus) ReadStatus() map[string]*ExportStatusItem {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.status
}

func (e *ExportStatus) PurgeFinished() {
	e.lock.Lock()
	filter := map[string]*ExportStatusItem{}
	for key, item := range e.status {
		if !(item.Started && item.Finished) {
			filter[key] = item
		}
	}
	e.status = filter
	e.lock.Unlock()
}

func (e *ExportStatus) GetExportStarted() bool {
	return e.started
}

func (e *ExportStatus) GetExportCompleted() bool {
	return e.finished
}
