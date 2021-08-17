package core

import (
	"context"
	"sort"
	"time"

	"github.com/Masterminds/semver"
	"github.com/fuseml/fuseml-core/pkg/domain"
	"k8s.io/apimachinery/pkg/util/rand"
)

// extensionRecord is the structure used to represent an extension in the extension store
type extensionRecord struct {
	domain.Extension
	// Map of services associated with the extension, indexed by ID
	services map[string]*extensionServiceRecord
}

type extensionServiceRecord struct {
	domain.ExtensionService
	// parent reference
	extension *extensionRecord
	// Map of endpoints associated with the service, indexed by URL
	endpoints map[string]*extensionEndpointRecord
	// Map of credentials associated with the service, indexed by ID
	credentials map[string]*extensionCredentialsRecord
}

type extensionEndpointRecord struct {
	domain.ExtensionEndpoint
	// parent reference
	service *extensionServiceRecord
}

type extensionCredentialsRecord struct {
	domain.ExtensionCredentials
	// parent reference
	service *extensionServiceRecord
}

// ExtensionStore is an in memory store for extensions
type ExtensionStore struct {
	// map of extensions indexed by ID
	items map[string]*extensionRecord
}

// NewExtensionStore returns an in-memory extension store instance
func NewExtensionStore() *ExtensionStore {
	return &ExtensionStore{
		items: make(map[string]*extensionRecord),
	}
}

// simple unique extension ID generator
func (store *ExtensionStore) generateExtensionID(extension *domain.Extension) string {
	prefix := extension.Product
	if prefix != "" {
		prefix = prefix + "-"
	}
	for {
		ID := prefix + rand.String(8)
		if store.items[ID] == nil {
			return ID
		}
	}
}

// simple unique extension service ID generator
func (store *ExtensionStore) generateExtensionServiceID(extension *extensionRecord, service *domain.ExtensionService) string {
	prefix := service.Resource
	if prefix == "" && extension.Product != "" {
		prefix = extension.Product + "-service"
	}
	if prefix != "" {
		prefix = prefix + "-"
	}
	for {
		ID := prefix + rand.String(8)
		if extension.services[ID] == nil {
			return ID
		}
	}
}

// simple unique extension credentials ID generator
func (store *ExtensionStore) generateExtensionCredentialsID(service *extensionServiceRecord, credentials *domain.ExtensionCredentials) string {
	prefix := service.Resource
	if prefix == "" {
		prefix = "creds"
	}
	if prefix != "" {
		prefix = prefix + "-"
	}
	for {
		ID := prefix + rand.String(8)
		if service.credentials[ID] == nil {
			return ID
		}
	}
}

// store an extension record
func (store *ExtensionStore) storeExtensionRecord(ctx context.Context, extension *domain.ExtensionRecord) (result *extensionRecord, err error) {
	if extension.ID == "" {
		extension.ID = store.generateExtensionID(&extension.Extension)
	}

	extension.Registered = time.Now()
	extension.Updated = extension.Registered

	// store a copy of the input extension
	extRecord := &extensionRecord{
		extension.Extension,
		make(map[string]*extensionServiceRecord),
	}
	store.items[extension.ID] = extRecord

	// next, store services
	for _, service := range extension.Services {
		service.ExtensionID = extension.ID
		_, err := store.storeServiceRecord(ctx, service, extRecord)
		if err != nil {
			return nil, err
		}
	}

	return extRecord, nil
}

// StoreExtension - store an extension, with all participating services, endpoints and credentials
func (store *ExtensionStore) StoreExtension(ctx context.Context, extension *domain.ExtensionRecord) (result *domain.ExtensionRecord, err error) {
	if store.items[extension.ID] != nil {
		return nil, domain.NewErrExtensionExists(extension.ID)
	}

	extRecord, err := store.storeExtensionRecord(ctx, extension)
	if err != nil {
		// rollback everything in case of error
		_ = store.deleteExtensionRecord(ctx, extRecord)
		return nil, err
	}
	return extension, nil
}

