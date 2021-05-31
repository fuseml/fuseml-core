package workflow

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	workflowc "github.com/fuseml/fuseml-core/gen/http/workflow/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type assignOptions struct {
	common.Clients
	global         *common.GlobalOptions
	name           string
	codesetName    string
	codesetProject string
}

func newAssignOptions(o *common.GlobalOptions) *assignOptions {
	return &assignOptions{global: o}
}

func newSubCmdAssign(c *common.GlobalOptions) *cobra.Command {
	o := newAssignOptions(c)
	cmd := &cobra.Command{
		Use:   "assign {-n|--name NAME} {-c|--codeset-name CODESET_NAME} {-p|--codeset-project CODESET_PROJECT}",
		Short: "Assigns a workflow to a codeset",
		Long: `Assigning a workflow to a codeset makes any change pushed to the codeset trigger the workflow(s) assigned to it.
Upon successfully assignment a workflow run is created using the workflow's default inputs and the assigned codeset.`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(c))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "name of the workflow to be assigned")
	cmd.Flags().StringVarP(&o.codesetName, "codeset-name", "c", "", "name of the codeset to assign the workflow to")
	cmd.Flags().StringVarP(&o.codesetProject, "codeset-project", "p", "", "name of the project to which the codeset belongs")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("codeset-name")
	cmd.MarkFlagRequired("codeset-project")

	return cmd
}

func (o *assignOptions) validate() error {
	return nil
}

func (o *assignOptions) run() error {
	request, err := workflowc.BuildAssignPayload(o.name, o.codesetName, o.codesetProject)
	if err != nil {
		return err
	}

	_, err = o.WorkflowClient.Assign()(context.Background(), request)
	if err != nil {
		return err
	}

	fmt.Printf("Workflow %q assigned to codeset \"%s/%s\"\n", o.name, o.codesetProject, o.codesetName)

	return nil
}