package design

import (
	"time"

	. "goa.design/goa/v3/dsl"
)

var _ = Service("extension", func() {
	Description("The extension registry service interfaces with the FuseML Extension Registry.")

	Method("registerExtension", func() {
		Description("Register an external utility as a FuseML extension with the FuseML extension registry.")

		Payload(Extension, "Extension registration request")

		Error("BadRequest", func() {
			Description("If the extension does not have the required fields, should return 400 Bad Request.")
		})
		Error("Conflict", func() {
			Description("If an extension with the same name already exists, should return 409 Conflict.")
		})

		Result(Extension)

		HTTP(func() {
			POST("/extensions")
			Response(StatusCreated)
			Response("BadRequest", StatusBadRequest)
			Response("Conflict", StatusConflict)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("BadRequest", CodeInvalidArgument)
			Response("Conflict", CodeAlreadyExists)
		})
	})

	Method("getExtension", func() {
		Description("Retrieve information about an extension.")

		Payload(func() {
			Field(1, "id", String, "Extension identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("kfserving-001")
			})
			Required("id")
		})

		Error("NotFound", func() {
			Description("If there is no extension with the given ID, should return 404 Not Found.")
		})

		Result(Extension)

		HTTP(func() {
			GET("/extensions/{id}")
			Response(StatusOK)
			Response("NotFound", StatusNotFound)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
		})
	})

	Method("listExtensions", func() {
		Description("List extensions registered in FuseML")

		Result(ArrayOf(Extension), "Return all registered extensions.")

		HTTP(func() {
			GET("/extensions")
			Response(StatusOK)
		})

		GRPC(func() {
			Response(CodeOK)
		})
	})

	Method("updateExtension", func() {
		Description("Update an extension registered in FuseML")

		Payload(Extension, "Extension update request")

		Result(Extension, "Return the updated extension.")

		Error("BadRequest", func() {
			Description("If the extension does not have the required fields, should return 400 Bad Request.")
		})

		Error("NotFound", func() {
			Description("If the extension is not found, should return 404 Not Found.")
		})

		HTTP(func() {
			PUT("/extensions/{id}")
			Response(StatusOK)
			Response("NotFound", StatusNotFound)
			Response("BadRequest", StatusBadRequest)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
			Response("BadRequest", CodeInvalidArgument)
		})
	})

	Method("deleteExtension", func() {
		Description("Delete an extension and its subtree of services, endpoints and credentials")
		Payload(func() {
			Field(1, "id", String, "Extension identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("kfserving-001")
			})
			Required("id")
		})

		Error("BadRequest", func() {
			Description("If the extension cannot be deleted, should return 400 Bad Request.")
		})

		Error("NotFound", func() {
			Description("If the extension is not found, should return 404 Not Found.")
		})

		HTTP(func() {
			DELETE("/extensions/{id}")
			Response(StatusNoContent)
			Response("NotFound", StatusNotFound)
			Response("BadRequest", StatusBadRequest)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
			Response("BadRequest", CodeInvalidArgument)
		})
	})

	Method("addService", func() {
		Description("Add a service to an existing extension registered with the FuseML extension registry.")

		Payload(ExtensionService, "Extension service add request")

		Error("NotFound", func() {
			Description("If the extension is not found, should return 404 Not Found.")
		})
		Error("BadRequest", func() {
			Description("If the service does not have the required fields, should return 400 Bad Request.")
		})
		Error("Conflict", func() {
			Description("If an extension service with the same name already exists, should return 409 Conflict.")
		})

		Result(ExtensionService)

		HTTP(func() {
			POST("/extensions/{extension_id}/services")
			Response(StatusCreated)
			Response("NotFound", StatusNotFound)
			Response("BadRequest", StatusBadRequest)
			Response("Conflict", StatusConflict)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
			Response("BadRequest", CodeInvalidArgument)
			Response("Conflict", CodeAlreadyExists)
		})
	})

	Method("getService", func() {
		Description("Retrieve information about a service belonging to an extension.")

		Payload(func() {
			Field(1, "extension_id", String, "Extension identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("kfserving-001")
			})
			Field(2, "id", String, "Uniquely identifies an extension service within the scope of an extension", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("s3")
			})
			Required("extension_id", "id")
		})

		Error("NotFound", func() {
			Description("If there is no extension or service with the given ID, should return 404 Not Found.")
		})

		Result(ExtensionService)

		HTTP(func() {
			GET("/extensions/{extension_id}/services/{id}")
			Response(StatusOK)
			Response("NotFound", StatusNotFound)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
		})
	})

	Method("listServices", func() {
		Description("List all services associated with an extension registered in FuseML")

		Payload(func() {
			Field(1, "extension_id", String, "Extension identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("kfserving-001")
			})
			Required("extension_id")
		})

		Result(ArrayOf(ExtensionService), "Return all services registered for an extension.")

		HTTP(func() {
			GET("/extensions/{extension_id}/services")
			Response(StatusOK)
		})

		GRPC(func() {
			Response(CodeOK)
		})
	})

	Method("updateService", func() {
		Description("Update a service belonging to an extension registered in FuseML")

		Payload(ExtensionService, "Extension service update request")

		Result(ExtensionService, "Return the updated extension service.")

		Error("BadRequest", func() {
			Description("If the extension service does not have the required fields, should return 400 Bad Request.")
		})

		Error("NotFound", func() {
			Description("If the extension or the service are not found, should return 404 Not Found.")
		})

		HTTP(func() {
			PUT("/extensions/{extension_id}/services/{id}")
			Response(StatusOK)
			Response("NotFound", StatusNotFound)
			Response("BadRequest", StatusBadRequest)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
			Response("BadRequest", CodeInvalidArgument)
		})
	})

	Method("deleteService", func() {
		Description("Delete an extension service and its subtree of endpoints and credentials")

		Payload(func() {
			Field(1, "extension_id", String, "Extension identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("kfserving-001")
			})
			Field(2, "id", String, "Uniquely identifies an extension service within the scope of an extension", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("s3")
			})
			Required("extension_id", "id")
		})

		Error("NotFound", func() {
			Description("If there is no extension or service with the given ID, should return 404 Not Found.")
		})

		Error("BadRequest", func() {
			Description("If the extension service cannot be deleted, should return 400 Bad Request.")
		})

		HTTP(func() {
			DELETE("/extensions/{extension_id}/services/{id}")
			Response(StatusNoContent)
			Response("NotFound", StatusNotFound)
			Response("BadRequest", StatusBadRequest)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
			Response("BadRequest", CodeInvalidArgument)
		})
	})

	Method("addEndpoint", func() {
		Description("Add an endpoint to an existing extension service registered with the FuseML extension registry.")

		Payload(ExtensionEndpoint, "Extension endpoint add request")

		Error("NotFound", func() {
			Description("If the extension or service are not found, should return 404 Not Found.")
		})
		Error("BadRequest", func() {
			Description("If the endpoint does not have the required fields, should return 400 Bad Request.")
		})
		Error("Conflict", func() {
			Description("If an extension endpoint with the same URL already exists, should return 409 Conflict.")
		})

		Result(ExtensionEndpoint)

		HTTP(func() {
			POST("/extensions/{extension_id}/services/{service_id}/endpoints")
			Response(StatusCreated)
			Response("NotFound", StatusNotFound)
			Response("BadRequest", StatusBadRequest)
			Response("Conflict", StatusConflict)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
			Response("BadRequest", CodeInvalidArgument)
			Response("Conflict", CodeAlreadyExists)
		})
	})

	Method("getEndpoint", func() {
		Description("Retrieve information about an endpoint belonging to an extension.")

		Payload(func() {
			Field(1, "extension_id", String, "Extension identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("kfserving-001")
			})
			Field(2, "service_id", String, "Extension service identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("s3")
			})
			Field(3, "url", String, "Endpoint URL", func() {
				Format(FormatURI)
				MaxLength(200)
				Example("https://mlflow.10.120.130.140.nip.io")
			})
			Required("extension_id", "service_id", "url")
		})

		Error("NotFound", func() {
			Description("If there is no extension or service with the given ID, should return 404 Not Found.")
		})

		Result(ExtensionEndpoint)

		HTTP(func() {
			GET("/extensions/{extension_id}/services/{service_id}/endpoints/{url}")
			Response(StatusOK)
			Response("NotFound", StatusNotFound)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
		})
	})

	Method("listEndpoints", func() {
		Description("List all endpoints associated with an extension service registered in FuseML")

		Payload(func() {
			Field(1, "extension_id", String, "Extension identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("kfserving-001")
			})
			Field(2, "service_id", String, "Extension service identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("s3")
			})
			Required("extension_id", "service_id")
		})

		Result(ArrayOf(ExtensionEndpoint), "Return all endpoints associated with an extension service.")

		HTTP(func() {
			GET("/extensions/{extension_id}/services/{service_id}/endpoints")
			Response(StatusOK)
		})

		GRPC(func() {
			Response(CodeOK)
		})
	})

	Method("updateEndpoint", func() {
		Description("Update an endpoint belonging to an extension service registered in FuseML")

		Payload(ExtensionEndpoint, "Extension endpoint update request")

		Result(ExtensionEndpoint, "Return the updated endpoint.")

		Error("BadRequest", func() {
			Description("If the extension endpoint does not have the required fields, should return 400 Bad Request.")
		})

		Error("NotFound", func() {
			Description("If the extension, the service or the endpoint are not found, should return 404 Not Found.")
		})

		HTTP(func() {
			PUT("/extensions/{extension_id}/services/{service_id}/endpoints/{url}")
			Response(StatusOK)
			Response("NotFound", StatusNotFound)
			Response("BadRequest", StatusBadRequest)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
			Response("BadRequest", CodeInvalidArgument)
		})
	})

	Method("deleteEndpoint", func() {
		Description("Delete an extension endpoint")

		Payload(func() {
			Field(1, "extension_id", String, "Extension identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("kfserving-001")
			})
			Field(2, "service_id", String, "Extension service identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("s3")
			})
			Field(3, "url", String, "Endpoint URL", func() {
				Format(FormatURI)
				MaxLength(200)
				Example("https://mlflow.10.120.130.140.nip.io")
			})
			Required("extension_id", "service_id", "url")
		})

		Error("NotFound", func() {
			Description("If there is no extension, service or endpoint with the given ID and URL, should return 404 Not Found.")
		})

		Error("BadRequest", func() {
			Description("If the extension endpoint cannot be deleted, should return 400 Bad Request.")
		})

		HTTP(func() {
			DELETE("/extensions/{extension_id}/services/{service_id}/endpoints/{url}")
			Response(StatusNoContent)
			Response("NotFound", StatusNotFound)
			Response("BadRequest", StatusBadRequest)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
			Response("BadRequest", CodeInvalidArgument)
		})
	})

	Method("addCredentials", func() {
		Description("Add a set of credentials to an existing extension service registered with the FuseML extension registry.")

		Payload(ExtensionCredentials, "Extension credentials add request")

		Error("NotFound", func() {
			Description("If the extension or service are not found, should return 404 Not Found.")
		})
		Error("BadRequest", func() {
			Description("If the credentials do not have the required fields, should return 400 Bad Request.")
		})
		Error("Conflict", func() {
			Description("If a set of extension credentials with the same ID already exists, should return 409 Conflict.")
		})

		Result(ExtensionCredentials)

		HTTP(func() {
			POST("/extensions/{extension_id}/services/{service_id}/credentials")
			Response(StatusCreated)
			Response("NotFound", StatusNotFound)
			Response("BadRequest", StatusBadRequest)
			Response("Conflict", StatusConflict)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
			Response("BadRequest", CodeInvalidArgument)
			Response("Conflict", CodeAlreadyExists)
		})
	})

	Method("getCredentials", func() {
		Description("Retrieve information about a set of credentials belonging to an extension.")

		Payload(func() {
			Field(1, "extension_id", String, "Extension identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("kfserving-001")
			})
			Field(2, "service_id", String, "Extension service identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("s3")
			})
			Field(3, "id", String, "Uniquely identifies a set of credentials within the scope of an extension service", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("cred-user-12bb")
			})
			Required("extension_id", "service_id", "id")
		})

		Error("NotFound", func() {
			Description("If there is no extension, service or set of credentials with the given ID, should return 404 Not Found.")
		})

		Result(ExtensionCredentials)

		HTTP(func() {
			GET("/extensions/{extension_id}/services/{service_id}/credentials/{id}")
			Response(StatusOK)
			Response("NotFound", StatusNotFound)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
		})
	})

	Method("listCredentials", func() {
		Description("List all credentials associated with an extension service registered in FuseML")

		Payload(func() {
			Field(1, "extension_id", String, "Extension identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("kfserving-001")
			})
			Field(2, "service_id", String, "Extension service identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("s3")
			})
			Required("extension_id", "service_id")
		})

		Result(ArrayOf(ExtensionCredentials), "Return all credentials associated with an extension service.")

		HTTP(func() {
			GET("/extensions/{extension_id}/services/{service_id}/credentials")
			Response(StatusOK)
		})

		GRPC(func() {
			Response(CodeOK)
		})
	})

	Method("updateCredentials", func() {
		Description("Update a set of credentials belonging to an extension service registered in FuseML")

		Payload(ExtensionCredentials, "Extension credentials update request")

		Result(ExtensionCredentials, "Return the updated set of credentials.")

		Error("BadRequest", func() {
			Description("If the set of extension credentials does not have the required fields, should return 400 Bad Request.")
		})

		Error("NotFound", func() {
			Description("If the extension, the service or the set of credentials are not found, should return 404 Not Found.")
		})

		HTTP(func() {
			PUT("/extensions/{extension_id}/services/{service_id}/credentials/{id}")
			Response(StatusOK)
			Response("NotFound", StatusNotFound)
			Response("BadRequest", StatusBadRequest)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
			Response("BadRequest", CodeInvalidArgument)
		})
	})

	Method("deleteCredentials", func() {
		Description("Delete a set of extension credentials")

		Payload(func() {
			Field(1, "extension_id", String, "Extension identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("kfserving-001")
			})
			Field(2, "service_id", String, "Extension service identifier", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("s3")
			})
			Field(3, "id", String, "Uniquely identifies a set of credentials within the scope of an extension service", func() {
				Pattern(identifierPattern)
				MaxLength(100)
				Example("cred-user-12bb")
			})
			Required("extension_id", "service_id", "id")
		})

		Error("NotFound", func() {
			Description("If there is no extension, service or set of credentials with the given ID, should return 404 Not Found.")
		})

		Error("BadRequest", func() {
			Description("If the set of extension credentials cannot be deleted, should return 400 Bad Request.")
		})

		HTTP(func() {
			DELETE("/extensions/{extension_id}/services/{service_id}/credentials/{id}")
			Response(StatusNoContent)
			Response("NotFound", StatusNotFound)
			Response("BadRequest", StatusBadRequest)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
			Response("BadRequest", CodeInvalidArgument)
		})
	})

})

