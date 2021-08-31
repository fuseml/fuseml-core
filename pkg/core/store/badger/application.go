package badger

import (
	"context"

	"github.com/fuseml/fuseml-core/pkg/domain"
	"github.com/timshannon/badgerhold/v3"
)

// ApplicationStore is a wrapper around a badgerhold.Store that implements the domain.ApplicationStore interface.
type ApplicationStore struct {
	store *badgerhold.Store
}

// NewApplicationStore creates a new ApplicationStore.
func NewApplicationStore(store *badgerhold.Store) *ApplicationStore {
	return &ApplicationStore{store: store}
}

// Find returns a application identified by id
func (as *ApplicationStore) Find(ctx context.Context, name string) *domain.Application {
	app := domain.Application{}
	err := as.store.Get(name, &app)
	if err != nil {
		return nil
	}
	return &app
}

// GetAll returns all applications of a given type.
// If type is not specified, return all applications.
func (as *ApplicationStore) GetAll(ctx context.Context, applicationType *string, applicationWorkflow *string) ([]*domain.Application, error) {
	result := []*domain.Application{}
	query := &badgerhold.Query{}

	if applicationType != nil && applicationWorkflow != nil {
		query = badgerhold.Where("Type").Eq(*applicationType).And("Workflow").Eq(*applicationWorkflow)
	} else if applicationType != nil {
		query = badgerhold.Where("Type").Eq(*applicationType)
	} else if applicationWorkflow != nil {
		query = badgerhold.Where("Workflow").Eq(*applicationWorkflow)
	}

	err := as.store.Find(&result, query)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Add adds a new application, based on the Application structure provided as argument
func (as *ApplicationStore) Add(ctx context.Context, a *domain.Application) (*domain.Application, error) {
	// TODO What if such application already exist? Should we thrown an error, or silently replace it?
	err := as.store.Insert(a.Name, a)
	if err != nil {
		as.store.Delete(a.Name, a)
		err := as.store.Insert(a.Name, a)
		if err != nil {
			return nil, err
		}
	}
	return a, nil
}

// Delete deletes the application registered by FuseML
func (as *ApplicationStore) Delete(ctx context.Context, name string) error {
	a := domain.Application{}
	return as.store.Delete(name, a)
}
