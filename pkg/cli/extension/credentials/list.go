package credentials

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type credentialsListOptions struct {
	client.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
}

func newCredentialsListOptions(o *common.GlobalOptions) (res *credentialsListOptions) {
	res = &credentialsListOptions{global: o}
	res.format = common.NewFormattingOptions(
		[]string{"ID", "Scope", "Projects", "Users", "Configuration"},
		[]table.SortBy{{Name: "ID", Mode: table.Asc}},
		common.OutputFormatters{
			"Configuration": common.FormatMapField,
			"Projects":      common.FormatSliceField,
			"Users":         common.FormatSliceField,
		},
	)

	return
}

func newSubCmdCredentialsList(gOpt *common.GlobalOptions) *cobra.Command {
	o := newCredentialsListOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `list {EXTENSION_ID} {SERVICE_ID}`,
		Short: "Lists credentials",
		Long:  `Display information about the credentials configured for an extension service.`,
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

func (o *credentialsListOptions) validate() error {
	return nil
}

func (o *credentialsListOptions) run(extensionID, serviceID string) error {
	creds, err := o.ExtensionClient.ListCredentials(extensionID, serviceID)
	if err != nil {
		return err
	}

	o.format.FormatValue(os.Stdout, creds)

	return nil
}
