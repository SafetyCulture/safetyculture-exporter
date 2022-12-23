package api

import "time"

// TemplateResponseItem simple representation of a template date
type TemplateResponseItem struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at"`
}

type ExportStatusResponse struct {
	Feeds []ExportStatusResponseItem `json:"feeds"`
}

type ExportStatusResponseItem struct {
	FeedName    string `json:"feed_name"`
	Started     bool   `json:"has_started"`
	DebugString string `json:"debug_string"` // TODO: THIS MUST BE DECOMPOSED
}
