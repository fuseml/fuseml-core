package codeset

import (
	"context"
	"os"

	codesetc "github.com/fuseml/fuseml-core/gen/http/codeset/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

// GetOptions holds the options for 'codeset get' sub command
type GetOptions struct {
	common.Clients
	global  *common.GlobalOptions
	format  *common.FormattingOptions
	Name    string
	Project string
}

// NewGetOptions creates a CodesetGetOptions struct
func NewGetOptions(o *common.GlobalOptions) *GetOptions {
	res := &GetOptions{global: o}
	res.format = common.NewSingleValueFormattingOptions()
	return res
}

// NewSubCmdCodesetGet creates and returns the cobra command for the `codeset get` CLI command
func NewSubCmdCodesetGet(gOpt *common.GlobalOptions) *cobra.Command {

	o := NewGetOptions(gOpt)

	cmd := &cobra.Command{
		Use:   `get {-n|--name NAME} {-p|--project PROJECT}`,
		Short: "Get codesets.",
		Long:  `Show details about a FuseML codeset`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "codeset name")
	cmd.Flags().StringVarP(&o.Project, "project", "p", "", "the project to which the codeset belongs")
	o.format.AddSingleValueFormattingFlags(cmd, common.FormatYAML)
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("project")
	return cmd
}

func (o *GetOptions) validate() error {
	return nil
}

func (o *GetOptions) run() error {
	request, err := codesetc.BuildGetPayload(o.Project, o.Name)
	if err != nil {
		return err
	}

	response, err := o.CodesetClient.Get()(context.Background(), request)
	if err != nil {
		return err
	}

	o.format.FormatValue(os.Stdout, response)

	return nil
}
