package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/semver"
	"github.com/fuseml/fuseml-core/pkg/util"
	"k8s.io/apimachinery/pkg/util/rand"
)

// ExtensionServiceEndpointType is the type used for the ExtensionServiceEndpoint Type field
type ExtensionServiceEndpointType string

// Valid values that can be used with ExtensionServiceEndpointType
const (
	// EETInternal is an internal endpoint that can only be accessed from the same zone
	EETInternal ExtensionServiceEndpointType = "internal"
	// EETExternal is an external endpoint that can be accessed from any zone
	EETExternal = "external"
)

// ExtensionServiceCredentialsScope is the type used for the ExtensionServiceCredentials Scope field
type ExtensionServiceCredentialsScope string

// Valid values that can be used with ExtensionCredentialScope
const (
	// ECSGlobal is a global scope indicating that credentials may be used for any project and user
	ECSGlobal ExtensionServiceCredentialsScope = "global"
	// ECSProject is a project scope indicating that credentials may only be used in the context of
	// a controlled list of projects
	ECSProject = "project"
	// ECSUser is a user scope indicating that credentials may only be used in the context of
	// a controlled list of users and projects
	ECSUser = "user"
)

// Extension is an entry in the extension registry that describes a particular installation of a
// framework/platform/service/product developed and released or hosted under a unique product name
type Extension struct {
	// Extension ID - used to uniquely identify an extension in the registry
	ID string
	// Universal product identifier that can be used to group and identify extensions according to the product
	// they belong to. Product values can be used to identify installations of the same product registered
	// with the same or different FuseML servers.
	Product string
	// Optional extension version. To support semantic version operations, such as matching lookup operations
	// that include a version requirement specifier, it should be formatted as [v]MAJOR[.MINOR[.PATCH[-PRERELEASE][+BUILD]]]
	Version string
	// Optional extension description
	Description string
	// Optional zone identifier. Can be used to group and lookup extensions according to the infrastructure
	// location / zone / area / domain where they are installed (e.g. kubernetes cluster).
	// Is used to automatically select between cluster-local and external endpoints when
	// running queries.
	Zone string
	// Configuration entries (e.g. configuration values required to configure all clients that connect to
	// this extension), expressed as set of key-value entries
	Configuration map[string]string
	// The time when the extension was registered
	Created time.Time
	// The time when the extension was last updated
	Updated time.Time
	// Services is a list of services that are part of this extension
	Services map[string]*ExtensionService
}

// ExtensionService is a service provided by an extension. A service is represented by a
// single API or UI. For extensions implemented as cloud-native applications, a service is the
// equivalent of a kubernetes service that is used to expose a public API or UI. Services are classified
// into known resource types (e.g. s3, git) encoded via the Resource attribute and service
// categories (e.g. model store, feature store, distributed training, serving) via the Category attribute
type ExtensionService struct {
	// Extension service ID - used to uniquely identify an extension service in the registry
	ID string
	// Universal service identifier that can be used to identify a service in any FuseML installation.
	// This identifier should uniquely identify the API or protocol (e.g. s3, git, mlflow) that the service
	// provides.
	Resource string
	// Universal service category. Used to classify services into well-known categories of AI/ML services
	// (e.g. model store, feature store, distributed training, serving).
	Category string
	// Optional extension service description
	Description string
	// Marks a service for which authentication is required. If set, a set of credentials is required
	// to access the service; if none of the provided credentials match the scope of the consumer,
	// this service will be excluded from queries
	AuthRequired bool
	// Configuration entries (e.g. configuration values required to configure the client to access this
	// service), expressed as set of key-value entries
	Configuration map[string]string
	// The time when the service was registered
	Created time.Time
	// The time when the service was last updated
	Updated time.Time
	// Endpoints is a list of endpoints that are part of this extension service
	Endpoints map[string]*ExtensionServiceEndpoint
	// Credentials is a list of credentials that are part of this extension service
	Credentials map[string]*ExtensionServiceCredentials
}

// ExtensionServiceEndpoint is an endpoint through which an extension service can be accessed. Having a list of
// endpoints associated with a single extension service is particularly important for representing k8s
// services, which can be exposed both internally (cluster IP) and externally (e.g. ingress). All endpoints
// belonging to the same extension service must be equivalent in the sense that they are backed by the same
// API and/or protocol and exhibit the same behavior
type ExtensionServiceEndpoint struct {
	// Endpoint URL. In case of k8s controllers and operators, the URL points to the cluster API.
	// Also used to uniquely identifies an endpoint within the scope of a service
	URL string
	// Endpoint type - internal/external. An internal endpoint can only be accessed when the consumer
	// is located in the same zone as the extension service
	Type ExtensionServiceEndpointType
	// Configuration entries (e.g. CA certificates), expressed as set of key-value entries
	Configuration map[string]string
	// The time when the extension was registered
	Created time.Time
	// The time when the extension was last updated
	Updated time.Time
}

