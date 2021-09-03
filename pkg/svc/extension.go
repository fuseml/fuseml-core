package svc

import (
	"context"
	"log"
	"net/url"
	"time"

	"github.com/fuseml/fuseml-core/gen/extension"
	"github.com/fuseml/fuseml-core/pkg/domain"
	"github.com/fuseml/fuseml-core/pkg/util"
)

// extension registry service implementation.
type extensionRegistrySvc struct {
	logger   *log.Logger
	registry domain.ExtensionRegistry
}

// NewExtensionRegistryService returns the extension registry service implementation.
func NewExtensionRegistryService(logger *log.Logger, registry domain.ExtensionRegistry) extension.Service {
	return &extensionRegistrySvc{logger, registry}
}

func extensionToDomain(extension *extension.Extension) (*domain.Extension, error) {
	ext := &domain.Extension{
		ID:            util.DerefString(extension.ID),
		Product:       util.DerefString(extension.Product),
		Version:       util.DerefString(extension.Version),
		Description:   util.DerefString(extension.Description),
		Zone:          util.DerefString(extension.Zone),
		Configuration: extension.Configuration,
	}
	if extension.Services != nil {
		err := setExtensionServices(ext, extension.Services)
		if err != nil {
			return nil, err
		}
	}
	return ext, nil
}

func extensionServiceToDomain(service *extension.ExtensionService) (*domain.ExtensionService, error) {
	svc := &domain.ExtensionService{
		ID:            util.DerefString(service.ID),
		Resource:      util.DerefString(service.Resource),
		Category:      util.DerefString(service.Category),
		Description:   util.DerefString(service.Description),
		AuthRequired:  util.DerefBool(service.AuthRequired),
		Configuration: service.Configuration,
	}
	if service.Endpoints != nil {
		err := setExtensionServiceEndpoints(svc, service.Endpoints)
		if err != nil {
			return nil, err
		}
	}
	if service.Credentials != nil {
		err := setExtensionServiceCredentials(svc, service.Credentials)
		if err != nil {
			return nil, err
		}
	}
	return svc, nil
}

func setExtensionServices(extension *domain.Extension, services []*extension.ExtensionService) error {
	for _, service := range services {
		svc, err := extensionServiceToDomain(service)
		if err != nil {
			return err
		}
		_, err = extension.AddService(svc)
		if err != nil {
			return err
		}
	}
	return nil
}

func extensionEndpointURLToDomain(URL *string) string {
	// the endpoint URL might be URL-encoded; attempt to decode it and ignore the error
	decodedURL := util.DerefString(URL)
	decodedURL, _ = url.PathUnescape(decodedURL)
	return decodedURL
}

func extensionEndpointToDomain(endpoint *extension.ExtensionEndpoint) *domain.ExtensionServiceEndpoint {
	return &domain.ExtensionServiceEndpoint{
		URL:           extensionEndpointURLToDomain(endpoint.URL),
		Type:          domain.ExtensionServiceEndpointType(util.DerefString(endpoint.Type, string(domain.EETExternal))),
		Configuration: endpoint.Configuration,
	}
}

func setExtensionServiceEndpoints(service *domain.ExtensionService, endpoints []*extension.ExtensionEndpoint) error {
	for _, endpoint := range endpoints {
		_, err := service.AddEndpoint(extensionEndpointToDomain(endpoint))
		if err != nil {
			return err
		}
	}
	return nil
}

func extensionCredentialsToDomain(credentials *extension.ExtensionCredentials) *domain.ExtensionServiceCredentials {
	return &domain.ExtensionServiceCredentials{
		ID:            util.DerefString(credentials.ID),
		Scope:         domain.ExtensionServiceCredentialsScope(util.DerefString(credentials.Scope, string(domain.ECSGlobal))),
		Default:       util.DerefBool(credentials.Default),
		Projects:      credentials.Projects,
		Users:         credentials.Users,
		Configuration: credentials.Configuration,
	}
}

