// Package cli contains helpers used by transport-specific command-line client
// generators for parsing the command-line flags to identify the service and
// the method to make a request along with the request payload to be sent.
package cli

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
	http "goa.design/goa/v3/http/codegen"
)

type (
	// CommandData contains the data needed to render a command.
	CommandData struct {
		// Name of command e.g. "cellar-storage"
		Name string
		// VarName is the name of the command variable e.g.
		// "cellarStorage"
		VarName string
		// Description is the help text.
		Description string
		// Subcommands is the list of endpoint commands.
		Subcommands []*SubcommandData
		// PkgName is the service HTTP client package import name,
		// e.g. "storagec".
		PkgName string
	}

	// SubcommandData contains the data needed to render a sub-command.
	SubcommandData struct {
		// Name is the sub-command name e.g. "add"
		Name string
		// FullName is the sub-command full name e.g. "storageAdd"
		FullName string
		// Description is the help text.
		Description string
		// MethodVarName is the endpoint method name, e.g. "Add"
		MethodVarName string
		// BuildFunction contains the data to generate a payload builder function
		// if any.
		BuildFunction *BuildFunctionData
	}

	// BuildFunctionData contains the data needed to generate a constructor
	// function that builds a service method payload type from values extracted
	// from command line flags.
	BuildFunctionData struct {
		// Name is the build payload function name.
		Name string
		// Description describes the payload function.
		Description string
		// Params is the list of build function parameters names.
		Params []*ParamData
		// ServiceName is the name of the service.
		ServiceName string
		// MethodName is the name of the method.
		MethodName string
		// ResultType is the fully qualified payload type name.
		ResultType string
		// Fields describes the payload fields.
		Fields []*FieldData
		// PayloadInit contains the data needed to render the function
		// body.
		PayloadInit *PayloadInitData
		// CheckErr is true if the payload initialization code requires an
		// "err error" variable that must be checked.
		CheckErr bool
	}

	// ParamData contains the data needed to generate the parameters accepted by
	// the payload function.
	ParamData struct {
		// Name is the name of the parameter.
		Name string
		// TypeName is the parameter data type.
		TypeName string
	}

	// FieldData contains the data needed to generate the code that initializes a
	// field in the method payload type.
	FieldData struct {
		// Name is the field name, e.g. "Vintage"
		Name string
		// VarName is the name of the local variable holding the field
		// value, e.g. "vintage"
		VarName string
		// TypeRef is the reference to the type.
		TypeRef string
		// Init is the code initializing the variable.
		Init string
	}

	// PayloadInitData contains the data needed to generate a constructor
	// function that initializes a service method payload type from the
	// command-ling arguments.
	PayloadInitData struct {
		// Code is the payload initialization code.
		Code string
		// ReturnTypeAttribute if non-empty returns an attribute in the payload
		// type that describes the shape of the method payload.
		ReturnTypeAttribute string
		// ReturnTypeAttributePointer is true if the return type attribute
		// generated struct field holds a pointer
		ReturnTypeAttributePointer bool
		// ReturnIsStruct if true indicates that the method payload is an object.
		ReturnIsStruct bool
		// ReturnTypeName is the fully-qualified name of the payload.
		ReturnTypeName string
		// ReturnTypePkg is the package name where the payload is present.
		ReturnTypePkg string
		// Args is the list of arguments for the constructor.
		Args []*codegen.InitArgData
	}
)

// BuildCommandData builds the data needed by CLI code generators to render the
// parsing of the service command.
func BuildCommandData(data *service.Data) *CommandData {
	description := data.Description
	if description == "" {
		description = fmt.Sprintf("Make requests to the %q service", data.Name)
	}
	return &CommandData{
		Name:        codegen.KebabCase(data.Name),
		VarName:     codegen.Goify(data.Name, false),
		Description: description,
		PkgName:     data.PkgName + "c",
	}
}

