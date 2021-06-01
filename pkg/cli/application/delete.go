package application

import (
	"context"
	"fmt"

	applicationc "github.com/fuseml/fuseml-core/gen/http/application/client"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

// deleteOptions holds the options for 'application delete' sub command
type deleteOptions struct {
	client.Clients
	global *common.GlobalOptions
	Name   string
}

func newDeleteOptions(o *common.GlobalOptions) *deleteOptions {
	return &deleteOptions{global: o}
}

// newSubCmdApplicationDelete creates and returns the cobra command for the `application delete` CLI command
func newSubCmdApplicationDelete(gOpt *common.GlobalOptions) *cobra.Command {

	o := newDeleteOptions(gOpt)

	cmd := &cobra.Command{
		Use:   `delete {-n|--name NAME}`,
		Short: "Delete an application.",
		Long:  `Delete an application registered by FuseML`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "application name")
	cmd.MarkFlagRequired("name")
	return cmd
}

func (o *deleteOptions) validate() error {
	return nil
}

func (o *deleteOptions) run() error {
	request, err := applicationc.BuildDeletePayload(o.Name)
	if err != nil {
		return err
	}

	_, err = o.ApplicationClient.Delete()(context.Background(), request)
	if err != nil {
		return err
	}

	fmt.Printf("Application %s successfully deleted\n", o.Name)

	return nil
}