func setExtensionServiceCredentials(service *domain.ExtensionService, credentials []*extension.ExtensionCredentials) error {
	for _, credential := range credentials {
		_, err := service.AddCredentials(extensionCredentialsToDomain(credential))
		if err != nil {
			return err
		}
	}
	return nil
}

func extensionQueryToDomain(query *extension.ExtensionQuery) (result *domain.ExtensionQuery) {
	result = &domain.ExtensionQuery{
		ExtensionID:        query.ExtensionID,
		Product:            query.Product,
		VersionConstraints: query.Version,
		Zone:               query.Zone,
		// If a zone is supplied, it must be used to do strict zone matching
		StrictZoneMatch: true,
		ServiceID:       query.ServiceID,
		ServiceResource: query.ServiceResource,
		ServiceCategory: query.ServiceCategory,
	}

	return result
}

func extensionToRest(ctx context.Context, ext *domain.Extension) *extension.Extension {
	restExt := &extension.Extension{
		ID:            util.RefString(ext.ID),
		Product:       util.RefString(ext.Product),
		Version:       util.RefString(ext.Version),
		Description:   util.RefString(ext.Description),
		Zone:          util.RefString(ext.Zone),
		Configuration: ext.Configuration,
		Status: &extension.ExtensionStatus{
			Registered: ext.Created.Format(time.RFC3339),
			Updated:    ext.Updated.Format(time.RFC3339),
		},
	}
	if ext.Services != nil {
		restExt.Services = extensionServiceListToRest(ext.ID, ext.Services)
	}
	return restExt
}

func extensionServiceToRest(extensionID string, service *domain.ExtensionService) *extension.ExtensionService {
	restSvc := &extension.ExtensionService{
		ID:            util.RefString(service.ID),
		ExtensionID:   util.RefString(extensionID),
		Resource:      util.RefString(service.Resource),
		Category:      util.RefString(service.Category),
		Description:   util.RefString(service.Description),
		AuthRequired:  &service.AuthRequired,
		Configuration: service.Configuration,
		Status: &extension.ExtensionServiceStatus{
			Registered: service.Created.Format(time.RFC3339),
			Updated:    service.Updated.Format(time.RFC3339),
		},
	}
	if service.Endpoints != nil {
		restSvc.Endpoints = extensionEndpointListToRest(extensionID, service.ID, service.Endpoints)
	}
	if service.Credentials != nil {
		restSvc.Credentials = extensionCredentialsListToRest(extensionID, service.ID, service.Credentials)
	}
	return restSvc
}

func extensionServiceListToRest(extensionID string, services map[string]*domain.ExtensionService) []*extension.ExtensionService {
	restServices := []*extension.ExtensionService{}
	for _, service := range services {
		restServices = append(restServices, extensionServiceToRest(extensionID, service))
	}
	return restServices
}

func extensionEndpointToRest(extensionID string, serviceID string, endpoint *domain.ExtensionServiceEndpoint) *extension.ExtensionEndpoint {
	return &extension.ExtensionEndpoint{
		URL:           util.RefString(endpoint.URL),
		ExtensionID:   util.RefString(extensionID),
		ServiceID:     util.RefString(serviceID),
		Type:          util.RefString(string(endpoint.Type)),
		Configuration: endpoint.Configuration,
		// TODO: Add registered and updated fields
		Status: &extension.ExtensionEndpointStatus{},
	}
}

func extensionEndpointListToRest(extensionID string, serviceID string, endpoints map[string]*domain.ExtensionServiceEndpoint) []*extension.ExtensionEndpoint {
	restEndpoints := []*extension.ExtensionEndpoint{}
	for _, endpoint := range endpoints {
		restEndpoints = append(restEndpoints, extensionEndpointToRest(extensionID, serviceID, endpoint))
	}
	return restEndpoints
}

func obfuscateCredentials(config map[string]string) (result map[string]string) {
	result = make(map[string]string)
	for k := range config {
		result[k] = "<hidden>"
	}
	return result
}

