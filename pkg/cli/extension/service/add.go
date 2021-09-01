package service

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/gen/extension"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type serviceAddOptions struct {
	client.Clients
	global       *common.GlobalOptions
	serviceID    string
	description  string
	resource     string
	category     string
	authRequired bool
	config       common.KeyValueArgs
}

func newServiceAddOptions(o *common.GlobalOptions) *serviceAddOptions {
	return &serviceAddOptions{global: o}
}

func newSubCmdServiceAdd(gOpt *common.GlobalOptions) *cobra.Command {
	o := newServiceAddOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `add [--id SERVICE_ID] [--desc DESCRIPTION] [-r|--resource SERVICE_RESOURCE] [-c|--category SERVICE_CATEGORY] [--auth-required={true|false}] [--configuration KEY:VALUE]... {EXTENSION_ID}`,
		Short: "Add a new service to an existing FuseML extension",
		Long:  `Add a service to a FuseML extension already registered with the extension registry`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run(cmd.Flags().Arg(0)))
		},
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().StringVar(&o.serviceID, "id", "", "service ID")
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

func (o *serviceAddOptions) validate() error {
	return nil
}

func (o *serviceAddOptions) run(extensionID string) error {
	service := extension.ExtensionService{
		ID:            &o.serviceID,
		ExtensionID:   &extensionID,
		Description:   &o.description,
		Resource:      &o.resource,
		Category:      &o.category,
		AuthRequired:  &o.authRequired,
		Configuration: o.config.Unpacked,
	}
	svc, err := o.ExtensionClient.AddService(&service)
	if err != nil {
		return err
	}

	fmt.Printf("Service %q successfully added to extension %q\n", *svc.ID, extensionID)

	return nil
}
