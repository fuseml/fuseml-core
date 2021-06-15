package project

import (
	"fmt"
	"os"

	project "github.com/fuseml/fuseml-core/gen/project"

	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// ListOptions holds the options for 'project list' sub command
type ListOptions struct {
	client.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
}

// custom formatting handler used to format project users
func formatUsers(object interface{}, column string, field interface{}) (formated string) {
	if project, ok := object.(*project.Project); ok {
		for _, user := range project.Users {
			if formated != "" {
				formated += "\n"
			}
			formated += fmt.Sprintf("- name: %s\n  email: %s", user.Name, user.Email)
		}
	}
	return
}

// NewListOptions initializes a ListOptions struct
func NewListOptions(o *common.GlobalOptions) (res *ListOptions) {
	res = &ListOptions{global: o}
	res.format = common.NewFormattingOptions(
		[]string{"Name", "Description", "Users"},
		[]table.SortBy{{Name: "Name", Mode: table.Asc}},
		common.OutputFormatters{"Users": formatUsers},
	)

	return
}

// NewSubCmdProjectList creates and returns the cobra command for the `project list` CLI command
func NewSubCmdProjectList(gOpt *common.GlobalOptions) *cobra.Command {

	o := NewListOptions(gOpt)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all projects.",
		Long:  `Retrieve information about Projects registered in FuseML`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}
	o.format.AddMultiValueFormattingFlags(cmd)

	return cmd
}

func (o *ListOptions) validate() error {
	return nil
}

func (o *ListOptions) run() error {
	projects, err := o.ProjectClient.List()
	if err != nil {
		return err
	}

	o.format.FormatValue(os.Stdout, projects)

	return nil
}
