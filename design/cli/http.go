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
		data []*CommandData
		svcs []*expr.HTTPServiceExpr
	)
	for _, svc := range root.API.HTTP.Services {
		sd := http.HTTPServices.Get(svc.Name())
		if len(sd.Endpoints) > 0 {
			command := BuildCommandData(sd.Service)

			for _, e := range sd.Endpoints {
				sub := buildSubcommandData(sd, e)
				command.Subcommands = append(command.Subcommands, sub)
			}

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
		file := payloadBuilders(genpkg, svc, data[i])
		files[file.Path] = file
	}
	return files
}

func buildSubcommandData(sd *http.ServiceData, e *http.EndpointData) *SubcommandData {
	buildFunction := buildPayloadBuildFunction(sd, e)

	sub := BuildSubcommandData(sd.Service.Name, e.Method, buildFunction)
	return sub
}

// endpointParser returns the file that implements the command line parser that
// builds the client endpoint and payload necessary to perform a request.
func endpointParser(genpkg string, root *expr.RootExpr, svr *expr.ServerExpr, data []*CommandData) *codegen.File {
	pkg := codegen.SnakeCase(codegen.Goify(svr.Name, true))
	path := filepath.Join(codegen.Gendir, "http", "cli", pkg, "cli.go")
	title := fmt.Sprintf("%s HTTP client CLI support package", svr.Name)
	specs := []*codegen.ImportSpec{}
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "cli", specs),
		{
			Name:   "parse-endpoint",
			Source: parseEndpointT,
		},
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
		{Name: "yaml", Path: "github.com/goccy/go-yaml"},
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

func buildPayloadBuildFunction(svc *http.ServiceData, e *http.EndpointData) *BuildFunctionData {
	var (
		buildFunction *BuildFunctionData
	)

	if e.Payload != nil {
		if e.Payload.Request.PayloadInit != nil {
			args := e.Payload.Request.PayloadInit.ClientArgs
			args = append(args, e.Payload.Request.PayloadInit.CLIArgs...)
			buildFunction = makePayloadBuildFunction(e, args, e.Payload.Request.PayloadType)
		}
	}

	return buildFunction
}

// Only some argument types need conversion. Primitive types can be passed
// directly using the data type of their corresponding payload field. Arrays
// of primitive types can also be passed directly. Everything else needs to
// be passed as a YAML or JSON formatted string value
func argTypeNeedsConversion(arg *http.InitArgData) bool {
	if expr.IsPrimitive(arg.Type) {
		return false
	}

	if at := expr.AsArray(arg.Type); at != nil && expr.IsPrimitive(at.ElemType.Type) {
		return false
	}

	if at := expr.AsMap(arg.Type); at != nil && expr.IsPrimitive(at.ElemType.Type) {
		return false
	}

	return true
}

func payloadParamType(arg *http.InitArgData) string {
	if !argTypeNeedsConversion(arg) {
		return arg.TypeName
	}

	return codegen.GoNativeTypeName(expr.String)
}

func makePayloadBuildFunction(e *http.EndpointData, args []*http.InitArgData, payload expr.DataType) *BuildFunctionData {
	var (
		fdata     []*FieldData
		params    = make([]*ParamData, len(args))
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

		fn := goifyTerms(e.ServiceName, e.Method.Name, arg.VarName)
		params[i] = &ParamData{
			Name:     fn,
			TypeName: payloadParamType(arg),
		}

		if arg.FieldName == "" && arg.VarName != "body" {
			continue
		}
		code, chek := FieldLoadCode(arg, fn, payload)
		check = check || chek
		tn := arg.TypeRef
		if !expr.IsPrimitive(arg.Type) {
			// We need to declare the variable without
			// a pointer to be able to unmarshal the YAML
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

	return &BuildFunctionData{
		Name:        "Build" + e.Method.VarName + "Payload",
		Params:      params,
		ServiceName: e.ServiceName,
		MethodName:  e.Method.Name,
		ResultType:  e.Payload.Ref,
		Fields:      fdata,
		PayloadInit: &pInit,
		CheckErr:    check,
	}
}

const parseEndpointT = `// This code is no longer generated. It's superseded by FuseML`
