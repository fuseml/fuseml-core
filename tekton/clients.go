package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1beta1"
	"k8s.io/client-go/tools/clientcmd"
)

// clients holds instances of interfaces for making requests to the Pipeline controllers.
type clients struct {
	PipelineClient v1beta1.PipelineInterface
	TaskClient     v1beta1.TaskInterface
}

// newClients instantiates and returns several clientsets required for making requests to the
// Pipeline cluster. Clients can make requests within namespace.
func newClients(namespace string) *clients {
	var err error
	c := &clients{}

	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	cs, err := versioned.NewForConfig(cfg)
	if err != nil {
		panic("failed to create pipeline clientset from config file")
	}
	c.PipelineClient = cs.TektonV1beta1().Pipelines(namespace)
	c.TaskClient = cs.TektonV1beta1().Tasks(namespace)
	return c
}
