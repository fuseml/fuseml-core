package codeset

import (
	"context"
	"fmt"

	"github.com/fuseml/fuseml-core/gen/codeset"
	codesetc "github.com/fuseml/fuseml-core/gen/http/codeset/client"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	gitc "github.com/fuseml/fuseml-core/pkg/cli/git"
	"github.com/spf13/cobra"
)

// RegisterOptions holds the options for 'codeset register' sub command
type RegisterOptions struct {
	client.Clients
	global      *common.GlobalOptions
	Name        string
	Project     string
	Description string
	Labels      []string
	Location    string
}

// NewRegisterOptions creates a CodesetRegisterOptions struct
func NewRegisterOptions(o *common.GlobalOptions) *RegisterOptions {
	return &RegisterOptions{global: o}
}

// NewSubCmdCodesetRegister creates and returns the cobra command for the `codeset register` CLI command
func NewSubCmdCodesetRegister(gOpt *common.GlobalOptions) *cobra.Command {

	o := NewRegisterOptions(gOpt)

	cmd := &cobra.Command{
		Use: `register {-n|--name NAME} {-p|--project PROJECT} {-d|--desc DESCRIPTION} [--label LABEL] LOCATION [flags]

LOCATION can be path to local directory or URL of a git repository`,
		Short: "Register codesets.",
		Long:  `Register a codeset with FuseML.`,
		Run: func(cmd *cobra.Command, args []string) {
			o.Location = cmd.Flags().Arg(0)
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "codeset name")
	cmd.Flags().StringVarP(&o.Project, "project", "p", "", "the project to which the codeset belongs")
	cmd.Flags().StringVarP(&o.Description, "desc", "d", "", "codeset description")
	cmd.Flags().StringSliceVar(&o.Labels, "label", []string{}, "one or more codeset labels associated with the codeset")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("project")
	return cmd
}

func (o *RegisterOptions) validate() error {
	return nil
}

func (o *RegisterOptions) run() error {
	request, err := codesetc.BuildRegisterPayload(o.Name, o.Project, o.Description, o.Labels)
	if err != nil {
		return err
	}

	response, err := o.CodesetClient.Register()(context.Background(), request)
	if err != nil {
		return err
	}

	result := response.(*codeset.RegisterResult)
	codeset := result.Codeset

	err = gitc.Push(o.Project, o.Name, o.Location, *codeset.URL, result.Username, result.Password, o.global.Verbose)
	if err != nil {
		return err
	}

	fmt.Printf("Codeset %s successfully registered\n", *codeset.URL)
	if result.Username != nil {
		fmt.Printf("Username for accessing the project %s\n", *result.Username)
	}
	if result.Password != nil {
		fmt.Printf("Password for accessing the project %s\n", *result.Password)
	}

	return nil
}