// Extension descriptor
var Extension = Type("Extension", func() {
	tag := 1
	Field(tag, "id", String, "Uniquely identifies an extension in the registry", func() {
		Pattern(identifierPattern)
		MaxLength(100)
		Example("s3-storage-axW45s")
	})
	tag++
	Field(tag, "product", String,
		`Universal product identifier that can be used to group and identify extensions according to the product
they belong to. Product values can be used to identify installations of the same product registered
with the same or different FuseML servers`, func() {
			MaxLength(100)
			Example("mlflow")
		})
	tag++
	Field(tag, "version", String,
		`Extension version. To support semantic version operations, such as listing extensions
that match a semantic version constraint, it should be formatted as [v]MAJOR[.MINOR[.PATCH[-PRERELEASE][+BUILD]]]`, func() {
			MaxLength(100)
			Example("1.0")
			Example("v10.3.1-prealpha+b10022020")
		})
	tag++
	Field(tag, "description", String, "Extension description", func() {
		MaxLength(1000)
	})
	tag++
	Field(tag, "zone", String,
		`Extension zone identifier. Can be used to group and lookup extensions according to the infrastructure
location / zone / area / domain where they are installed (e.g. kubernetes cluster).
Is used to automatically select between local and external endpoints when
running queries that specify a zone filter.`, func() {
			MaxLength(100)
			Example("eu-central-01")
			Example("kube-cluster-dev-00126")
		})
	tag++
	Field(tag, "configuration", MapOf(String, String),
		`Configuration entries (e.g. configuration values required to configure all clients that connect to
this extension), expressed as set of key-value entries`, func() {
			Key(func() {
				Pattern(identifierPattern)
			})
			Example(map[string]string{
				"authentication": "enabled",
				"api_version":    "v2",
			})
		})
	tag++
	Field(tag, "status", ExtensionStatus, "Extension status")
	tag++
	Field(tag, "services", ArrayOf(ExtensionService), "List of services provided by this extension")
})