// store an extension service record
func (store *ExtensionStore) storeServiceRecord(
	ctx context.Context, service *domain.ExtensionServiceRecord, extRecord *extensionRecord) (result *extensionServiceRecord, err error) {

	if service.ID == "" {
		service.ID = store.generateExtensionServiceID(extRecord, &service.ExtensionService)
	}

	service.Registered = time.Now()
	service.Updated = service.Registered

	// store a copy of the input extension service
	svcRecord := &extensionServiceRecord{
		service.ExtensionService,
		extRecord,
		make(map[string]*extensionEndpointRecord),
		make(map[string]*extensionCredentialsRecord),
	}
	extRecord.services[service.ID] = svcRecord

	// next, store endpoints
	for _, endpoint := range service.Endpoints {
		endpoint.ExtensionID = extRecord.ID
		endpoint.ServiceID = service.ID
		_, err = store.storeEndpointRecord(ctx, endpoint, svcRecord)
		if err != nil {
			return nil, err
		}
	}

	// next, store credentials
	for _, credentials := range service.Credentials {
		credentials.ExtensionID = extRecord.ID
		credentials.ServiceID = service.ID
		_, err = store.storeCredentialsRecord(ctx, credentials, svcRecord)
		if err != nil {
			return nil, err
		}
	}

	return svcRecord, nil
}

// StoreService - store an extension service, with all participating endpoints and credentials
func (store *ExtensionStore) StoreService(
	ctx context.Context, service *domain.ExtensionServiceRecord) (result *domain.ExtensionServiceRecord, err error) {

	extRecord := store.items[service.ExtensionID]
	if extRecord == nil {
		return nil, domain.NewErrExtensionNotFound(service.ExtensionID)
	}

	if extRecord.services[service.ID] != nil {
		return nil, domain.NewErrExtensionServiceExists(extRecord.ID, service.ID)
	}

	svcRecord, err := store.storeServiceRecord(ctx, service, extRecord)
	if err != nil {
		// rollback everything in case of error
		_ = store.deleteServiceRecord(ctx, svcRecord)
		return nil, err
	}
	return service, nil

}

// store an extension endpoint record
func (store *ExtensionStore) storeEndpointRecord(
	ctx context.Context, endpoint *domain.ExtensionEndpoint, svcRecord *extensionServiceRecord) (result *extensionEndpointRecord, err error) {

	if svcRecord.endpoints[endpoint.URL] != nil {
		return nil, domain.NewErrExtensionEndpointExists(endpoint.ExtensionID, endpoint.ServiceID, endpoint.URL)
	}

	// store a copy of the input extension endpoint
	endpointRecord := &extensionEndpointRecord{
		*endpoint,
		svcRecord,
	}
	svcRecord.endpoints[endpoint.URL] = endpointRecord
	return endpointRecord, nil
}

// StoreEndpoint - store an extension endpoint
func (store *ExtensionStore) StoreEndpoint(ctx context.Context, endpoint *domain.ExtensionEndpoint) (result *domain.ExtensionEndpoint, err error) {
	extRecord := store.items[endpoint.ExtensionID]
	if extRecord == nil {
		return nil, domain.NewErrExtensionNotFound(endpoint.ExtensionID)
	}
	svcRecord := extRecord.services[endpoint.ServiceID]
	if svcRecord == nil {
		return nil, domain.NewErrExtensionServiceNotFound(endpoint.ExtensionID, endpoint.ServiceID)
	}

	_, err = store.storeEndpointRecord(ctx, endpoint, svcRecord)
	return endpoint, err
}

// store an extension credentials record
func (store *ExtensionStore) storeCredentialsRecord(
	ctx context.Context, credentials *domain.ExtensionCredentials, svcRecord *extensionServiceRecord) (result *extensionCredentialsRecord, err error) {

	if credentials.ID == "" {
		credentials.ID = store.generateExtensionCredentialsID(svcRecord, credentials)
	}

	if svcRecord.credentials[credentials.ID] != nil {
		return nil, domain.NewErrExtensionCredentialsExists(credentials.ExtensionID, credentials.ServiceID, credentials.ID)
	}

	credentials.Created = time.Now()
	credentials.Updated = credentials.Created

	// store a copy of the input extension credentials
	credsRecord := &extensionCredentialsRecord{
		*credentials,
		svcRecord,
	}
	svcRecord.credentials[credentials.ID] = credsRecord

	return credsRecord, nil
}

