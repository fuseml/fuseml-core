package codeset

import (
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

func NewCmdCodeset(c *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "codeset",
		Short: "codeset management",
		Long:  `Perform operations on codesets`,
		//Args:          cobra.ExactArgs(0),
		//RunE:          Install,
		//SilenceErrors: true,
		//SilenceUsage:  true,
	}

	cmd.AddCommand(NewSubCmdCodesetList(c))
	cmd.AddCommand(NewSubCmdCodesetRegister(c))

	return cmd
}
