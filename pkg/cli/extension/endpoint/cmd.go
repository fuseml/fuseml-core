package endpoint

import (
	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

// NewSubCmdExtensionEndpoint creates and returns the cobra command that acts as a root for all other extension endpoint CLI sub-commands
func NewSubCmdExtensionEndpoint(c *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "endpoint",
		Short: "Extension endpoint management",
		Long:  `Perform operations on extension endpoints`,
	}

	cmd.AddCommand(newSubCmdEndpointAdd(c))
	cmd.AddCommand(newSubCmdEndpointUpdate(c))
	cmd.AddCommand(newSubCmdEndpointDelete(c))
	cmd.AddCommand(newSubCmdEndpointList(c))

	return cmd
}