// Extension status descriptor
var ExtensionStatus = Type("ExtensionStatus", func() {
	tag := 1
	Field(tag, "registered", String, "The time when the extension was registered", func() {
		Format(FormatDateTime)
		Default(time.Now().Format(time.RFC3339))
		Example(time.Now().Format(time.RFC3339))
	})
	tag++
	Field(tag, "updated", String, "The time when the extension was last updated", func() {
		Format(FormatDateTime)
		Default(time.Now().Format(time.RFC3339))
		Example(time.Now().Format(time.RFC3339))
	})
})

// Extension service descriptor
var ExtensionService = Type("ExtensionService", func() {
	tag := 1
	Field(tag, "id", String, "Uniquely identifies an extension service within the scope of an extension", func() {
		Pattern(identifierPattern)
		MaxLength(100)
		Example("s3")
	})
	tag++
	Field(tag, "extension_id", String, "Reference to the extension this service belongs to", func() {
		Pattern(identifierPattern)
		MaxLength(100)
		Example("s3-storage-axW45s")
	})
	tag++
	Field(tag, "resource", String,
		`Universal service identifier that can be used to identify a service in any FuseML installation.
This identifier uniquely identifies the API or protocol (e.g. s3, git, mlflow) that the service
provides`, func() {
			MaxLength(100)
			Example("s3")
			Example("git")
			Example("mlflow-tracker")
		})
	tag++
	Field(tag, "category", String,
		`Universal service category. Used to classify services into well-known categories of AI/ML services
(e.g. model store, feature store, distributed training, serving)`, func() {
			MaxLength(100)
			Example("model-store")
			Example("serving-platform")
		})
	tag++
	Field(tag, "auth_required", Boolean,
		`Marks a service for which authentication is required. If set, a set of credentials is required
to access the service; if none of the provided credentials match the scope of the consumer,
this service will be excluded from queries`, func() {
		})
	tag++
	Field(tag, "description", String, "Service description", func() {
		MaxLength(1000)
	})
	tag++
	Field(tag, "configuration", MapOf(String, String),
		`Configuration entries (e.g. configuration values required to configure all clients that connect to
this service), expressed as set of key-value entries`, func() {
			Key(func() {
				Pattern(identifierPattern)
			})
			Example(map[string]string{
				"authentication": "enabled",
				"api_version":    "v2",
			})
		})
	tag++
	Field(tag, "status", ExtensionServiceStatus, "Service status")
	tag++
	tag++
	Field(tag, "endpoints", ArrayOf(ExtensionEndpoint), "List of endpoints through which this service can be accessed")
	tag++
	Field(tag, "credentials", ArrayOf(ExtensionCredentials), "List of credentials required to access this service")
})

