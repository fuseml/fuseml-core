package extension

import (
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/tektoncd/cli/pkg/formatted"

	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

const extensionGetTemplate = `{{decorate "bold" "ID"}}:	{{ .ID }}
{{decorate "bold" "Registered"}}:	{{ .Status.Registered }}
{{- if ne (deref .Description) "" }}
{{decorate "bold" "Description"}}:	{{ deref .Description }}
{{- end }}
{{- if ne (deref .Product) "" }}
{{decorate "bold" "Product"}}:	{{ deref .Product }}
{{- end }}
{{- if ne (deref .Version) "" }}
{{decorate "bold" "Version"}}:	{{ deref .Version }}
{{- end }}
{{- if ne (deref .Zone) "" }}
{{decorate "bold" "Zone"}}:	{{ deref .Zone }}
{{- end }}
{{- $l := len .Configuration }}{{ if ne $l 0 }}
{{decorate "bold" "Configuration"}}:
{{- range $k, $v := .Configuration }}
 {{decorate "bullet" $k }} = {{ $v }}
{{- end }}
{{- end }}
{{- $l := len .Services }}{{ if ne $l 0 }}
{{decorate "pipelineruns" ""}}{{decorate "underline bold" "Services\n"}}
{{- range $s := .Services }}
 {{decorate "bullet" ""}}{{decorate "bold" "ID"}}:  {{ deref $s.ID }}
 {{- if ne (deref $s.Description) "" }}
   {{decorate "bold" "Description"}}:	{{ deref $s.Description }}
 {{- end }}
 {{- if ne (deref $s.Resource) "" }}
   {{decorate "bold" "Resource"}}:	{{ deref $s.Resource }}
 {{- end }}
 {{- if ne (deref $s.Category) "" }}
   {{decorate "bold" "Category"}}:	{{ deref $s.Category }}
 {{- end }}
   {{decorate "bold" "Authentication required"}}:	{{ $s.AuthRequired }}
 
   {{- $l := len $s.Endpoints }}{{ if ne $l 0 }}
   {{decorate "pipelineruns" ""}}{{decorate "underline bold" "Endpoints\n"}}
   {{- range $e := $s.Endpoints }}
    {{decorate "bullet" ""}}{{decorate "bold" "URL"}}:  {{ deref $e.URL }}
    {{- if ne (deref $e.Type) "" }}
      {{decorate "bold" "Type"}}:	{{ deref $e.Type }}
    {{- end }}
	{{- $l := len $e.Configuration }}{{ if ne $l 0 }}
      {{decorate "bold" "Configuration"}}:
	  {{- range $k, $v := $e.Configuration }}
       {{decorate "bullet" $k }} = {{ $v }}
	  {{- end }}
	{{- end }}
   {{- end }}
 {{- end }}
 
   {{- $l := len $s.Credentials }}{{ if ne $l 0 }}
   {{decorate "pipelineruns" ""}}{{decorate "underline bold" "Credentials\n"}}
   {{- range $c := $s.Credentials }}
    {{decorate "bullet" ""}}{{decorate "bold" "ID"}}:  {{ deref $c.ID }}
    {{- if ne (deref $c.Scope) "" }}
      {{decorate "bold" "Scope"}}:	{{ deref $c.Scope }}
    {{- end }}
    {{- $l := len $c.Projects }}{{ if ne $l 0 }}
      {{decorate "bold" "Projects"}}:	{{ join $c.Projects ", " }}
    {{- end }}
    {{- $l := len $c.Users }}{{ if ne $l 0 }}
      {{decorate "bold" "Users"}}:	{{ join $c.Users ", " }}
    {{- end }}
    {{- $l := len $c.Configuration }}{{ if ne $l 0 }}
      {{decorate "bold" "Configuration"}}:
       {{- range $k, $v := $c.Configuration }}
      {{decorate "bullet" $k }} = {{ $v }}
      {{- end }}
    {{- end }}
  {{- end }}
{{- end }}

{{- end }}
{{- end }}
`

type extensionGetOptions struct {
	client.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
}

func newExtensionGetOptions(o *common.GlobalOptions) *extensionGetOptions {
	res := &extensionGetOptions{global: o}
	res.format = common.NewSingleValueFormattingOptions()
	return res
}

func newSubCmdExtensionGet(gOpt *common.GlobalOptions) *cobra.Command {
	o := newExtensionGetOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `get {EXTENSION_ID}`,
		Short: "Get a extension",
		Long:  `Show detailed information about an extension`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run(cmd.Flags().Arg(0)))
		},
		Args: cobra.ExactArgs(1),
	}
	o.format.AddSingleValueFormattingFlags(cmd, common.FormatText)
	return cmd
}

func (o *extensionGetOptions) validate() error {
	return nil
}

func (o *extensionGetOptions) run(extensionID string) error {
	ext, err := o.ExtensionClient.GetExtension(extensionID)
	if err != nil {
		return err
	}

	if o.format.Format == common.FormatText {
		funcMap := template.FuncMap{
			"decorate":    formatted.DecorateAttr,
			"formatParam": formatted.Param,
			"colorStatus": formatted.ColorStatus,
			"join":        strings.Join,
			"deref": func(s *string) string {
				if s != nil {
					return *s
				}
				return ""
			},
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 5, 3, ' ', tabwriter.TabIndent)
		t := template.Must(template.New("Describe Extension").Funcs(funcMap).Parse(extensionGetTemplate))
		err = t.Execute(w, ext)
		if err != nil {
			return err
		}

		w.Flush()
	} else {
		o.format.FormatValue(os.Stdout, ext)
	}

	return nil
}
