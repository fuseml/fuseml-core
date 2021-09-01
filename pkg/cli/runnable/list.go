package runnable

import (
	"context"
	"fmt"
	"os"
	"strings"

	runnablec "github.com/fuseml/fuseml-core/gen/http/runnable/client"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// ListOptions holds the options for 'runnable list' sub command
type ListOptions struct {
	client.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
	ID     string
	Kind   string
	Labels common.KeyValueArgs
}

// custom formatting handler used to format runnable labels
func formatLabels(object interface{}, column string, field interface{}) string {
	if labels, ok := field.(map[string]interface{}); ok {
		labelStr := make([]string, 0)
		for k, v := range labels {
			labelStr = append(labelStr, fmt.Sprintf("%s: %s", k, v))
		}
		return strings.Join(labelStr, "\n")
	}
	return ""
}

// NewListOptions initializes a ListOptions struct
func NewListOptions(o *common.GlobalOptions) (res *ListOptions) {
	res = &ListOptions{global: o}
	res.format = common.NewFormattingOptions(
		[]string{"ID", "Kind", "Description", "Labels"},
		[]table.SortBy{{Name: "ID", Mode: table.Asc}},
		common.OutputFormatters{"Labels": formatLabels},
	)

	return
}

// NewSubCmdRunnableList creates and returns the cobra command for the `runnable list` CLI command
func NewSubCmdRunnableList(gOpt *common.GlobalOptions) *cobra.Command {

	o := NewListOptions(gOpt)
	cmd := &cobra.Command{
		Use:   "list [-i|--id ID] [-k|--kind KIND] [-l|--label LABEL_KEY:LABEL_VALUE]...",
		Short: "List runnables.",
		Long:  `Retrieve information about Runnables registered in FuseML`,
		Run: func(cmd *cobra.Command, args []string) {
			o.Labels.Unpack()
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.ID, "id", "i", "", "ID value or regular expression used to filter runnables")
	cmd.Flags().StringVarP(&o.Kind, "kind", "k", "", "kind value or regular expression used to filter runnables")
	cmd.Flags().StringSliceVar(&o.Labels.Packed, "label", []string{}, "label value or regular expression used to filter runnables. One or more may be supplied.")
	o.format.AddMultiValueFormattingFlags(cmd)

	return cmd
}

func (o *ListOptions) validate() error {
	return nil
}

func (o *ListOptions) run() error {
	request, err := runnablec.BuildListPayload(o.ID, o.Kind, o.Labels.Unpacked)
	if err != nil {
		return err
	}

	response, err := o.RunnableClient.List()(context.Background(), request)
	if err != nil {
		return err
	}

	o.format.FormatValue(os.Stdout, response)

	return nil
}