// BuildSubcommandData builds the data needed by CLI code generators to render
// the CLI parsing of the service sub-command.
func BuildSubcommandData(svcName string, m *service.MethodData, buildFunction *BuildFunctionData) *SubcommandData {
	var (
		name        string
		fullName    string
		description string
	)
	{
		en := m.Name
		name = codegen.KebabCase(en)
		fullName = goifyTerms(svcName, en)
		description = m.Description
		if description == "" {
			description = fmt.Sprintf("Make request to the %q endpoint", m.Name)
		}
	}
	sub := &SubcommandData{
		Name:          name,
		FullName:      fullName,
		Description:   description,
		MethodVarName: m.VarName,
		BuildFunction: buildFunction,
	}

	return sub
}

// PayloadBuilderSection builds the section template that can be used to
// generate the payload builder code.
func PayloadBuilderSection(buildFunction *BuildFunctionData) *codegen.SectionTemplate {
	return &codegen.SectionTemplate{
		Name:   "cli-build-payload",
		Source: buildPayloadT,
		Data:   buildFunction,
		FuncMap: map[string]interface{}{
			"fieldCode": fieldCode,
		},
	}
}

// FieldLoadCode returns the code used in the build payload function that
// initializes one of the payload object fields. It returns the initialization
// code and a boolean indicating whether the code requires an "err" variable.
func FieldLoadCode(arg *http.InitArgData, fullName string, payload expr.DataType) (string, bool) {

	var (
		code    string
		declErr bool
		startIf string
		endIf   string
		rval    string
	)
	{
		if !arg.Required && argTypeNeedsConversion(arg) {
			startIf = fmt.Sprintf("if %s != \"\" {\n", fullName)
			endIf = "\n}"
		}
		if expr.IsPrimitive(payload) {
			switch payload {
			case expr.Boolean:
				rval = "false"
			case expr.String:
				rval = "\"\""
			case expr.Bytes, expr.Any:
				rval = "nil"
			default:
				rval = "0"
			}
		} else {
			rval = "nil"
		}
		if !argTypeNeedsConversion(arg) {
			ref := "&"
			if arg.Required || arg.DefaultValue != nil || expr.IsArray(arg.Type) || expr.IsMap(arg.Type) {
				ref = ""
			}
			code = arg.VarName + " = " + ref + fullName
			declErr = arg.Validate != ""
		} else {
			var checkErr bool
			code, declErr, checkErr = conversionCode(fullName, arg.VarName, arg.TypeName, !arg.Required && arg.DefaultValue == nil)
			if checkErr {
				code += "\nif err != nil {\n"
				if flagType(arg.TypeName) == "YAML" {
					code += fmt.Sprintf(`return %v, fmt.Errorf("invalid format for %s, \nerror: %%s", err)`,
						rval, arg.VarName)
				} else {
					code += fmt.Sprintf(`return %v, fmt.Errorf("invalid value for %s, must be %s")`,
						rval, arg.VarName, flagType(arg.TypeName))
				}
				code += "\n}"
			}
		}
		if arg.Validate != "" {
			code += "\n" + arg.Validate + "\n" + fmt.Sprintf("if err != nil {\n\treturn %v, err\n}", rval)
		}
	}

	return fmt.Sprintf("%s%s%s", startIf, code, endIf), declErr
}

// flagType calculates the type of a flag
func flagType(tname string) string {
	switch tname {
	case boolN, intN, int32N, int64N, uintN, uint32N, uint64N, float32N, float64N, stringN:
		return strings.ToUpper(tname)
	case bytesN:
		return "STRING"
	default: // Any, Array, Map, Object, User
		return "YAML"
	}
}

