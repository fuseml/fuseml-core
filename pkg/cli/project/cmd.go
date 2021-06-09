package project

import (
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

// NewCmdProject creates and returns the cobra command that acts as a root for all other project CLI sub-commands
func NewCmdProject(c *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Project management",
		Long:  `Perform operations on projects`,
	}

	cmd.AddCommand(NewSubCmdProjectDelete(c))
	cmd.AddCommand(NewSubCmdProjectGet(c))
	cmd.AddCommand(NewSubCmdProjectList(c))
	cmd.AddCommand(NewSubCmdProjectSet(c))

	return cmd
}
