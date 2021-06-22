package codeset

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

// SetOptions holds the options for 'codeset set' sub command
type SetOptions struct {
	global *common.GlobalOptions
	format *common.FormattingOptions
	Name   string
}

// NewSetOptions creates a CodesetSetOptions struct
func NewSetOptions(o *common.GlobalOptions) *SetOptions {
	res := &SetOptions{global: o}
	res.format = common.NewSingleValueFormattingOptions()
	return res
}

// NewSubCmdCodesetSet creates and returns the cobra command for the `codeset set` CLI command
func NewSubCmdCodesetSet(gOpt *common.GlobalOptions) *cobra.Command {

	o := NewSetOptions(gOpt)

	cmd := &cobra.Command{
		Use:   `set {-n|--name NAME}`,
		Short: "Set current codeset.",
		Long:  `Set codeset as a current one`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "codeset name")
	o.format.AddSingleValueFormattingFlags(cmd, common.FormatYAML)
	cmd.MarkFlagRequired("name")
	return cmd
}

func (o *SetOptions) validate() error {
	return nil
}

func (o *SetOptions) run() error {

	if viper.GetString("CurrentCodeset") == o.Name {
		fmt.Printf("Codeset %s is already current one.\n", o.Name)
		return nil
	}

	viper.Set("CurrentCodeset", o.Name)

	if err := common.WriteConfigFile(); err != nil {
		return errors.Wrap(err, "Error writing config file")
	}

	fmt.Printf("Current codeset set to %s.\n", o.Name)
	return nil
}
