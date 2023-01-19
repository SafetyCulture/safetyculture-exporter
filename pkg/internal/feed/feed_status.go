package feed

import "sync"

func NewExportStatus() *ExportStatus {
	return &ExportStatus{
		status: map[string]*ExportStatusItem{},
	}
}

type ExportStatus struct {
	lock     sync.Mutex
	status   map[string]*ExportStatusItem
	finished bool
	started  bool
}

const (
	InProgress string = "In Progress"
	Failed     string = "Failed"
	Completed  string = "Complete"
)

type ExportStatusItem struct {
	Name          string
	Status        string
	EstRemaining  int64
	StatusMessage string
}

func (e *ExportStatus) UpdateStatus(feedName string, remaining int64, err error) {
	e.lock.Lock()
	if err != nil {
		e.status[feedName] = &ExportStatusItem{
			Name:          feedName,
			Status:        Failed,
			StatusMessage: err.Error(),
		}
	} else if remaining == 0 {
		e.status[feedName] = &ExportStatusItem{
			Name:         feedName,
			Status:       Completed,
			EstRemaining: remaining,
		}
	} else {
		e.status[feedName] = &ExportStatusItem{
			Name:         feedName,
			Status:       InProgress,
			EstRemaining: remaining,
		}
	}
	e.lock.Unlock()
}

func (e *ExportStatus) ReadStatus() map[string]*ExportStatusItem {
	return e.status
}

func (e *ExportStatus) GetExportStarted() bool {
	return e.started
}

func (e *ExportStatus) GetExportCompleted() bool {
	return e.finished
}

func (e *ExportStatus) ResetStatus() {
	e.lock.Lock()
	e.status = map[string]*ExportStatusItem{}
	e.lock.Unlock()
}
