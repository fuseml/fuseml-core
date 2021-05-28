package workflow

import (
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

// NewCmdWorkflow creates and returns the cobra command that acts as a root for all other workflow CLI sub-commands
func NewCmdWorkflow(c *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Workflow management",
		Long:  `Perform operations on workflows`,
	}

	cmd.AddCommand(NewSubCmdList(c))
	cmd.AddCommand(NewSubCmdCreate(c))
	cmd.AddCommand(NewSubCmdGet(c))
	cmd.AddCommand(newSubCmdAssign(c))

	return cmd
}
