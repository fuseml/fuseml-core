package endpoint

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/fuseml/fuseml-core/gen/extension"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type endpointUpdateOptions struct {
	client.Clients
	global       *common.GlobalOptions
	endpointType string
	config       common.KeyValueArgs
}

func newEndpointUpdateOptions(o *common.GlobalOptions) *endpointUpdateOptions {
	return &endpointUpdateOptions{global: o}
}

func newSubCmdEndpointUpdate(gOpt *common.GlobalOptions) *cobra.Command {
	o := newEndpointUpdateOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `update [--type {internal|external}] [-c|--configuration KEY:VALUE]... {EXTENSION_ID} {SERVICE_ID} {ENDPOINT_URL}`,
		Short: "Update the attributes of an existing FuseML extension endpoint",
		Long:  `Update the attributes of a FuseML extension endpoint already registered with the extension registry`,
		Run: func(cmd *cobra.Command, args []string) {
			o.config.Unpack()
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run(cmd.Flags().Arg(0), cmd.Flags().Arg(1), cmd.Flags().Arg(2), cmd.Flags()))
		},
		Args: cobra.ExactArgs(3),
	}
	cmd.Flags().StringVar(&o.endpointType, "type", "external", "endpoint type (internal/external). Internal endpoints cannot be accessed from outside the zone")
	cmd.Flags().StringSliceVarP(&o.config.Packed, "configuration", "c", []string{}, "endpoint configuration data. One or more may be supplied")

	return cmd
}

func (o *endpointUpdateOptions) validate() error {
	return common.ValidateEnumArgument("endpoint type", o.endpointType, []string{"internal", "external"})
}

func (o *endpointUpdateOptions) run(extensionID, serviceID, URL string, flags *pflag.FlagSet) error {
	endpoint := extension.ExtensionEndpoint{
		URL:         &URL,
		ExtensionID: &extensionID,
		ServiceID:   &serviceID,
	}
	if flags.Changed("type") {
		endpoint.Type = &o.endpointType
	}
	if flags.Changed("configuration") {
		endpoint.Configuration = o.config.Unpacked
	}
	_, err := o.ExtensionClient.UpdateEndpoint(&endpoint)
	if err != nil {
		return err
	}

	fmt.Printf("Endpoint %q from service %s and extension %q successfully updated\n", URL, serviceID, extensionID)

	return nil
}
