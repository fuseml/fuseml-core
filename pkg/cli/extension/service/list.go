package service

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type serviceListOptions struct {
	client.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
}

func newServiceListOptions(o *common.GlobalOptions) (res *serviceListOptions) {
	res = &serviceListOptions{global: o}
	res.format = common.NewFormattingOptions(
		[]string{"ID", "Resource", "Category", "Configuration"},
		[]table.SortBy{{Name: "ID", Mode: table.Asc}},
		common.OutputFormatters{"Configuration": common.FormatMapField},
	)

	return
}

func newSubCmdServiceList(gOpt *common.GlobalOptions) *cobra.Command {
	o := newServiceListOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `list {EXTENSION_ID}`,
		Short: "Lists services",
		Long:  `Display information about the services configured for an extension.`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run(cmd.Flags().Arg(0)))
		},
		Args: cobra.ExactArgs(1),
	}
	o.format.AddMultiValueFormattingFlags(cmd)

	return cmd
}

func (o *serviceListOptions) validate() error {
	return nil
}

func (o *serviceListOptions) run(extensionID string) error {
	svcs, err := o.ExtensionClient.ListServices(extensionID)
	if err != nil {
		return err
	}

	o.format.FormatValue(os.Stdout, svcs)

	return nil
}
