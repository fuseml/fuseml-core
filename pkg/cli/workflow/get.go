package workflow

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	workflowc "github.com/fuseml/fuseml-core/gen/http/workflow/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

// GetOptions holds the options for 'workflow get' sub command
type GetOptions struct {
	common.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
	Name   string
}

// NewGetOptions creates a GetOptions struct
func NewGetOptions(o *common.GlobalOptions) *GetOptions {
	res := &GetOptions{global: o}
	res.format = common.NewSingleValueFormattingOptions()
	return res
}

// NewSubCmdGet creates and returns the cobra command for the `workflow get` CLI command
func NewSubCmdGet(gOpt *common.GlobalOptions) *cobra.Command {
	o := NewGetOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `get {-n|--name NAME}`,
		Short: "Get a workflow",
		Long:  `Show detailed information from a workflow`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "workflow name")
	o.format.AddSingleValueFormattingFlags(cmd)
	cmd.MarkFlagRequired("name")
	return cmd
}

func (o *GetOptions) validate() error {
	return nil
}

func (o *GetOptions) run() error {
	request, err := workflowc.BuildGetPayload(o.Name)
	if err != nil {
		return err
	}

	response, err := o.WorkflowClient.Get()(context.Background(), request)
	if err != nil {
		return err
	}

	o.format.FormatValue(os.Stdout, response)

	return nil
}
