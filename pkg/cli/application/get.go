package application

import (
	"context"
	"os"

	applicationc "github.com/fuseml/fuseml-core/gen/http/application/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

// GetOptions holds the options for 'application get' sub command
type getOptions struct {
	common.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
	Name   string
}

func newGetOptions(o *common.GlobalOptions) *getOptions {
	res := &getOptions{global: o}
	res.format = common.NewSingleValueFormattingOptions()
	return res
}

// newSubCmdApplicationGet creates and returns the cobra command for the `application get` CLI command
func newSubCmdApplicationGet(gOpt *common.GlobalOptions) *cobra.Command {

	o := newGetOptions(gOpt)

	cmd := &cobra.Command{
		Use:   `get {-n|--name NAME}`,
		Short: "Get an application.",
		Long:  `Show details about a FuseML application`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "application name")
	o.format.AddSingleValueFormattingFlags(cmd)
	cmd.MarkFlagRequired("name")
	return cmd
}

func (o *getOptions) validate() error {
	return nil
}

func (o *getOptions) run() error {
	request, err := applicationc.BuildGetPayload(o.Name)
	if err != nil {
		return err
	}

	response, err := o.ApplicationClient.Get()(context.Background(), request)
	if err != nil {
		return err
	}

	o.format.FormatValue(os.Stdout, response)

	return nil
}
