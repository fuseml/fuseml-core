package kubernetes

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// Cluster holds the config information for Kubernetes cluster
type Cluster struct {
	restConfig *rest.Config
	logger     *log.Logger
}

// GetClientConfig fetchs the kubernetes config of current cluster
func GetClientConfig() (*rest.Config, error) {
	if _, inCluster := os.LookupEnv("KUBERNETES_SERVICE_HOST"); inCluster {
		return rest.InClusterConfig()
	}
	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)
	if configEnv, fromEnv := os.LookupEnv("KUBECONFIG"); fromEnv {
		kubeconfig = configEnv
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

// NewCluster returns new cluster struct initialized with KUBECONFIG from environment
func NewCluster(logger *log.Logger) (*Cluster, error) {

	config, err := GetClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes client config: %w", err)
	}

	return &Cluster{config, logger}, nil
}

// DeleteResource deletes kuberneres resource from current cluster, identified by name, namespace and kind
func (c *Cluster) DeleteResource(ctx context.Context, name, namespace, kind string) error {
	c.logger.Printf("want to delete resource %s of kind %s in %s namespace", name, kind, namespace)
	// Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(c.restConfig)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	gvk, err := mapper.KindFor(schema.GroupVersionResource{Resource: kind})
	if err != nil {
		return err
	}

	// Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	// create the dynamic client
	dynClient, err := dynamic.NewForConfig(c.restConfig)
	if err != nil {
		return err
	}
	// get REST interface
	dr := dynClient.Resource(mapping.Resource).Namespace(namespace)

	err = dr.Delete(ctx, name, metav1.DeleteOptions{})
	if !k8serr.IsNotFound(err) {
		return err
	}
	c.logger.Print("resource not found, no need to delete")
	return nil
}
