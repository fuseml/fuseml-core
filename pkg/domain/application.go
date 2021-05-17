package domain

import (
	"context"
)

// ApplicationStore is an inteface to application stores
type ApplicationStore interface {
	Find(context.Context, string) *Application
	GetAll(context.Context, *string, *string) ([]*Application, error)
	Add(context.Context, *Application) (*Application, error)
	Delete(context.Context, string) error
}

// Application holds the information about the application
type Application struct {
	// The name of the Application
	Name string
	// The type of the Application
	Type string
	// Application description
	Description string
	// The public URL for accessing the Application
	URL string
	// Name of the Workflow used to create Application
	Workflow string
	// Kubernetes resources describing the Application
	K8sResources []*KubernetesResource
	// Kubernetes namespace where the resources are located
	K8sNamespace string
}

// KubernetesResource describes the Kubernetes resource that forms the application
type KubernetesResource struct {
	// The name of the Kubernetes resource
	Name string
	// The kind of Kubernetes resource
	Kind string
}
