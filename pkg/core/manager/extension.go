package manager

import (
	"context"
	"sort"

	"github.com/fuseml/fuseml-core/pkg/domain"
)

// ExtensionRegistry implements the domain.ExtensionRegistry interface
type ExtensionRegistry struct {
	extensionStore domain.ExtensionStore
}

// NewExtensionRegistry initializes an extension registry
func NewExtensionRegistry(extensionStore domain.ExtensionStore) *ExtensionRegistry {
	return &ExtensionRegistry{extensionStore}
}

// RegisterExtension - register a new extension, with all participating services, endpoints and credentials
func (registry *ExtensionRegistry) RegisterExtension(ctx context.Context, extension *domain.Extension) (*domain.Extension, error) {
	return registry.extensionStore.AddExtension(ctx, extension)
}

// AddService - add a service to an existing extension
func (registry *ExtensionRegistry) AddService(ctx context.Context, extensionID string, service *domain.ExtensionService) (*domain.ExtensionService, error) {
	return registry.extensionStore.AddExtensionService(ctx, extensionID, service)
}

// AddEndpoint - add an endpoint to an existing extension service
func (registry *ExtensionRegistry) AddEndpoint(ctx context.Context, extensionID string, serviceID string,
	endpoint *domain.ExtensionServiceEndpoint) (*domain.ExtensionServiceEndpoint, error) {
	if endpoint.URL == "" {
		return nil, domain.NewErrMissingField("endpoint", "URL")
	}
	return registry.extensionStore.AddExtensionServiceEndpoint(ctx, extensionID, serviceID, endpoint)
}

// AddCredentials - add a set of credentials to an existing extension service
func (registry *ExtensionRegistry) AddCredentials(ctx context.Context, extensionID string, serviceID string,
	credentials *domain.ExtensionServiceCredentials) (*domain.ExtensionServiceCredentials, error) {
	return registry.extensionStore.AddExtensionServiceCredentials(ctx, extensionID, serviceID, credentials)
}

// ListExtensions - list all registered extensions that match the supplied query parameters
func (registry *ExtensionRegistry) ListExtensions(ctx context.Context, query *domain.ExtensionQuery) (result []*domain.Extension, err error) {
	return registry.extensionStore.ListExtensions(ctx, query), nil
}

// GetExtension - retrieve an extension by ID and, optionally, its entire service/endpoint/credentials subtree
func (registry *ExtensionRegistry) GetExtension(ctx context.Context, extensionID string) (*domain.Extension, error) {
	return registry.extensionStore.GetExtension(ctx, extensionID)
}

// GetService - retrieve an extension service by ID and, optionally, its entire endpoint/credentials subtree
func (registry *ExtensionRegistry) GetService(ctx context.Context, extensionID, serviceID string) (*domain.ExtensionService, error) {
	return registry.extensionStore.GetExtensionService(ctx, extensionID, serviceID)
}

// GetEndpoint - retrieve an extension endpoint by ID
func (registry *ExtensionRegistry) GetEndpoint(ctx context.Context, extensionID, serviceID, endpointURL string) (*domain.ExtensionServiceEndpoint, error) {
	return registry.extensionStore.GetExtensionServiceEndpoint(ctx, extensionID, serviceID, endpointURL)
}

// GetCredentials - retrieve a set of extension credentials by ID
func (registry *ExtensionRegistry) GetCredentials(ctx context.Context, extensionID, serviceID, credentialsID string) (*domain.ExtensionServiceCredentials, error) {
	return registry.extensionStore.GetExtensionServiceCredentials(ctx, extensionID, serviceID, credentialsID)
}

// UpdateExtension - update an extension
func (registry *ExtensionRegistry) UpdateExtension(ctx context.Context, extension *domain.Extension) error {
	if extension.ID == "" {
		return domain.NewErrMissingField("extension", "extension ID")
	}
	return registry.extensionStore.UpdateExtension(ctx, extension)
}

