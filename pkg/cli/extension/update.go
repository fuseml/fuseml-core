package extension

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/fuseml/fuseml-core/gen/extension"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type extensionUpdateOptions struct {
	client.Clients
	global      *common.GlobalOptions
	description string
	product     string
	version     string
	zone        string
	config      common.KeyValueArgs
}

func newExtensionUpdateOptions(o *common.GlobalOptions) *extensionUpdateOptions {
	return &extensionUpdateOptions{global: o}
}

func newSubCmdExtensionUpdate(gOpt *common.GlobalOptions) *cobra.Command {
	o := newExtensionUpdateOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `update [--desc DESCRIPTION] [-p|--product PRODUCT] [--version VERSION] [-z|--zone ZONE] [-c|--configuration KEY:VALUE]... {EXTENSION_ID}`,
		Short: "Update the attributes of an existing FuseML extension",
		Long:  `Update the attributes of a FuseML extension already registered with the extension registry`,
		Run: func(cmd *cobra.Command, args []string) {
			o.config.Unpack()
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run(cmd.Flags().Arg(0), cmd.Flags()))
		},
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().StringVar(&o.description, "desc", "", "extension description")
	cmd.Flags().StringVarP(&o.product, "product", "p", "",
		`universal product identifier. Product values can be used to identify
installations of the same product registered with the same or different FuseML servers`)
	cmd.Flags().StringVar(&o.version, "version", "", "extension version")
	cmd.Flags().StringVarP(&o.zone, "zone", "z", "", "zone where the extension is installed")
	cmd.Flags().StringSliceVarP(&o.config.Packed, "configuration", "c", []string{}, "extension configuration data. One or more may be supplied")

	return cmd
}

func (o *extensionUpdateOptions) validate() error {
	return nil
}

func (o *extensionUpdateOptions) run(extensionID string, flags *pflag.FlagSet) error {
	extension := extension.Extension{
		ID: &extensionID,
	}
	if flags.Changed("desc") {
		extension.Description = &o.description
	}
	if flags.Changed("product") {
		extension.Product = &o.product
	}
	if flags.Changed("version") {
		extension.Version = &o.version
	}
	if flags.Changed("zone") {
		extension.Zone = &o.zone
	}
	if flags.Changed("configuration") {
		extension.Configuration = o.config.Unpacked
	}
	_, err := o.ExtensionClient.UpdateExtension(&extension)
	if err != nil {
		return err
	}

	fmt.Printf("Extension %q successfully updated\n", extensionID)

	return nil
}
