package core

import (
	"context"

	"github.com/fuseml/fuseml-core/pkg/domain"
)

// ApplicationStore describes in memory store for applications
type ApplicationStore struct {
	items map[string]*domain.Application
}

// NewApplicationStore returns an in-memory application store instance
func NewApplicationStore() *ApplicationStore {
	return &ApplicationStore{
		items: make(map[string]*domain.Application),
	}
}

// Find returns a application identified by id
func (as *ApplicationStore) Find(ctx context.Context, name string) *domain.Application {
	return as.items[name]
}

// GetAll returns all applications of a given type.
// If type is not specified, return all applications.
func (as *ApplicationStore) GetAll(ctx context.Context, applicationType *string, applicationWorkflow *string) ([]*domain.Application, error) {
	result := make([]*domain.Application, 0, len(as.items))
	for _, app := range as.items {
		if applicationWorkflow != nil && app.Workflow != *applicationWorkflow {
			continue
		}
		if applicationType != nil && app.Type != *applicationType {
			continue
		}
		result = append(result, app)
	}
	return result, nil
}

// Add adds a new application, based on the Application structure provided as argument
func (as *ApplicationStore) Add(ctx context.Context, a *domain.Application) (*domain.Application, error) {
	// TODO What if such application already exist? Should we thrown an error, or silently replace it?
	as.items[a.Name] = a
	return a, nil
}

// Delete deletes the application registered by FuseML
func (as *ApplicationStore) Delete(ctx context.Context, name string) error {
	if _, exists := as.items[name]; exists {
		delete(as.items, name)
	}
	return nil
}
