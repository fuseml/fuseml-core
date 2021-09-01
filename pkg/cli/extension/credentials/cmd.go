package credentials

import (
	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

// NewSubCmdExtensionCredentials creates and returns the cobra command that acts as a root for all other extension credentials CLI sub-commands
func NewSubCmdExtensionCredentials(c *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "credentials",
		Short: "Extension credentials management",
		Long:  `Perform operations on extension credentials`,
	}

	cmd.AddCommand(newSubCmdCredentialsAdd(c))
	cmd.AddCommand(newSubCmdCredentialsUpdate(c))
	cmd.AddCommand(newSubCmdCredentialsDelete(c))
	cmd.AddCommand(newSubCmdCredentialsList(c))

	return cmd
}
