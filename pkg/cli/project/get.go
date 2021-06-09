package project

import (
	"context"
	"os"

	projectc "github.com/fuseml/fuseml-core/gen/http/project/client"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

// GetOptions holds the options for 'project get' sub command
type GetOptions struct {
	client.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
	Name   string
}

// NewGetOptions creates a ProjectGetOptions struct
func NewGetOptions(o *common.GlobalOptions) *GetOptions {
	res := &GetOptions{global: o}
	res.format = common.NewSingleValueFormattingOptions()
	return res
}

// NewSubCmdProjectGet creates and returns the cobra command for the `project get` CLI command
func NewSubCmdProjectGet(gOpt *common.GlobalOptions) *cobra.Command {

	o := NewGetOptions(gOpt)

	cmd := &cobra.Command{
		Use:   `get {-n|--name NAME}`,
		Short: "Get projects.",
		Long:  `Show details about a FuseML project`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "project name")
	o.format.AddSingleValueFormattingFlags(cmd, common.FormatYAML)
	cmd.MarkFlagRequired("name")
	return cmd
}

func (o *GetOptions) validate() error {
	return nil
}

func (o *GetOptions) run() error {
	request, err := projectc.BuildGetPayload(o.Name)
	if err != nil {
		return err
	}

	response, err := o.ProjectClient.Get()(context.Background(), request)
	if err != nil {
		return err
	}

	o.format.FormatValue(os.Stdout, response)

	return nil
}
