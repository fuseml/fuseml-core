package credentials

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type credentialsDeleteOptions struct {
	client.Clients
	global *common.GlobalOptions
}

func newCredentialsDeleteOptions(o *common.GlobalOptions) (res *credentialsDeleteOptions) {
	return &credentialsDeleteOptions{global: o}
}

func newSubCmdCredentialsDelete(gOpt *common.GlobalOptions) *cobra.Command {
	o := newCredentialsDeleteOptions(gOpt)
	cmd := &cobra.Command{
		Use:   "delete {EXTENSION_ID} {SERVICE_ID} {CREDENTIALS_ID}",
		Short: "Deletes a set of credentials from an extension service",
		Long:  `Delete a set of  credentials from an extension service registered with the FuseML extension registry.`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run(cmd.Flags().Arg(0), cmd.Flags().Arg(1), cmd.Flags().Arg(2)))
		},
		Args: cobra.ExactArgs(3),
	}

	return cmd
}

func (o *credentialsDeleteOptions) validate() error {
	return nil
}

func (o *credentialsDeleteOptions) run(extensionID, serviceID, credentialsID string) error {
	err := o.ExtensionClient.DeleteCredentials(extensionID, serviceID, credentialsID)
	if err != nil {
		return err
	}

	fmt.Printf("Credentials %q successfully deleted from service %q and extension %q\n", credentialsID, serviceID, extensionID)

	return nil
}
