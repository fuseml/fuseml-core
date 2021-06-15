package project

import (
	"fmt"

	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

// DeleteOptions holds the options for 'project delete' sub command
type DeleteOptions struct {
	client.Clients
	global *common.GlobalOptions
	Name   string
}

// NewDeleteOptions creates a ProjectDeleteOptions struct
func NewDeleteOptions(o *common.GlobalOptions) *DeleteOptions {
	return &DeleteOptions{global: o}
}

// NewSubCmdProjectDelete creates and returns the cobra command for the `project delete` CLI command
func NewSubCmdProjectDelete(gOpt *common.GlobalOptions) *cobra.Command {

	o := NewDeleteOptions(gOpt)

	cmd := &cobra.Command{
		Use:   `delete {-n|--name NAME}`,
		Short: "Delete projects.",
		Long:  `Delete a project from FuseML`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "project name")
	cmd.MarkFlagRequired("name")
	return cmd
}

func (o *DeleteOptions) validate() error {
	return nil
}

func (o *DeleteOptions) run() error {
	err := o.ProjectClient.Delete(o.Name)
	if err != nil {
		return err
	}

	fmt.Printf("Project %s successfully deleted\n", o.Name)

	return nil
}
