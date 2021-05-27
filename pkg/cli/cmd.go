package cli

import (
	"os"

	"github.com/fuseml/fuseml-core/pkg/cli/codeset"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/fuseml/fuseml-core/pkg/cli/runnable"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewCmdRoot creates and returns the cobra command that acts as a root for all other CLI sub-commands
func NewCmdRoot() *cobra.Command {
	o := &common.GlobalOptions{}

	cmd := &cobra.Command{
		Use:   os.Args[0] + " [--url HOST | -s HOST] [--timeout SECONDS] [--verbose|-v]",
		Short: "FuseML CLI",
		Long:  "FuseML command line client",
	}

	pf := cmd.PersistentFlags()
	pf.StringVarP(&o.URL, "url", "u", "http://localhost:8000", "URL where the FuseML service is running")
	viper.BindPFlag("url", pf.Lookup("url"))
	viper.BindEnv("url", "FUSEML_SERVER_URL")

	pf.IntVar(&o.Timeout, "timeout", 30, "maximum number of seconds to wait for response")
	viper.BindPFlag("timeout", pf.Lookup("timeout"))
	viper.BindEnv("timeout", "FUSEML_HTTP_TIMEOUT")

	pf.BoolVarP(&o.Verbose, "verbose", "v", false, "print verbose information, such as HTTP request and response details")
	viper.BindPFlag("verbose", pf.Lookup("verbose"))
	viper.BindEnv("verbose")

	cmd.AddCommand(codeset.NewCmdCodeset(o))
	cmd.AddCommand(runnable.NewCmdRunnable(o))

	return cmd
}

// Execute creates the root cobra command, which in turn creates all sub-commands and sets all flags appropriately.
func Execute() {

	rootCmd := NewCmdRoot()

	// Errors caught here should only be those that come from the cobra framework regarding
	// incorrect command line arguments. They don't need to be printed again, as it's already
	// done by cobra
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
