package workflow

import (
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/tektoncd/cli/pkg/formatted"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/fuseml/fuseml-core/pkg/util"
)

const getTemplate = `{{decorate "bold" "Name"}}:	{{ .Workflow.Name }}
{{decorate "bold" "Created"}}:	{{ .Workflow.Created }}
{{- if ne (deref .Workflow.Description) "" }}
{{decorate "bold" "Description"}}:	{{ deref .Workflow.Description }}
{{- end }}

{{decorate "params" ""}}{{decorate "underline bold" "Inputs\n"}}
{{- $l := len .Workflow.Inputs }}{{ if eq $l 0 }}
 No inputs
{{- else }}
 NAME	TYPE	DESCRIPTION	DEFAULT
{{- range $input := .Workflow.Inputs }}
{{- if not $input.Default }}
 {{decorate "bullet" $input.Name }}	{{ $input.Type }}	{{ formatDesc $input.Description }}	{{ "---" }}
{{- else }}
 {{decorate "bullet" $input.Name }}	{{ $input.Type }}	{{ formatDesc $input.Description }}	{{ $input.Default }}
{{- end }}
{{- end }}
{{- end }}

{{decorate "results" ""}}{{decorate "underline bold" "Outputs\n"}}
{{- if eq (len .Workflow.Outputs) 0 }}
 No outputs
{{- else }}
 NAME	TYPE	DESCRIPTION
{{- range $output := .Workflow.Outputs }}
 {{ decorate "bullet" $output.Name }}	{{ $output.Type }}	{{ formatDesc $output.Description }}
{{- end }}
{{- end }}

{{decorate "steps" ""}}{{decorate "underline bold" "Steps\n"}}
{{- $tl := len .Workflow.Steps }}{{ if eq $tl 0 }}
 No steps
{{- else }}
 NAME	IMAGE
{{- range $s := .Workflow.Steps }}
 {{decorate "bullet" $s.Name }}	{{ $s.Image }}
{{- end }}
{{- end }}

{{decorate "pipelineruns" ""}}{{decorate "underline bold" "Workflow Runs\n"}}
{{- $rl := len .WorkflowRuns }}{{ if eq $rl 0 }}
 No workflow runs
{{- else }}
 NAME	STARTED	DURATION	STATUS
{{- range $wr := .WorkflowRuns }}
 {{decorate "bullet" $wr.Name }}	{{ formatAge $wr.StartTime }}	{{ formatDuration $wr.StartTime $wr.CompletionTime }}	{{ colorStatus $wr.Status }}
{{- end }}
{{- end }}
`

type getOptions struct {
	client.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
	name   string
}

func newGetOptions(o *common.GlobalOptions) *getOptions {
	res := &getOptions{global: o}
	res.format = common.NewSingleValueFormattingOptions()
	return res
}

func newSubCmdGet(gOpt *common.GlobalOptions) *cobra.Command {
	o := newGetOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `get {-n|--name NAME}`,
		Short: "Get a workflow",
		Long:  `Show detailed information from a workflow`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "workflow name")
	o.format.AddSingleValueFormattingFlags(cmd, common.FormatText)
	cmd.MarkFlagRequired("name")
	return cmd
}

func (o *getOptions) validate() error {
	return nil
}

func (o *getOptions) run() error {
	wf, err := o.WorkflowClient.Get(o.name)
	if err != nil {
		return err
	}

	if o.format.Format == common.FormatText {
		wfRuns, err := o.WorkflowClient.ListRuns(o.name, "", "", "")
		if err != nil {
			return err
		}

		var data = struct {
			Workflow     *workflow.Workflow
			WorkflowRuns []*workflow.WorkflowRun
		}{
			Workflow:     wf,
			WorkflowRuns: wfRuns,
		}

		funcMap := template.FuncMap{
			"decorate":       formatted.DecorateAttr,
			"formatDesc":     formatDesc,
			"formatParam":    formatted.Param,
			"formatAge":      formatAge,
			"formatDuration": formatDuration,
			"colorStatus":    formatted.ColorStatus,
			"join":           strings.Join,
			"deref":          util.DerefString,
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 5, 3, ' ', tabwriter.TabIndent)
		t := template.Must(template.New("Describe Pipeline").Funcs(funcMap).Parse(getTemplate))
		err = t.Execute(w, data)
		if err != nil {
			return err
		}

		w.Flush()
	} else {
		o.format.FormatValue(os.Stdout, wf)
	}

	return nil
}