func extensionCredentialsToRest(extensionID string, serviceID string, credentials *domain.ExtensionServiceCredentials) *extension.ExtensionCredentials {
	return &extension.ExtensionCredentials{
		ID:            util.RefString(credentials.ID),
		ExtensionID:   util.RefString(extensionID),
		ServiceID:     util.RefString(serviceID),
		Scope:         util.RefString(string(credentials.Scope)),
		Default:       &credentials.Default,
		Projects:      credentials.Projects,
		Users:         credentials.Users,
		Configuration: obfuscateCredentials(credentials.Configuration),
		Status: &extension.ExtensionCredentialsStatus{
			Created: credentials.Created.Format(time.RFC3339),
			Updated: credentials.Updated.Format(time.RFC3339),
		},
	}
}

func extensionCredentialsListToRest(extensionID string, serviceID string, credentials map[string]*domain.ExtensionServiceCredentials) []*extension.ExtensionCredentials {
	restCredentials := []*extension.ExtensionCredentials{}
	for _, credential := range credentials {
		restCredentials = append(restCredentials, extensionCredentialsToRest(extensionID, serviceID, credential))
	}
	return restCredentials
}

func errToRest(err error) error {
	switch err.(type) {
	case *domain.ErrExtensionNotFound:
		return extension.MakeNotFound(err)
	case *domain.ErrExtensionServiceNotFound:
		return extension.MakeNotFound(err)
	case *domain.ErrExtensionServiceEndpointNotFound:
		return extension.MakeNotFound(err)
	case *domain.ErrExtensionServiceCredentialsNotFound:
		return extension.MakeNotFound(err)
	case *domain.ErrExtensionExists:
		return extension.MakeConflict(err)
	case *domain.ErrExtensionServiceExists:
		return extension.MakeConflict(err)
	case *domain.ErrExtensionServiceEndpointExists:
		return extension.MakeConflict(err)
	case *domain.ErrExtensionServiceCredentialsExists:
		return extension.MakeConflict(err)
	default:
		return extension.MakeBadRequest(err)
	}
}

// Register an extension with the FuseML extension registry.
func (s *extensionRegistrySvc) RegisterExtension(ctx context.Context, req *extension.Extension) (*extension.Extension, error) {
	s.logger.Print("extension.registerExtension")
	domainExt, err := extensionToDomain(req)
	if err != nil {
		return nil, errToRest(err)
	}
	extension, err := s.registry.RegisterExtension(ctx, domainExt)
	if err != nil {
		return nil, errToRest(err)
	}
	return extensionToRest(ctx, extension), nil
}

// Retrieve information about an extension.
func (s *extensionRegistrySvc) GetExtension(ctx context.Context, req *extension.GetExtensionPayload) (res *extension.Extension, err error) {
	s.logger.Print("extension.getExtension")
	extension, err := s.registry.GetExtension(ctx, req.ID)
	if err != nil {
		return nil, errToRest(err)
	}
	return extensionToRest(ctx, extension), nil
}

// List extensions registered in FuseML
func (s *extensionRegistrySvc) ListExtensions(ctx context.Context, query *extension.ExtensionQuery) (res []*extension.Extension, err error) {
	s.logger.Print("extension.listExtensions")
	extensions, err := s.registry.ListExtensions(ctx, extensionQueryToDomain(query))
	if err != nil {
		return nil, errToRest(err)
	}

	res = make([]*extension.Extension, len(extensions))
	for i, extension := range extensions {
		res[i] = extensionToRest(ctx, extension)
	}

	return res, nil
}