// jsonExample generates a json example
func jsonExample(v interface{}) string {
	// In JSON, keys must be a string. But goa allows map keys to be anything.
	r := reflect.ValueOf(v)
	if r.Kind() == reflect.Map {
		keys := r.MapKeys()
		if keys[0].Kind() != reflect.String {
			a := make(map[string]interface{}, len(keys))
			var kstr string
			for _, k := range keys {
				switch t := k.Interface().(type) {
				case bool:
					kstr = strconv.FormatBool(t)
				case int32:
					kstr = strconv.FormatInt(int64(t), 10)
				case int64:
					kstr = strconv.FormatInt(t, 10)
				case int:
					kstr = strconv.Itoa(t)
				case float32:
					kstr = strconv.FormatFloat(float64(t), 'f', -1, 32)
				case float64:
					kstr = strconv.FormatFloat(t, 'f', -1, 64)
				default:
					kstr = k.String()
				}
				a[kstr] = r.MapIndex(k).Interface()
			}
			v = a
		}
	}
	b, err := json.MarshalIndent(v, "   ", "   ")
	ex := "?"
	if err == nil {
		ex = string(b)
	}
	if strings.Contains(ex, "\n") {
		ex = "'" + strings.Replace(ex, "'", "\\'", -1) + "'"
	}
	return ex
}

// yamlExample generates a yaml example
func yamlExample(v interface{}) string {
	// Scalars are represented on a single line
	r := reflect.ValueOf(v)
	if r.Kind() != reflect.Map && r.Kind() != reflect.Array {
		return fmt.Sprintf("\"%s\"", r)
	}
	b, err := yaml.Marshal(v)
	ex := "?"
	if err == nil {
		ex = "\"" + string(b) + "\""
	}
	return ex
}

var (
	boolN    = codegen.GoNativeTypeName(expr.Boolean)
	intN     = codegen.GoNativeTypeName(expr.Int)
	int32N   = codegen.GoNativeTypeName(expr.Int32)
	int64N   = codegen.GoNativeTypeName(expr.Int64)
	uintN    = codegen.GoNativeTypeName(expr.UInt)
	uint32N  = codegen.GoNativeTypeName(expr.UInt32)
	uint64N  = codegen.GoNativeTypeName(expr.UInt64)
	float32N = codegen.GoNativeTypeName(expr.Float32)
	float64N = codegen.GoNativeTypeName(expr.Float64)
	stringN  = codegen.GoNativeTypeName(expr.String)
	bytesN   = codegen.GoNativeTypeName(expr.Bytes)
)

