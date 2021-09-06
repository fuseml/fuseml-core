package service

import (
	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

// NewSubCmdExtensionService creates and returns the cobra command that acts as a root for all other extension service CLI sub-commands
func NewSubCmdExtensionService(c *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Extension service management",
		Long:  `Perform operations on extension services`,
	}

	cmd.AddCommand(newSubCmdServiceAdd(c))
	cmd.AddCommand(newSubCmdServiceUpdate(c))
	cmd.AddCommand(newSubCmdServiceDelete(c))
	cmd.AddCommand(newSubCmdServiceList(c))

	return cmd
}
