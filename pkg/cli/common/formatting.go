package common

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag"
)

type OutputFormat enumflag.Flag

const (
	FormatTable OutputFormat = iota
	FormatJSON
	FormatYAML
	FormatCSV
)

type OutputSortFields map[string]table.SortMode

type OutputFormatFunc func(object interface{}, column string, field interface{}) string

type OutputFormatters map[string]OutputFormatFunc

// Map format values to their textual representations
var OutputFormatIDs = map[OutputFormat][]string{
	FormatTable: {"table"},
	FormatJSON:  {"json"},
	FormatYAML:  {"yaml"},
	FormatCSV:   {"csv"},
}

// FormattingOptions contains output formatting parameters
type FormattingOptions struct {
	Format OutputFormat
	Fields []string
	SortBy OutputSortFields
	// Custom formatting functions
	Formatters OutputFormatters
}

func NewFormattingOptions(defaultFields []string, sortFields OutputSortFields, formatters OutputFormatters) (o *FormattingOptions) {
	o = &FormattingOptions{}
	if sortFields != nil {
		o.SortBy = sortFields
	}
	if defaultFields != nil {
		o.Fields = defaultFields
	}
	if formatters != nil {
		o.Formatters = formatters
	}

	return
}

func (o *FormattingOptions) addFormattingFlags(cmd *cobra.Command, withTable bool) {

	formats := make([]string, 0)
	for i, f := range OutputFormatIDs {
		if !withTable && (i == FormatTable || i == FormatCSV) {
			continue
		}
		formats = append(formats, f[0])
	}
	cmd.Flags().Var(
		enumflag.New(&o.Format, "format", OutputFormatIDs, enumflag.EnumCaseInsensitive),
		"format",
		fmt.Sprintf("specify the output format. Possible values are: %s", strings.Join(formats, ", ")),
	)

	cmd.Use = fmt.Sprintf("%s [--format {%s}] ", cmd.Use, strings.Join(formats, ","))

	if withTable {
		cmd.Flags().StringSliceVar(&o.Fields, "field", o.Fields,
			`specify one or more columns to include in the output.
The field name may also be specified explicitly if different than the column name.
This option only has effect with the 'table' and 'csv' formats.`)

		cmd.Use = fmt.Sprintf("%s [--field COLUMN[:FIELD]]...", cmd.Use)
	}
}

func (o *FormattingOptions) AddMultiValueFormattingFlags(cmd *cobra.Command) {
	o.addFormattingFlags(cmd, true)
}

func (o *FormattingOptions) AddSingleValueFormattingFlags(cmd *cobra.Command) {
	o.addFormattingFlags(cmd, false)
}

func getFieldValue(valueMap map[string]interface{}, fields []string) interface{} {

	fieldValue, hasField := valueMap[fields[0]]
	// field not found
	if !hasField {
		return nil
	}

	// field found and no sub-fields specified
	if len(fields) == 1 {
		return fieldValue
	}

	// field found and sub-fields specified and field is a sub-map
	if subMap, isMap := fieldValue.(map[string]interface{}); isMap {
		return getFieldValue(subMap, fields[1:])
	}

	// field found, but sub-fields specified and field is not a sub-map
	return nil
}

func (o *FormattingOptions) formatTable(out io.Writer, values []interface{}) {

	columns := make([]string, len(o.Fields))
	fields := make([][]string, len(o.Fields))
	for i, f := range o.Fields {
		f = strings.TrimSpace(f)

		// Split the field name and extract the explicit field name, if supplied, then
		// split each field name into sub-fields.
		s := strings.Split(f, ":")
		if len(s) > 1 {
			columns[i] = s[0]
			fields[i] = strings.Split(strings.ToLower(s[len(s)-1]), ".")
		} else if len(s) == 1 {
			columns[i] = s[0]
			fields[i] = strings.Split(strings.ToLower(s[0]), ".")
		} else {
			columns[i] = f
			fields[i] = strings.Split(strings.ToLower(f), ".")
		}
	}

	t := table.NewWriter()
	t.SetOutputMirror(out)
	t.SetStyle(table.Style{
		Name: "fuseml",
		Format: table.FormatOptions{
			Footer: text.FormatUpper,
			Header: text.FormatDefault,
			Row:    text.FormatDefault,
		},
		Box: table.StyleBoxDefault,
		Options: table.Options{
			DrawBorder:      true,
			SeparateColumns: true,
			SeparateFooter:  true,
			SeparateHeader:  true,
			SeparateRows:    false,
		},
	})
	header := table.Row{}
	for _, h := range columns {
		header = append(header, h)
	}
	t.AppendHeader(header)
	for _, value := range values {

		// First, we encode the value as JSON, then decode it back as a generic unstructured
		// data into a map. We do this to be able to more easily extract its fields.
		// Another nice side-effect of doing the conversion is we don't need to worry about
		// upper-case chars in the key names, given that all keys are lower-cased.
		var valueMap map[string]interface{}
		jsonValue, err := yaml.MarshalWithOptions(value, yaml.JSON())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to encode value %v as JSON: %s", value, err.Error())
		}
		err = yaml.Unmarshal(jsonValue, &valueMap)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to decode value %s as map: %s", jsonValue, err.Error())
		}

		row := make([]interface{}, len(o.Fields))

		for i, f := range fields {
			v := getFieldValue(valueMap, f)

			// if a custom formatter is configured for the column, use it
			if fmt, hasFmt := o.Formatters[columns[i]]; hasFmt {
				v = fmt(value, columns[i], v)
			}

			if v == nil {
				// the table formatter doesn't handle nil values very well
				v = ""
			}
			row[i] = v

		}

		t.AppendRow(row)
	}
	sort := make([]table.SortBy, 0)
	for n, m := range o.SortBy {
		sort = append(sort, table.SortBy{Name: n, Mode: m})
	}
	t.SortBy(sort)

	if o.Format == FormatCSV {
		t.RenderCSV()
	} else {
		t.Render()
	}
}

func (o *FormattingOptions) formatObject(out io.Writer, value interface{}) {
	var m []byte

	if o.Format == FormatYAML {
		fmt.Fprintln(out, "---")
		m, _ = yaml.Marshal(value)
	} else {
		m, _ = yaml.MarshalWithOptions(value, yaml.JSON())
	}
	fmt.Fprintln(out, string(m))
}

func (o *FormattingOptions) FormatValue(out io.Writer, value interface{}) {
	switch o.Format {
	case FormatTable, FormatCSV:
		s := reflect.ValueOf(value)
		if s.Kind() == reflect.Array || s.Kind() == reflect.Slice {
			valueList := make([]interface{}, s.Len())
			for i := 0; i < s.Len(); i++ {
				valueList[i] = s.Index(i).Interface()
			}
			o.formatTable(out, valueList)
		} else {
			o.formatObject(out, value)
		}
	case FormatYAML, FormatJSON:
		o.formatObject(out, value)
	}
}