// StoreCredentials - store a set of extension credentials
func (store *ExtensionStore) StoreCredentials(ctx context.Context, credentials *domain.ExtensionCredentials) (result *domain.ExtensionCredentials, err error) {
	extRecord := store.items[credentials.ExtensionID]
	if extRecord == nil {
		return nil, domain.NewErrExtensionNotFound(credentials.ExtensionID)
	}
	svcRecord := extRecord.services[credentials.ServiceID]
	if svcRecord == nil {
		return nil, domain.NewErrExtensionServiceNotFound(credentials.ExtensionID, credentials.ServiceID)
	}

	_, err = store.storeCredentialsRecord(ctx, credentials, svcRecord)
	return credentials, err
}

// Retrieve an extension record by ID
func (store *ExtensionStore) getExtensionRecord(ctx context.Context, extensionID string) (result *extensionRecord, err error) {
	extRecord := store.items[extensionID]
	if extRecord == nil {
		return nil, domain.NewErrExtensionNotFound(extensionID)
	}
	return extRecord, nil
}

// GetAllExtensionss - retrieve all registered extensions, with all participating services, endpoints and credentials
func (store *ExtensionStore) GetAllExtensions(ctx context.Context) (result []*domain.ExtensionRecord, err error) {
	result = make([]*domain.ExtensionRecord, 0)
	for extID := range store.items {
		extRecord, err := store.GetExtension(ctx, extID, true)
		if err != nil {
			return nil, err
		}
		result = append(result, extRecord)
	}
	return result, nil
}

// GetExtension - retrieve an extension by ID
func (store *ExtensionStore) GetExtension(ctx context.Context, extensionID string, fullTree bool) (result *domain.ExtensionRecord, err error) {
	extRecord, err := store.getExtensionRecord(ctx, extensionID)
	if err != nil {
		return nil, err
	}

	result = &domain.ExtensionRecord{
		Extension: extRecord.Extension,
		Services:  make([]*domain.ExtensionServiceRecord, 0),
	}
	if !fullTree {
		return result, nil
	}

	// sort services by ID
	svcIDs := make([]string, 0, len(extRecord.services))
	for svcID := range extRecord.services {
		svcIDs = append(svcIDs, svcID)
	}
	sort.Strings(svcIDs)

	for _, svcID := range svcIDs {
		svcRecord := extRecord.services[svcID]
		service := domain.ExtensionServiceRecord{
			ExtensionService: svcRecord.ExtensionService,
			Endpoints:        make([]*domain.ExtensionEndpoint, 0),
			Credentials:      make([]*domain.ExtensionCredentials, 0),
		}
		result.Services = append(result.Services, &service)

		// sort endpoints by URL
		URLs := make([]string, 0, len(svcRecord.endpoints))
		for URL := range svcRecord.endpoints {
			URLs = append(URLs, URL)
		}
		sort.Strings(URLs)

		for _, URL := range URLs {
			endpointRecord := svcRecord.endpoints[URL]
			service.Endpoints = append(service.Endpoints, &endpointRecord.ExtensionEndpoint)
		}

		// sort credentials by ID
		credIDs := make([]string, 0, len(svcRecord.credentials))
		for credID := range svcRecord.credentials {
			credIDs = append(credIDs, credID)
		}
		sort.Strings(credIDs)

		for _, credID := range credIDs {
			credsRecord := svcRecord.credentials[credID]
			service.Credentials = append(service.Credentials, &credsRecord.ExtensionCredentials)
		}
	}
	return result, nil
}

