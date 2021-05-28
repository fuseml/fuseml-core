package workflow

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	workflowc "github.com/fuseml/fuseml-core/gen/http/workflow/client"
	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type createOptions struct {
	common.Clients
	global   *common.GlobalOptions
	workflow string
}

func newCreateOptions(o *common.GlobalOptions) *createOptions {
	return &createOptions{global: o}
}

func newSubCmdCreate(gOpt *common.GlobalOptions) *cobra.Command {
	o := newCreateOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `create WORKFLOW_FILE`,
		Short: "Creates a workflow",
		Long:  `Creates a workflow from a file`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt))
			common.CheckErr(common.LoadFileIntoVar(cmd.Flags().Arg(0), &o.workflow))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(1),
	}

	return cmd
}

func (o *createOptions) validate() error {
	// TODO: schema validation for the workflow
	return nil
}

func (o *createOptions) run() error {
	request, err := workflowc.BuildRegisterPayload(o.workflow)
	if err != nil {
		return err
	}

	response, err := o.WorkflowClient.Register()(context.Background(), request)
	if err != nil {
		return err
	}

	workflow := response.(*workflow.Workflow)

	fmt.Printf("Workflow %q successfully created\n", workflow.Name)

	return nil
}
