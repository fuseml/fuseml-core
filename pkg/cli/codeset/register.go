package codeset

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fuseml/fuseml-core/gen/codeset"
	codesetc "github.com/fuseml/fuseml-core/gen/http/codeset/client"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	gitc "github.com/fuseml/fuseml-core/pkg/cli/git"
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
	Password    string
	User        string
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

	cmd.Flags().StringVarP(&o.Password, "password", "", "", "(FUSEML_PROJECT_PASSWORD) Password of the user accessing a project")
	viper.BindEnv("password", "FUSEML_PROJECT_PASSWORD")

	cmd.Flags().StringVarP(&o.User, "user", "", "", "(FUSEML_PROJECT_USER) Username of the user accessing a project")
	viper.BindEnv("user", "FUSEML_PROJECT_USER")

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

	// priority have username/password from the registering (when the new user was created)
	password := result.Password
	username := result.Username
	if username == nil && o.User != "" {
		username = &o.User
	}
	if password == nil && o.Password != "" {
		password = &o.Password
	}

	err = gitc.Push(o.Project, o.Name, o.Location, *codeset.URL, username, password, o.global.Verbose)
	if err != nil {
		return err
	}

	fmt.Printf("Codeset %s successfully registered\n", *codeset.URL)

	if result.Username != nil {
		if viper.GetString("Username") != *result.Username {
			fmt.Println("Saving new username into config file as current username.")
			viper.Set("Username", *result.Username)
		}
	}
	if result.Password != nil {
		if viper.GetString("Password") != *result.Password {
			fmt.Println("Saving new password into config file as current password.")
			viper.Set("Password", *result.Password)
		}
	}

	if viper.GetString("CurrentCodeset") != o.Name {
		fmt.Printf("Setting %s as current codeset.\n", o.Name)
		viper.Set("CurrentCodeset", o.Name)
	}

	if viper.GetString("CurrentProject") != o.Project {
		fmt.Printf("Setting %s as current project.\n", o.Project)
		viper.Set("CurrentProject", o.Project)
	}

	if err := common.WriteConfigFile(); err != nil {
		return errors.Wrap(err, "Error writing config file")
	}

	return nil
}
