package version

import (
	"fmt"
	"os"

	versionc "github.com/fuseml/fuseml-core/gen/version"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/fuseml/fuseml-core/pkg/version"

	"github.com/spf13/cobra"
)

// versionOptions holds the options for the 'version' sub command
type versionOptions struct {
	client.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
}

// newVersionOptions initializes a versionOptions struct
func newVersionOptions(gOpt *common.GlobalOptions) *versionOptions {
	res := &versionOptions{global: gOpt}
	res.format = common.NewSingleValueFormattingOptions()
	return res
}

// NewCmdVersion creates and returns the cobra command for the `version` CLI command
func NewCmdVersion(gOpt *common.GlobalOptions) *cobra.Command {

	o := newVersionOptions(gOpt)

	cmd := &cobra.Command{
		Use:   "version",
		Short: "display version information",
		Long:  `Display version information about the CLI and server`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	o.format.AddSingleValueFormattingFlags(cmd, common.FormatYAML)
	return cmd
}

func (o *versionOptions) run() error {
	var data = struct {
		Client *version.Info
		Server *versionc.VersionInfo
	}{}

	data.Client = version.GetInfo()
	err := o.InitializeClients(o.global.URL, o.global.Timeout, o.global.Verbose)

	if err == nil {
		data.Server, err = o.VersionClient.Get()
	}

	o.format.FormatValue(os.Stdout, data)

	if err != nil {
		return fmt.Errorf("could not retrieve server version information: %s", err.Error())
	}

	return nil
}
