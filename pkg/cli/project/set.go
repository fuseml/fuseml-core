package project

import (
	"fmt"

	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SetOptions holds the options for 'project set' sub command
type SetOptions struct {
	global *common.GlobalOptions
	format *common.FormattingOptions
	Name   string
}

// NewSetOptions creates a ProjectSetOptions struct
func NewSetOptions(o *common.GlobalOptions) *SetOptions {
	res := &SetOptions{global: o}
	res.format = common.NewSingleValueFormattingOptions()
	return res
}

// NewSubCmdProjectSet creates and returns the cobra command for the `project set` CLI command
func NewSubCmdProjectSet(gOpt *common.GlobalOptions) *cobra.Command {

	o := NewSetOptions(gOpt)

	cmd := &cobra.Command{
		Use:   `set {-n|--name NAME}`,
		Short: "Set current project.",
		Long:  `Set project as a current one`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "project name")
	o.format.AddSingleValueFormattingFlags(cmd, common.FormatYAML)
	cmd.MarkFlagRequired("name")
	return cmd
}

func (o *SetOptions) validate() error {
	return nil
}

func (o *SetOptions) run() error {

	if viper.GetString("CurrentProject") == o.Name {
		fmt.Printf("Project %s is already current one.\n", o.Name)
		return nil
	}

	viper.Set("CurrentProject", o.Name)

	if err := viper.WriteConfig(); err != nil {
		return err
	}

	fmt.Printf("Current project set to %s.\n", o.Name)
	return nil
}
