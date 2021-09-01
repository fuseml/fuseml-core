package extension

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/fuseml/fuseml-core/gen/extension"
	"github.com/fuseml/fuseml-core/pkg/cli/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

type extensionListOptions struct {
	client.Clients
	global *common.GlobalOptions
	format *common.FormattingOptions
	query  extension.ExtensionQuery
}

func newExtensionListOptions(o *common.GlobalOptions) (res *extensionListOptions) {
	res = &extensionListOptions{global: o}
	res.format = common.NewFormattingOptions(
		[]string{"ID", "Product", "Version", "Zone", "Services", "Endpoints", "Credentials"},
		[]table.SortBy{{Name: "ID", Mode: table.Asc}},
		common.OutputFormatters{"Services": formatServices, "Endpoints": formatEndpoints, "Credentials": formatCredentials},
	)

	return
}

func derefString(s *string, defaultValue ...string) string {
	ds := ""
	if len(defaultValue) > 0 {
		ds = defaultValue[0]
	}
	if s != nil {
		return *s
	}
	return ds
}

func formatServices(object interface{}, column string, field interface{}) (formated string) {
	if ext, ok := object.(*extension.Extension); ok {
		for _, svc := range ext.Services {
			formated += fmt.Sprintf(`[ %s ]
resource: %s
category: %s

`, derefString(svc.ID), derefString(svc.Resource, "N/A"), derefString(svc.Category, "N/A"))
		}
	}
	return
}

func formatEndpoints(object interface{}, column string, field interface{}) (formated string) {
	if ext, ok := object.(*extension.Extension); ok {
		for _, svc := range ext.Services {
			if len(svc.Endpoints) == 0 {
				continue
			}
			formated += fmt.Sprintf("[ %s ]\n", derefString(svc.ID))
			for _, ep := range svc.Endpoints {
				formated += fmt.Sprintf("%s: %s\n", derefString(ep.Type, "external"), derefString(ep.URL))
			}
			formated += "\n"
		}
	}
	return
}

func formatCredentials(object interface{}, column string, field interface{}) (formated string) {
	if ext, ok := object.(*extension.Extension); ok {
		for _, svc := range ext.Services {
			if len(svc.Credentials) == 0 {
				continue
			}
			formated += fmt.Sprintf("[ %s ]\n", derefString(svc.ID))
			for _, cred := range svc.Credentials {
				projects := strings.Join(cred.Projects, ", ")
				users := strings.Join(cred.Users, ", ")
				scope := derefString(cred.Scope)
				formated += fmt.Sprintf("%s: %s\n", derefString(cred.ID), scope)
				if scope == "project" {
					formated += fmt.Sprintf("  projects: %s\n", projects)
				}
				if scope == "user" {
					formated += fmt.Sprintf("  users: %s\n", users)
				}
			}
			formated += "\n"
		}
	}
	return
}

func newSubCmdExtensionList(gOpt *common.GlobalOptions) *cobra.Command {
	o := newExtensionListOptions(gOpt)
	cmd := &cobra.Command{
		Use: `list [--id EXTENSION_ID] [--product|-p PRODUCT] [--version VERSION] 
[--zone|-z ZONE] [--service-id SERVICE_ID] [--service-resource|-r SERVICE_RESOURCE] 
[--service-category|-r SERVICE_CATEGORY]`,
		Short: "Lists one or more extensions",
		Long:  `Display information about registered extensions matching supplied criteria.`,
		Run: func(cmd *cobra.Command, args []string) {
			common.CheckErr(o.InitializeClients(gOpt.URL, gOpt.Timeout, gOpt.Verbose))
			common.CheckErr(o.validate())
			common.CheckErr(o.run())
		},
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVar(&o.query.ExtensionID, "id", "", "match an extension by explicit extension ID")
	cmd.Flags().StringVarP(&o.query.Product, "product", "p", "",
		`match extensions by a universal product identifier. Product values can be used to identify
installations of the same product registered with the same or different FuseML servers`)
	cmd.Flags().StringVar(&o.query.Version, "version", "",
		"match extensions by version or by semantic version constraints")
	cmd.Flags().StringVarP(&o.query.Zone, "zone", "z", "",
		"return only extensions installed in a particular zone")
	cmd.Flags().StringVar(&o.query.ServiceID, "service-id", "",
		"Match extensions that provide services identified by an explicit service ID")
	cmd.Flags().StringVarP(&o.query.ServiceResource, "service-resource", "r", "",
		"match only extensions providing a particular API or protocol (e.g. s3, git, mlflow)")
	cmd.Flags().StringVarP(&o.query.ServiceCategory, "service-category", "c", "",
		`match only extensions providing one of the well-known categories of AI/ML services
(e.g. model store, feature store, distributed training, serving)`)
	o.format.AddMultiValueFormattingFlags(cmd)

	return cmd
}

func (o *extensionListOptions) validate() error {
	return nil
}

func (o *extensionListOptions) run() error {
	exts, err := o.ExtensionClient.ListExtension(&o.query)
	if err != nil {
		return err
	}

	o.format.FormatValue(os.Stdout, exts)

	return nil
}
