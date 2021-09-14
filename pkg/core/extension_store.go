package core

import (
	"context"
	"time"

	"github.com/fuseml/fuseml-core/pkg/domain"
)

// ExtensionStore is an in memory store for extensions.
type ExtensionStore struct {
	// map of extensions indexed by ID
	items map[string]*domain.Extension
}

// NewExtensionStore returns an in-memory extension store instance.
func NewExtensionStore() *ExtensionStore {
	return &ExtensionStore{
		items: make(map[string]*domain.Extension),
	}
}

// AddExtension adds a new extension to the store.
func (store *ExtensionStore) AddExtension(ctx context.Context, extension *domain.Extension) (*domain.Extension, error) {
	if store.items[extension.ID] != nil {
		return nil, domain.NewErrExtensionExists(extension.ID)
	}

	extension.EnsureID(ctx, store)
	extension.SetCreated(ctx)
	store.items[extension.ID] = extension
	return extension, nil
}

// GetExtension retrieves an extension by its ID.
func (store *ExtensionStore) GetExtension(ctx context.Context, extensionID string) (*domain.Extension, error) {
	extension := store.items[extensionID]
	if extension == nil {
		return nil, domain.NewErrExtensionNotFound(extensionID)
	}
	return extension, nil
}

// ListExtensions retrieves all stored extensions.
func (store *ExtensionStore) ListExtensions(ctx context.Context, query *domain.ExtensionQuery) (result []*domain.Extension) {
	result = make([]*domain.Extension, 0, len(store.items))

	if query != nil {
		if query.ExtensionID != "" {
			fullExtension, err := store.GetExtension(ctx, query.ExtensionID)
			if err == nil {
				matchingExtension := fullExtension.GetExtensionIfMatch(query)
				if matchingExtension != nil {
					result = append(result, matchingExtension)
				}
			}
			return
		}

		for _, extension := range store.items {
			matchingExtension := extension.GetExtensionIfMatch(query)
			if matchingExtension != nil {
				result = append(result, matchingExtension)
			}
		}
		return
	}

	for _, extension := range store.items {
		result = append(result, extension)
	}
	return
}

// UpdateExtension updates an existing extension.
func (store *ExtensionStore) UpdateExtension(ctx context.Context, newExtension *domain.Extension) error {
	extension, err := store.GetExtension(ctx, newExtension.ID)
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

	store.items[newExtension.ID] = newExtension
	return nil
}

// DeleteExtension deletes an extension from the store.
func (store *ExtensionStore) DeleteExtension(ctx context.Context, extensionID string) error {
	_, err := store.GetExtension(ctx, extensionID)
	if err != nil {
		return err
	}
	delete(store.items, extensionID)
	return nil
}

// AddExtensionService adds a new extension service to an extension.
func (store *ExtensionStore) AddExtensionService(ctx context.Context, extensionID string, service *domain.ExtensionService) (*domain.ExtensionService, error) {
	extension, err := store.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	return extension.AddService(service)
}

// GetExtensionService retrieves an extension service by its ID.
func (store *ExtensionStore) GetExtensionService(ctx context.Context, extensionID string, serviceID string) (*domain.ExtensionService, error) {
	extension, err := store.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	return extension.GetService(serviceID)
}

// ListExtensionServices retrieves all services belonging to an extension.
func (store *ExtensionStore) ListExtensionServices(ctx context.Context, extensionID string) ([]*domain.ExtensionService, error) {
	extension, err := store.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	return extension.ListServices(), nil
}

// UpdateExtensionService updates a service belonging to an extension.
func (store *ExtensionStore) UpdateExtensionService(ctx context.Context, extensionID string, newService *domain.ExtensionService) error {
	extension, err := store.GetExtension(ctx, extensionID)
	if err != nil {
		return err
	}
	return extension.UpdateService(newService)
}

// DeleteExtensionService deletes an extension service from an extension.
func (store *ExtensionStore) DeleteExtensionService(ctx context.Context, extensionID, serviceID string) error {
	extension, err := store.GetExtension(ctx, extensionID)
	if err != nil {
		return err
	}
	return extension.DeleteService(serviceID)
}

