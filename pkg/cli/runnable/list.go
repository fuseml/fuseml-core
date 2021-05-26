package runnable

import (
	"context"
	"fmt"
	"os"
	"strings"

	runnablec "github.com/fuseml/fuseml-core/gen/http/runnable/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// RunnableListOptions holds the options for 'runnable list' sub command
type RunnableListOptions struct {
	common.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
	ID     string
	Kind   string
	Labels map[string]string
}

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

// NewRunnableListOptionsOptions creates a RunnableListOptions struct
func NewRunnableListOptions(o *common.GlobalOptions) (res *RunnableListOptions) {
	res = &RunnableListOptions{global: o}
	res.format = common.NewFormattingOptions(
		[]string{"ID", "Kind", "Description", "Labels"},
		common.OutputSortFields{"ID": table.Asc},
		common.OutputFormatters{"Labels": formatLabels},
	)

	return
}

func NewSubCmdRunnableList(c *common.GlobalOptions) *cobra.Command {

	o := NewRunnableListOptions(c)
	// local variable used to collect the label arguments and then unpack them
	var labels []string

	cmd := &cobra.Command{
		Use:   "list [-i|--id ID] [-k|--kind KIND] [-l|--label LABEL_KEY:LABEL_VALUE]...",
		Short: "List runnables.",
		Long:  `Retrieve information about Runnables registered in FuseML`,
		Run: func(cmd *cobra.Command, args []string) {
			common.UnpackLabelArgs(labels, o.Labels)
			common.CheckErr(o.InitializeClients(c))
			common.CheckErr(o.Validate())
			common.CheckErr(o.Run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVarP(&o.ID, "id", "i", "", "ID value or regular expression used to filter runnables")
	cmd.Flags().StringVarP(&o.Kind, "kind", "k", "", "kind value or regular expression used to filter runnables")
	cmd.Flags().StringSliceVar(&labels, "label", []string{}, "label value or regular expression used to filter runnables. One or more may be supplied.")
	o.format.AddMultiValueFormattingFlags(cmd)

	return cmd
}

func (o *RunnableListOptions) Validate() error {
	return nil
}

func (o *RunnableListOptions) Run() error {
	request, err := runnablec.BuildListPayload(o.ID, o.Kind, o.Labels)
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
