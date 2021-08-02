package domain

import (
	"context"
)

// ExtensionEndpointType is the type used for the ExtensionEndpoint EndpointType field
type ExtensionEndpointType string

// Valid values that can be used with ExtensionEndpointType
const (
	EETInternal ExtensionEndpointType = "internal"
	EETExternal                       = "external"
)

// ExtensionCredentialScope is the type used for the ExtensionCredentials Scope field
type ExtensionCredentialScope string

// Valid values that can be used with ExtensionCredentialScope
const (
	ECSGlobal  ExtensionCredentialScope = "global"
	ECSProject                          = "project"
	ECSUser                             = "user"
)

// ExtensionID is a unique extension identifier
type ExtensionID struct {
	// Extension ID - used to uniquely identify an extension in the registry
	ExtensionID string
}

// Extension is an entry in the extension registry that describes a particular installation of a
// framework/platform/service/product developed and released or hosted under a unique product name
type Extension struct {
	// Extension ID - used to uniquely identify an extension in the registry
	ExtensionID
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
}

// ExtensionServiceID is a unique extension service identifier
type ExtensionServiceID struct {
	// Extension ID - references the extension this service belongs to
	ExtensionID string
	// Extension service ID - used to uniquely identify a service within the scope of an extension
	ServiceID string
}

// ExtensionService is a service provided by an extension. A service is represented by a
// single API or UI. For extensions implemented as cloud-native applications, a service is the
// equivalent of a kubernetes service that is used to expose a public API or UI. Services are classified
// into known resource types (e.g. s3, git) encoded via the Resource attribute and service
// categories (e.g. model store, feature store, distributed training, serving) via the Category attribute
type ExtensionService struct {
	// Extension service ID - used to uniquely identify an extension service in the registry
	ExtensionServiceID
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
	// List of endpoints
	Endpoints []*ExtensionEndpoint
	// Configuration entries (e.g. configuration values required to configure the client to access this
	// service), expressed as set of key-value entries
	Configuration map[string]string
}

// ExtensionEndpointID is a unique extension endpoint identifier
type ExtensionEndpointID struct {
	// Extension ID - references the extension this endpoint belongs to
	ExtensionID string
	// Extension service ID - references the extension service this endpoint belongs to
	ServiceID string
	// Endpoint URL. In case of k8s controllers and operators, the URL points to the cluster API.
	// Also used to uniquely identifies an endpoint within the scope of a service
	URL string
}

// ExtensionEndpoint is an endpoint through which an extension service can be accessed. Having a list of
// endpoints associated with a single extension service is particularly important for representing k8s
// services, which can be exposed both internally (cluster IP) and externally (e.g. ingress). All endpoints
// belonging to the same extension service must be equivalent in the sense that they are backed by the same
// API and/or protocol and exhibit the same behavior
type ExtensionEndpoint struct {
	// Extension endpoint ID - used to uniquely identify an extension endpoint in the registry
	ExtensionEndpointID
	// Endpoint type - internal/external. An internal endpoint can only be accessed when the consumer
	// is located in the same zone as the extension service
	EndpointType ExtensionEndpointType
	// Configuration entries (e.g. CA certificates), expressed as set of key-value entries
	Configuration map[string]string
}

// ExtensionCredentialsID is a unique extension credentials identifier
type ExtensionCredentialsID struct {
	// Extension ID - references the extension this set of credentials belongs to
	ExtensionID string
	// Extension service ID - references the extension service this set of credentials belongs to
	ServiceID string
	// Extension credentials ID - used to uniquely identify a set of credentials within the scope
	// of an extension service
	CredentialsID string
}

