package exporter_test

import (
	"log"
	"os"

	"github.com/SafetyCulture/safetyculture-exporter/pkg/internal/exporter"
)

// getTemporaryJSONExporter creates a JSONExporter that writes to a temp folder
func getTemporaryJSONExporter() exporter.SafetyCultureJSONExporter {
	dir, err := os.MkdirTemp("", "export")
	if err != nil {
		log.Fatal(err)
	}

	return exporter.NewJSONExporter(dir)
}
