package endpoint

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type endpointListOptions struct {
	client.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
}

func newEndpointListOptions(o *common.GlobalOptions) (res *endpointListOptions) {
	res = &endpointListOptions{global: o}
	res.format = common.NewFormattingOptions(
		[]string{"URL", "Type", "Configuration"},
		[]table.SortBy{{Name: "URL", Mode: table.Asc}},
		common.OutputFormatters{"Configuration": common.FormatMapField},
	)

	return
}

func newSubCmdEndpointList(gOpt *common.GlobalOptions) *cobra.Command {
	o := newEndpointListOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `list {EXTENSION_ID} {SERVICE_ID}`,
		Short: "Lists endpoints",
		Long:  `Display information about the endpoints configured for an extension service.`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run(cmd.Flags().Arg(0), cmd.Flags().Arg(1)))
		},
		Args: cobra.ExactArgs(2),
	}
	o.format.AddMultiValueFormattingFlags(cmd)

	return cmd
}

func (o *endpointListOptions) validate() error {
	return nil
}

func (o *endpointListOptions) run(extensionID, serviceID string) error {
	eps, err := o.ExtensionClient.ListEndpoints(extensionID, serviceID)
	if err != nil {
		return err
	}

	o.format.FormatValue(os.Stdout, eps)

	return nil
}
