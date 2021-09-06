package service

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/fuseml/fuseml-core/gen/extension"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type serviceUpdateOptions struct {
	client.Clients
	global       *common.GlobalOptions
	description  string
	resource     string
	category     string
	authRequired bool
	config       common.KeyValueArgs
}

func newServiceUpdateOptions(o *common.GlobalOptions) *serviceUpdateOptions {
	return &serviceUpdateOptions{global: o}
}

func newSubCmdServiceUpdate(gOpt *common.GlobalOptions) *cobra.Command {
	o := newServiceUpdateOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `update [--desc DESCRIPTION] [-r|--resource SERVICE_RESOURCE] [-c|--category SERVICE_CATEGORY] [--auth-required={true|false}] [--configuration KEY:VALUE]... {EXTENSION_ID} {SERVICE_ID}`,
		Short: "Update the attributes of an existing FuseML extension service",
		Long:  `Update the attributes of FuseML extension service already registered with the extension registry`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run(cmd.Flags().Arg(0), cmd.Flags().Arg(1), cmd.Flags()))
		},
		Args: cobra.ExactArgs(2),
	}
	cmd.Flags().StringVar(&o.description, "desc", "", "service description")
	cmd.Flags().StringVarP(&o.resource, "resource", "r", "",
		"particular API or protocol (e.g. s3, git, mlflow) provided by the extension service")
	cmd.Flags().StringVarP(&o.category, "category", "c", "",
		`a well-known category of AI/ML services provided by the extension service
(e.g. model store, feature store, distributed training, serving)`)
	cmd.Flags().BoolVar(&o.authRequired, "auth-required", false, "determines if the service requires authentication credentials to be accessed (default: false)")
	cmd.Flags().StringSliceVar(&o.config.Packed, "configuration", []string{}, "service configuration data. One or more may be supplied.")

	return cmd
}

func (o *serviceUpdateOptions) validate() error {
	return nil
}

func (o *serviceUpdateOptions) run(extensionID, serviceID string, flags *pflag.FlagSet) error {
	service := extension.ExtensionService{
		ID:          &serviceID,
		ExtensionID: &extensionID,
	}
	if flags.Changed("desc") {
		service.Description = &o.description
	}
	if flags.Changed("resource") {
		service.Resource = &o.resource
	}
	if flags.Changed("category") {
		service.Category = &o.category
	}
	if flags.Changed("auth-required") {
		service.AuthRequired = &o.authRequired
	}
	if flags.Changed("configuration") {
		service.Configuration = o.config.Unpacked
	}
	_, err := o.ExtensionClient.UpdateService(&service)
	if err != nil {
		return err
	}

	fmt.Printf("Service %q from extension %q successfully updated\n", serviceID, extensionID)

	return nil
}
