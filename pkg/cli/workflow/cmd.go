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

	cmd.AddCommand(newSubCmdList(c))
	cmd.AddCommand(newSubCmdCreate(c))
	cmd.AddCommand(newSubCmdGet(c))
	cmd.AddCommand(newSubCmdAssign(c))

	return cmd
}
