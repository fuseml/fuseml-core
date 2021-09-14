package badger

import (
	"context"
	"time"

	"github.com/fuseml/fuseml-core/pkg/domain"
	"github.com/timshannon/badgerhold/v3"
)

// ExtensionStore is a wrapper around a badgerhold.Store that implements the domain.ExtensionStore interface.
type ExtensionStore struct {
	store *badgerhold.Store
}

// NewExtensionStore creates a new ExtensionStore.
func NewExtensionStore(store *badgerhold.Store) *ExtensionStore {
	return &ExtensionStore{store: store}
}

// AddExtension adds a new extension to the store.
func (es *ExtensionStore) AddExtension(ctx context.Context, extension *domain.Extension) (*domain.Extension, error) {
	extension.EnsureID(ctx, es)
	extension.SetCreated(ctx)

	err := es.store.Insert(extension.ID, extension)
	if err != nil {
		return nil, domain.NewErrExtensionExists(extension.ID)
	}
	return extension, nil
}

// GetExtension retrieves an extension by its ID.
func (es *ExtensionStore) GetExtension(ctx context.Context, extensionID string) (*domain.Extension, error) {
	extension := &domain.Extension{}
	err := es.store.Get(extensionID, extension)
	if err != nil {
		return nil, domain.NewErrExtensionNotFound(extensionID)
	}
	return extension, nil
}

// ListExtensions retrieves all stored extensions.
func (es *ExtensionStore) ListExtensions(ctx context.Context, query *domain.ExtensionQuery) (result []*domain.Extension) {
	result = []*domain.Extension{}

	// TODO: Replace with a badgerhold query.
	if query != nil {
		if query.ExtensionID != "" {
			fullExtension, err := es.GetExtension(ctx, query.ExtensionID)
			if err == nil {
				matchingExtension := fullExtension.GetExtensionIfMatch(query)
				if matchingExtension != nil {
					result = append(result, matchingExtension)
				}
			}
			return
		}

		allExtensions := []*domain.Extension{}
		es.store.Find(&allExtensions, nil)

		for _, extension := range allExtensions {
			matchingExtension := extension.GetExtensionIfMatch(query)
			if matchingExtension != nil {
				result = append(result, matchingExtension)
			}
		}
		return
	}

	es.store.Find(&result, nil)
	return
}

// UpdateExtension updates an existing extension.
func (es *ExtensionStore) UpdateExtension(ctx context.Context, newExtension *domain.Extension) error {
	extension, err := es.GetExtension(ctx, newExtension.ID)
	if err != nil {
		return err
	}
	newExtension.Created = extension.Created
	newExtension.Updated = time.Now()

	for _, newExtService := range newExtension.ListServices() {
		_, err := extension.GetService(newExtService.ID)
		if err != nil {
			// If the service is new, set the creation time
			newExtService.SetCreated(newExtension.Updated)
		}
	}

	err = es.store.Update(newExtension.ID, newExtension)
	if err != nil {
		return domain.NewErrExtensionNotFound(newExtension.ID)
	}
	return nil
}

// DeleteExtension deletes an extension from the store.
func (es *ExtensionStore) DeleteExtension(ctx context.Context, extensionID string) error {
	extension, err := es.GetExtension(ctx, extensionID)
	if err != nil {
		return err
	}
	return es.store.Delete(extension.ID, extension)
}

