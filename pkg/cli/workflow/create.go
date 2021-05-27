package workflow

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	workflowc "github.com/fuseml/fuseml-core/gen/http/workflow/client"
	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

// CreateOptions holds the options for 'workflow create' sub command
type CreateOptions struct {
	common.Clients
	global   *common.GlobalOptions
	Workflow string
}

// NewCreateOptions initializes a CreateOptions struct
func NewCreateOptions(o *common.GlobalOptions) *CreateOptions {
	return &CreateOptions{global: o}
}

// NewSubCmdCreate creates and returns the cobra command for the `workflow create` CLI command
func NewSubCmdCreate(gOpt *common.GlobalOptions) *cobra.Command {

	o := NewCreateOptions(gOpt)

	cmd := &cobra.Command{
		Use:   `create WORKFLOW_FILE`,
		Short: "Creates a workflow",
		Long:  `Creates a workflow from a file`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt))
			common.CheckErr(common.LoadFileIntoVar(cmd.Flags().Arg(0), &o.Workflow))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(1),
	}

	return cmd
}

func (o *CreateOptions) validate() error {
	// TODO: schema validation for the workflow
	return nil
}

func (o *CreateOptions) run() error {
	request, err := workflowc.BuildRegisterPayload(o.Workflow)
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
