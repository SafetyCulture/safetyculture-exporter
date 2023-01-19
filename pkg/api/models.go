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
	FeedName      string `json:"feed_name"`
	Status        string `json:"status"`
	Remaining     int64  `json:"remaining"`
	StatusMessage string `json:"status_message"`
}
