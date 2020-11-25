package feed

import (
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"

	"go.uber.org/zap"
)

// SchemaExporter is an interface to export data feeds to CSV files
type SchemaExporter struct {
	*SQLExporter
	Logger *zap.SugaredLogger
	Output io.Writer
}

type Schema struct {
	Index        int    `gorm:"column:cid"`
	Name         string `gorm:"column:name"`
	Type         string `gorm:"column:type"`
	NotNull      int    `gorm:"column:notnull"`
	DefaultValue string `gorm:"column:dflt_value"`
	PrimaryKey   int    `gorm:"column:pk"`
}

func IsPrimaryKey(pk int) string {
	if pk > 0 {
		return "true"
	}
	return ""
}

func (e *SchemaExporter) WriteSchema(feed Feed) error {
	e.Logger.Infof("Schema for %s:", feed.Name())

	schema := &[]*Schema{}

	resp := e.DB.Raw(fmt.Sprintf("PRAGMA table_info('%s') ", feed.Name())).Scan(schema)
	if resp.Error != nil {
		return resp.Error
	}

	table := tablewriter.NewWriter(e.Output)
	table.SetHeader([]string{"Name", "Type", "Primary Key"})

	for _, v := range *schema {
		table.Append([]string{v.Name, v.Type, IsPrimaryKey(v.PrimaryKey)})
	}
	table.Render()

	return nil
}

func NewSchemaExporter(output io.Writer) (*SchemaExporter, error) {
	sqlExporter, err := NewSQLExporter("sqlite", "file::memory:?cache=shared", true)
	if err != nil {
		return nil, err
	}

	return &SchemaExporter{
		SQLExporter: sqlExporter,
		Logger:      sqlExporter.Logger,
		Output:      output,
	}, nil
}