// conversionCode produces the code that converts the string contained in the
// variable named from to the value stored in the variable "to" of type
// typeName. The second return value indicates whether the "err" variable must
// be declared prior to the conversion code being rendered. The last return
// value indicates whether the generated code can produce errors (i.e.
// initialize the err variable).
func conversionCode(from, to, typeName string, pointer bool) (string, bool, bool) {
	var (
		parse string
		cast  string

		target   = to
		needCast = typeName != stringN && typeName != bytesN && flagType(typeName) != "YAML"
		declErr  = true
		checkErr = true
		decl     = ""
	)
	if needCast && pointer {
		target = "val"
		decl = ":"
	}
	switch typeName {
	case boolN:
		if pointer {
			parse = fmt.Sprintf("var %s bool\n", target)
		}
		parse += fmt.Sprintf("%s, err = strconv.ParseBool(%s)", target, from)
	case intN:
		parse = fmt.Sprintf("var v int64\nv, err = strconv.ParseInt(%s, 10, 64)", from)
		cast = fmt.Sprintf("%s %s= int(v)", target, decl)
	case int32N:
		parse = fmt.Sprintf("var v int64\nv, err = strconv.ParseInt(%s, 10, 32)", from)
		cast = fmt.Sprintf("%s %s= int32(v)", target, decl)
	case int64N:
		parse = fmt.Sprintf("%s, err %s= strconv.ParseInt(%s, 10, 64)", target, decl, from)
		declErr = decl == ""
	case uintN:
		parse = fmt.Sprintf("var v uint64\nv, err = strconv.ParseUint(%s, 10, 64)", from)
		cast = fmt.Sprintf("%s %s= uint(v)", target, decl)
	case uint32N:
		parse = fmt.Sprintf("var v uint64\nv, err = strconv.ParseUint(%s, 10, 32)", from)
		cast = fmt.Sprintf("%s %s= uint32(v)", target, decl)
	case uint64N:
		parse = fmt.Sprintf("%s, err %s= strconv.ParseUint(%s, 10, 64)", target, decl, from)
		declErr = decl == ""
	case float32N:
		parse = fmt.Sprintf("var v float64\nv, err = strconv.ParseFloat(%s, 32)", from)
		cast = fmt.Sprintf("%s %s= float32(v)", target, decl)
	case float64N:
		parse = fmt.Sprintf("%s, err %s= strconv.ParseFloat(%s, 64)", target, decl, from)
		declErr = decl == ""
	case stringN:
		parse = fmt.Sprintf("%s %s= %s", target, decl, from)
		declErr = false
		checkErr = false
	case bytesN:
		parse = fmt.Sprintf("%s %s= []byte(%s)", target, decl, from)
		declErr = false
		checkErr = false
	default:
		parse = fmt.Sprintf("err = yaml.UnmarshalWithOptions([]byte(%s), &%s, yaml.Strict())", from, target)
	}
	if !needCast {
		return parse, declErr, checkErr
	}
	if cast != "" {
		parse = parse + "\n" + cast
	}
	if to != target {
		ref := ""
		if pointer {
			ref = "&"
		}
		parse = parse + fmt.Sprintf("\n%s = %s%s", to, ref, target)
	}
	return parse, declErr, checkErr
}

// goifyTerms makes valid go identifiers out of the supplied terms
func goifyTerms(terms ...string) string {
	res := codegen.Goify(terms[0], false)
	if len(terms) == 1 {
		return res
	}
	for _, t := range terms[1:] {
		res += codegen.Goify(t, true)
	}
	return res
}

// fieldCode generates code to initialize the data structures fields
// from the given args. It is used only in templates.
func fieldCode(init *PayloadInitData) string {
	varn := "res"
	if init.ReturnTypeAttribute == "" {
		varn = "v"
	}
	// We can ignore the transform helpers as there won't be any generated
	// because the args cannot be user types.
	c, _, err := codegen.InitStructFields(init.Args, init.ReturnTypeName, varn, "", init.ReturnTypePkg, init.Code == "")
	if err != nil {
		panic(err) //bug
	}
	return c
}

// input: buildFunctionData
const buildPayloadT = `{{ printf "%s builds the payload for the %s %s endpoint from CLI flags." .Name .ServiceName .MethodName | comment }}
func {{ .Name }}({{ range .Params }}{{ .Name }} {{ .TypeName }}, {{ end }}) ({{ .ResultType }}, error) {
{{- if .CheckErr }}
	var err error
{{- end }}
{{- range .Fields }}
	{{- if .VarName }}
		var {{ .VarName }} {{ .TypeRef }}
		{
			{{ .Init }}
		}
	{{- end }}
{{- end }}
{{- with .PayloadInit }}
	{{- if .Code }}
		{{ .Code }}
		{{- if .ReturnTypeAttribute }}
			res := &{{ .ReturnTypeName }}{
				{{ .ReturnTypeAttribute }}: {{ if .ReturnTypeAttributePointer }}&{{ end }}v,
			}
		{{- end }}
	{{- end }}
	{{- if .ReturnIsStruct }}
		{{- if not .Code }}
		{{ if .ReturnTypeAttribute }}res{{ else }}v{{ end }} := &{{ .ReturnTypeName }}{}
		{{- end }}
		{{ fieldCode . }}
	{{- end }}
	return {{ if .ReturnTypeAttribute }}res{{ else }}v{{ end }}, nil
{{- end }}
}
`
