package codeset

import (
	"context"
	"os"
	"strings"

	codeset "github.com/fuseml/fuseml-core/gen/codeset"
	codesetc "github.com/fuseml/fuseml-core/gen/http/codeset/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// ListOptions holds the options for 'codeset list' sub command
type ListOptions struct {
	common.Clients
	global  *common.GlobalOptions
	format  *common.FormattingOptions
	Project string
	Label   string
}

// custom formatting handler used to format codeset labels
func formatLabels(object interface{}, column string, field interface{}) string {
	if codeset, ok := object.(*codeset.Codeset); ok {
		return strings.Join(codeset.Labels, "\n")
	}
	return ""
}

// NewListOptions initializes a ListOptions struct
func NewListOptions(o *common.GlobalOptions) (res *ListOptions) {
	res = &ListOptions{global: o}
	res.format = common.NewFormattingOptions(
		[]string{"Name", "Project", "Description", "Labels", "URL"},
		[]table.SortBy{{Name: "Name", Mode: table.Asc}, {Name: "Project", Mode: table.Asc}},
		common.OutputFormatters{"Labels": formatLabels},
	)

	return
}

// NewSubCmdCodesetList creates and returns the cobra command for the `codeset list` CLI command
func NewSubCmdCodesetList(c *common.GlobalOptions) *cobra.Command {

	o := NewListOptions(c)

	cmd := &cobra.Command{
		Use:   "list [-p|--project PROJECT] [-l|--label LABEL]",
		Short: "List codesets.",
		Long:  `Retrieve information about Codesets registered in FuseML`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(c))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.Project, "project", "p", "", "filter codesets by project")
	cmd.Flags().StringVarP(&o.Label, "label", "l", "", "filter codesets by label")
	o.format.AddMultiValueFormattingFlags(cmd)

	return cmd
}

func (o *ListOptions) validate() error {
	return nil
}

func (o *ListOptions) run() error {
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