// Update an extension registered in FuseML
func (s *extensionRegistrySvc) UpdateExtension(ctx context.Context, req *extension.Extension) (res *extension.Extension, err error) {
	s.logger.Print("extension.updateExtension")
	domainExt, err := extensionToDomain(req)
	if err != nil {
		return nil, errToRest(err)
	}
	extension, err := s.registry.GetExtension(ctx, domainExt.ID)
	if err != nil {
		return nil, errToRest(err)
	}
	// update only attributes present in the update request
	extUpdate := domain.Extension{
		ID:            extension.ID,
		Product:       util.DerefString(req.Product, extension.Product),
		Version:       util.DerefString(req.Version, extension.Version),
		Description:   util.DerefString(req.Description, extension.Description),
		Zone:          util.DerefString(req.Zone, extension.Zone),
		Configuration: extension.Configuration,
		Services:      extension.Services,
	}
	if req.Configuration != nil {
		extUpdate.Configuration = req.Configuration
	}

	if req.Services != nil {
		err = setExtensionServices(&extUpdate, req.Services)
		if err != nil {
			return nil, errToRest(err)
		}
	}

	err = s.registry.UpdateExtension(ctx, &extUpdate)
	if err != nil {
		return nil, errToRest(err)
	}
	return extensionToRest(ctx, &extUpdate), nil
}

// Delete an extension and its subtree of services, endpoints and credentials
func (s *extensionRegistrySvc) DeleteExtension(ctx context.Context, req *extension.DeleteExtensionPayload) (err error) {
	s.logger.Print("extension.deleteExtension")
	err = s.registry.RemoveExtension(ctx, req.ID)
	if err != nil {
		return errToRest(err)
	}
	return nil
}

// Add a service to an existing extension registered with the FuseML extension
// registry.
func (s *extensionRegistrySvc) AddService(ctx context.Context, service *extension.ExtensionService) (res *extension.ExtensionService, err error) {
	svc, err := extensionServiceToDomain(service)
	if err != nil {
		return nil, errToRest(err)
	}
	svc, err = s.registry.AddService(ctx, util.DerefString(service.ExtensionID), svc)
	if err != nil {
		return nil, errToRest(err)
	}
	return extensionServiceToRest(util.DerefString(service.ExtensionID), svc), nil
}

// Retrieve information about a service belonging to an extension.
func (s *extensionRegistrySvc) GetService(ctx context.Context, req *extension.GetServicePayload) (res *extension.ExtensionService, err error) {
	svc, err := s.registry.GetService(ctx, req.ExtensionID, req.ID)
	if err != nil {
		return nil, errToRest(err)
	}
	return extensionServiceToRest(req.ExtensionID, svc), nil
}

// List all services associated with an extension registered in FuseML
func (s *extensionRegistrySvc) ListServices(ctx context.Context, req *extension.ListServicesPayload) (res []*extension.ExtensionService, err error) {
	ext, err := s.registry.GetExtension(ctx, req.ExtensionID)
	if err != nil {
		return nil, errToRest(err)
	}
	res = make([]*extension.ExtensionService, len(ext.Services))
	for i, svc := range ext.ListServices() {
		res[i] = extensionServiceToRest(req.ExtensionID, svc)
	}
	return res, nil
}

// Update a service belonging to an extension registered in FuseML
func (s *extensionRegistrySvc) UpdateService(ctx context.Context, req *extension.ExtensionService) (res *extension.ExtensionService, err error) {
	s.logger.Print("extension.updateService")
	service, err := extensionServiceToDomain(req)
	if err != nil {
		return nil, errToRest(err)
	}
	svc, err := s.registry.GetService(ctx, util.DerefString(req.ExtensionID), service.ID)
	if err != nil {
		return nil, errToRest(err)
	}

	// update only attributes present in the update request
	svcUpdate := domain.ExtensionService{
		ID:            service.ID,
		Resource:      util.DerefString(req.Resource, svc.Resource),
		Category:      util.DerefString(req.Category, svc.Resource),
		Description:   util.DerefString(req.Description, svc.Description),
		AuthRequired:  util.DerefBool(req.AuthRequired, svc.AuthRequired),
		Configuration: svc.Configuration,
		Endpoints:     svc.Endpoints,
		Credentials:   svc.Credentials,
	}
	if req.Configuration != nil {
		svcUpdate.Configuration = req.Configuration
	}
	if req.Endpoints != nil {
		err = setExtensionServiceEndpoints(&svcUpdate, req.Endpoints)
		if err != nil {
			return nil, errToRest(err)
		}
	}
	if req.Credentials != nil {
		err = setExtensionServiceCredentials(&svcUpdate, req.Credentials)
		if err != nil {
			return nil, errToRest(err)
		}
	}

	err = s.registry.UpdateService(ctx, util.DerefString(req.ExtensionID), &svcUpdate)
	if err != nil {
		return nil, errToRest(err)
	}
	return extensionServiceToRest(util.DerefString(req.ExtensionID), &svcUpdate), nil
}

