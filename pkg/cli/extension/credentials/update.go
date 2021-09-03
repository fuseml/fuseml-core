package credentials

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/fuseml/fuseml-core/gen/extension"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type credentialsUpdateOptions struct {
	client.Clients
	global           *common.GlobalOptions
	credentialsScope string
	projects         []string
	users            []string
	config           common.KeyValueArgs
}

func newCredentialsUpdateOptions(o *common.GlobalOptions) *credentialsUpdateOptions {
	return &credentialsUpdateOptions{global: o}
}

func newSubCmdCredentialsUpdate(gOpt *common.GlobalOptions) *cobra.Command {
	o := newCredentialsUpdateOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `update [--scope {global|project|user}] [--proj PROJECT]... [--user USER]... [-c|--configuration KEY:VALUE]... {EXTENSION_ID} {SERVICE_ID} {CREDENTIALS_ID}`,
		Short: "Update the attributes of an existing set of FuseML extension credentials",
		Long:  `Update the attributes of a set of FuseML extension service credentials already registered with the extension registry`,
		Run: func(cmd *cobra.Command, args []string) {
			o.config.Unpack()
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run(cmd.Flags().Arg(0), cmd.Flags().Arg(1), cmd.Flags().Arg(2), cmd.Flags()))
		},
		Args: cobra.ExactArgs(3),
	}
	cmd.Flags().StringVar(&o.credentialsScope, "scope", "global", "credentials scope (global/project/user)")
	cmd.Flags().StringSliceVar(&o.projects, "proj", []string{}, "project that has access to the credentials. One or more may be supplied")
	cmd.Flags().StringSliceVar(&o.users, "user", []string{}, "user that has access to the credentials. One or more may be supplied")
	cmd.Flags().StringSliceVarP(&o.config.Packed, "configuration", "c", []string{}, "credentials configuration data. One or more may be supplied")

	return cmd
}

func (o *credentialsUpdateOptions) validate() error {
	return common.ValidateEnumArgument("credentials scope", o.credentialsScope, []string{"global", "project", "user"})
}

func (o *credentialsUpdateOptions) run(extensionID, serviceID, credentialsID string, flags *pflag.FlagSet) error {
	credentials := extension.ExtensionCredentials{
		ID:          &credentialsID,
		ExtensionID: &extensionID,
		ServiceID:   &serviceID,
	}
	if flags.Changed("scope") {
		credentials.Scope = &o.credentialsScope
	}
	if flags.Changed("project") {
		credentials.Projects = o.projects
	}
	if flags.Changed("user") {
		credentials.Users = o.users
	}
	if flags.Changed("configuration") {
		credentials.Configuration = o.config.Unpacked
	}
	cred, err := o.ExtensionClient.UpdateCredentials(&credentials)
	if err != nil {
		return err
	}

	fmt.Printf("Credentials %q from service %s and extension %q successfully updated\n", *cred.ID, serviceID, extensionID)

	return nil
}
