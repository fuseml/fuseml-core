package codeset

import (
	"context"
	"fmt"

	codesetc "github.com/fuseml/fuseml-core/gen/http/codeset/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

// DeleteOptions holds the options for 'codeset delete' sub command
type DeleteOptions struct {
	common.Clients
	global  *common.GlobalOptions
	Name    string
	Project string
}

// NewDeleteOptions creates a CodesetDeleteOptions struct
func NewDeleteOptions(o *common.GlobalOptions) *DeleteOptions {
	return &DeleteOptions{global: o}
}

// NewSubCmdCodesetDelete creates and returns the cobra command for the `codeset delete` CLI command
func NewSubCmdCodesetDelete(gOpt *common.GlobalOptions) *cobra.Command {

	o := NewDeleteOptions(gOpt)

	cmd := &cobra.Command{
		Use:   `delete {-n|--name NAME} {-p|--project PROJECT}`,
		Short: "Delete codesets.",
		Long:  `Delete a codeset from FuseML`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "codeset name")
	cmd.Flags().StringVarP(&o.Project, "project", "p", "", "the project to which the codeset belongs")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("project")
	return cmd
}

func (o *DeleteOptions) validate() error {
	return nil
}

func (o *DeleteOptions) run() error {
	request, err := codesetc.BuildDeletePayload(o.Project, o.Name)
	if err != nil {
		return err
	}

	_, err = o.CodesetClient.Delete()(context.Background(), request)
	if err != nil {
		return err
	}

	fmt.Printf("Codeset %s successfully deleted\n", o.Name)

	return nil
}
