package codeset

import (
	"context"
	"os"
	"strings"

	codesetc "github.com/fuseml/fuseml-core/gen/http/codeset/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// CodesetListOptions holds the options for 'codeset list' sub command
type CodesetListOptions struct {
	common.Clients
	global  *common.GlobalOptions
	format  *common.FormattingOptions
	Project string
	Label   string
}

func formatLabels(object interface{}, column string, field interface{}) string {
	if labels, ok := field.([]string); ok {
		return strings.Join(labels, "\n")
	}
	return ""
}

// NewCodesetListOptionsOptions creates a CodesetListOptions struct
func NewCodesetListOptions(o *common.GlobalOptions) (res *CodesetListOptions) {
	res = &CodesetListOptions{global: o}
	res.format = common.NewFormattingOptions(
		[]string{"Name", "Project", "Description", "Labels", "URL"},
		common.OutputSortFields{"Name": table.Asc, "Project": table.Asc},
		common.OutputFormatters{"Labels": formatLabels},
	)

	return
}

func NewSubCmdCodesetList(c *common.GlobalOptions) *cobra.Command {

	o := NewCodesetListOptions(c)

	cmd := &cobra.Command{
		Use:   "list [-p|--project PROJECT] [-l|--label LABEL]",
		Short: "List codesets.",
		Long:  `Retrieve information about Codesets registered in FuseML`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(c))
			common.CheckErr(o.Validate())
			common.CheckErr(o.Run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.Project, "project", "p", "", "filter codesets by project")
	cmd.Flags().StringVarP(&o.Label, "label", "l", "", "filter codesets by label")
	o.format.AddMultiValueFormattingFlags(cmd)

	return cmd
}

func (o *CodesetListOptions) Validate() error {
	return nil
}

func (o *CodesetListOptions) Run() error {
	request, err := codesetc.BuildListPayload(o.Project, o.Label)
	if err != nil {
		return err
	}

	response, err := o.CodesetClient.List()(context.Background(), request)
	if err != nil {
		return err
	}

	o.format.FormatValue(os.Stdout, response)

	return nil
}