// ExtensionServiceCredentials is a group of configuration values that can be generally used to embed information
// pertaining to the authentication and authorization features supported by a service. This descriptor allows
// administrators and operators of 3rd party tools integrated with FuseML to configure different accounts
// and credentials (tokens, certificates, passwords etc.) to be associated with different FuseML organization
// entities (users, projects, groups etc.). All information embedded in a credentials descriptor entry is
// treated as sensitive information. Each credentials entry has an associated scope that controls who has
// access to this information (e.g. global, project, user, workflow). This is the equivalent of a k8s secret.
type ExtensionServiceCredentials struct {
	// Extension credentials ID - used to uniquely identify a set of credentials in the registry
	ID string
	// The scope associated with this set of credentials. Global scoped credentials can be used by any
	// user/project. Project scoped credentials can be used only in the context of one of the projects
	// supplied in the Projects list. User scoped credentials can only be used by the users in the Users
	// list and, optionally, in the context of the projects supplied in the Projects list.
	Scope ExtensionServiceCredentialsScope
	// Use as default credentials. Used to automatically select one of several credentials with the same
	// scope matching the same query.
	Default bool
	// List of projects allowed to use these credentials
	Projects []string
	// List of users allowed to use these credentials
	Users []string
	// Configuration entries (e.g. usernames, passwords, tokens, keys), expressed as set of key-value entries
	Configuration map[string]string
	// The time when the credential set was created
	Created time.Time
	// The time when the credential set was last updated
	Updated time.Time
}

// ExtensionAccessDescriptor is a structure that contains all the information needed to access an extension:
// a service, an endpoint and an optional set of credentials. It's returned as result when running access queries
// against the extension registry.
type ExtensionAccessDescriptor struct {
	Extension   Extension
	Service     ExtensionService
	Endpoint    ExtensionServiceEndpoint
	Credentials *ExtensionServiceCredentials
}

// ExtensionQuery is a query that can be run against the extension registry to retrieve
// a list of extension endpoints and credentials that meet all supplied criteria
type ExtensionQuery struct {
	// Search by explicit extension ID
	ExtensionID string
	// Search by product name. Leave empty to include all products
	Product string
	// Search by version or by semantic version constraints. Leave empty to include all available versions
	VersionConstraints string
	// Match extensions installed in a given zone.
	Zone string
	// Use strict filtering when a zone query field is supplied. When set, only extensions
	// installed in the supplied zone are returned.
	StrictZoneMatch bool
	// Search by explicit service ID
	ServiceID string
	// Search by service resource type. Leave empty to include all resource types
	ServiceResource string
	// Search by service category. Leave empty to include all services
	ServiceCategory string
	// Search by explicit endpoint URL
	EndpointURL string
	// Search by endpoint type. If not explicitly specified, the endpoint type will be
	// determined automatically by StrictZoneMatch and the Zone value, if a Zone is supplied:
	//  - if the extension is in the same zone as the query, both internal and external endpoints will match
	//  - otherwise, only external endpoints will match
	Type *ExtensionServiceEndpointType
	// Search by explicit credentials ID
	CredentialsID string
	// Match credentials by scope
	CredentialsScope ExtensionServiceCredentialsScope
	// Match credentials allowed for a given user. CredentialsScope must be set to ECSUser
	// for this to have effect
	User string
	// Match credentials allowed for a given project. CredentialsScope must be set to ECSUser
	// or ECSProject for this to have effect
	Project string
}

// Errors returned by the methods in the ExtensionRegistry and ExtensionStore interfaces
// ---------------------------

// ErrExtensionExists is the error returned during registration, when an extension with the same ID
// already exists in the registry
type ErrExtensionExists string

// NewErrExtensionExists creates a new ErrExtensionExists error
func NewErrExtensionExists(extensionID string) *ErrExtensionExists {
	err := ErrExtensionExists(extensionID)
	return &err
}

func (e *ErrExtensionExists) Error() string {
	return fmt.Sprintf("an extension with the same ID already exists: %s", string(*e))
}

// ErrExtensionNotFound is the error returned by various registry methods when an extension with
// a given ID is not found in the registry
type ErrExtensionNotFound string

// NewErrExtensionNotFound creates a new ErrExtensionNotFound error
func NewErrExtensionNotFound(extensionID string) *ErrExtensionNotFound {
	err := ErrExtensionNotFound(extensionID)
	return &err
}

func (e *ErrExtensionNotFound) Error() string {
	return fmt.Sprintf("an extension with the given ID could not be found: %s", string(*e))
}

// ErrMissingField is the error returned by various registry methods if a required field has not been
// filled in the supplied object
type ErrMissingField struct {
	Element string
	Field   string
}

// NewErrMissingField creates a new ErrMissingField error
func NewErrMissingField(element, field string) *ErrMissingField {
	return &ErrMissingField{element, field}
}

func (e *ErrMissingField) Error() string {
	return fmt.Sprintf("required field is missing from '%s' structure: %s", e.Element, e.Field)
}

// ErrExtensionServiceExists is the error returned during registration or service addition, when an
// extension service with the same ID already exists under the parent extension
type ErrExtensionServiceExists struct {
	ExtensionID string
	ServiceID   string
}

