package endpoint

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/gen/extension"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type endpointAddOptions struct {
	client.Clients
	global       *common.GlobalOptions
	endpointType string
	config       common.KeyValueArgs
}

func newEndpointAddOptions(o *common.GlobalOptions) *endpointAddOptions {
	return &endpointAddOptions{global: o}
}

func newSubCmdEndpointAdd(gOpt *common.GlobalOptions) *cobra.Command {
	o := newEndpointAddOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `add [--type {internal|external}] [-c|--configuration KEY:VALUE]... {EXTENSION_ID} {SERVICE_ID} {ENDPOINT_URL}`,
		Short: "Add a new endpoint to an existing FuseML extension service",
		Long:  `Add an endpoint to a FuseML extension service already registered with the extension registry`,
		Run: func(cmd *cobra.Command, args []string) {
			o.config.Unpack()
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run(cmd.Flags().Arg(0), cmd.Flags().Arg(1), cmd.Flags().Arg(2)))
		},
		Args: cobra.ExactArgs(3),
	}
	cmd.Flags().StringVar(&o.endpointType, "type", "external", "endpoint type (internal/external). Internal endpoints cannot be accessed from outside the zone")
	cmd.Flags().StringSliceVarP(&o.config.Packed, "configuration", "c", []string{}, "endpoint configuration data. One or more may be supplied")

	return cmd
}

func (o *endpointAddOptions) validate() error {
	return common.ValidateEnumArgument("endpoint type", o.endpointType, []string{"internal", "external"})
}

func (o *endpointAddOptions) run(extensionID, serviceID, URL string) error {
	endpoint := extension.ExtensionEndpoint{
		URL:           &URL,
		ExtensionID:   &extensionID,
		ServiceID:     &serviceID,
		Type:          &o.endpointType,
		Configuration: o.config.Unpacked,
	}
	_, err := o.ExtensionClient.AddEndpoint(&endpoint)
	if err != nil {
		return err
	}

	fmt.Printf("Endpoint %q successfully added to service %s from extension %q\n", URL, serviceID, extensionID)

	return nil
}
