package core

import (
	"context"

	"github.com/fuseml/fuseml-core/gen/application"
)

// ApplicationStore describes in memory store for applications
type ApplicationStore struct {
	items map[string]*application.Application
}

// NewApplicationStore returns an in-memory application store instance
func NewApplicationStore() *ApplicationStore {
	return &ApplicationStore{
		items: make(map[string]*application.Application),
	}
}

// Find returns a application identified by id
func (as *ApplicationStore) Find(ctx context.Context, name string) *application.Application {
	return as.items[name]
}

// GetAll returns all applications of a given type.
// If type is not specified, return all applications.
func (as *ApplicationStore) GetAll(ctx context.Context, applicationType *string) ([]*application.Application, error) {
	result := make([]*application.Application, 0, len(as.items))
	for _, app := range as.items {
		if applicationType == nil || app.Type == *applicationType {
			result = append(result, app)
		}
	}
	return result, nil
}

// Add adds a new application, based on the Application structure provided as argument
func (as *ApplicationStore) Add(ctx context.Context, a *application.Application) (*application.Application, error) {
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
