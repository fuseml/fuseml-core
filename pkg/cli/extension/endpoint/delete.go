package endpoint

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type endpointDeleteOptions struct {
	client.Clients
	global *common.GlobalOptions
}

func newEndpointDeleteOptions(o *common.GlobalOptions) (res *endpointDeleteOptions) {
	return &endpointDeleteOptions{global: o}
}

func newSubCmdEndpointDelete(gOpt *common.GlobalOptions) *cobra.Command {
	o := newEndpointDeleteOptions(gOpt)
	cmd := &cobra.Command{
		Use:   "delete {EXTENSION_ID} {SERVICE_ID} {ENDPOINT_URL}",
		Short: "Deletes a endpoint from an extension service",
		Long:  `Delete a endpoint from an extension service registered with the FuseML extension registry.`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run(cmd.Flags().Arg(0), cmd.Flags().Arg(1), cmd.Flags().Arg(2)))
		},
		Args: cobra.ExactArgs(3),
	}

	return cmd
}

func (o *endpointDeleteOptions) validate() error {
	return nil
}

func (o *endpointDeleteOptions) run(extensionID, serviceID, URL string) error {
	err := o.ExtensionClient.DeleteEndpoint(extensionID, serviceID, URL)
	if err != nil {
		return err
	}

	fmt.Printf("Endpoint %q successfully deleted from service %q and extension %q\n", URL, serviceID, extensionID)

	return nil
}
