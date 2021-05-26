package runnable

import (
	"context"
	"fmt"

	runnablec "github.com/fuseml/fuseml-core/gen/http/runnable/client"
	"github.com/fuseml/fuseml-core/gen/runnable"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/spf13/cobra"
)

// RunnableRegisterOptions holds the options for 'runnable register' sub command
type RunnableRegisterOptions struct {
	common.Clients
	global       *common.GlobalOptions
	RunnableDesc string
}

// NewRunnableRegisterOptionsOptions creates a RunnableRegisterOptions struct
func NewRunnableRegisterOptions(o *common.GlobalOptions) *RunnableRegisterOptions {
	return &RunnableRegisterOptions{global: o}
}

func NewSubCmdRunnableRegister(gOpt *common.GlobalOptions) *cobra.Command {

	o := NewRunnableRegisterOptions(gOpt)

	cmd := &cobra.Command{
		Use:   `register RUNNABLE_FILE`,
		Short: "Register runnables.",
		Long:  `Register a runnable with FuseML`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt))
			common.CheckErr(common.LoadFileIntoVar(cmd.Flags().Arg(0), &o.RunnableDesc))
			common.CheckErr(o.Validate())
			common.CheckErr(o.Run())
		},
		Args: cobra.ExactArgs(1),
	}

	return cmd
}

func (o *RunnableRegisterOptions) Validate() error {
	return nil
}

func (o *RunnableRegisterOptions) Run() error {
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
