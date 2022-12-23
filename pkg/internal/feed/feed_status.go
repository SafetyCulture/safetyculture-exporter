package feed

import "sync"

func NewExportStatus() *ExportStatus {
	return &ExportStatus{
		status: map[string]ExportStatusItem{},
	}
}

type ExportStatus struct {
	lock     sync.Mutex
	status   map[string]ExportStatusItem
	finished bool
	started  bool
}

type ExportStatusItem struct {
	Name         string
	Started      bool
	EstRemaining int64
}

func (e *ExportStatus) UpdateStatus(feedName string, status ExportStatusItem) {
	e.lock.Lock()
	e.status[feedName] = status
	e.lock.Unlock()
}

func (e *ExportStatus) ReadStatus() map[string]ExportStatusItem {
	e.lock.Lock()
	temp := make(map[string]ExportStatusItem, len(e.status))
	for k, v := range e.status {
		temp[k] = v
	}
	e.lock.Unlock()
	return temp
}

func (e *ExportStatus) GetExportStarted() bool {
	return e.started
}

func (e *ExportStatus) GetExportCompleted() bool {
	return e.finished
}
