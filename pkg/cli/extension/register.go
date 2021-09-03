package extension

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/fuseml/fuseml-core/gen/extension"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type extensionRegisterOptions struct {
	client.Clients
	global      *common.GlobalOptions
	extensionID string
	description string
	product     string
	version     string
	zone        string
	config      common.KeyValueArgs
	fromFile    string
}

func newExtensionRegisterOptions(o *common.GlobalOptions) *extensionRegisterOptions {
	return &extensionRegisterOptions{global: o}
}

func newSubCmdExtensionRegister(gOpt *common.GlobalOptions) *cobra.Command {
	o := newExtensionRegisterOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `register [-f|--file EXTENSION_FILE] [--id EXTENSION_ID] [--desc DESCRIPTION] [-p|--product PRODUCT] [--version VERSION] [-z|--zone ZONE] [-c|--configuration KEY:VALUE]...`,
		Short: "Registers a FuseML extension",
		Long: `Registers an external application as a FuseML extension

Use this command to register an extension from a YAML or JSON file. For example:

  fuseml extension register -f mlflow.yaml

Alternatively, use this command to create an empty extension from command line
arguments and subsequently add services, endpoints and credentials by running
'fuseml extension service add', 'fuseml extension endpoint add' and
'fuseml extension credentials add'. For example:

  fuseml extension register --id mlflow-devel --product mlflow --version 1.19.0 --zone local
  fuseml extension service add --id mlflow-store --resource s3 --auth-required mlflow-devel
  fuseml extension endpoint add --type internal -c MLFLOW_S3_ENDPOINT_URL:http://mlflow-minio:9000 mlflow-devel mlflow-store http://mlflow-minio:9000
  fuseml extension credentials add --scope global -c AWS_ACCESS_KEY_ID:v4Us74XUtkuEGd10yS05,AWS_SECRET_ACCESS_KEY:MJtLeytp72bpnq2XtSqpRTlB3MXTV8Am5ASjED4x mlflow-devel mlflow-store

The '-f|--file' argument may be used in combination with the other options. In this
case, the attribute values supplied as command line arguments are used to override
those that are read from the YAML/JSON extension descriptor file. For example, the
following command reads the extension from the 'mlflow.yaml' file but overrides the
ID and zone attribues with the values supplied as command line arguments:

  fuseml extension register --id mlflow-devel --zone local -f mlflow.yaml

`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate(cmd.Flags()))
			common.CheckErr(o.run(cmd.Flags()))
		},
		Args: cobra.ExactArgs(0),
	}
	cmd.Flags().StringVarP(&o.fromFile, "file", "f", "", "read the extension descriptor from a YAML or JSON file")
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

func (o *extensionRegisterOptions) validate(flags *pflag.FlagSet) error {
	if !flags.Changed("id") && !flags.Changed("product") && !flags.Changed("file") {
		return fmt.Errorf("an ID or a product must be configured for the extension")
	}

	return nil
}

func (o *extensionRegisterOptions) run(flags *pflag.FlagSet) error {
	extension := extension.Extension{}

	if flags.Changed("file") {
		ext, err := o.ExtensionClient.ReadExtensionFromFile(o.fromFile)
		if err != nil {
			return err
		}
		extension = *ext
	}

	if flags.Changed("id") {
		extension.ID = &o.extensionID
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

	ext, err := o.ExtensionClient.RegisterExtension(&extension)
	if err != nil {
		return err
	}

	fmt.Printf("Extension %q successfully registered\n", *ext.ID)

	return nil
}
