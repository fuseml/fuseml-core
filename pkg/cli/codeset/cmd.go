package codeset

import (
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

// NewCmdCodeset creates and returns the cobra command that acts as a root for all other codeset CLI sub-commands
func NewCmdCodeset(c *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "codeset",
		Short: "codeset management",
		Long:  `Perform operations on codesets`,
	}

	cmd.AddCommand(NewSubCmdCodesetList(c))
	cmd.AddCommand(NewSubCmdCodesetRegister(c))

	return cmd
}
