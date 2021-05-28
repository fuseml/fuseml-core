package application

import (
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

// NewCmdApplication creates and returns the cobra command that acts as a root for all other application CLI sub-commands
func NewCmdApplication(c *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "application",
		Short: "application management",
		Long:  `Perform operations on applications`,
	}

	cmd.AddCommand(newSubCmdApplicationList(c))
	cmd.AddCommand(newSubCmdApplicationGet(c))
	cmd.AddCommand(newSubCmdApplicationDelete(c))

	return cmd
}
