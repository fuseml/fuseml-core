package extension

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type extensionDeleteOptions struct {
	client.Clients
	global *common.GlobalOptions
}

func newExtensionDeleteOptions(o *common.GlobalOptions) (res *extensionDeleteOptions) {
	return &extensionDeleteOptions{global: o}
}

func newSubCmdExtensionDelete(gOpt *common.GlobalOptions) *cobra.Command {
	o := newExtensionDeleteOptions(gOpt)
	cmd := &cobra.Command{
		Use:   "delete {EXTENSION_ID}",
		Short: "Deletes an extension",
		Long:  `Delete an extension from the FuseML extension registry, along with all services, endpoints and credentials.`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run(cmd.Flags().Arg(0)))
		},
		Args: cobra.ExactArgs(1),
	}

	return cmd
}

func (o *extensionDeleteOptions) validate() error {
	return nil
}

func (o *extensionDeleteOptions) run(extensionID string) error {
	err := o.ExtensionClient.DeleteExtension(extensionID)
	if err != nil {
		return err
	}

	fmt.Printf("Extension %s successfully deleted\n", extensionID)

	return nil
}
