package project

import (
	"context"
	"fmt"

	projectc "github.com/fuseml/fuseml-core/gen/http/project/client"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

// CreateOptions holds the options for 'project get' sub command
type CreateOptions struct {
	client.Clients
	global      *common.GlobalOptions
	format      *common.FormattingOptions
	Name        string
	Description string
}

// NewCreateOptions creates a ProjectCreateOptions struct
func NewCreateOptions(o *common.GlobalOptions) *CreateOptions {
	res := &CreateOptions{global: o}
	res.format = common.NewSingleValueFormattingOptions()
	return res
}

// NewSubCmdProjectCreate creates and returns the cobra command for the `project get` CLI command
func NewSubCmdProjectCreate(gOpt *common.GlobalOptions) *cobra.Command {

	o := NewCreateOptions(gOpt)

	cmd := &cobra.Command{
		Use:   `create {-n|--name NAME} {-d|--desc DESCRIPTION} [flags]`,
		Short: "Create projects.",
		Long:  `Create new FuseML project`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "project name")
	cmd.Flags().StringVarP(&o.Description, "desc", "d", "", "project description")
	o.format.AddSingleValueFormattingFlags(cmd, common.FormatYAML)
	cmd.MarkFlagRequired("name")
	return cmd
}

func (o *CreateOptions) validate() error {
	return nil
}

func (o *CreateOptions) run() error {
	request, err := projectc.BuildCreatePayload(o.Name, o.Description)
	if err != nil {
		return err
	}

	_, err = o.ProjectClient.Create()(context.Background(), request)
	if err != nil {
		return err
	}

	fmt.Printf("Project %s successfully created.\n", o.Name)

	return nil
}
