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

// OutputFormat encodes the available formats that can be used to display structured data
type OutputFormat enumflag.Flag

// Supported output formats
const (
	FormatTable OutputFormat = iota
	FormatJSON
	FormatYAML
	FormatCSV
)

// OutputFormatFunc is a handler used to customize how a tabular field is formatted.
// The formatting utility calls the handler with the following arguments:
//  - object: the golang struct that represents the table row being formatted
//  - colume: the table column for which this handler is called
//  - field: the object that needs formatting
//
// The handler must return a textual representation of the `field` object.
type OutputFormatFunc func(object interface{}, column string, field interface{}) string

// OutputFormatters is a map of output formatters, indexed by column name
type OutputFormatters map[string]OutputFormatFunc

// OutputFormatIDs maps format values to their textual representations
var OutputFormatIDs = map[OutputFormat][]string{
	FormatTable: {"table"},
	FormatJSON:  {"json"},
	FormatYAML:  {"yaml"},
	FormatCSV:   {"csv"},
}

// FormattingOptions contains output formatting parameters
type FormattingOptions struct {
	// Output format
	Format OutputFormat
	// List of field specifiers controlling how information is converted from structured data into tabular format.
	// Each value can be formatted using the following syntax:
	//
	//  <column-name>[:<field-name>[.<subfield-name>[...]]]
	//
	// The <column-name> token represents the name used for the header. If a field is not specified, it is also
	// interpreted as a field name and its value is not case-sensitive as far as it concerns matching the field names
	// in the structure information being formatting.
	//
	// The <field-name> and subsequent <subfield-name> tokens are used to identify the exact (sub)field in the hierarchically
	// structured information that the table column maps to. Their values are not case-sensitive.
	Fields []string
	// List of column names and their associated sorting mode, in sorting order.
	SortBy []table.SortBy
	// Custom formatting functions
	Formatters OutputFormatters
}

// NewFormattingOptions initializes formatting options for a cobra command. It accepts the following arguments:
//  - fields: list of field specifiers controlling how information is converted from structured data into tabular format
//  (see FormattingOptions/Fields).
//  - sortFields: list of sort specifiers. Each specifier should indicate the column name and sort mode. The order is significant
//  and will determine the order in which columns will be sorted.
//  - formatters: map of custom formatters. Use this to attach custom formatting handlers to columns that are not
//  handled properly by the default formatting.
func NewFormattingOptions(fields []string, sortFields []table.SortBy, formatters OutputFormatters) (o *FormattingOptions) {
	o = &FormattingOptions{}
	if sortFields != nil {
		o.SortBy = sortFields
	}
	if fields != nil {
		o.Fields = fields
	}
	if formatters != nil {
		o.Formatters = formatters
	}

	return
}

// NewSingleValueFormattingOptions initializes formatting options for a cobra command. Use this method instead of NewTableFormattingOptions
// if your command doesn't need to format list of objects in a table layout.
func NewSingleValueFormattingOptions() (o *FormattingOptions) {
	return &FormattingOptions{}
}

func (o *FormattingOptions) addFormattingFlags(cmd *cobra.Command, withTable bool) {

	mapping := make(map[OutputFormat][]string)

	// remove table formats (table and CSV) from the enum values if table layout is not required
	formats := make([]string, 0)
	for i, f := range OutputFormatIDs {
		if !withTable && (i == FormatTable || i == FormatCSV) {
			continue
		}
		formats = append(formats, f[0])
		mapping[i] = f
	}
	if !withTable {
		// default to YAML if not using table formatting
		o.Format = FormatYAML
	}
	cmd.Flags().Var(
		enumflag.New(&o.Format, "format", mapping, enumflag.EnumCaseInsensitive),
		"format",
		fmt.Sprintf("specify the output format. Possible values are: %s", strings.Join(formats, ", ")),
	)

	cmd.Use = fmt.Sprintf("%s [--format {%s}]", cmd.Use, strings.Join(formats, ","))

	if withTable {
		cmd.Flags().StringSliceVar(&o.Fields, "field", o.Fields,
			`specify one or more columns to include in the output.
The field name may also be specified explicitly if different than the column name.
This option only has effect with the 'table' and 'csv' formats.`)

		cmd.Use = fmt.Sprintf("%s [--field COLUMN[:FIELD]]...", cmd.Use)
	}
}

// AddMultiValueFormattingFlags adds formatting command line flags to a cobra command.
// This function includes tabular formatting parameters. If the command only outputs
// single objects, use AddSingleValueFormattingFlags instead.
func (o *FormattingOptions) AddMultiValueFormattingFlags(cmd *cobra.Command) {
	o.addFormattingFlags(cmd, true)
}

// AddSingleValueFormattingFlags adds formatting command line flags to a cobra command.
// This function does not include tabular formatting parameters. If the command also outputs
// lists of objects that can be formatted using a tabular layout, use AddMultiValueFormattingFlags
// instead.
func (o *FormattingOptions) AddSingleValueFormattingFlags(cmd *cobra.Command) {
	o.addFormattingFlags(cmd, false)
}

// Recursive function that extracts a subfield from a generic hierarchical structure.
// This really a simple and far less powerful alternative to JSONPath and YAML path.
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

	// Convert the formatting field specifiers into column names and field/subfield names
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
			Header: text.FormatUpper,
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

			// if the value is not a scalar, format it using YAML
			t := reflect.TypeOf(v)
			if t.Kind() == reflect.Array || t.Kind() == reflect.Slice || t.Kind() == reflect.Map {
				if j, err := yaml.Marshal(v); err == nil {
					v = string(j)
				}
			}

			row[i] = v
		}

		t.AppendRow(row)
	}
	t.SortBy(o.SortBy)

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

// FormatValue formats any go struct or list of structs that can be converted into JSON,
// according to the configured formatting options.
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