// NewErrExtensionServiceExists creates a new ErrExtensionServiceExists error
func NewErrExtensionServiceExists(extensionID, serviceID string) *ErrExtensionServiceExists {
	return &ErrExtensionServiceExists{extensionID, serviceID}
}

func (e *ErrExtensionServiceExists) Error() string {
	return fmt.Sprintf("a service with the same ID already exists under the '%s' extension: %s", e.ExtensionID, e.ServiceID)
}

// ErrExtensionServiceNotFound is the error returned by various registry methods when an extension service
// with a given ID is not found under an extension
type ErrExtensionServiceNotFound struct {
	ExtensionID string
	ServiceID   string
}

// NewErrExtensionServiceNotFound creates a new ErrExtensionServiceNotFound error
func NewErrExtensionServiceNotFound(extensionID, serviceID string) *ErrExtensionServiceNotFound {
	return &ErrExtensionServiceNotFound{extensionID, serviceID}
}

func (e *ErrExtensionServiceNotFound) Error() string {
	return fmt.Sprintf("a service with the given ID could not be found under the '%s' extension: %s", e.ExtensionID, e.ServiceID)
}

// ErrExtensionServiceEndpointExists is the error returned during registration or endpoint addition, when an
// extension endpoint with the same URL already exists under the parent extension service
type ErrExtensionServiceEndpointExists struct {
	ExtensionID string
	ServiceID   string
	URL         string
}

// NewErrExtensionServiceEndpointExists creates a new ErrExtensionEndpointExists error
func NewErrExtensionServiceEndpointExists(extensionID, serviceID, URL string) *ErrExtensionServiceEndpointExists {
	return &ErrExtensionServiceEndpointExists{extensionID, serviceID, URL}
}

func (e *ErrExtensionServiceEndpointExists) Error() string {
	reference := e.ServiceID
	if e.ExtensionID != "" {
		reference = fmt.Sprintf("%s/%s", e.ExtensionID, e.ServiceID)
	}
	return fmt.Sprintf(
		"an endpoint with the same URL already exists under the %q extension service: %s",
		reference, e.URL)
}

// ErrExtensionServiceEndpointNotFound is the error returned by various registry methods when an extension endpoint
// with a given URL is not found under an extension service
type ErrExtensionServiceEndpointNotFound struct {
	ExtensionID string
	ServiceID   string
	URL         string
}

// NewErrExtensionServiceEndpointNotFound creates a new ErrExtensionEndpointNotFound error
func NewErrExtensionServiceEndpointNotFound(extensionID, serviceID, URL string) *ErrExtensionServiceEndpointNotFound {
	return &ErrExtensionServiceEndpointNotFound{extensionID, serviceID, URL}
}

func (e *ErrExtensionServiceEndpointNotFound) Error() string {
	return fmt.Sprintf(
		"an endpoint with the given URL could not be found under the '%s/%s' extension service: %s",
		e.ExtensionID, e.ServiceID, e.URL)
}

// ErrExtensionServiceCredentialsExists is the error returned during registration or credential addition, when a
// set of extension credentials with the same ID already exists under the parent extension service
type ErrExtensionServiceCredentialsExists struct {
	ExtensionID   string
	ServiceID     string
	CredentialsID string
}

// NewErrExtensionServiceCredentialsExists creates a new ErrExtensionCredentialsExists error
func NewErrExtensionServiceCredentialsExists(extensionID, serviceID, credentialsID string) *ErrExtensionServiceCredentialsExists {
	return &ErrExtensionServiceCredentialsExists{extensionID, serviceID, credentialsID}
}

func (e *ErrExtensionServiceCredentialsExists) Error() string {
	return fmt.Sprintf(
		"a set of credentials with the same ID already exists under the '%s/%s' extension service: %s",
		e.ExtensionID, e.ServiceID, e.CredentialsID)
}

// ErrExtensionServiceCredentialsNotFound is the error returned by various registry methods when a set of extension
// credentials with a given ID is not found under an extension service
type ErrExtensionServiceCredentialsNotFound struct {
	ExtensionID   string
	ServiceID     string
	CredentialsID string
}

// NewErrExtensionServiceCredentialsNotFound creates a new ErrExtensionCredentialsNotFound error
func NewErrExtensionServiceCredentialsNotFound(extensionID, serviceID, credentialsID string) *ErrExtensionServiceCredentialsNotFound {
	return &ErrExtensionServiceCredentialsNotFound{extensionID, serviceID, credentialsID}
}

func (e *ErrExtensionServiceCredentialsNotFound) Error() string {
	return fmt.Sprintf(
		"a set of credentials with the given ID could not be found under the '%s/%s' extension service: %s",
		e.ExtensionID, e.ServiceID, e.CredentialsID)
}

