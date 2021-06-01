package workflow

import (
	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/pkg/cli/common"
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
	cmd.AddCommand(newSubCmdListAssignments(c))
	cmd.AddCommand(newSubCmdListRuns(c))
	cmd.AddCommand(newSubCmdUnassign(c))
	cmd.AddCommand(newSubCmdDelete(c))

	return cmd
}
