package cli

import (
	"fmt"
	"os"

	"github.com/fuseml/fuseml-core/pkg/cli/codeset"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCmdRoot() *cobra.Command {
	o := &common.GlobalOptions{}

	cmd := &cobra.Command{
		Use:   os.Args[0] + " [flags]", //"[--url HOST | -s HOST] [--timeout SECONDS] [--verbose|-v]",
		Short: "FuseML CLI",
		Long:  "FuseML command line client",
		// Run: func(cmd *cobra.Command, args []string) {
		// 	fmt.Println("fuseml ok")
		// },
	}

	pf := cmd.PersistentFlags()
	pf.StringVarP(&o.Url, "url", "u", "http://localhost:8000", "URL where the FuseML service is running")
	viper.BindPFlag("url", pf.Lookup("url"))
	viper.BindEnv("url", "FUSEML_SERVER_URL")

	pf.IntVar(&o.Timeout, "timeout", 30, "maximum number of seconds to wait for response")
	viper.BindPFlag("timeout", pf.Lookup("timeout"))
	viper.BindEnv("timeout", "FUSEML_HTTP_TIMEOUT")

	pf.BoolVarP(&o.Verbose, "verbose", "v", false, "print request and response details")
	viper.BindPFlag("verbose", pf.Lookup("verbose"))
	viper.BindEnv("verbose")

	cmd.AddCommand(codeset.NewCmdCodeset(o))

	return cmd
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	rootCmd := NewCmdRoot()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
