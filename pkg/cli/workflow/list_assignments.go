package workflow

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type listAssignmentsOptions struct {
	client.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
	name   string
}

func formatAssignmentStatus(object interface{}, column string, field interface{}) string {
	if wa, ok := object.(*workflow.WorkflowAssignment); ok {
		if wa.Status.Available {
			return color.New(color.FgHiGreen).Sprint("UP")
		}
		return color.New(color.FgHiRed).Sprint("DOWN")
	}
	return "N/A"
}

func formatAssignedCodesets(object interface{}, column string, field interface{}) (formated string) {
	if wa, ok := object.(*workflow.WorkflowAssignment); ok {
		for i, c := range wa.Codesets {
			formated += fmt.Sprintf("- name: %s\n  project: %s", c.Name, c.Project)
			if i != len(wa.Codesets)-1 {
				formated += "\n"
			}
		}
	}
	return
}

func newListAssignmentsOptions(o *common.GlobalOptions) (res *listAssignmentsOptions) {
	res = &listAssignmentsOptions{global: o}
	res.format = common.NewFormattingOptions(
		[]string{"Workflow", "Codesets", "Status"},
		[]table.SortBy{{Name: "Workflow", Mode: table.Asc}},
		common.OutputFormatters{"Status": formatAssignmentStatus, "Codesets": formatAssignedCodesets},
	)

	return
}

func newSubCmdListAssignments(gOpt *common.GlobalOptions) *cobra.Command {
	o := newListAssignmentsOptions(gOpt)
	cmd := &cobra.Command{
		Use:   "list-assignments [-n|--name NAME]",
		Short: "Lists one or more workflow assignments",
		Long:  `Prints a table of the most important information about workflow assignments. You can filter the list by the workflow name.`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "filter workflow assignments by the workflow name")
	o.format.AddMultiValueFormattingFlags(cmd)

	return cmd
}

func (o *listAssignmentsOptions) validate() error {
	return nil
}

func (o *listAssignmentsOptions) run() error {
	al, err := o.WorkflowClient.ListAssignments(o.name)
	if err != nil {
		return err
	}

	o.format.FormatValue(os.Stdout, al)

	return nil
}
