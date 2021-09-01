package extension

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type extensionRegisterOptions struct {
	client.Clients
	global *common.GlobalOptions
}

func newExtensionRegisterOptions(o *common.GlobalOptions) *extensionRegisterOptions {
	return &extensionRegisterOptions{global: o}
}

func newSubCmdExtensionRegister(gOpt *common.GlobalOptions) *cobra.Command {
	o := newExtensionRegisterOptions(gOpt)
	cmd := &cobra.Command{
		Use:   `register {WORKFLOW_FILE}`,
		Short: "Registers a FuseML extension",
		Long:  `Registers an external application as a FuseML extension`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run(cmd.Flags().Arg(0)))
		},
		Args: cobra.ExactArgs(1),
	}

	return cmd
}

func (o *extensionRegisterOptions) validate() error {
	return nil
}

func (o *extensionRegisterOptions) run(extensionDesc string) error {
	ext, err := o.ExtensionClient.RegisterExtensionFromFile(extensionDesc)
	if err != nil {
		return err
	}

	fmt.Printf("Extension %q successfully registered\n", *ext.ID)

	return nil
}
