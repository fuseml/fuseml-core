package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fuseml/fuseml-core/pkg/cli/codeset"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/fuseml/fuseml-core/pkg/cli/runnable"
	"github.com/fuseml/fuseml-core/pkg/cli/workflow"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// NewCmdRoot creates and returns the cobra command that acts as a root for all other CLI sub-commands
func NewCmdRoot() *cobra.Command {
	o := &common.GlobalOptions{}

	cmd := &cobra.Command{
		Use:   os.Args[0] + " [--url HOST | -s HOST] [--timeout SECONDS] [--verbose|-v]",
		Short: "FuseML CLI",
		Long:  "FuseML command line client",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
			return initializeConfig(cmd)
		},
	}

	pf := cmd.PersistentFlags()
	pf.StringVarP(&o.URL, "url", "u", common.DefaultFuseMLURL, "URL where the FuseML service is running")
	viper.BindEnv("url", "FUSEML_SERVER_URL")

	pf.IntVar(&o.Timeout, "timeout", common.DefaultHTTPTimeout, "maximum number of seconds to wait for response")
	viper.BindEnv("timeout", "FUSEML_HTTP_TIMEOUT")

	pf.BoolVarP(&o.Verbose, "verbose", "v", false, "print verbose information, such as HTTP request and response details")
	viper.BindEnv("verbose")

	cmd.AddCommand(codeset.NewCmdCodeset(o))
	cmd.AddCommand(runnable.NewCmdRunnable(o))
	cmd.AddCommand(workflow.NewCmdWorkflow(o))

	return cmd
}

func initializeConfig(cmd *cobra.Command) error {

	// Set the base name of the config file, without the file extension.
	viper.SetConfigName(common.ConfigFileName)

	// Set the config format and extension
	viper.SetConfigType("yaml")

	// Set paths where viper should look for the config file.
	if dirname, err := os.Getwd(); err == nil {
		viper.AddConfigPath(filepath.Join(dirname, common.ConfigHomeSubdir))
	}
	if dirname, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(filepath.Join(dirname, common.ConfigHomeSubdir))
	}

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// Bind the current command's flags to viper
	bindFlags(cmd)

	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
// This is required because viper doesn't work with cobra flags that are also bound to a variable
// (e.g. using StringVar to bind a flag to a string variable). See https://github.com/spf13/viper/issues/671.
func bindFlags(cmd *cobra.Command) {

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && viper.IsSet(f.Name) {
			val := viper.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})

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
