package exporter_test

import (
	"io/ioutil"
	"log"

	"github.com/SafetyCulture/iauditor-exporter/internal/app/exporter"
)

// getTemporaryJSONExporter creates a JSONExporter that writes to a temp folder
func getTemporaryJSONExporter() exporter.Exporter {
	dir, err := ioutil.TempDir("", "export")
	if err != nil {
		log.Fatal(err)
	}

	return exporter.NewJSONExporter(dir)
}