// AddExtensionServiceEndpoint adds a new endpoint to an extension service.
func (store *ExtensionStore) AddExtensionServiceEndpoint(ctx context.Context, extensionID string, serviceID string, endpoint *domain.ExtensionServiceEndpoint) (*domain.ExtensionServiceEndpoint, error) {
	extension, err := store.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	return extension.AddEndpoint(serviceID, endpoint)
}

// GetExtensionServiceEndpoint retrieves an extension endpoint by its ID.
func (store *ExtensionStore) GetExtensionServiceEndpoint(ctx context.Context, extensionID string, serviceID string, endpointID string) (*domain.ExtensionServiceEndpoint, error) {
	extension, err := store.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	return extension.GetServiceEndpoint(serviceID, endpointID)
}

// ListExtensionServiceEndpoints retrieves all endpoints belonging to an extension service.
func (store *ExtensionStore) ListExtensionServiceEndpoints(ctx context.Context, extensionID string, serviceID string) ([]*domain.ExtensionServiceEndpoint, error) {
	extension, err := store.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	return extension.ListServiceEndpoints(serviceID)
}

// UpdateExtensionServiceEndpoint updates an endpoint belonging to an extension service.
func (store *ExtensionStore) UpdateExtensionServiceEndpoint(ctx context.Context, extensionID string, serviceID string, endpoint *domain.ExtensionServiceEndpoint) error {
	extension, err := store.GetExtension(ctx, extensionID)
	if err != nil {
		return err
	}
	return extension.UpdateServiceEndpoint(serviceID, endpoint)
}

// DeleteExtensionServiceEndpoint deletes an extension endpoint from an extension service.
func (store *ExtensionStore) DeleteExtensionServiceEndpoint(ctx context.Context, extensionID, serviceID, endpointID string) error {
	extension, err := store.GetExtension(ctx, extensionID)
	if err != nil {
		return err
	}
	return extension.DeleteServiceEndpoint(serviceID, endpointID)
}

// AddExtensionServiceCredentials adds a new credential to an extension service.
func (store *ExtensionStore) AddExtensionServiceCredentials(ctx context.Context, extensionID string, serviceID string,
	credentials *domain.ExtensionServiceCredentials) (*domain.ExtensionServiceCredentials, error) {
	extension, err := store.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	return extension.AddCredentials(serviceID, credentials)
}

// GetExtensionServiceCredentials retrieves an extension credential by its ID.
func (store *ExtensionStore) GetExtensionServiceCredentials(ctx context.Context, extensionID string, serviceID string, credentialsID string) (*domain.ExtensionServiceCredentials, error) {
	extension, err := store.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	return extension.GetServiceCredentials(serviceID, credentialsID)
}

// ListExtensionServiceCredentials retrieves all credentials belonging to an extension service.
func (store *ExtensionStore) ListExtensionServiceCredentials(ctx context.Context, extensionID string, serviceID string) ([]*domain.ExtensionServiceCredentials, error) {
	extension, err := store.GetExtension(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	return extension.ListServiceCredentials(serviceID)
}

// UpdateExtensionServiceCredentials updates an extension credential.
func (store *ExtensionStore) UpdateExtensionServiceCredentials(ctx context.Context, extensionID string, serviceID string, credentials *domain.ExtensionServiceCredentials) (err error) {
	extension, err := store.GetExtension(ctx, extensionID)
	if err != nil {
		return err
	}
	return extension.UpdateServiceCredentials(serviceID, credentials)
}

// DeleteExtensionServiceCredentials deletes an extension credential from an extension service.
func (store *ExtensionStore) DeleteExtensionServiceCredentials(ctx context.Context, extensionID, serviceID, credentialsID string) error {
	extension, err := store.GetExtension(ctx, extensionID)
	if err != nil {
		return err
	}
	return extension.DeleteServiceCredentials(serviceID, credentialsID)
}

// GetExtensionAccessDescriptors retrieves access descriptors belonging to an extension that matches the query.
func (store *ExtensionStore) GetExtensionAccessDescriptors(ctx context.Context, query *domain.ExtensionQuery) (result []*domain.ExtensionAccessDescriptor, err error) {
	result = make([]*domain.ExtensionAccessDescriptor, 0)

	for _, extension := range store.ListExtensions(ctx, query) {
		result = append(result, extension.GetAccessDescriptors()...)
	}
	return result, nil
}
