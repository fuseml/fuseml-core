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

// ListExtensions - list all registered extensions that match the supplied query parameters
func (registry *ExtensionRegistry) ListExtensions(ctx context.Context, query *domain.ExtensionQuery) (result []*domain.ExtensionRecord, err error) {
	if query == nil {
		return registry.extensionStore.GetAllExtensions(ctx)
	}
	return registry.extensionStore.RunExtensionQuery(ctx, query)
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

// UpdateExtension - update an extension
func (registry *ExtensionRegistry) UpdateExtension(ctx context.Context, extension *domain.Extension) (err error) {
	if extension.ID == "" {
		return domain.NewErrMissingField("extension", "extension ID")
	}
	return registry.extensionStore.UpdateExtension(ctx, extension)
}

// UpdateService - update a service belonging to an extension
func (registry *ExtensionRegistry) UpdateService(ctx context.Context, service *domain.ExtensionService) (err error) {
	if service.ExtensionID == "" {
		return domain.NewErrMissingField("service", "extension ID")
	}
	if service.ID == "" {
		return domain.NewErrMissingField("service", "service ID")
	}
	return registry.extensionStore.UpdateService(ctx, service)
}

// UpdateEndpoint - update an endpoint belonging to a service
func (registry *ExtensionRegistry) UpdateEndpoint(ctx context.Context, endpoint *domain.ExtensionEndpoint) (err error) {
	if endpoint.ExtensionID == "" {
		return domain.NewErrMissingField("endpoint", "extension ID")
	}
	if endpoint.ServiceID == "" {
		return domain.NewErrMissingField("endpoint", "service ID")
	}
	if endpoint.URL == "" {
		return domain.NewErrMissingField("endpoint", "URL")
	}
	return registry.extensionStore.UpdateEndpoint(ctx, endpoint)
}

// UpdateCredentials - update a set of credentials belonging to a service
func (registry *ExtensionRegistry) UpdateCredentials(ctx context.Context, credentials *domain.ExtensionCredentials) (err error) {
	if credentials.ExtensionID == "" {
		return domain.NewErrMissingField("credentials", "extension ID")
	}
	if credentials.ServiceID == "" {
		return domain.NewErrMissingField("credentials", "service ID")
	}
	if credentials.ID == "" {
		return domain.NewErrMissingField("credentials", "credentials ID")
	}
	return registry.extensionStore.UpdateCredentials(ctx, credentials)
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
							Extension:   extension.Extension,
							Service:     service.ExtensionService,
							Endpoint:    *endpoint,
							Credentials: credentials,
						}
						result = append(result, &accessDesc)
					}
				} else {
					accessDesc := domain.ExtensionAccessDescriptor{
						Extension:   extension.Extension,
						Service:     service.ExtensionService,
						Endpoint:    *endpoint,
						Credentials: nil,
					}
					result = append(result, &accessDesc)
				}
			}
		}
	}

	sort.Sort(byID{result})
	return result, nil
}
