package service

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type serviceDeleteOptions struct {
	client.Clients
	global *common.GlobalOptions
}

func newServiceDeleteOptions(o *common.GlobalOptions) (res *serviceDeleteOptions) {
	return &serviceDeleteOptions{global: o}
}

func newSubCmdServiceDelete(gOpt *common.GlobalOptions) *cobra.Command {
	o := newServiceDeleteOptions(gOpt)
	cmd := &cobra.Command{
		Use:   "delete {EXTENSION_ID} {SERVICE_ID}",
		Short: "Deletes a service from an extension",
		Long:  `Delete a service from an extension registered with the FuseML extension registry.`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run(cmd.Flags().Arg(0), cmd.Flags().Arg(1)))
		},
		Args: cobra.ExactArgs(2),
	}

	return cmd
}

func (o *serviceDeleteOptions) validate() error {
	return nil
}

func (o *serviceDeleteOptions) run(extensionID, serviceID string) error {
	err := o.ExtensionClient.DeleteService(extensionID, serviceID)
	if err != nil {
		return err
	}

	fmt.Printf("Service %q successfully deleted from extension %q\n", serviceID, extensionID)

	return nil
}
