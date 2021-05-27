package runnable

import (
	"context"
	"os"

	runnablec "github.com/fuseml/fuseml-core/gen/http/runnable/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

// GetOptions holds the options for 'runnable get' sub command
type GetOptions struct {
	common.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
	ID     string
}

// NewGetOptions creates a RunnableGetOptions struct
func NewGetOptions(o *common.GlobalOptions) *GetOptions {
	res := &GetOptions{global: o}
	res.format = common.NewSingleValueFormattingOptions()
	return res
}

// NewSubCmdRunnableGet creates and returns the cobra command for the `runnable get` CLI command
func NewSubCmdRunnableGet(gOpt *common.GlobalOptions) *cobra.Command {

	o := NewGetOptions(gOpt)

	cmd := &cobra.Command{
		Use:   `get {-n|--name NAME} {-p|--project PROJECT}`,
		Short: "Get runnables.",
		Long:  `Show details about a FuseML runnable`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.ID, "id", "n", "", "runnable ID")
	o.format.AddSingleValueFormattingFlags(cmd)
	cmd.MarkFlagRequired("id")
	return cmd
}

func (o *GetOptions) validate() error {
	return nil
}

func (o *GetOptions) run() error {
	request, err := runnablec.BuildGetPayload(o.ID)
	if err != nil {
		return err
	}

	response, err := o.RunnableClient.Get()(context.Background(), request)
	if err != nil {
		return err
	}

	o.format.FormatValue(os.Stdout, response)

	return nil
}
