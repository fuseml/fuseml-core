package cli

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	http "goa.design/goa/v3/http/codegen"
)

// init registers the plugin generator function.
func init() {
	codegen.RegisterPlugin("fuseml-cli", "gen", nil, Generate)
}

// commandData wraps the common CommandData and adds HTTP-specific fields.
type commandData struct {
	*CommandData
	// Subcommands is the list of endpoint commands.
	Subcommands []*subcommandData
	// NeedStream if true initializes the websocket dialer.
	NeedStream bool
}

// commandData wraps the common SubcommandData and adds HTTP-specific fields.
type subcommandData struct {
	*SubcommandData
	// MultipartFuncName is the name of the function used to render a multipart
	// request encoder.
	MultipartFuncName string
	// MultipartFuncName is the name of the variabl used to render a multipart
	// request encoder.
	MultipartVarName string
	// StreamFlag is the flag used to identify the file to be streamed when
	// the endpoint uses SkipRequestBodyEncodeDecode.
	StreamFlag *FlagData
	// BuildStreamPayload is the name of the generated function that builds the
	// request data structure that wraps the payload and the file stream for
	// endpoints that use SkipRequestBodyEncodeDecode.
	BuildStreamPayload string
}

// Generate produces the fuseml modified CLI code files
func Generate(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	for _, root := range roots {
		if root, ok := root.(*expr.RootExpr); ok {
			fuseml_cli_files := ClientCLIFiles(genpkg, root)
			for _, file := range files {
				if f, isModified := fuseml_cli_files[file.Path]; isModified {
					*file = *f
					fmt.Printf("FuseML overriden file: %s\n", file.Path)
				}
				// fmt.Printf("FUSEML: file %s\n", file.Path)
				// for _, template := range file.SectionTemplates {
				// 	fmt.Printf("FUSEML:     template %s\n", template.Name)
				// 	fmt.Printf("FUSEML:         source %s\n", template.Source)
				// 	for fnc, _ := range template.FuncMap {
				// 		fmt.Printf("FUSEML:         function %s\n", fnc)
				// 	}

				// }
			}
		}
	}
	return files, nil
}

// ClientCLIFiles returns the client HTTP CLI support files mapped according to their full path.
func ClientCLIFiles(genpkg string, root *expr.RootExpr) map[string]*codegen.File {
	if len(root.API.HTTP.Services) == 0 {
		return nil
	}
	var (
		data []*commandData
		svcs []*expr.HTTPServiceExpr
	)
	for _, svc := range root.API.HTTP.Services {
		sd := http.HTTPServices.Get(svc.Name())
		if len(sd.Endpoints) > 0 {
			command := &commandData{
				CommandData: BuildCommandData(sd.Service),
				NeedStream:  false, // hasWebSocket(sd),
			}

			for _, e := range sd.Endpoints {
				sub := buildSubcommandData(sd, e)
				command.Subcommands = append(command.Subcommands, sub)
				command.CommandData.Subcommands = append(command.CommandData.Subcommands, sub.SubcommandData)
			}

			command.Example = command.Subcommands[0].Example

			data = append(data, command)
			svcs = append(svcs, svc)
		}
	}
	files := make(map[string]*codegen.File)
	for _, svr := range root.API.Servers {
		file := endpointParser(genpkg, root, svr, data)
		files[file.Path] = file
	}
	for i, svc := range svcs {
		file := payloadBuilders(genpkg, svc, data[i].CommandData)
		files[file.Path] = file
	}
	return files
}

func buildSubcommandData(sd *http.ServiceData, e *http.EndpointData) *subcommandData {
	flags, buildFunction := buildFlags(sd, e)

	sub := &subcommandData{
		SubcommandData: BuildSubcommandData(sd.Service.Name, e.Method, buildFunction, flags),
	}
	if e.MultipartRequestEncoder != nil {
		sub.MultipartVarName = e.MultipartRequestEncoder.VarName
		sub.MultipartFuncName = e.MultipartRequestEncoder.FuncName
	}
	if e.Method.SkipRequestBodyEncodeDecode {
		sub.StreamFlag = streamFlag(sd.Service.Name, e.Method.Name)
		sub.BuildStreamPayload = e.BuildStreamPayload
	}
	return sub
}

