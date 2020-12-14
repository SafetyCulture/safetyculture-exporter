package exporter_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLastModifiedAt(t *testing.T) {
	tmpExporter := getTemporaryJSONExporter()
	now := time.Now()
	tmpExporter.SetLastModifiedAt(now)

	lastModified := tmpExporter.GetLastModifiedAt()
	assert.NotNil(t, lastModified)

	expected := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second())
	actual := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		lastModified.Year(), lastModified.Month(), lastModified.Day(),
		lastModified.Hour(), lastModified.Minute(), lastModified.Second())
	assert.Equal(t, expected, actual)
}

func TestWriteRow(t *testing.T) {
	tmpExporter := getTemporaryJSONExporter()
	str := "sample-string"
	var tmp json.RawMessage
	tmp = []byte(str)
	tmpExporter.WriteRow("tmp-file", &tmp)
}