// Extension service status descriptor
var ExtensionServiceStatus = Type("ExtensionServiceStatus", func() {
	tag := 1
	Field(tag, "registered", String, "The time when the service was registered", func() {
		Format(FormatDateTime)
		Default(time.Now().Format(time.RFC3339))
		Example(time.Now().Format(time.RFC3339))
	})
	tag++
	Field(tag, "updated", String, "The time when the service was last updated", func() {
		Format(FormatDateTime)
		Default(time.Now().Format(time.RFC3339))
		Example(time.Now().Format(time.RFC3339))
	})
})

// Extension endpoint descriptor
var ExtensionEndpoint = Type("ExtensionEndpoint", func() {
	tag := 1
	Field(tag, "url", String,
		`Endpoint URL. In case of k8s controllers and operators, the URL points to the cluster API.
Also used to uniquely identifies an endpoint within the scope of a service`, func() {
			Format(FormatURI)
			MaxLength(200)
			Example("https://mlflow.10.120.130.140.nip.io")
		})
	tag++
	Field(tag, "extension_id", String, "Reference to the extension this endpoint belongs to", func() {
		Pattern(identifierPattern)
		MaxLength(100)
		Example("s3-storage-axW45s")
	})
	tag++
	Field(tag, "service_id", String, "Reference to the service this endpoint belongs to", func() {
		Pattern(identifierPattern)
		MaxLength(100)
		Example("s3")
	})
	tag++
	Field(tag, "type", String,
		`Endpoint type - internal/external. An internal endpoint can only be accessed when the consumer
is located in the same zone as the extension service`, func() {
			Enum("internal", "external")
			Example("internal")
			Example("external")
		})
	tag++
	Field(tag, "configuration", MapOf(String, String),
		`Configuration entries (e.g. configuration values required to configure all clients that connect to
this endpoint), expressed as set of key-value entries`, func() {
			Key(func() {
				Pattern(identifierPattern)
			})
			Example(map[string]string{
				"ca_cert":  "A7sCSdd7879sDFDSj872jkcis7",
				"insecure": "true",
			})
		})
	tag++
	Field(tag, "status", ExtensionEndpointStatus, "Endpoint status")
})