// endpointParser returns the file that implements the command line parser that
// builds the client endpoint and payload necessary to perform a request.
func endpointParser(genpkg string, root *expr.RootExpr, svr *expr.ServerExpr, data []*commandData) *codegen.File {
	pkg := codegen.SnakeCase(codegen.Goify(svr.Name, true))
	path := filepath.Join(codegen.Gendir, "http", "cli", pkg, "cli.go")
	title := fmt.Sprintf("%s HTTP client CLI support package", svr.Name)
	specs := []*codegen.ImportSpec{
		{Path: "encoding/json"},
		{Path: "flag"},
		{Path: "fmt"},
		{Path: "net/http"},
		{Path: "os"},
		{Path: "strconv"},
		{Path: "unicode/utf8"},
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
	}
	for _, sv := range svr.Services {
		svc := root.Service(sv)
		sd := http.HTTPServices.Get(svc.Name)
		if sd == nil {
			continue
		}
		specs = append(specs, &codegen.ImportSpec{
			Path: genpkg + "/http/" + codegen.SnakeCase(sd.Service.VarName) + "/client",
			Name: sd.Service.PkgName + "c",
		})
	}

	cliData := make([]*CommandData, len(data))
	for i, cmd := range data {
		cliData[i] = cmd.CommandData
	}

	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "cli", specs),
		UsageCommands(cliData),
		UsageExamples(cliData),
		{
			Name:   "parse-endpoint",
			Source: parseEndpointT,
			Data: struct {
				FlagsCode string
				Commands  []*commandData
			}{
				FlagsCode(cliData),
				data,
			},
			FuncMap: map[string]interface{}{"streamingCmdExists": streamingCmdExists},
		},
	}
	for _, cmd := range cliData {
		sections = append(sections, CommandUsage(cmd))
	}
	return &codegen.File{Path: path, SectionTemplates: sections}
}

