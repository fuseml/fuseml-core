package extension

import (
	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/fuseml/fuseml-core/pkg/cli/extension/credentials"
	"github.com/fuseml/fuseml-core/pkg/cli/extension/endpoint"
	"github.com/fuseml/fuseml-core/pkg/cli/extension/service"
)

// NewCmdExtension creates and returns the cobra command that acts as a root for all other extension CLI sub-commands
func NewCmdExtension(c *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extension",
		Short: "Extension management",
		Long:  `Perform operations on extensions`,
	}

	cmd.AddCommand(service.NewSubCmdExtensionService(c))
	cmd.AddCommand(endpoint.NewSubCmdExtensionEndpoint(c))
	cmd.AddCommand(credentials.NewSubCmdExtensionCredentials(c))
	cmd.AddCommand(newSubCmdExtensionRegister(c))
	cmd.AddCommand(newSubCmdExtensionGet(c))
	cmd.AddCommand(newSubCmdExtensionList(c))
	cmd.AddCommand(newSubCmdExtensionDelete(c))
	cmd.AddCommand(newSubCmdExtensionAdd(c))
	cmd.AddCommand(newSubCmdExtensionUpdate(c))

	return cmd
}