// ExtensionRegistry defines the public interface implemented by the extension registry
type ExtensionRegistry interface {
	// Register a new extension, with all participating services, endpoints and credentials
	RegisterExtension(ctx context.Context, extension *Extension) (*Extension, error)
	// Add a service to an existing extension
	AddService(ctx context.Context, extensionID string, service *ExtensionService) (*ExtensionService, error)
	// Add an endpoint to an existing extension service
	AddEndpoint(ctx context.Context, extensionID string, serviceID string, endpoint *ExtensionServiceEndpoint) (*ExtensionServiceEndpoint, error)
	// Add a set of credentials to an existing extension service
	AddCredentials(ctx context.Context, extensionID string, serviceID string, credentials *ExtensionServiceCredentials) (*ExtensionServiceCredentials, error)
	// List all registered extensions that match the supplied query parameters
	ListExtensions(ctx context.Context, query *ExtensionQuery) (result []*Extension, err error)
	// Retrieve an extension by ID and, optionally, its entire service/endpoint/credentials subtree
	GetExtension(ctx context.Context, extensionID string) (*Extension, error)
	// Retrieve an extension service by ID and, optionally, its entire endpoint/credentials subtree
	GetService(ctx context.Context, extensionID, serviceID string) (*ExtensionService, error)
	// Retrieve an extension endpoint by ID
	GetEndpoint(ctx context.Context, extensionID, serviceID, endpointURL string) (*ExtensionServiceEndpoint, error)
	// Retrieve a set of extension credentials by ID
	GetCredentials(ctx context.Context, extensionID, serviceID, credentialsID string) (*ExtensionServiceCredentials, error)
	// Update an extension
	UpdateExtension(ctx context.Context, extension *Extension) error
	// Update a service belonging to an extension
	UpdateService(ctx context.Context, extensionID string, service *ExtensionService) error
	// Update an endpoint belonging to a service
	UpdateEndpoint(ctx context.Context, extensionID string, serviceID string, endpoint *ExtensionServiceEndpoint) error
	// Update a set of credentials belonging to a service
	UpdateCredentials(ctx context.Context, extensionID string, serviceID string, credentials *ExtensionServiceCredentials) error
	// Remove an extension from the registry, along with all its services, endpoints and credentials
	RemoveExtension(ctx context.Context, extensionID string) error
	// Remove an extension service from the registry, along with all its endpoints and credentials
	RemoveService(ctx context.Context, extensionID, serviceID string) error
	// Remove an extension endpoint from the registry
	RemoveEndpoint(ctx context.Context, extensionID, serviceID, endpointID string) error
	// Remove a set of extension credentials from the registry
	RemoveCredentials(ctx context.Context, extensionID, serviceID, credentialsID string) error
	// Run a query on the extension registry to find one or more ways to access extensions matching given search parameters
	GetExtensionAccessDescriptors(ctx context.Context, query *ExtensionQuery) ([]*ExtensionAccessDescriptor, error)
}

// ExtensionStore defines the interface required to store extensions.
type ExtensionStore interface {
	// AddExtension adds a new extension to the store.
	AddExtension(ctx context.Context, extension *Extension) (*Extension, error)
	// GetExtension retrieves an extension by its ID.
	GetExtension(ctx context.Context, extensionID string) (*Extension, error)
	// ListExtensions retrieves all stored extensions.
	ListExtensions(ctx context.Context, query *ExtensionQuery) []*Extension
	// UpdateExtension updates an existing extension.
	UpdateExtension(ctx context.Context, newExtension *Extension) error
	// DeleteExtension deletes an extension from the store.
	DeleteExtension(ctx context.Context, extensionID string) error
	// AddExtensionService adds a new extension service to an extension.
	AddExtensionService(ctx context.Context, extensionID string, service *ExtensionService) (*ExtensionService, error)
	// GetExtensionService retrieves an extension service by its ID.
	GetExtensionService(ctx context.Context, extensionID string, serviceID string) (*ExtensionService, error)
	// ListExtensionServices retrieves all services belonging to an extension.
	ListExtensionServices(ctx context.Context, extensionID string) ([]*ExtensionService, error)
	// UpdateExtensionService updates a service belonging to an extension.
	UpdateExtensionService(ctx context.Context, extensionID string, newService *ExtensionService) error
	// DeleteExtensionService deletes an extension service from an extension.
	DeleteExtensionService(ctx context.Context, extensionID, serviceID string) error
	// AddExtensionServiceEndpoint adds a new endpoint to an extension service.
	AddExtensionServiceEndpoint(ctx context.Context, extensionID string, serviceID string, endpoint *ExtensionServiceEndpoint) (*ExtensionServiceEndpoint, error)
	// GetExtensionServiceEndpoint retrieves an extension endpoint by its ID.
	GetExtensionServiceEndpoint(ctx context.Context, extensionID string, serviceID string, endpointID string) (*ExtensionServiceEndpoint, error)
	// ListExtensionServiceEndpoints retrieves all endpoints belonging to an extension service.
	ListExtensionServiceEndpoints(ctx context.Context, extensionID string, serviceID string) ([]*ExtensionServiceEndpoint, error)
	// UpdateExtensionServiceEndpoint updates an endpoint belonging to an extension service.
	UpdateExtensionServiceEndpoint(ctx context.Context, extensionID string, serviceID string, newEndpoint *ExtensionServiceEndpoint) error
	// DeleteExtensionServiceEndpoint deletes an extension endpoint from an extension service.
	DeleteExtensionServiceEndpoint(ctx context.Context, extensionID, serviceID, endpointID string) error
	// AddExtensionServiceCredentials adds a new credential to an extension service.
	AddExtensionServiceCredentials(ctx context.Context, extensionID string, serviceID string, credentials *ExtensionServiceCredentials) (*ExtensionServiceCredentials, error)
	// GetExtensionServiceCredentials retrieves an extension credential by its ID.
	GetExtensionServiceCredentials(ctx context.Context, extensionID string, serviceID string, credentialsID string) (*ExtensionServiceCredentials, error)
	// ListExtensionServiceCredentials retrieves all credentials belonging to an extension service.
	ListExtensionServiceCredentials(ctx context.Context, extensionID string, serviceID string) ([]*ExtensionServiceCredentials, error)
	// UpdateExtensionServiceCredentials updates an extension credential.
	UpdateExtensionServiceCredentials(ctx context.Context, extensionID string, serviceID string, credentials *ExtensionServiceCredentials) (err error)
	// DeleteExtensionServiceCredentials deletes an extension credential from an extension service.
	DeleteExtensionServiceCredentials(ctx context.Context, extensionID, serviceID, credentialsID string) error
	// GetExtensionAccessDescriptors retrieves access descriptors belonging to an extension that match the query.
	GetExtensionAccessDescriptors(ctx context.Context, query *ExtensionQuery) (result []*ExtensionAccessDescriptor, err error)
}