// payloadBuilders returns the file that contains the payload constructors that
// use flag values as arguments.
func payloadBuilders(genpkg string, svc *expr.HTTPServiceExpr, data *CommandData) *codegen.File {
	sd := http.HTTPServices.Get(svc.Name())
	path := filepath.Join(codegen.Gendir, "http", codegen.SnakeCase(sd.Service.VarName), "client", "cli.go")
	title := fmt.Sprintf("%s HTTP client CLI support package", svc.Name())
	specs := []*codegen.ImportSpec{
		{Path: "encoding/json"},
		{Path: "fmt"},
		{Path: "net/http"},
		{Path: "os"},
		{Path: "strconv"},
		{Path: "unicode/utf8"},
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
		{Path: genpkg + "/" + codegen.SnakeCase(sd.Service.VarName), Name: sd.Service.PkgName},
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", specs),
	}
	for _, sub := range data.Subcommands {
		if sub.BuildFunction != nil {
			sections = append(sections, PayloadBuilderSection(sub.BuildFunction))
		}
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func buildFlags(svc *http.ServiceData, e *http.EndpointData) ([]*FlagData, *BuildFunctionData) {
	var (
		flags         []*FlagData
		buildFunction *BuildFunctionData
	)

	svcn := svc.Service.Name
	en := e.Method.Name
	if e.Payload != nil {
		if e.Payload.Request.PayloadInit != nil {
			args := e.Payload.Request.PayloadInit.ClientArgs
			args = append(args, e.Payload.Request.PayloadInit.CLIArgs...)
			flags, buildFunction = makeFlags(e, args, e.Payload.Request.PayloadType)
		} else if e.Payload.Ref != "" {
			flags = append(flags, NewFlagData(svcn, en, "p", e.Method.PayloadRef, e.Method.PayloadDesc, true, e.Method.PayloadEx, e.Method.PayloadDefault))
		}
	}
	if e.Method.SkipRequestBodyEncodeDecode {
		flags = append(flags, streamFlag(svcn, en))
	}

	return flags, buildFunction
}

func makeFlags(e *http.EndpointData, args []*http.InitArgData, payload expr.DataType) ([]*FlagData, *BuildFunctionData) {
	var (
		fdata     []*FieldData
		flags     = make([]*FlagData, len(args))
		params    = make([]string, len(args))
		pInitArgs = make([]*codegen.InitArgData, len(args))
		check     bool
	)
	for i, arg := range args {
		pInitArgs[i] = &codegen.InitArgData{
			Name:         arg.VarName,
			Pointer:      arg.Pointer,
			FieldName:    arg.FieldName,
			FieldPointer: arg.FieldPointer,
			FieldType:    arg.FieldType,
			Type:         arg.Type,
		}

		f := NewFlagData(e.ServiceName, e.Method.Name, arg.VarName, arg.TypeName, arg.Description, arg.Required, arg.Example, arg.DefaultValue)
		flags[i] = f
		params[i] = f.FullName
		if arg.FieldName == "" && arg.VarName != "body" {
			continue
		}
		code, chek := FieldLoadCode(f, arg.VarName, arg.TypeName, arg.Validate, arg.DefaultValue, payload)
		check = check || chek
		tn := arg.TypeRef
		if f.Type == "JSON" {
			// We need to declare the variable without
			// a pointer to be able to unmarshal the JSON
			// using its address.
			tn = arg.TypeName
		}
		fdata = append(fdata, &FieldData{
			Name:    arg.VarName,
			VarName: arg.VarName,
			TypeRef: tn,
			Init:    code,
		})
	}

	pInit := PayloadInitData{
		Code:                       e.Payload.Request.PayloadInit.ClientCode,
		ReturnTypeAttribute:        e.Payload.Request.PayloadInit.ReturnTypeAttribute,
		ReturnTypeAttributePointer: e.Payload.Request.PayloadInit.ReturnIsPrimitivePointer,
		ReturnIsStruct:             e.Payload.Request.PayloadInit.ReturnIsStruct,
		ReturnTypeName:             e.Payload.Request.PayloadInit.ReturnTypeName,
		ReturnTypePkg:              e.Payload.Request.PayloadInit.ReturnTypePkg,
		Args:                       pInitArgs,
	}

	return flags, &BuildFunctionData{
		Name:         "Build" + e.Method.VarName + "Payload",
		ActualParams: params,
		FormalParams: params,
		ServiceName:  e.ServiceName,
		MethodName:   e.Method.Name,
		ResultType:   e.Payload.Ref,
		Fields:       fdata,
		PayloadInit:  &pInit,
		CheckErr:     check,
	}
}

// streamFlag returns the flag used to specify the upload file for endpoints
// that use SkipRequestBodyEncodeDecode.
func streamFlag(svcn, en string) *FlagData {
	return NewFlagData(svcn, en, "stream", "string", "path to file containing the streamed request body", true, "goa.png", nil)
}

// streamingCmdExists returns true if at least one command in the list of commands
// uses stream for sending payload/result.
func streamingCmdExists(data []*commandData) bool {
	for _, c := range data {
		if c.NeedStream {
			return true
		}
	}
	return false
}

const parseEndpointT = `// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(
	scheme, host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restore bool,
	{{- if streamingCmdExists .Commands }}
	dialer goahttp.Dialer,
		{{- range .Commands }}
			{{- if .NeedStream }}
				{{ .VarName }}Configurer *{{ .PkgName }}.ConnConfigurer,
			{{- end }}
		{{- end }}
	{{- end }}
	{{- range $c := .Commands }}
	{{- range .Subcommands }}
		{{- if .MultipartVarName }}
	{{ .MultipartVarName }} {{ $c.PkgName }}.{{ .MultipartFuncName }},
		{{- end }}
	{{- end }}
	{{- end }}
) (goa.Endpoint, interface{}, error) {
	{{ .FlagsCode }}
    var (
		data     interface{}
		endpoint goa.Endpoint
		err      error
	)
	{
		switch svcn {
	{{- range .Commands }}
		case "{{ .Name }}":
			c := {{ .PkgName }}.NewClient(scheme, host, doer, enc, dec, restore{{ if .NeedStream }}, dialer, {{ .VarName }}Configurer{{ end }})
			switch epn {
		{{- $pkgName := .PkgName }}{{ range .Subcommands }}
			case "{{ .Name }}":
				endpoint = c.{{ .MethodVarName }}({{ if .MultipartVarName }}{{ .MultipartVarName }}{{ end }})
			{{- if .BuildFunction }}
				data, err = {{ $pkgName}}.{{ .BuildFunction.Name }}({{ range .BuildFunction.ActualParams }}*{{ . }}Flag, {{ end }})
			{{- else if .Conversion }}
				{{ .Conversion }}
			{{- else }}
				data = nil
			{{- end }}
			{{- if .StreamFlag }}
				{{- if .BuildFunction }}
				if err == nil {
				{{- end }}
					data, err = {{ $pkgName }}.{{ .BuildStreamPayload }}({{ if or .BuildFunction .Conversion }}data, {{ end }}*{{ .StreamFlag.FullName }}Flag)
				{{- if .BuildFunction }}
				}
				{{- end }}
			{{- end }}
		{{- end }}
			}
	{{- end }}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}
`
