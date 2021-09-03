package credentials

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/gen/extension"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type credentialsAddOptions struct {
	client.Clients
	global           *common.GlobalOptions
	credentialsID    string
	credentialsScope string
	projects         []string
	users            []string
	config           common.KeyValueArgs
}

func newCredentialsAddOptions(o *common.GlobalOptions) *credentialsAddOptions {
	return &credentialsAddOptions{global: o}
}

func newSubCmdCredentialsAdd(gOpt *common.GlobalOptions) *cobra.Command {
	o := newCredentialsAddOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `add [--id CREDENTIALS_ID] [--scope {global|project|user}] [--proj PROJECT]... [--user USER]... [-c|--configuration KEY:VALUE]... {EXTENSION_ID} {SERVICE_ID}`,
		Short: "Add a new credentials to an existing FuseML extension service",
		Long:  `Add an credentials to a FuseML extension service already registered with the extension registry`,
		Run: func(cmd *cobra.Command, args []string) {
			o.config.Unpack()
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run(cmd.Flags().Arg(0), cmd.Flags().Arg(1)))
		},
		Args: cobra.ExactArgs(2),
	}
	cmd.Flags().StringVar(&o.credentialsID, "id", "", "credentials ID")
	cmd.Flags().StringVar(&o.credentialsScope, "scope", "global", "credentials scope (global/project/user)")
	cmd.Flags().StringSliceVar(&o.projects, "proj", []string{}, "project that has access to the credentials. One or more may be supplied")
	cmd.Flags().StringSliceVar(&o.users, "user", []string{}, "user that has access to the credentials. One or more may be supplied")
	cmd.Flags().StringSliceVarP(&o.config.Packed, "configuration", "c", []string{}, "credentials configuration data. One or more may be supplied")

	return cmd
}

func (o *credentialsAddOptions) validate() error {
	return common.ValidateEnumArgument("credentials scope", o.credentialsScope, []string{"global", "project", "user"})
}

func (o *credentialsAddOptions) run(extensionID, serviceID string) error {
	credentials := extension.ExtensionCredentials{
		ID:            &o.credentialsID,
		ExtensionID:   &extensionID,
		ServiceID:     &serviceID,
		Scope:         &o.credentialsScope,
		Projects:      o.projects,
		Users:         o.users,
		Configuration: o.config.Unpacked,
	}
	cred, err := o.ExtensionClient.AddCredentials(&credentials)
	if err != nil {
		return err
	}

	fmt.Printf("Credentials %q successfully added to service %s from extension %q\n", *cred.ID, serviceID, extensionID)

	return nil
}