// AddExtensionService adds a new extension service to an extension.
func (es *ExtensionStore) AddExtensionService(ctx context.Context, extensionID string, service *domain.ExtensionService) (*domain.ExtensionService, error) {
	extension, err := es.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	svc, err := extension.AddService(service)
	if err != nil {
		return nil, err
	}
	err = es.UpdateExtension(ctx, extension)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

// GetExtensionService retrieves an extension service by its ID.
func (es *ExtensionStore) GetExtensionService(ctx context.Context, extensionID string, serviceID string) (*domain.ExtensionService, error) {
	extension, err := es.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	return extension.GetService(serviceID)
}

// ListExtensionServices retrieves all services belonging to an extension.
func (es *ExtensionStore) ListExtensionServices(ctx context.Context, extensionID string) ([]*domain.ExtensionService, error) {
	extension, err := es.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	return extension.ListServices(), nil
}

// UpdateExtensionService updates a service belonging to an extension.
func (es *ExtensionStore) UpdateExtensionService(ctx context.Context, extensionID string, newService *domain.ExtensionService) error {
	extension, err := es.GetExtension(ctx, extensionID)
	if err != nil {
		return err
	}
	err = extension.UpdateService(newService)
	if err != nil {
		return err
	}
	return es.UpdateExtension(ctx, extension)
}

// DeleteExtensionService deletes an extension service from an extension.
func (es *ExtensionStore) DeleteExtensionService(ctx context.Context, extensionID string, serviceID string) error {
	extension, err := es.GetExtension(ctx, extensionID)
	if err != nil {
		return err
	}
	err = extension.DeleteService(serviceID)
	if err != nil {
		return err
	}
	return es.UpdateExtension(ctx, extension)
}

// AddExtensionServiceEndpoint adds a new endpoint to an extension service.
func (es *ExtensionStore) AddExtensionServiceEndpoint(ctx context.Context, extensionID string, serviceID string, endpoint *domain.ExtensionServiceEndpoint) (*domain.ExtensionServiceEndpoint, error) {
	extension, err := es.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	svc, err := extension.GetService(serviceID)
	if err != nil {
		return nil, err
	}
	endpoint, err = svc.AddEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	err = es.UpdateExtension(ctx, extension)
	if err != nil {
		return nil, err
	}
	return endpoint, nil
}

// GetExtensionServiceEndpoint retrieves an extension endpoint by its ID.
func (es *ExtensionStore) GetExtensionServiceEndpoint(ctx context.Context, extensionID string, serviceID string, endpointID string) (*domain.ExtensionServiceEndpoint, error) {
	extension, err := es.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	svc, err := extension.GetService(serviceID)
	if err != nil {
		return nil, err
	}
	return svc.GetEndpoint(endpointID)
}

// ListExtensionServiceEndpoints retrieves all endpoints belonging to an extension service.
func (es *ExtensionStore) ListExtensionServiceEndpoints(ctx context.Context, extensionID string, serviceID string) ([]*domain.ExtensionServiceEndpoint, error) {
	extension, err := es.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	svc, err := extension.GetService(serviceID)
	if err != nil {
		return nil, err
	}
	return svc.ListEndpoints(), nil
}

// UpdateExtensionServiceEndpoint updates an endpoint belonging to an extension service.
func (es *ExtensionStore) UpdateExtensionServiceEndpoint(ctx context.Context, extensionID string, serviceID string, newEndpoint *domain.ExtensionServiceEndpoint) error {
	extension, err := es.GetExtension(ctx, extensionID)
	if err != nil {
		return err
	}
	svc, err := extension.GetService(serviceID)
	if err != nil {
		return err
	}
	err = svc.UpdateEndpoint(newEndpoint)
	if err != nil {
		return err
	}
	return es.UpdateExtension(ctx, extension)
}

// DeleteExtensionServiceEndpoint deletes an extension endpoint from an extension service.
func (es *ExtensionStore) DeleteExtensionServiceEndpoint(ctx context.Context, extensionID string, serviceID string, endpointID string) error {
	extension, err := es.GetExtension(ctx, extensionID)
	if err != nil {
		return err
	}
	svc, err := extension.GetService(serviceID)
	if err != nil {
		return err
	}
	err = svc.DeleteEndpoint(endpointID)
	if err != nil {
		return err
	}
	return es.UpdateExtension(ctx, extension)
}

// AddExtensionServiceCredentials adds a new credential to an extension service.
func (es *ExtensionStore) AddExtensionServiceCredentials(ctx context.Context, extensionID string, serviceID string, credentials *domain.ExtensionServiceCredentials) (*domain.ExtensionServiceCredentials, error) {
	extension, err := es.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	svc, err := extension.GetService(serviceID)
	if err != nil {
		return nil, err
	}
	credentials, err = svc.AddCredentials(credentials)
	if err != nil {
		return nil, err
	}
	err = es.UpdateExtension(ctx, extension)
	if err != nil {
		return nil, err
	}
	return credentials, nil
}

// GetExtensionServiceCredentials retrieves an extension credential by its ID.
func (es *ExtensionStore) GetExtensionServiceCredentials(ctx context.Context, extensionID string, serviceID string, credentialsID string) (*domain.ExtensionServiceCredentials, error) {
	extension, err := es.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	svc, err := extension.GetService(serviceID)
	if err != nil {
		return nil, err
	}
	return svc.GetCredentials(credentialsID)
}

// ListExtensionServiceCredentials retrieves all credentials belonging to an extension service.
func (es *ExtensionStore) ListExtensionServiceCredentials(ctx context.Context, extensionID string, serviceID string) ([]*domain.ExtensionServiceCredentials, error) {
	extension, err := es.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	svc, err := extension.GetService(serviceID)
	if err != nil {
		return nil, err
	}
	return svc.ListCredentials(), nil
}

// UpdateExtensionServiceCredentials updates an extension credential.
func (es *ExtensionStore) UpdateExtensionServiceCredentials(ctx context.Context, extensionID string, serviceID string, newCredentials *domain.ExtensionServiceCredentials) error {
	extension, err := es.GetExtension(ctx, extensionID)
	if err != nil {
		return err
	}
	svc, err := extension.GetService(serviceID)
	if err != nil {
		return err
	}
	err = svc.UpdateCredentials(newCredentials)
	if err != nil {
		return err
	}
	return es.UpdateExtension(ctx, extension)
}

// DeleteExtensionServiceCredentials deletes an extension credential from an extension service.
func (es *ExtensionStore) DeleteExtensionServiceCredentials(ctx context.Context, extensionID string, serviceID string, credentialsID string) error {
	extension, err := es.GetExtension(ctx, extensionID)
	if err != nil {
		return err
	}
	svc, err := extension.GetService(serviceID)
	if err != nil {
		return err
	}
	err = svc.DeleteCredentials(credentialsID)
	if err != nil {
		return err
	}
	return es.UpdateExtension(ctx, extension)
}

// GetExtensionAccessDescriptors retrieves access descriptors belonging to an extension that matches the query.
func (es *ExtensionStore) GetExtensionAccessDescriptors(ctx context.Context, query *domain.ExtensionQuery) (result []*domain.ExtensionAccessDescriptor, err error) {
	result = make([]*domain.ExtensionAccessDescriptor, 0)

	for _, extension := range es.ListExtensions(ctx, query) {
		result = append(result, extension.GetAccessDescriptors()...)
	}
	return result, nil
}
