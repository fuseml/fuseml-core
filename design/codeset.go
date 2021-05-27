package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("codeset", func() {
	Description("The codeset service performs operations on Codesets.")

	// Method describes a service method (endpoint)
	Method("list", func() {
		Description("Retrieve information about Codesets registered in FuseML.")
		// Payload describes the method payload.
		// Here the payload is an object that consists of two fields.
		Payload(func() {
			// Field describes an object field given a field index, a field
			// name, a type and a description.
			Field(1, "project", String, "List only Codesets that belong to given project", func() {
				Example("mlflow-project-01")
			})
			Field(2, "label", String, "List only Codesets with matching label", func() {
				Example("mlflow")
			})
		})

		// Result describes the method result.
		// Here the result is a collection of codeset value.
		Result(ArrayOf(Codeset), "Return all registered Codesets matching the query.")

		Error("NotFound", func() {
			Description("If the Codeset is not found, should return 404 Not Found.")
		})

		// HTTP describes the HTTP transport mapping.
		HTTP(func() {
			// Requests to the service consist of HTTP GET requests.
			// The payload fields are encoded as path parameters.
			GET("/codesets")
			Param("project", String, "List only Codesets that belong to given project", func() {
				Example("mlflow-project-01")
			})
			Param("label", String, "List only Codesets with matching label", func() {
				Example("mlflow")
			})
			// Responses use a "200 OK" HTTP status.
			// The result is encoded in the response body (default).
			Response(StatusOK)
			Response("NotFound", StatusNotFound)
		})

		// GRPC describes the gRPC transport mapping.
		GRPC(func() {
			// Responses use a "OK" gRPC code.
			// The result is encoded in the response message (default).
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
		})

	})

	Method("register", func() {
		Description("Register a Codeset with the FuseML codeset store.")

		Payload(func() {
			Field(1, "name", String, "The name of the Codeset", func() {
				Example("mlflow-app-01")
				Pattern(`^[A-Za-z0-9_][A-Za-z0-9-_]*$`)
			})
			Field(2, "project", String, "The project this Codeset belongs to", func() {
				Example("mlflow-project-01")
				Pattern(`^[A-Za-z0-9_][A-Za-z0-9-_]*$`)
			})
			Field(3, "description", String, "Codeset description", func() {
				Example("My first MLFlow application with FuseML")
				Default("")
			})
			Field(4, "labels", ArrayOf(String), "Additional Codeset labels that helps with identifying the type", func() {
				Elem(func() {
					Pattern(`^[A-Za-z0-9_][A-Za-z0-9-_]*$`)
				})
				Example([]string{"mlflow", "playground"})
			})
			Required("name", "project")
		})

		Error("BadRequest", func() {
			Description("If the Codeset does not have the required fields, should return 400 Bad Request.")
		})

		Result(Codeset)

		HTTP(func() {
			POST("/codesets")
			Param("name")
			Param("project")
			Param("description")
			Param("labels")
			Response(StatusCreated)
			Response("BadRequest", StatusBadRequest)
		})
		GRPC(func() {
			Response(CodeOK)
			Response("BadRequest", CodeInvalidArgument)
		})
	})

	Method("get", func() {
		Description("Retrieve a Codeset from FuseML.")

		Payload(func() {
			Field(1, "project", String, "Project name", func() {
				Example("mlflow-project-01")
			})
			Field(2, "name", String, "Codeset name", func() {
				Example("mlflow-app-01")
			})
			Required("project", "name")
		})

		Error("BadRequest", func() {
			Description("If neither name or project is not given, should return 400 Bad Request.")
		})
		Error("NotFound", func() {
			Description("If there is no codeset with the given name and project, should return 404 Not Found.")
		})

		Result(Codeset)

		HTTP(func() {
			GET("/codesets/{project}/{name}")
			Response(StatusOK)
			Response("BadRequest", StatusBadRequest)
			Response("NotFound", StatusNotFound)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("BadRequest", CodeInvalidArgument)
			Response("NotFound", CodeNotFound)
		})
	})

	Method("delete", func() {
		Description("Delete a Codeset registered by FuseML.")

		Payload(func() {
			Field(1, "project", String, "Project name", func() {
				Example("mlflow-project-01")
			})
			Field(2, "name", String, "Codeset name", func() {
				Example("mlflow-app-01")
			})
			Required("project", "name")
		})

		Error("BadRequest", func() {
			Description("If neither name or project is not given, should return 400 Bad Request.")
		})
		Error("NotFound", func() {
			Description("If there is no codeset with the given name and project, should return 404 Not Found.")
		})

		HTTP(func() {
			DELETE("/codesets/{project}/{name}")
			Response(StatusNoContent)
			Response("BadRequest", StatusBadRequest)
		})
		GRPC(func() {
			Response(CodeOK)
			Response("BadRequest", CodeInvalidArgument)
		})
	})

})

// Codeset describes the Codeset
var Codeset = Type("Codeset", func() {

	Field(1, "name", String, "The name of the Codeset", func() {
		Example("mlflow-app-01")
	})
	Field(2, "project", String, "The project this Codeset belongs to", func() {
		Example("mlflow-project-01")
	})
	Field(3, "description", String, "Codeset description", func() {
		Example("My first MLFlow application with FuseML")
		Default("")
	})
	Field(4, "labels", ArrayOf(String), "Additional Codeset labels that helps with identifying the type", func() {
		Example([]string{"mlflow", "playground"})
	})
	Field(5, "url", String, "Full URL to the Codeset", func() {
		Example("http://my-gitea.server/project/repository.git")
	})
	Required("name", "project")
})