// Delete an extension service and its subtree of endpoints and credentials
func (s *extensionRegistrySvc) DeleteService(ctx context.Context, req *extension.DeleteServicePayload) (err error) {
	s.logger.Print("extension.deleteService")
	err = s.registry.RemoveService(ctx, req.ExtensionID, req.ID)
	if err != nil {
		return errToRest(err)
	}
	return nil
}

// Add an endpoint to an existing extension service registered with the FuseML
// extension registry.
func (s *extensionRegistrySvc) AddEndpoint(ctx context.Context, req *extension.ExtensionEndpoint) (res *extension.ExtensionEndpoint, err error) {
	s.logger.Print("extension.addEndpoint")
	endpoint, err := s.registry.AddEndpoint(ctx, util.DerefString(req.ExtensionID), util.DerefString(req.ServiceID), extensionEndpointToDomain(req))
	if err != nil {
		return nil, errToRest(err)
	}
	return extensionEndpointToRest(util.DerefString(req.ExtensionID), util.DerefString(req.ServiceID), endpoint), nil
}

// Retrieve information about an endpoint belonging to an extension.
func (s *extensionRegistrySvc) GetEndpoint(ctx context.Context, req *extension.GetEndpointPayload) (res *extension.ExtensionEndpoint, err error) {
	s.logger.Print("extension.getEndpoint")
	endpoint, err := s.registry.GetEndpoint(ctx, req.ExtensionID, req.ServiceID, extensionEndpointURLToDomain(&req.URL))
	if err != nil {
		return nil, errToRest(err)
	}
	return extensionEndpointToRest(req.ExtensionID, req.ServiceID, endpoint), nil
}

// List all endpoints associated with an extension service registered in FuseML
func (s *extensionRegistrySvc) ListEndpoints(ctx context.Context, req *extension.ListEndpointsPayload) (res []*extension.ExtensionEndpoint, err error) {
	s.logger.Print("extension.listEndpoints")
	svc, err := s.registry.GetService(ctx, req.ExtensionID, req.ServiceID)
	if err != nil {
		return nil, errToRest(err)
	}
	res = extensionEndpointListToRest(req.ExtensionID, req.ServiceID, svc.Endpoints)
	return res, nil
}

// Update an endpoint belonging to an extension service registered in FuseML
func (s *extensionRegistrySvc) UpdateEndpoint(ctx context.Context, req *extension.ExtensionEndpoint) (res *extension.ExtensionEndpoint, err error) {
	s.logger.Print("extension.updateEndpoint")
	endpoint := extensionEndpointToDomain(req)
	ep, err := s.registry.GetEndpoint(ctx, util.DerefString(req.ExtensionID), util.DerefString(req.ServiceID), extensionEndpointURLToDomain(&endpoint.URL))
	if err != nil {
		return nil, errToRest(err)
	}

	// update only attributes present in the update request
	epUpdate := domain.ExtensionServiceEndpoint{
		URL:           extensionEndpointURLToDomain(&endpoint.URL),
		Type:          domain.ExtensionServiceEndpointType(util.DerefString(req.Type, string(ep.Type))),
		Configuration: ep.Configuration,
	}
	if req.Configuration != nil {
		epUpdate.Configuration = req.Configuration
	}

	err = s.registry.UpdateEndpoint(ctx, util.DerefString(req.ExtensionID), util.DerefString(req.ServiceID), &epUpdate)
	if err != nil {
		return nil, errToRest(err)
	}
	return extensionEndpointToRest(util.DerefString(req.ExtensionID), util.DerefString(req.ServiceID), &epUpdate), nil
}

