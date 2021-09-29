package codeset

import (
	"context"
	"fmt"
	"os"
	"strings"

	codeset "github.com/fuseml/fuseml-core/gen/codeset"
	codesetc "github.com/fuseml/fuseml-core/gen/http/codeset/client"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// ListOptions holds the options for 'codeset list' sub command
type ListOptions struct {
	client.Clients
	global  *common.GlobalOptions
	format  *common.FormattingOptions
	Project string
	Label   string
	All     bool
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
func NewSubCmdCodesetList(gOpt *common.GlobalOptions) *cobra.Command {

	o := NewListOptions(gOpt)

	cmd := &cobra.Command{
		Use:   "list [-p|--project PROJECT] [-l|--label LABEL] [--all]",
		Short: "List codesets.",
		Long:  `Retrieve information about Codesets registered in FuseML`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.Project, "project", "p", "", "filter codesets by project (filled by CurrentProject config value if present)")
	cmd.Flags().StringVarP(&o.Label, "label", "l", "", "filter codesets by label")
	cmd.Flags().BoolVar(&o.All, "all", false, "show all codesets; ignores 'label' and 'project' options (default: false)")
	o.format.AddMultiValueFormattingFlags(cmd)

	return cmd
}

func (o *ListOptions) validate() error {
	return nil
}

func (o *ListOptions) run() error {
	if o.All {
		o.Project = ""
		o.Label = ""
	}
	request, err := codesetc.BuildListPayload(o.Project, o.Label)
	if err != nil {
		return err
	}

	response, err := o.CodesetClient.List()(context.Background(), request)
	if err != nil {
		return err
	}

	if o.Project == "" && o.Label == "" {
		fmt.Println("Listing all Codesets:")
	} else if o.Project != "" && o.Label == "" {
		fmt.Printf("Listing Codesets for project %s:\n", o.Project)
	} else if o.Project == "" && o.Label != "" {
		fmt.Printf("Listing Codesets with label %s:\n", o.Label)
	} else {
		fmt.Printf("Listing Codesets for project %s and with label %s:\n", o.Project, o.Label)
	}
	o.format.FormatValue(os.Stdout, response)

	return nil
}