// EnsureID sets the ID of the extension when not set.
func (e *Extension) EnsureID(ctx context.Context, store ExtensionStore) {
	if e.ID == "" {
		e.ID = e.generateExtensionID(ctx, store)
	}
}

// AddService adds a new service to the extension.
func (e *Extension) AddService(service *ExtensionService) (*ExtensionService, error) {
	service.EnsureID(e)

	if e.Services == nil {
		e.Services = make(map[string]*ExtensionService)
	} else {
		if _, err := e.GetService(service.ID); err == nil {
			return nil, NewErrExtensionServiceExists(e.ID, service.ID)
		}
	}

	service.SetCreated(time.Now())
	e.Services[service.ID] = service

	return service, nil
}

// AddEndpoint adds a new endpoint to a service.
func (e *Extension) AddEndpoint(serviceID string, endpoint *ExtensionServiceEndpoint) (*ExtensionServiceEndpoint, error) {
	service, error := e.GetService(serviceID)
	if error != nil {
		return nil, error
	}
	return service.AddEndpoint(endpoint)
}

// AddCredentials adds a new credential to a service.
func (e *Extension) AddCredentials(serviceID string, credential *ExtensionServiceCredentials) (*ExtensionServiceCredentials, error) {
	service, error := e.GetService(serviceID)
	if error != nil {
		return nil, error
	}
	return service.AddCredentials(credential)
}

// ListServices returns all services belonging to the extension.
func (e *Extension) ListServices() []*ExtensionService {
	if e.Services == nil {
		return []*ExtensionService{}
	}

	services := make([]*ExtensionService, 0, len(e.Services))
	for _, service := range e.Services {
		services = append(services, service)
	}

	return services
}

// GetService returns a service belonging to the extension.
func (e *Extension) GetService(serviceID string) (*ExtensionService, error) {
	if e.Services != nil {
		service, ok := e.Services[serviceID]
		if ok {
			return service, nil
		}
	}

	return nil, NewErrExtensionServiceNotFound(e.ID, serviceID)
}

// ListServiceEndpoints returns all endpoints belonging to a service.
func (e *Extension) ListServiceEndpoints(serviceID string) ([]*ExtensionServiceEndpoint, error) {
	service, error := e.GetService(serviceID)
	if error != nil {
		return nil, error
	}

	return service.ListEndpoints(), nil
}

// GetServiceEndpoint returns an endpoint belonging to a service.
func (e *Extension) GetServiceEndpoint(serviceID string, endpointID string) (*ExtensionServiceEndpoint, error) {
	service, err := e.GetService(serviceID)
	if err != nil {
		return nil, err
	}

	endpoint, err := service.GetEndpoint(endpointID)
	if err != nil {
		return nil, NewErrExtensionServiceEndpointNotFound(e.ID, serviceID, endpointID)
	}

	return endpoint, nil
}

// ListServiceCredentials returns all credentials belonging to a service.
func (e *Extension) ListServiceCredentials(serviceID string) ([]*ExtensionServiceCredentials, error) {
	service, err := e.GetService(serviceID)
	if err != nil {
		return nil, err
	}

	return service.ListCredentials(), nil
}

// GetServiceCredentials returns a credential belonging to a service.
func (e *Extension) GetServiceCredentials(serviceID string, credentialsID string) (*ExtensionServiceCredentials, error) {
	service, err := e.GetService(serviceID)
	if err != nil {
		return nil, err
	}

	credentials, err := service.GetCredentials(credentialsID)
	if err != nil {
		return nil, NewErrExtensionServiceCredentialsNotFound(e.ID, serviceID, credentialsID)
	}

	return credentials, nil
}

