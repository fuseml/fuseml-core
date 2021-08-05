package manager

import (
	"context"

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
func (registry *ExtensionRegistry) RegisterExtension(ctx context.Context, extension *domain.ExtensionRecord) (result *domain.ExtensionRecord, err error) {
	return registry.extensionStore.StoreExtension(ctx, extension)
}

// AddService - add a service to an existing extension
func (registry *ExtensionRegistry) AddService(ctx context.Context, service *domain.ExtensionServiceRecord) (result *domain.ExtensionServiceRecord, err error) {
	if service.ExtensionID == "" {
		return nil, domain.NewErrMissingField("service", "extension ID")
	}
	return registry.extensionStore.StoreService(ctx, service)
}

// AddEndpoint - add an endpoint to an existing extension service
func (registry *ExtensionRegistry) AddEndpoint(
	ctx context.Context, endpoint *domain.ExtensionEndpoint) (result *domain.ExtensionEndpoint, err error) {

	if endpoint.ExtensionID == "" {
		return nil, domain.NewErrMissingField("endpoint", "extension ID")
	}
	if endpoint.ServiceID == "" {
		return nil, domain.NewErrMissingField("endpoint", "service ID")
	}
	if endpoint.URL == "" {
		return nil, domain.NewErrMissingField("endpoint", "URL")
	}
	return registry.extensionStore.StoreEndpoint(ctx, endpoint)
}

// AddCredentials - add a set of credentials to an existing extension service
func (registry *ExtensionRegistry) AddCredentials(
	ctx context.Context, credentials *domain.ExtensionCredentials) (result *domain.ExtensionCredentials, err error) {

	if credentials.ExtensionID == "" {
		return nil, domain.NewErrMissingField("credentials", "extension ID")
	}
	if credentials.ServiceID == "" {
		return nil, domain.NewErrMissingField("credentials", "service ID")
	}

	return registry.extensionStore.StoreCredentials(ctx, credentials)
}

// GetExtension - retrieve an extension by ID and, optionally, its entire service/endpoint/credentials subtree
func (registry *ExtensionRegistry) GetExtension(ctx context.Context, extensionID string, fullTree bool) (result *domain.ExtensionRecord, err error) {
	return registry.extensionStore.GetExtension(ctx, extensionID, fullTree)
}

// GetService - retrieve an extension service by ID and, optionally, its entire endpoint/credentials subtree
func (registry *ExtensionRegistry) GetService(ctx context.Context, serviceID domain.ExtensionServiceID, fullTree bool) (result *domain.ExtensionServiceRecord, err error) {
	return registry.extensionStore.GetService(ctx, serviceID, fullTree)
}

// GetEndpoint - retrieve an extension endpoint by ID
func (registry *ExtensionRegistry) GetEndpoint(ctx context.Context, endpointID domain.ExtensionEndpointID) (result *domain.ExtensionEndpoint, err error) {
	return registry.extensionStore.GetEndpoint(ctx, endpointID)
}

// GetCredentials - retrieve a set of extension credentials by ID
func (registry *ExtensionRegistry) GetCredentials(ctx context.Context, credentialsID domain.ExtensionCredentialsID) (result *domain.ExtensionCredentials, err error) {
	return registry.extensionStore.GetCredentials(ctx, credentialsID)
}

// RemoveExtension - remove an extension from the registry
func (registry *ExtensionRegistry) RemoveExtension(ctx context.Context, extensionID string) error {
	return registry.extensionStore.DeleteExtension(ctx, extensionID)
}

// RemoveService - remove an extension service from the registry
func (registry *ExtensionRegistry) RemoveService(ctx context.Context, serviceID domain.ExtensionServiceID) error {
	return registry.extensionStore.DeleteService(ctx, serviceID)
}

// RemoveEndpoint - remove an extension endpoint from the registry
func (registry *ExtensionRegistry) RemoveEndpoint(ctx context.Context, endpointID domain.ExtensionEndpointID) error {
	return registry.extensionStore.DeleteEndpoint(ctx, endpointID)
}

// RemoveCredentials - remove a set of extension credentials from the registry
func (registry *ExtensionRegistry) RemoveCredentials(ctx context.Context, credentialsID domain.ExtensionCredentialsID) error {
	return registry.extensionStore.DeleteCredentials(ctx, credentialsID)
}

// RunExtensionAccessQuery - run a query on the extension registry to find one or more ways to access extensions matching given search parameters
func (registry *ExtensionRegistry) RunExtensionAccessQuery(ctx context.Context, query *domain.ExtensionQuery) (result []*domain.ExtensionAccessDescriptor, err error) {

	result = make([]*domain.ExtensionAccessDescriptor, 0)

	extensions, err := registry.extensionStore.RunExtensionQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	for _, extension := range extensions {
		for _, service := range extension.Services {
			for _, endpoint := range service.Endpoints {
				if len(service.Credentials) > 0 || service.AuthRequired {
					for _, credentials := range service.Credentials {
						accessDesc := domain.ExtensionAccessDescriptor{
							Extension:            extension.Extension,
							ExtensionService:     service.ExtensionService,
							ExtensionEndpoint:    *endpoint,
							ExtensionCredentials: credentials,
						}
						result = append(result, &accessDesc)
					}
				} else {
					accessDesc := domain.ExtensionAccessDescriptor{
						Extension:            extension.Extension,
						ExtensionService:     service.ExtensionService,
						ExtensionEndpoint:    *endpoint,
						ExtensionCredentials: nil,
					}
					result = append(result, &accessDesc)
				}
			}
		}
	}
	return result, nil
}
