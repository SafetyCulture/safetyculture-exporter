package feed

import (
	"fmt"
	"io"
	"time"

	"github.com/olekukonko/tablewriter"

	"go.uber.org/zap"
)

// SchemaExporter is an interface to export data feeds to CSV files
type SchemaExporter struct {
	*SQLExporter
	Logger *zap.SugaredLogger
	Output io.Writer
}

type schema struct {
	Index        int    `gorm:"column:cid"`
	Name         string `gorm:"column:name"`
	Type         string `gorm:"column:type"`
	NotNull      int    `gorm:"column:notnull"`
	DefaultValue string `gorm:"column:dflt_value"`
	PrimaryKey   int    `gorm:"column:pk"`
}

func isPrimaryKey(pk int) string {
	if pk > 0 {
		return "true"
	}
	return ""
}

// WriteSchema writes schema of a feed to output in tabular format
func (e *SchemaExporter) WriteSchema(feed Feed) error {
	e.Logger.With("feed", feed.Name()).Info("writing schema")

	schema := &[]*schema{}

	resp := e.DB.Raw(fmt.Sprintf("PRAGMA table_info('%s') ", feed.Name())).Scan(schema)
	if resp.Error != nil {
		return resp.Error
	}

	table := tablewriter.NewWriter(e.Output)
	table.SetHeader([]string{"Name", "Type", "Primary Key"})

	for _, v := range *schema {
		table.Append([]string{v.Name, v.Type, isPrimaryKey(v.PrimaryKey)})
	}
	table.Render()

	return nil
}

// GetDuration will return the duration for exporting a batch
func (e *SchemaExporter) GetDuration() time.Duration {
	// NOT IMPLEMENTED
	return 0
}

// NewSchemaExporter creates a new instance of SchemaExporter
func NewSchemaExporter(output io.Writer) (*SchemaExporter, error) {
	sqlExporter, err := NewSQLExporter("sqlite", "file::memory:?cache=shared", true, "")
	if err != nil {
		return nil, err
	}

	return &SchemaExporter{
		SQLExporter: sqlExporter,
		Logger:      sqlExporter.Logger,
		Output:      output,
	}, nil
}