// UpdateService updates a service from the extension.
func (e *Extension) UpdateService(newService *ExtensionService) error {
	service, err := e.GetService(newService.ID)
	if err != nil {
		return err
	}

	newService.Created = service.Created
	newService.Updated = time.Now()

	for _, endpoint := range newService.ListEndpoints() {
		_, err := service.GetEndpoint(endpoint.URL)
		if err != nil {
			// If the endpoint is new, set its creation time
			endpoint.Created = newService.Updated
			endpoint.Updated = newService.Updated
		}
	}

	for _, credential := range newService.ListCredentials() {
		_, err := service.GetCredentials(credential.ID)
		if err != nil {
			// If the credential is new, set its creation time
			credential.Created = newService.Updated
			credential.Updated = newService.Updated
		}
	}

	e.Services[service.ID] = newService
	return nil
}

// UpdateServiceEndpoint updates an endpoint from a service.
func (e *Extension) UpdateServiceEndpoint(serviceID string, endpoint *ExtensionServiceEndpoint) error {
	service, err := e.GetService(serviceID)
	if err != nil {
		return err
	}

	return service.UpdateEndpoint(endpoint)
}

// UpdateServiceCredentials updates a credential from a service.
func (e *Extension) UpdateServiceCredentials(serviceID string, credentials *ExtensionServiceCredentials) error {
	service, err := e.GetService(serviceID)
	if err != nil {
		return err
	}

	return service.UpdateCredentials(credentials)
}

// DeleteService deletes a service from the extension.
func (e *Extension) DeleteService(serviceID string) error {
	_, err := e.GetService(serviceID)
	if err != nil {
		return err
	}

	delete(e.Services, serviceID)
	return nil
}

// DeleteServiceEndpoint deletes an endpoint from a service.
func (e *Extension) DeleteServiceEndpoint(serviceID string, endpointID string) error {
	service, err := e.GetService(serviceID)
	if err != nil {
		return err
	}

	return service.DeleteEndpoint(endpointID)
}

// DeleteServiceCredentials deletes a credential from a service.
func (e *Extension) DeleteServiceCredentials(serviceID string, credentialsID string) error {
	service, err := e.GetService(serviceID)
	if err != nil {
		return err
	}

	return service.DeleteCredentials(credentialsID)
}

// GetExtensionIfMatch returns an extension where the extension, services, endpoints and credentials match the given query.
func (e *Extension) GetExtensionIfMatch(query *ExtensionQuery) *Extension {
	if query.ExtensionID != "" && query.ExtensionID != e.ID {
		return nil
	}
	if query.Zone != "" && query.StrictZoneMatch && query.Zone != e.Zone {
		return nil
	}
	if query.Product != "" && query.Product != e.Product {
		return nil
	}
	if query.VersionConstraints != "" && query.VersionConstraints != e.Version {
		// try interpreting the version as a semantic version constraint

		version, err := semver.NewVersion(e.Version)
		if err != nil {
			// Version not parseable or doesn't respect semantic versioning format
			return nil
		}

		constraints, err := semver.NewConstraint(query.VersionConstraints)
		if err != nil {
			// Constraints not parseable or not a constraints string
			return nil
		}

		// Check if the version meets the constraints
		if !constraints.Check(version) {
			return nil
		}
	}

	extensionCopy := *e
	if query.ServiceID != "" || query.ServiceResource != "" || query.ServiceCategory != "" ||
		query.EndpointURL != "" || query.Type != nil || query.CredentialsID != "" || query.CredentialsScope != "" ||
		query.User != "" || query.Project != "" {
		services, err := e.FindServices(query)
		if err != nil || len(services) == 0 {
			return nil
		}
		extensionCopy.Services = services
	}
	return &extensionCopy
}

// GetAccessDescriptors returns access descriptors for the extension.
func (e *Extension) GetAccessDescriptors() []*ExtensionAccessDescriptor {
	result := make([]*ExtensionAccessDescriptor, 0)

	for _, service := range e.ListServices() {
		for _, endpoint := range service.Endpoints {
			if len(service.Credentials) > 0 || service.AuthRequired {
				for _, credential := range service.Credentials {
					accessDesc := ExtensionAccessDescriptor{
						Extension:   *e,
						Service:     *service,
						Endpoint:    *endpoint,
						Credentials: credential,
					}
					result = append(result, &accessDesc)
				}
			} else {
				accessDesc := ExtensionAccessDescriptor{
					Extension:   *e,
					Service:     *service,
					Endpoint:    *endpoint,
					Credentials: nil,
				}
				result = append(result, &accessDesc)
			}
		}
	}

	return result
}

