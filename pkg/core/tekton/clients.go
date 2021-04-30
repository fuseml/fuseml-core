package tekton

import (
	"log"
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
	TaskClient            v1beta1.TaskInterface
	TriggerTemplateClient v1alpha1.TriggerTemplateInterface
	TriggerBindingClient  v1alpha1.TriggerBindingInterface
	EventListenerClient   v1alpha1.EventListenerInterface
}

// NewClients instantiates and returns several clientsets required for making requests to
// tekton. Clients can make requests within namespace.
func newClients(namespace string) *clients {
	var err error
	c := &clients{}

	cfg, err := getClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	cs, err := pipelineclient.NewForConfig(cfg)
	if err != nil {
		panic("failed to create pipeline clientset from config file")
	}
	c.PipelineClient = cs.TektonV1beta1().Pipelines(namespace)
	c.TaskClient = cs.TektonV1beta1().Tasks(namespace)

	cst, err := triggersclient.NewForConfig(cfg)
	if err != nil {
		panic("failed to create triggers clientset from config file")
	}
	c.TriggerTemplateClient = cst.TriggersV1alpha1().TriggerTemplates(namespace)
	c.TriggerBindingClient = cst.TriggersV1alpha1().TriggerBindings(namespace)
	c.EventListenerClient = cst.TriggersV1alpha1().EventListeners(namespace)

	return c
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
