package api

import "time"

// TemplateResponseItem simple representation of a template date
type TemplateResponseItem struct {
	ID         string
	Name       string
	ModifiedAt time.Time
}
