package workflow

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	workflowc "github.com/fuseml/fuseml-core/gen/http/workflow/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type getOptions struct {
	common.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
	name   string
}

func newGetOptions(o *common.GlobalOptions) *getOptions {
	res := &getOptions{global: o}
	res.format = common.NewSingleValueFormattingOptions()
	return res
}

func newSubCmdGet(gOpt *common.GlobalOptions) *cobra.Command {
	o := newGetOptions(gOpt)
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

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "workflow name")
	o.format.AddSingleValueFormattingFlags(cmd)
	cmd.MarkFlagRequired("name")
	return cmd
}

func (o *getOptions) validate() error {
	return nil
}

func (o *getOptions) run() error {
	request, err := workflowc.BuildGetPayload(o.name)
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
