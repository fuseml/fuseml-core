package workflow

import (
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jonboulle/clockwork"
	"github.com/spf13/cobra"
	"github.com/tektoncd/cli/pkg/formatted"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type listRunsOptions struct {
	client.Clients
	global         *common.GlobalOptions
	format         *common.FormattingOptions
	name           string
	codesetName    string
	codesetProject string
	status         string
}

func formatRunDuration(object interface{}, column string, field interface{}) string {
	if wr, ok := object.(*workflow.WorkflowRun); ok {
		return formatDuration(wr.StartTime, wr.CompletionTime)
	}
	return ""
}

func formatRunStatus(object interface{}, column string, field interface{}) string {
	if wr, ok := object.(*workflow.WorkflowRun); ok {
		return formatted.ColorStatus(wr.Status)
	}
	return ""
}

func formatRunWorkflowRef(object interface{}, column string, field interface{}) string {
	if wr, ok := object.(*workflow.WorkflowRun); ok {
		return wr.WorkflowRef
	}
	return ""
}

func formatRunStartTime(object interface{}, column string, field interface{}) string {
	if wr, ok := object.(*workflow.WorkflowRun); ok {
		startTime, _ := time.Parse(time.RFC3339, wr.StartTime)
		return formatted.Age(&v1.Time{Time: startTime}, clockwork.NewRealClock())
	}
	return "---"
}

func newListRunsOptions(o *common.GlobalOptions) (res *listRunsOptions) {
	res = &listRunsOptions{global: o}
	res.format = common.NewFormattingOptions(
		[]string{"Name", "Workflow", "Started", "Duration", "Status"},
		[]table.SortBy{},
		common.OutputFormatters{"Duration": formatRunDuration, "Status": formatRunStatus,
			"Workflow": formatRunWorkflowRef, "Started": formatRunStartTime},
	)

	return
}

func newSubCmdListRuns(gOpt *common.GlobalOptions) *cobra.Command {
	o := newListRunsOptions(gOpt)
	cmd := &cobra.Command{
		Use:   "list-runs [-n|--name NAME] [-p|--codeset-project CODESET_PROJECT] [-c|--codeset-name CODESET_NAME] [-s|--status STATUS]",
		Short: "Lists one or more workflow runs",
		Long:  `Prints a table of the most important information about workflow runs. You can filter the list by the workflow name, codeset name, codeset project or status.`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "filter workflow runs by the workflow name")
	cmd.Flags().StringVarP(&o.codesetProject, "codeset-project", "p", "", "filter workflow runs by the codeset project")
	cmd.Flags().StringVarP(&o.codesetName, "codeset-name", "c", "", "filter workflow runs by the codeset name")
	cmd.Flags().StringVarP(&o.status, "status", "s", "", "filter workflow runs by the workflow run status")
	o.format.AddMultiValueFormattingFlags(cmd)

	return cmd
}

func (o *listRunsOptions) validate() error {
	return nil
}

func (o *listRunsOptions) run() error {
	wfRuns, err := o.WorkflowClient.ListRuns(o.name, o.codesetProject, o.codesetName, o.status)
	if err != nil {
		return err
	}

	o.format.FormatValue(os.Stdout, wfRuns)

	return nil
}
