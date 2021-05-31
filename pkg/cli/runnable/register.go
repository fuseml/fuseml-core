package runnable

import (
	"context"
	"fmt"

	runnablec "github.com/fuseml/fuseml-core/gen/http/runnable/client"
	"github.com/fuseml/fuseml-core/gen/runnable"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

// RegisterOptions holds the options for 'runnable register' sub command
type RegisterOptions struct {
	client.Clients
	global       *common.GlobalOptions
	RunnableDesc string
}

// NewRegisterOptions initializes a RegisterOptions struct
func NewRegisterOptions(o *common.GlobalOptions) *RegisterOptions {
	return &RegisterOptions{global: o}
}

// NewSubCmdRunnableRegister creates and returns the cobra command for the `runnable register` CLI command
func NewSubCmdRunnableRegister(gOpt *common.GlobalOptions) *cobra.Command {

	o := NewRegisterOptions(gOpt)

	cmd := &cobra.Command{
		Use:   `register RUNNABLE_FILE`,
		Short: "Register runnables.",
		Long:  `Register a runnable with FuseML`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(common.LoadFileIntoVar(cmd.Flags().Arg(0), &o.RunnableDesc))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(1),
	}

	return cmd
}

func (o *RegisterOptions) validate() error {
	return nil
}

func (o *RegisterOptions) run() error {
	request, err := runnablec.BuildRegisterPayload(o.RunnableDesc)
	if err != nil {
		return err
	}

	response, err := o.RunnableClient.Register()(context.Background(), request)
	if err != nil {
		return err
	}

	runnable := response.(*runnable.Runnable)

	fmt.Printf("Runnable %s successfully registered\n", runnable.ID)

	return nil
}