// GetExtensionServices - retrieve the list of services belonging to an extension
func (store *ExtensionStore) GetExtensionServices(ctx context.Context, extensionID string) (result []*domain.ExtensionService, err error) {
	extRecord, err := store.getExtensionRecord(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	result = make([]*domain.ExtensionService, 0)
	for _, svcRecord := range extRecord.services {
		result = append(result, &svcRecord.ExtensionService)
	}
	return result, nil
}

// Retrieve an extension service record by ID
func (store *ExtensionStore) getServiceRecord(ctx context.Context, serviceID domain.ExtensionServiceID) (result *extensionServiceRecord, err error) {
	extRecord, err := store.getExtensionRecord(ctx, serviceID.ExtensionID)
	if err != nil {
		return nil, err
	}
	svcRecord := extRecord.services[serviceID.ID]
	if svcRecord == nil {
		return nil, domain.NewErrExtensionServiceNotFound(serviceID.ExtensionID, serviceID.ID)
	}
	return svcRecord, nil
}

// GetService - retrieve an extension service by ID
func (store *ExtensionStore) GetService(ctx context.Context, serviceID domain.ExtensionServiceID, fullTree bool) (result *domain.ExtensionServiceRecord, err error) {
	svcRecord, err := store.getServiceRecord(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	result = &domain.ExtensionServiceRecord{
		ExtensionService: svcRecord.ExtensionService,
		Endpoints:        make([]*domain.ExtensionEndpoint, 0),
		Credentials:      make([]*domain.ExtensionCredentials, 0),
	}

	if !fullTree {
		return result, nil
	}

	// sort endpoints by URL
	URLs := make([]string, 0, len(svcRecord.endpoints))
	for URL := range svcRecord.endpoints {
		URLs = append(URLs, URL)
	}
	sort.Strings(URLs)

	for _, URL := range URLs {
		endpointRecord := svcRecord.endpoints[URL]
		result.Endpoints = append(result.Endpoints, &endpointRecord.ExtensionEndpoint)
	}

	// sort credentials by ID
	credIDs := make([]string, 0, len(svcRecord.credentials))
	for credID := range svcRecord.credentials {
		credIDs = append(credIDs, credID)
	}
	sort.Strings(credIDs)

	for _, credID := range credIDs {
		credsRecord := svcRecord.credentials[credID]
		result.Credentials = append(result.Credentials, &credsRecord.ExtensionCredentials)
	}

	return result, nil
}

// GetServiceEndpoints - retrieve the list of endpoints belonging to an extension service
func (store *ExtensionStore) GetServiceEndpoints(ctx context.Context, serviceID domain.ExtensionServiceID) (result []*domain.ExtensionEndpoint, err error) {
	svcRecord, err := store.getServiceRecord(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	result = make([]*domain.ExtensionEndpoint, 0)
	for _, endpointRecord := range svcRecord.endpoints {
		result = append(result, &endpointRecord.ExtensionEndpoint)
	}
	return result, nil
}

// GetServiceCredentials - retrieve the list of credentials belonging to an extension service
func (store *ExtensionStore) GetServiceCredentials(ctx context.Context, serviceID domain.ExtensionServiceID) (result []*domain.ExtensionCredentials, err error) {
	svcRecord, err := store.getServiceRecord(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	result = make([]*domain.ExtensionCredentials, 0)
	for _, credentialsRecord := range svcRecord.credentials {
		result = append(result, &credentialsRecord.ExtensionCredentials)
	}
	return result, nil
}

// Retrieve an extension endpoint record by ID
func (store *ExtensionStore) getEndpointRecord(ctx context.Context, endpointID domain.ExtensionEndpointID) (result *extensionEndpointRecord, err error) {
	svcRecord, err := store.getServiceRecord(ctx,
		domain.ExtensionServiceID{
			ExtensionID: endpointID.ExtensionID,
			ID:          endpointID.ServiceID,
		})
	if err != nil {
		return nil, err
	}
	endpointRecord := svcRecord.endpoints[endpointID.URL]
	if endpointRecord == nil {
		return nil, domain.NewErrExtensionEndpointNotFound(endpointID.ExtensionID, endpointID.ServiceID, endpointID.URL)
	}
	return endpointRecord, nil
}

// GetEndpoint - retrieve an extension endpoint by ID
func (store *ExtensionStore) GetEndpoint(ctx context.Context, endpointID domain.ExtensionEndpointID) (result *domain.ExtensionEndpoint, err error) {
	endpointRecord, err := store.getEndpointRecord(ctx, endpointID)
	if err != nil {
		return nil, err
	}
	return &endpointRecord.ExtensionEndpoint, nil
}

// Retrieve an extension credentials record by ID
func (store *ExtensionStore) getCredentialsRecord(
	ctx context.Context, credentialsID domain.ExtensionCredentialsID) (result *extensionCredentialsRecord, err error) {
	svcRecord, err := store.getServiceRecord(ctx,
		domain.ExtensionServiceID{
			ExtensionID: credentialsID.ExtensionID,
			ID:          credentialsID.ServiceID,
		})
	if err != nil {
		return nil, err
	}
	credentialsRecord := svcRecord.credentials[credentialsID.ID]
	if credentialsRecord == nil {
		return nil, domain.NewErrExtensionCredentialsNotFound(credentialsID.ExtensionID, credentialsID.ServiceID, credentialsID.ID)
	}
	return credentialsRecord, nil
}

// GetCredentials - retrieve a set of extension credentials by ID
func (store *ExtensionStore) GetCredentials(ctx context.Context, credentialsID domain.ExtensionCredentialsID) (result *domain.ExtensionCredentials, err error) {
	credentialsRecord, err := store.getCredentialsRecord(ctx, credentialsID)
	if err != nil {
		return nil, err
	}
	return &credentialsRecord.ExtensionCredentials, nil
}

// UpdateExtension - update an extension
func (store *ExtensionStore) UpdateExtension(ctx context.Context, extension *domain.Extension) (err error) {
	extRecord, err := store.getExtensionRecord(ctx, extension.ID)
	if err != nil {
		return err
	}
	extension.Registered = extRecord.Registered
	extension.Updated = time.Now()
	extRecord.Extension = *extension
	return nil
}

// UpdateService - update a service belonging to an extension
func (store *ExtensionStore) UpdateService(ctx context.Context, service *domain.ExtensionService) (err error) {
	svcRecord, err := store.getServiceRecord(ctx, service.ExtensionServiceID)
	if err != nil {
		return err
	}
	service.Registered = svcRecord.Registered
	service.Updated = time.Now()
	svcRecord.ExtensionService = *service
	return nil
}

// UpdateEndpoint - update an endpoint belonging to a service
func (store *ExtensionStore) UpdateEndpoint(ctx context.Context, endpoint *domain.ExtensionEndpoint) (err error) {
	endpointRecord, err := store.getEndpointRecord(ctx, endpoint.ExtensionEndpointID)
	if err != nil {
		return err
	}
	endpointRecord.ExtensionEndpoint = *endpoint
	return nil
}

// UpdateCredentials - update a set of credentials belonging to a service
func (store *ExtensionStore) UpdateCredentials(ctx context.Context, credentials *domain.ExtensionCredentials) (err error) {
	credRecord, err := store.getCredentialsRecord(ctx, credentials.ExtensionCredentialsID)
	if err != nil {
		return err
	}
	credentials.Created = credRecord.Created
	credentials.Updated = time.Now()
	credRecord.ExtensionCredentials = *credentials
	return nil
}

// Delete an extension record
func (store *ExtensionStore) deleteExtensionRecord(ctx context.Context, extRecord *extensionRecord) error {
	for _, svcRecord := range extRecord.services {
		err := store.deleteServiceRecord(ctx, svcRecord)
		if err != nil {
			return err
		}
	}
	delete(store.items, extRecord.ID)
	return nil
}

// DeleteExtension - remove an extension and all its services, endpoints and credentials
func (store *ExtensionStore) DeleteExtension(ctx context.Context, extensionID string) error {
	extRecord, err := store.getExtensionRecord(ctx, extensionID)
	if err != nil {
		return err
	}
	return store.deleteExtensionRecord(ctx, extRecord)
}

// Delete an extension service record
func (store *ExtensionStore) deleteServiceRecord(ctx context.Context, svcRecord *extensionServiceRecord) error {
	for _, endpointRecord := range svcRecord.endpoints {
		err := store.deleteEndpointRecord(ctx, endpointRecord)
		if err != nil {
			return err
		}
	}
	for _, credsRecord := range svcRecord.credentials {
		err := store.deleteCredentialsRecord(ctx, credsRecord)
		if err != nil {
			return err
		}
	}
	delete(svcRecord.extension.services, svcRecord.ID)
	return nil
}

// DeleteService - remove an extension service and all its endpoints and credentials
func (store *ExtensionStore) DeleteService(ctx context.Context, serviceID domain.ExtensionServiceID) error {
	svcRecord, err := store.getServiceRecord(ctx, serviceID)
	if err != nil {
		return err
	}
	return store.deleteServiceRecord(ctx, svcRecord)
}

// Delete an extension endpoint record
func (store *ExtensionStore) deleteEndpointRecord(ctx context.Context, endpointRecord *extensionEndpointRecord) error {
	delete(endpointRecord.service.endpoints, endpointRecord.URL)
	return nil
}

// DeleteEndpoint - remove an extension endpoint
func (store *ExtensionStore) DeleteEndpoint(ctx context.Context, endpointID domain.ExtensionEndpointID) error {
	endpointRecord, err := store.getEndpointRecord(ctx, endpointID)
	if err != nil {
		return err
	}
	return store.deleteEndpointRecord(ctx, endpointRecord)
}

// Delete an extension credentials record
func (store *ExtensionStore) deleteCredentialsRecord(ctx context.Context, credentialsRecord *extensionCredentialsRecord) error {
	delete(credentialsRecord.service.credentials, credentialsRecord.ID)
	return nil
}

// DeleteCredentials - remove a set of extension credentials
func (store *ExtensionStore) DeleteCredentials(ctx context.Context, credentialsID domain.ExtensionCredentialsID) error {
	credentialsRecord, err := store.getCredentialsRecord(ctx, credentialsID)
	if err != nil {
		return err
	}
	return store.deleteCredentialsRecord(ctx, credentialsRecord)
}

func (store *ExtensionStore) findExtensionRecords(ctx context.Context, query *domain.ExtensionQuery) (result []*domain.ExtensionRecord, err error) {
	result = make([]*domain.ExtensionRecord, 0)

	addIfMatch := func(extRecord *extensionRecord) {
		if query.Zone != "" && query.StrictZoneMatch && query.Zone != extRecord.Zone {
			return
		}
		if query.Product != "" && query.Product != extRecord.Product {
			return
		}
		if query.VersionConstraints != "" && query.VersionConstraints != extRecord.Version {
			// try interpreting the version as a semantic version constraint

			version, err := semver.NewVersion(extRecord.Version)
			if err != nil {
				// Version not parseable or doesn't respect semantic versioning format
				return
			}

			constraints, err := semver.NewConstraint(query.VersionConstraints)
			if err != nil {
				// Constraints not parseable or not a constraints string
				return
			}

			// Check if the version meets the constraints
			if !constraints.Check(version) {
				return
			}
		}

		services, err := store.findServiceRecords(ctx, extRecord, query)

		if err == nil && len(services) > 0 {
			result = append(result,
				&domain.ExtensionRecord{
					Extension: extRecord.Extension,
					Services:  services,
				})
		}
	}

	if query.ExtensionID != "" {
		extRecord := store.items[query.ExtensionID]
		if extRecord != nil {
			addIfMatch(extRecord)
		}
		return result, nil
	}

	for _, extRecord := range store.items {
		addIfMatch(extRecord)
	}
	return result, nil
}

func (store *ExtensionStore) findServiceRecords(
	ctx context.Context, extRecord *extensionRecord, query *domain.ExtensionQuery) (result []*domain.ExtensionServiceRecord, err error) {
	result = make([]*domain.ExtensionServiceRecord, 0)

	addIfMatch := func(svcRecord *extensionServiceRecord) {
		if query.ServiceID != "" && query.ServiceID != svcRecord.ID {
			return
		}
		if query.ServiceResource != "" && query.ServiceResource != svcRecord.Resource {
			return
		}
		if query.ServiceCategory != "" && query.ServiceCategory != svcRecord.Category {
			return
		}

		endpoints, err := store.findEndpoints(ctx, svcRecord, query)
		if err == nil && len(endpoints) > 0 {
			credentials, err := store.findCredentials(ctx, svcRecord, query)
			if err == nil && (len(credentials) > 0 || !svcRecord.AuthRequired) {
				result = append(result,
					&domain.ExtensionServiceRecord{
						ExtensionService: svcRecord.ExtensionService,
						Endpoints:        endpoints,
						Credentials:      credentials,
					})
			}
		}
	}

	for _, svcRecord := range extRecord.services {
		addIfMatch(svcRecord)
	}

	return result, nil
}

func (store *ExtensionStore) findEndpoints(
	ctx context.Context, svcRecord *extensionServiceRecord, query *domain.ExtensionQuery) (result []*domain.ExtensionEndpoint, err error) {
	result = make([]*domain.ExtensionEndpoint, 0)

	addIfMatch := func(endpoint *domain.ExtensionEndpoint) {

		if query.EndpointURL != "" && query.EndpointURL != endpoint.URL {
			return
		}
		if query.EndpointType != nil {
			// match endpoint type, if supplied with the query
			if *query.EndpointType != endpoint.EndpointType {
				return
			}
		} else if query.Zone != "" {
			// if endpoint type is not supplied with the query, determine the endpoint type
			// by comparing the query zone (if supplied) with the extension record zone
			//  - if the extension is in the same zone as the query, both internal and external endpoints will match
			//  - otherwise, only external endpoints will match
			if query.Zone != svcRecord.extension.Zone && endpoint.EndpointType == domain.EETInternal {
				return
			}
		}

		result = append(result, endpoint)
	}

	for _, endpointRecord := range svcRecord.endpoints {
		addIfMatch(&endpointRecord.ExtensionEndpoint)
	}

	return result, nil
}

func stringInSlice(s string, slice []string) bool {
	for _, v := range slice {
		if s == v {
			return true
		}
	}
	return false
}

func (store *ExtensionStore) findCredentials(
	ctx context.Context, svcRecord *extensionServiceRecord, query *domain.ExtensionQuery) (result []*domain.ExtensionCredentials, err error) {
	result = make([]*domain.ExtensionCredentials, 0)

	addIfMatch := func(credentials *domain.ExtensionCredentials) {
		if query.CredentialsID != "" && query.CredentialsID != credentials.ID {
			return
		}
		// if the query is for global scoped credentials, only credentials with a global scope match
		if query.CredentialsScope == domain.ECSGlobal && credentials.Scope != domain.ECSGlobal {
			return
		}
		// if the query is for project scoped credentials, only credentials with a global scope
		// and those project scoped to the supplied project match
		if query.CredentialsScope == domain.ECSProject {
			if credentials.Scope == domain.ECSUser {
				return
			}
			if credentials.Scope == domain.ECSProject && !stringInSlice(query.Project, credentials.Projects) {
				return
			}
		}
		// if the query is for a user scoped credential, only credentials with a global scope,
		// those project scoped to the supplied project, and those user scoped to the supplied user and project
		// are a match
		if query.CredentialsScope == domain.ECSUser {
			if credentials.Scope != domain.ECSGlobal && query.Project != "" && !stringInSlice(query.Project, credentials.Projects) {
				return
			}
			if credentials.Scope == domain.ECSUser && !stringInSlice(query.User, credentials.Users) {
				return
			}
		}
		result = append(result, credentials)
	}

	for _, credentialsRecord := range svcRecord.credentials {
		addIfMatch(&credentialsRecord.ExtensionCredentials)
	}

	return result, nil
}

// RunExtensionQuery - run a query on the extension store to find one or more extensions, services, endpoints and credentials matching
// the supplied criteria
func (store *ExtensionStore) RunExtensionQuery(ctx context.Context, query *domain.ExtensionQuery) (result []*domain.ExtensionRecord, err error) {
	return store.findExtensionRecords(ctx, query)
}