// Delete an extension endpoint
func (s *extensionRegistrySvc) DeleteEndpoint(ctx context.Context, req *extension.DeleteEndpointPayload) (err error) {
	s.logger.Print("extension.deleteEndpoint")
	err = s.registry.RemoveEndpoint(ctx, req.ExtensionID, req.ServiceID, extensionEndpointURLToDomain(&req.URL))
	if err != nil {
		return errToRest(err)
	}
	return nil
}

// Add a set of credentials to an existing extension service registered with
// the FuseML extension registry.
func (s *extensionRegistrySvc) AddCredentials(ctx context.Context, req *extension.ExtensionCredentials) (res *extension.ExtensionCredentials, err error) {
	credentials, err := s.registry.AddCredentials(ctx, util.DerefString(req.ExtensionID), util.DerefString(req.ServiceID), extensionCredentialsToDomain(req))
	if err != nil {
		return nil, errToRest(err)
	}
	return extensionCredentialsToRest(util.DerefString(req.ExtensionID), util.DerefString(req.ServiceID), credentials), nil
}

// Retrieve information about a set of credentials belonging to an extension.
func (s *extensionRegistrySvc) GetCredentials(ctx context.Context, req *extension.GetCredentialsPayload) (res *extension.ExtensionCredentials, err error) {
	s.logger.Print("extension.getCredentials")
	credentials, err := s.registry.GetCredentials(ctx, req.ExtensionID, req.ServiceID, req.ID)
	if err != nil {
		return nil, errToRest(err)
	}
	return extensionCredentialsToRest(req.ExtensionID, req.ServiceID, credentials), nil
}

// List all credentials associated with an extension service registered in
// FuseML
func (s *extensionRegistrySvc) ListCredentials(ctx context.Context, req *extension.ListCredentialsPayload) (res []*extension.ExtensionCredentials, err error) {
	s.logger.Print("extension.listCredentials")
	svc, err := s.registry.GetService(ctx, req.ExtensionID, req.ServiceID)
	if err != nil {
		return nil, errToRest(err)
	}
	res = extensionCredentialsListToRest(req.ExtensionID, req.ServiceID, svc.Credentials)
	return res, nil
}

// Update a set of credentials belonging to an extension service registered in
// FuseML
func (s *extensionRegistrySvc) UpdateCredentials(ctx context.Context, req *extension.ExtensionCredentials) (res *extension.ExtensionCredentials, err error) {
	s.logger.Print("extension.updateCredentials")
	credentials := extensionCredentialsToDomain(req)
	cred, err := s.registry.GetCredentials(ctx, util.DerefString(req.ExtensionID), util.DerefString(req.ServiceID), credentials.ID)
	if err != nil {
		return nil, errToRest(err)
	}

	// update only attributes present in the update request
	credUpdate := domain.ExtensionServiceCredentials{
		ID:            cred.ID,
		Scope:         domain.ExtensionServiceCredentialsScope(util.DerefString(req.Scope, string(cred.Scope))),
		Default:       util.DerefBool(req.Default, cred.Default),
		Projects:      cred.Projects,
		Users:         cred.Users,
		Configuration: cred.Configuration,
	}

	if req.Configuration != nil {
		credUpdate.Configuration = req.Configuration
	}
	if req.Projects != nil {
		credUpdate.Projects = req.Projects
	}
	if req.Users != nil {
		credUpdate.Users = req.Users
	}

	err = s.registry.UpdateCredentials(ctx, util.DerefString(req.ExtensionID), util.DerefString(req.ServiceID), &credUpdate)
	if err != nil {
		return nil, errToRest(err)
	}
	return extensionCredentialsToRest(util.DerefString(req.ExtensionID), util.DerefString(req.ServiceID), &credUpdate), nil
}

// Delete a set of extension credentials
func (s *extensionRegistrySvc) DeleteCredentials(ctx context.Context, req *extension.DeleteCredentialsPayload) (err error) {
	s.logger.Print("extension.deleteCredentials")
	err = s.registry.RemoveCredentials(ctx, req.ExtensionID, req.ServiceID, req.ID)
	if err != nil {
		return errToRest(err)
	}
	return nil
}
