package workflow

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type deleteOptions struct {
	client.Clients
	global *common.GlobalOptions
	name   string
}

func newDeleteOptions(o *common.GlobalOptions) (res *deleteOptions) {
	return &deleteOptions{global: o}
}

func newSubCmdDelete(gOpt *common.GlobalOptions) *cobra.Command {
	o := newDeleteOptions(gOpt)
	cmd := &cobra.Command{
		Use:   "delete [-n|--name NAME]",
		Short: "Deletes a workflow",
		Long:  `Delete a workflow and all existing assignments to it.`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", "name of the workflow to be deleted")
	cmd.MarkFlagRequired("name")

	return cmd
}

func (o *deleteOptions) validate() error {
	return nil
}

func (o *deleteOptions) run() error {
	err := o.WorkflowClient.Delete(o.name)
	if err != nil {
		return err
	}

	fmt.Printf("Workflow %s successfully deleted\n", o.name)

	return nil
}
