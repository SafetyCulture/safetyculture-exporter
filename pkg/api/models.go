package api

import (
	"time"
)

// TemplateResponseItem simple representation of a template date
type TemplateResponseItem struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at"`
}

type ExportStatusResponse struct {
	ExportStarted   bool                       `json:"export_started"`
	ExportCompleted bool                       `json:"export_completed"`
	Feeds           []ExportStatusResponseItem `json:"feeds"`
}

// ExportStatusResponseItem representation of Feed Export Status
type ExportStatusResponseItem struct {
	FeedName           string `json:"feed_name"`
	Counter            int64  `json:"counter"`
	CounterDecremental bool   `json:"counter_decremental"`
	StatusMessage      string `json:"status_message"`
	Started            bool   `json:"started"`
	Finished           bool   `json:"finished"`
	Stage              string `json:"stage"`
	HasError           bool   `json:"has_error"`
	DurationMs         int64  `json:"duration_ms"`
}
