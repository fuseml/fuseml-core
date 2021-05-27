package workflow

import (
	"context"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	workflowc "github.com/fuseml/fuseml-core/gen/http/workflow/client"
	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

// ListOptions holds the options for 'workflow list' sub command
type ListOptions struct {
	common.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
	Name   string
}

func formatInputs(object interface{}, column string, field interface{}) (formated string) {
	if workflow, ok := object.(*workflow.Workflow); ok {
		for i, input := range workflow.Inputs {
			formated += fmt.Sprintf("- name: %s\n  type: %s", *input.Name, *input.Type)
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
			formated += fmt.Sprintf("- name: %s\n  type: %s", *input.Name, *input.Type)
			if i != len(workflow.Inputs)-1 {
				formated += "\n"
			}
		}
	}
	return
}

// NewListOptions creates a ListOptions struct
func NewListOptions(o *common.GlobalOptions) (res *ListOptions) {
	res = &ListOptions{global: o}
	res.format = common.NewFormattingOptions(
		[]string{"Name", "Description", "Inputs", "Outputs"},
		[]table.SortBy{{Name: "Name", Mode: table.Asc}},
		common.OutputFormatters{"Inputs": formatInputs, "Outputs": formatOutputs},
	)

	return
}

// NewSubCmdList creates and returns the cobra command for the `workflow list` CLI command
func NewSubCmdList(c *common.GlobalOptions) *cobra.Command {
	o := NewListOptions(c)
	cmd := &cobra.Command{
		Use:   "list [-n|--name NAME]",
		Short: "Display one or more workflows",
		Long:  `Prints a table of the most important information about workflows. You can filter the list by the workflow name.`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(c))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "filter workflows by name")
	o.format.AddMultiValueFormattingFlags(cmd)

	return cmd
}

func (o *ListOptions) validate() error {
	return nil
}

func (o *ListOptions) run() error {
	request, err := workflowc.BuildListPayload(o.Name)
	if err != nil {
		return err
	}

	response, err := o.WorkflowClient.List()(context.Background(), request)
	if err != nil {
		return err
	}

	o.format.FormatValue(os.Stdout, response)

	return nil
}