// Extension endpoint status descriptor
var ExtensionEndpointStatus = Type("ExtensionEndpointStatus", func() {
	// empty for now, but can be extended to include future operational status attrs
	// e.g. whether the endpoint is reachable or not, or number of consumers
})

// Extension credentials descriptor
var ExtensionCredentials = Type("ExtensionCredentials", func() {
	tag := 1
	Field(tag, "id", String, "Uniquely identifies a set of credentials within the scope of an extension service", func() {
		Pattern(identifierPattern)
		MaxLength(100)
		Example("dev-token-1353411")
	})
	tag++
	Field(tag, "extension_id", String, "Reference to the extension this set of credentials belongs to", func() {
		Pattern(identifierPattern)
		MaxLength(100)
		Example("s3-storage-axW45s")
	})
	tag++
	Field(tag, "service_id", String, "Reference to the service this set of credentials belongs to", func() {
		Pattern(identifierPattern)
		MaxLength(100)
		Example("s3")
	})
	tag++
	Field(tag, "default", Boolean,
		`Use as default credentials. Used to automatically select one of several credentials with the same
scope matching the same query.`, func() {
		})
	tag++
	Field(tag, "scope", String,
		`The scope associated with this set of credentials. Global scoped credentials can be used by any
user/project. Project scoped credentials can be used only in the context of one of the projects
supplied in the Projects list. User scoped credentials can only be used by the users in the Users
list and, optionally, in the context of the projects supplied in the Projects list`, func() {
			Enum("global", "project", "user")
			Example("global")
			Example("project")
		})
	tag++
	Field(tag, "projects", ArrayOf(String), "List of projects allowed to use these credentials", func() {
		Example([]string{"prj-prototype-001", "prj-core"})
	})
	tag++
	Field(tag, "users", ArrayOf(String), "List of users allowed to use these credentials", func() {
		Example([]string{"bobthemagicpeanut", "core-admin"})
	})
	tag++
	Field(tag, "configuration", MapOf(String, String),
		`Configuration entries (e.g. usernames, passwords, tokens, keys), expressed as set of key-value entries`, func() {
			Key(func() {
				Pattern(identifierPattern)
			})
			Example(map[string]string{
				"username": "bobthemagicpeanut",
				"key":      "sEsnFT4#F03",
			})
		})
	tag++
	Field(tag, "status", ExtensionCredentialsStatus, "Credentials status")
})

// Extension credentials status descriptor
var ExtensionCredentialsStatus = Type("ExtensionCredentialsStatus", func() {
	tag := 1
	Field(tag, "created", String, "The time when the set of credentials was created", func() {
		Format(FormatDateTime)
		Default(time.Now().Format(time.RFC3339))
		Example(time.Now().Format(time.RFC3339))
	})
	tag++
	Field(tag, "updated", String, "The time when the set of credentials was last updated", func() {
		Format(FormatDateTime)
		Default(time.Now().Format(time.RFC3339))
		Example(time.Now().Format(time.RFC3339))
	})
})