// UpdateService - update a service belonging to an extension
func (registry *ExtensionRegistry) UpdateService(ctx context.Context, extensionID string, service *domain.ExtensionService) error {
	if service.ID == "" {
		return domain.NewErrMissingField("service", "service ID")
	}
	return registry.extensionStore.UpdateExtensionService(ctx, extensionID, service)
}

// UpdateEndpoint - update an endpoint belonging to a service
func (registry *ExtensionRegistry) UpdateEndpoint(ctx context.Context, extensionID string, serviceID string, endpoint *domain.ExtensionServiceEndpoint) error {
	if endpoint.URL == "" {
		return domain.NewErrMissingField("endpoint", "URL")
	}
	return registry.extensionStore.UpdateExtensionServiceEndpoint(ctx, extensionID, serviceID, endpoint)
}

// UpdateCredentials - update a set of credentials belonging to a service
func (registry *ExtensionRegistry) UpdateCredentials(ctx context.Context, extensionID string, serviceID string, credentials *domain.ExtensionServiceCredentials) error {
	if credentials.ID == "" {
		return domain.NewErrMissingField("credentials", "credentials ID")
	}
	return registry.extensionStore.UpdateExtensionServiceCredentials(ctx, extensionID, serviceID, credentials)
}

// RemoveExtension - remove an extension from the registry
func (registry *ExtensionRegistry) RemoveExtension(ctx context.Context, extensionID string) error {
	return registry.extensionStore.DeleteExtension(ctx, extensionID)
}

// RemoveService - remove an extension service from the registry
func (registry *ExtensionRegistry) RemoveService(ctx context.Context, extensionID, serviceID string) error {
	return registry.extensionStore.DeleteExtensionService(ctx, extensionID, serviceID)
}

// RemoveEndpoint - remove an extension endpoint from the registry
func (registry *ExtensionRegistry) RemoveEndpoint(ctx context.Context, extensionID, serviceID, endpointID string) error {
	return registry.extensionStore.DeleteExtensionServiceEndpoint(ctx, extensionID, serviceID, endpointID)
}

// RemoveCredentials - remove a set of extension credentials from the registry
func (registry *ExtensionRegistry) RemoveCredentials(ctx context.Context, extensionID, serviceID, credentialsID string) error {
	return registry.extensionStore.DeleteExtensionServiceCredentials(ctx, extensionID, serviceID, credentialsID)
}

type queryResults []*domain.ExtensionAccessDescriptor

func (r queryResults) Len() int      { return len(r) }
func (r queryResults) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

type byID struct{ queryResults }

func (s byID) Less(i, j int) bool {
	if s.queryResults[i].Extension.ID != s.queryResults[j].Extension.ID {
		return s.queryResults[i].Extension.ID < s.queryResults[j].Extension.ID
	}
	if s.queryResults[i].Service.ID != s.queryResults[j].Service.ID {
		return s.queryResults[i].Service.ID < s.queryResults[j].Service.ID
	}
	if s.queryResults[i].Endpoint.URL != s.queryResults[j].Endpoint.URL {
		return s.queryResults[i].Endpoint.URL < s.queryResults[j].Endpoint.URL
	}
	if s.queryResults[i].Credentials == nil {
		return true
	}
	if s.queryResults[j].Credentials == nil {
		return false
	}
	return s.queryResults[i].Credentials.ID < s.queryResults[j].Credentials.ID
}

// GetExtensionAccessDescriptors - returns access descriptors for extensions that matches the query
func (registry *ExtensionRegistry) GetExtensionAccessDescriptors(ctx context.Context, query *domain.ExtensionQuery) (result []*domain.ExtensionAccessDescriptor, err error) {
	result, err = registry.extensionStore.GetExtensionAccessDescriptors(ctx, query)
	if err != nil {
		return nil, err
	}
	sort.Sort(byID{result})
	return result, nil
}
