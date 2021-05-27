package runnable

import (
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

// NewCmdRunnable creates and returns the cobra command that acts as a root for all other runnable CLI sub-commands
func NewCmdRunnable(c *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runnable",
		Short: "runnable management",
		Long:  `Perform operations on runnables`,
	}

	cmd.AddCommand(NewSubCmdRunnableRegister(c))
	cmd.AddCommand(NewSubCmdRunnableGet(c))
	cmd.AddCommand(NewSubCmdRunnableList(c))

	return cmd
}