// ExtensionCredentials is a group of configuration values that can be generally used to embed information
// pertaining to the authentication and authorization features supported by a service. This descriptor allows
// administrators and operators of 3rd party tools integrated with FuseML to configure different accounts
// and credentials (tokens, certificates, passwords etc.) to be associated with different FuseML organization
// entities (users, projects, groups etc.). All information embedded in a credentials descriptor entry is
// treated as sensitive information. Each credentials entry has an associated scope that controls who has
// access to this information (e.g. global, project, user, workflow). This is the equivalent of a k8s secret.
type ExtensionCredentials struct {
	// Extension credentials ID - used to uniquely identify a set of credentials in the registry
	ExtensionCredentialsID
	// The scope associated with this set of credentials. Global scoped credentials can be used by any
	// user/project. Project scoped credentials can be used only in the context of one of the projects
	// supplied in the Projects list. User scoped credentials can only be used by the users in the Users
	// list and, optionally, in the context of the projects supplied in the Projects list.
	Scope ExtensionCredentialScope
	// Use as default credentials. Used to automatically select one of several credentials with the same
	// scope matching the same query.
	Default bool
	// List of projects allowed to use these credentials
	Projects []string
	// List of users allowed to use these credentials
	Users []string
	// Configuration entries (e.g. usernames, passwords, tokens, keys), expressed as set of key-value entries
	Configuration map[string]string
}

// ExtensionRecord is used to associate an extension with a list of all provided services along with the
// endpoints and credentials that can be used to access them
type ExtensionRecord struct {
	Extension
	// Map of services associated with the extension, indexed by service ID
	Services map[string]*ExtensionServiceRecord
}

// ExtensionServiceRecord is used to associate an extension service with a list of endpoints that can be used
// to access the service and a set of credentials configured for it
type ExtensionServiceRecord struct {
	ExtensionService
	// Map of endpoints associated with the service, indexed by endpoint URL
	Endpoints map[string]*ExtensionEndpoint
	// Map of credentials associated with the service, indexed by credentials ID
	Credentials map[string]*ExtensionCredentials
}

// ExtensionAccessDescriptor is a structure that contains all the information needed to access an extension:
// a service, an endpoint and an optional set of credentials. It's returned as result when running access queries
// against the extension registry.
type ExtensionAccessDescriptor struct {
	Extension
	ExtensionService
	ExtensionEndpoint
	*ExtensionCredentials
}

// ExtensionAccessQuery is a query that can be run against the extension registry to retrieve
// a list of extension endpoints and credentials that meet all supplied criteria
type ExtensionAccessQuery struct {
	// Search by explicit extension ID
	ExtensionID string
	// Search by product name. Leave empty to include all products
	Product string
	// Search by explicit service ID
	ServiceID string
	// Search by service resource type. Leave empty to include all resource types
	ServiceResource string
	// Search by service category. Leave empty to include all services
	ServiceCategory string
	// Match extensions installed in a given zone.
	Zone string
	// Use strict filtering when a zone query field is supplied. When set, only extensions
	// installed in the supplied zone are returned.
	StrictZoneMatch bool
	// Search by explicit endpoint URL
	EndpointURL string
	// Search by explicit credentials ID
	CredentialsID string
	// Match credentials by scope
	CredentialsScope ExtensionCredentialScope
	// Match credentials allowed for a given user. CredentialsScope must be set to ECSUser
	// for this to have effect
	User string
	// Match credentials allowed for a given project. CredentialsScope must be set to ECSUser
	// or ECSProject for this to have effect
	Project string
}

