package extension

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/fuseml/fuseml-core/gen/extension"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type extensionAddOptions struct {
	client.Clients
	global      *common.GlobalOptions
	extensionID string
	description string
	product     string
	version     string
	zone        string
	config      common.KeyValueArgs
}

func newExtensionAddOptions(o *common.GlobalOptions) *extensionAddOptions {
	return &extensionAddOptions{global: o}
}

func newSubCmdExtensionAdd(gOpt *common.GlobalOptions) *cobra.Command {
	o := newExtensionAddOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `add [--id EXTENSION_ID] [--desc DESCRIPTION] [-r|--resource SERVICE_RESOURCE] [-c|--category SERVICE_CATEGORY] [--auth-required={true|false}] [--configuration KEY:VALUE]...`,
		Short: "Add a new extension to an existing FuseML extension",
		Long:  `Add a extension to a FuseML extension already registered with the extension registry`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate(cmd.Flags()))
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}
	cmd.Flags().StringVar(&o.extensionID, "id", "", "extension ID")
	cmd.Flags().StringVar(&o.description, "desc", "", "extension description")
	cmd.Flags().StringVarP(&o.product, "product", "p", "",
		`universal product identifier. Product values can be used to identify
installations of the same product registered with the same or different FuseML servers`)
	cmd.Flags().StringVar(&o.version, "version", "", "extension version")
	cmd.Flags().StringVarP(&o.zone, "zone", "z", "", "zone where the extension is installed")
	cmd.Flags().StringSliceVarP(&o.config.Packed, "configuration", "c", []string{}, "extension configuration data. One or more may be supplied")

	return cmd
}

func (o *extensionAddOptions) validate(flags *pflag.FlagSet) error {
	return nil
}

func (o *extensionAddOptions) run() error {
	extension := extension.Extension{
		ID:            &o.extensionID,
		Description:   &o.description,
		Product:       &o.product,
		Version:       &o.version,
		Zone:          &o.zone,
		Configuration: o.config.Unpacked,
	}
	ext, err := o.ExtensionClient.RegisterExtension(&extension)
	if err != nil {
		return err
	}

	fmt.Printf("Extension %q successfully added to registry\n", *ext.ID)

	return nil
}
