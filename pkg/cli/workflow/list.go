package workflow

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type listOptions struct {
	client.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
	name   string
}

func formatInputs(object interface{}, column string, field interface{}) (formated string) {
	if workflow, ok := object.(*workflow.Workflow); ok {
		for i, input := range workflow.Inputs {
			formated += fmt.Sprintf("- name: %s\n  type: %s", input.Name, *input.Type)
			if input.Default != nil {
				formated += fmt.Sprintf("\n  default: %s", *input.Default)
			}
			if i != len(workflow.Inputs)-1 {
				formated += "\n"
			}
		}
	}
	return
}

func formatOutputs(object interface{}, column string, field interface{}) (formated string) {
	if workflow, ok := object.(*workflow.Workflow); ok {
		for i, input := range workflow.Outputs {
			formated += fmt.Sprintf("- name: %s\n  type: %s", input.Name, *input.Type)
			if i != len(workflow.Inputs)-1 {
				formated += "\n"
			}
		}
	}
	return
}

func newListOptions(o *common.GlobalOptions) (res *listOptions) {
	res = &listOptions{global: o}
	res.format = common.NewFormattingOptions(
		[]string{"Name", "Description", "Inputs", "Outputs"},
		[]table.SortBy{{Name: "Name", Mode: table.Asc}},
		common.OutputFormatters{"Inputs": formatInputs, "Outputs": formatOutputs},
	)

	return
}

func newSubCmdList(gOpt *common.GlobalOptions) *cobra.Command {
	o := newListOptions(gOpt)
	cmd := &cobra.Command{
		Use:   "list [-n|--name NAME]",
		Short: "Lists one or more workflows",
		Long:  `Prints a table of the most important information about workflows. You can filter the list by the workflow name.`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "filter workflows by name")
	o.format.AddMultiValueFormattingFlags(cmd)

	return cmd
}

func (o *listOptions) validate() error {
	return nil
}

func (o *listOptions) run() error {
	wfs, err := o.WorkflowClient.List(o.name)
	if err != nil {
		return err
	}

	o.format.FormatValue(os.Stdout, wfs)

	return nil
}