// ExtensionRegistry defines the public interface implemented by the extension registry
type ExtensionRegistry interface {
	// Register a new extension, with all participating services, endpoints and credentials
	RegisterExtension(ctx context.Context, extension *ExtensionRecord) (result *ExtensionRecord, err error)
	// Add a service to an existing extension
	AddService(ctx context.Context, service *ExtensionService) (result *ExtensionService, err error)
	// Add an endpoint to an existing extension service
	AddEndpoint(ctx context.Context, endpoint *ExtensionEndpoint) (result *ExtensionEndpoint, err error)
	// Add a set of credentials to an existing extension service
	AddCredentials(ctx context.Context, credentials *ExtensionCredentials) (result *ExtensionCredentials, err error)
	// Retrieve an extension by ID
	GetExtension(ctx context.Context, ID ExtensionID) (result *ExtensionRecord, err error)
	// Retrieve an extension service by ID
	GetService(ctx context.Context, ID ExtensionServiceID) (result *ExtensionServiceRecord, err error)
	// Retrieve an extension endpoint by ID
	GetEndpoint(ctx context.Context, ID ExtensionEndpointID) (result *ExtensionEndpoint, err error)
	// Retrieve a set of extension credentials by ID
	GetCredentials(ctx context.Context, ID ExtensionCredentialsID) (result *ExtensionCredentials, err error)
	// Remove an extension from the registry
	RemoveExtension(ctx context.Context, ID ExtensionID) error
	// Remove an extension service from the registry
	RemoveService(ctx context.Context, ID ExtensionServiceID) error
	// Remove an extension endpoint from the registry
	RemoveEndpoint(ctx context.Context, ID ExtensionEndpointID) error
	// Remove a set of extension credentials from the registry
	RemoveCredentials(ctx context.Context, ID ExtensionCredentialsID) error
	// Run a query on the extension registry to find one or more ways to access extensions matching given search parameters
	RunAccessQuery(ctx context.Context, query *ExtensionAccessQuery) (result []*ExtensionAccessDescriptor, err error)
}

// ExtensionRegistryPersistentStorage defines the interface implemented by the extension registry persistent storage backend
type ExtensionRegistryPersistentStorage interface {
	// Add an extension
	CreateExtension(ctx context.Context, extension *Extension) (result *Extension, err error)
	// Add an extension service
	CreateService(ctx context.Context, service *ExtensionService) (result *ExtensionService, err error)
	// Add an extension endpoint
	CreateEndpoint(ctx context.Context, endpoint *ExtensionEndpoint) (result *ExtensionEndpoint, err error)
	// Add a set of extension credentials
	CreateCredentials(ctx context.Context, credentials *ExtensionCredentials) (result *ExtensionCredentials, err error)
	// Retrieve an extension by ID
	GetExtension(ctx context.Context, ID ExtensionID) (result *Extension, err error)
	// Retrieve an extension service by ID
	GetService(ctx context.Context, ID ExtensionServiceID) (result *ExtensionService, err error)
	// Retrieve an extension endpoint by ID
	GetEndpoint(ctx context.Context, ID ExtensionEndpointID) (result *ExtensionEndpoint, err error)
	// Retrieve a set of extension credentials by ID
	GetCredentials(ctx context.Context, ID ExtensionCredentialsID) (result *ExtensionCredentials, err error)
	// Remove an extension
	DeleteExtension(ctx context.Context, ID ExtensionID) error
	// Remove an extension service
	DeleteService(ctx context.Context, ID ExtensionServiceID) error
	// Remove an extension endpoint
	DeleteEndpoint(ctx context.Context, ID ExtensionEndpointID) error
	// Remove a set of extension credentials
	DeleteCredentials(ctx context.Context, ID ExtensionCredentialsID) error

	// Find extensions matching an ID or product
	FindExtensions(ctx context.Context, ID ExtensionID, product string) (result []*Extension, err error)
	// Find services matching an ID, zone, service name or category
	FindServices(ctx context.Context, ID ExtensionServiceID, zone, service, serviceCategory string) (result []*ExtensionService, err error)
	// Find endpoints matching an ID or type
	FindEndpoints(ctx context.Context, ID ExtensionEndpointID, internal *bool) (result *ExtensionEndpoint, err error)
	// Find endpoints matching an ID or scope
	FindCredentials(ctx context.Context, ID ExtensionCredentialsID, CredentialsScope ExtensionCredentialScope, user, project string) (result *ExtensionCredentials, err error)
}
