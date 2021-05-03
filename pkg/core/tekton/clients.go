package tekton

import (
	"fmt"
	"os"
	"path/filepath"

	pipelineclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1beta1"
	triggersclient "github.com/tektoncd/triggers/pkg/client/clientset/versioned"
	"github.com/tektoncd/triggers/pkg/client/clientset/versioned/typed/triggers/v1alpha1"
	restClient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Clients holds instances of interfaces for making requests to the tekton controllers.
type clients struct {
	PipelineClient        v1beta1.PipelineInterface
	PipelineRunClient     v1beta1.PipelineRunInterface
	TaskClient            v1beta1.TaskInterface
	TriggerTemplateClient v1alpha1.TriggerTemplateInterface
	TriggerBindingClient  v1alpha1.TriggerBindingInterface
	EventListenerClient   v1alpha1.EventListenerInterface
}

// NewClients instantiates and returns several clientsets required for making requests to
// tekton. Clients can make requests within namespace.
func newClients(namespace string) (*clients, error) {
	var err error
	c := &clients{}

	cfg, err := getClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes client config: %w", err)
	}

	cs, err := pipelineclient.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating tekton pipeline client set: %w", err)
	}
	c.PipelineClient = cs.TektonV1beta1().Pipelines(namespace)
	c.TaskClient = cs.TektonV1beta1().Tasks(namespace)
	c.PipelineRunClient = cs.TektonV1beta1().PipelineRuns(namespace)

	cst, err := triggersclient.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating tekton triggers client set: %w", err)
	}
	c.TriggerTemplateClient = cst.TriggersV1alpha1().TriggerTemplates(namespace)
	c.TriggerBindingClient = cst.TriggersV1alpha1().TriggerBindings(namespace)
	c.EventListenerClient = cst.TriggersV1alpha1().EventListeners(namespace)

	return c, nil
}

func getClientConfig() (*restClient.Config, error) {
	if _, inCluster := os.LookupEnv("KUBERNETES_SERVICE_HOST"); inCluster {
		return restClient.InClusterConfig()
	}
	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)
	if configEnv, fromEnv := os.LookupEnv("KUBECONFIG"); fromEnv {
		kubeconfig = configEnv
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}