// FindServices returns services belonging to the extension that match the query.
func (e *Extension) FindServices(query *ExtensionQuery) (map[string]*ExtensionService, error) {
	result := make(map[string]*ExtensionService)

	addIfMatch := func(service *ExtensionService) {
		if query.ServiceID != "" && query.ServiceID != service.ID {
			return
		}
		if query.ServiceResource != "" && query.ServiceResource != service.Resource {
			return
		}
		if query.ServiceCategory != "" && query.ServiceCategory != service.Category {
			return
		}

		endpoints, err := service.FindEndpoints(e.Zone, query)
		if err == nil {
			credentials, err := service.FindCredentials(query)
			if err == nil {
				// Create a copy of the service but containing only the endpoints and credentials that match the query
				serviceCopy := *service
				serviceCopy.Endpoints = endpoints
				serviceCopy.Credentials = credentials
				result[service.ID] = &serviceCopy
			}
		}
	}

	for _, service := range e.ListServices() {
		addIfMatch(service)
	}

	return result, nil
}

// SetCreated sets the created time of the extension and its services, credentials, endpoints.
func (e *Extension) SetCreated(ctx context.Context) {
	e.Created = time.Now()
	e.Updated = e.Created

	for _, service := range e.ListServices() {
		service.SetCreated(e.Created)
	}
}

// simple unique extension ID generator
func (e *Extension) generateExtensionID(ctx context.Context, store ExtensionStore) string {
	prefix := e.Product
	if prefix != "" {
		prefix = prefix + "-"
	}
	for {
		ID := prefix + rand.String(8)
		if _, err := store.GetExtension(ctx, ID); err != nil {
			return ID
		}
	}
}

// simple unique extension service ID generator
func (e *Extension) generateExtensionServiceID(service *ExtensionService) string {
	prefix := service.Resource
	if prefix == "" && e.Product != "" {
		prefix = e.Product + "-service"
	}
	if prefix != "" {
		prefix = prefix + "-"
	}
	for {
		ID := prefix + rand.String(8)
		if e.Services[ID] == nil {
			return ID
		}
	}
}

// EnsureID sets the ID of the service when empty.
func (es *ExtensionService) EnsureID(ext *Extension) {
	if es.ID == "" {
		es.ID = ext.generateExtensionServiceID(es)
	}
}

// ListEndpoints returns all endpoints belonging to the service.
func (es *ExtensionService) ListEndpoints() []*ExtensionServiceEndpoint {
	if es.Endpoints == nil {
		return make([]*ExtensionServiceEndpoint, 0)
	}

	endpoints := make([]*ExtensionServiceEndpoint, 0, len(es.Endpoints))
	for _, endpoint := range es.Endpoints {
		endpoints = append(endpoints, endpoint)
	}

	return endpoints
}

// GetEndpoint returns the endpoint with the given ID.
func (es *ExtensionService) GetEndpoint(endpointID string) (*ExtensionServiceEndpoint, error) {
	if es.Endpoints != nil {
		endpoint, ok := es.Endpoints[endpointID]
		if ok {
			return endpoint, nil
		}
	}

	return nil, NewErrExtensionServiceEndpointNotFound("", es.ID, endpointID)
}

// GetCredentials returns the credential with the given ID.
func (es *ExtensionService) GetCredentials(credentialsID string) (*ExtensionServiceCredentials, error) {
	if es.Credentials != nil {
		credential, ok := es.Credentials[credentialsID]
		if ok {
			return credential, nil
		}
	}

	return nil, NewErrExtensionServiceCredentialsNotFound("", es.ID, credentialsID)
}

// ListCredentials returns all credentials belonging to the service.
func (es *ExtensionService) ListCredentials() []*ExtensionServiceCredentials {
	if es.Credentials == nil {
		return make([]*ExtensionServiceCredentials, 0)
	}

	result := make([]*ExtensionServiceCredentials, 0, len(es.Credentials))
	for _, credentials := range es.Credentials {
		result = append(result, credentials)
	}
	return result
}

// AddEndpoint adds the given endpoint to the service.
func (es *ExtensionService) AddEndpoint(endpoint *ExtensionServiceEndpoint) (*ExtensionServiceEndpoint, error) {
	if es.Endpoints == nil {
		es.Endpoints = make(map[string]*ExtensionServiceEndpoint)
	}

	if es.Endpoints[endpoint.URL] != nil {
		return nil, NewErrExtensionServiceEndpointExists("", es.ID, endpoint.URL)
	}

	endpoint.Created = time.Now()
	endpoint.Updated = endpoint.Created

	es.Endpoints[endpoint.URL] = endpoint
	return endpoint, nil
}

// AddCredentials adds the given credential to the service.
func (es *ExtensionService) AddCredentials(credentials *ExtensionServiceCredentials) (*ExtensionServiceCredentials, error) {
	credentials.EnsureID(es)

	if es.Credentials == nil {
		es.Credentials = make(map[string]*ExtensionServiceCredentials)
	} else {
		if _, err := es.GetCredentials(credentials.ID); err == nil {
			return nil, NewErrExtensionServiceCredentialsExists("", es.ID, credentials.ID)
		}
	}

	credentials.Created = time.Now()
	credentials.Updated = credentials.Created

	es.Credentials[credentials.ID] = credentials
	return credentials, nil
}

