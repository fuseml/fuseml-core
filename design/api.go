package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = API("fuseml", func() {
	Title("FuseML core API")
	Description("Provides an API for the core operations of FuseML")

	// Server describes a single process listening for client requests. The DSL
	// defines the set of services that the server hosts as well as hosts details.
	Server("fuseml-core", func() {
		Description("fuseml-core hosts the core services")

		// List the services hosted by this server.
		Services("runnable", "openapi")

		// List the Hosts and their transport URLs.
		Host("localhost", func() {
			Description("Run the service on localhost.")
			// Transport specific URLs, supported schemes are:
			// 'http', 'https', 'grpc' and 'grpcs' with the respective default
			// ports: 80, 443, 8080, 8443.
			URI("http://localhost:8000")
			URI("grpc://localhost:8080")
		})
	})
})
