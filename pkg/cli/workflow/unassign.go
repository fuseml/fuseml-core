package workflow

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type unassignOptions struct {
	client.Clients
	global         *common.GlobalOptions
	name           string
	codesetName    string
	codesetProject string
}

func newUnassignOptions(o *common.GlobalOptions) *unassignOptions {
	return &unassignOptions{global: o}
}

func newSubCmdUnassign(gOpt *common.GlobalOptions) *cobra.Command {
	o := newUnassignOptions(gOpt)
	cmd := &cobra.Command{
		Use:   "unassign {-n|--name NAME} {-p|--codeset-project CODESET_PROJECT}  {-c|--codeset-name CODESET_NAME}",
		Short: "Unassign a workflow from a codeset",
		Long:  `Removes the assignment between a workflow and a codeset.`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "workflow name")
	cmd.Flags().StringVarP(&o.codesetProject, "codeset-project", "p", "", "name of the project to which the codeset belongs")
	cmd.Flags().StringVarP(&o.codesetName, "codeset-name", "c", "", "name of the codeset assigned to the workflow")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("codeset-project")
	cmd.MarkFlagRequired("codeset-name")

	return cmd
}

func (o *unassignOptions) validate() error {
	return nil
}

func (o *unassignOptions) run() error {
	err := o.WorkflowClient.Unassign(o.name, o.codesetProject, o.codesetName)
	if err != nil {
		return err
	}

	fmt.Printf("Workflow %q unassigned from codeset \"%s/%s\"\n", o.name, o.codesetProject, o.codesetName)

	return nil
}