// UpdateEndpoint updates an endpoint in the service.
func (es *ExtensionService) UpdateEndpoint(newEndpoint *ExtensionServiceEndpoint) error {
	endpoint, err := es.GetEndpoint(newEndpoint.URL)
	if err != nil {
		return err
	}

	newEndpoint.Created = endpoint.Created
	newEndpoint.Updated = time.Now()

	es.Endpoints[endpoint.URL] = newEndpoint
	return nil
}

// UpdateCredentials updates a credential in the service.
func (es *ExtensionService) UpdateCredentials(newCredentials *ExtensionServiceCredentials) error {
	credential, err := es.GetCredentials(newCredentials.ID)
	if err != nil {
		return err
	}

	newCredentials.Created = credential.Created
	newCredentials.Updated = time.Now()

	es.Credentials[credential.ID] = newCredentials
	return nil
}

// DeleteEndpoint deletes an endpoint from the service.
func (es *ExtensionService) DeleteEndpoint(endpointID string) error {
	_, err := es.GetEndpoint(endpointID)
	if err != nil {
		return err
	}

	delete(es.Endpoints, endpointID)
	return nil
}

// DeleteCredentials deletes a credential from the service.
func (es *ExtensionService) DeleteCredentials(credentialsID string) error {
	_, err := es.GetCredentials(credentialsID)
	if err != nil {
		return err
	}

	delete(es.Credentials, credentialsID)
	return nil
}

// FindEndpoints returns all endpoints matching the given query.
func (es *ExtensionService) FindEndpoints(extensionZone string, query *ExtensionQuery) (map[string]*ExtensionServiceEndpoint, error) {
	result := make(map[string]*ExtensionServiceEndpoint)

	addIfMatch := func(endpoint *ExtensionServiceEndpoint) {
		if query.EndpointURL != "" && query.EndpointURL != endpoint.URL {
			return
		}
		if query.Type != nil {
			// match endpoint type, if supplied with the query
			if *query.Type != endpoint.Type {
				return
			}
		} else if query.Zone != "" {
			// if endpoint type is not supplied with the query, determine the endpoint type
			// by comparing the query zone (if supplied) with the extension record zone
			//  - if the extension is in the same zone as the query, both internal and external endpoints will match
			//  - otherwise, only external endpoints will match
			if query.Zone != extensionZone && endpoint.Type == EETInternal {
				return
			}
		}
		endpointCopy := *endpoint
		result[endpointCopy.URL] = &endpointCopy
	}

	for _, endpoint := range es.ListEndpoints() {
		addIfMatch(endpoint)
	}

	return result, nil
}

// FindCredentials returns all credentials matching the given query.
func (es *ExtensionService) FindCredentials(query *ExtensionQuery) (map[string]*ExtensionServiceCredentials, error) {
	result := make(map[string]*ExtensionServiceCredentials)
	addIfMatch := func(credentials *ExtensionServiceCredentials) {
		if query.CredentialsID != "" && query.CredentialsID != credentials.ID {
			return
		}
		// if the query is for global scoped credentials, only credentials with a global scope match
		if query.CredentialsScope == ECSGlobal && credentials.Scope != ECSGlobal {
			return
		}
		// if the query is for project scoped credentials, only credentials with a global scope
		// and those project scoped to the supplied project match
		if query.CredentialsScope == ECSProject {
			if credentials.Scope == ECSUser {
				return
			}
			if credentials.Scope == ECSProject && !util.StringInSlice(query.Project, credentials.Projects) {
				return
			}
		}
		// if the query is for a user scoped credential, only credentials with a global scope,
		// those project scoped to the supplied project, and those user scoped to the supplied user and project
		// are a match
		if query.CredentialsScope == ECSUser {
			if credentials.Scope != ECSGlobal && query.Project != "" && !util.StringInSlice(query.Project, credentials.Projects) {
				return
			}
			if credentials.Scope == ECSUser && !util.StringInSlice(query.User, credentials.Users) {
				return
			}
		}
		credentialCopy := *credentials
		result[credentialCopy.ID] = &credentialCopy
	}

	for _, credential := range es.ListCredentials() {
		addIfMatch(credential)
	}

	return result, nil
}

// SetCreated sets the created time of the extension service, endpoint and credentials.
func (es *ExtensionService) SetCreated(date time.Time) {
	es.Created = date
	es.Updated = date

	for _, endpoint := range es.ListEndpoints() {
		endpoint.Created = date
		endpoint.Updated = date
	}
	for _, credential := range es.ListCredentials() {
		credential.Created = date
		credential.Updated = date
	}
}

// simple unique extension credentials ID generator.
func (es *ExtensionService) generateExtensionCredentialsID(credentials *ExtensionServiceCredentials) string {
	prefix := es.Resource
	if prefix == "" {
		prefix = "creds"
	}
	if prefix != "" {
		prefix = prefix + "-"
	}
	for {
		ID := prefix + rand.String(8)
		if es.Credentials[ID] == nil {
			return ID
		}
	}
}

// EnsureID sets the ID of the credential if empty.
func (ec *ExtensionServiceCredentials) EnsureID(svc *ExtensionService) {
	if ec.ID == "" {
		ec.ID = svc.generateExtensionCredentialsID(ec)
	}
}
