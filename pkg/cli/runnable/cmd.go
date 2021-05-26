package runnable

import (
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

func NewCmdRunnable(c *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runnable",
		Short: "runnable management",
		Long:  `Perform operations on runnables`,
	}

	cmd.AddCommand(NewSubCmdRunnableList(c))
	cmd.AddCommand(NewSubCmdRunnableRegister(c))

	return cmd
}
