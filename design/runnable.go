package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("runnable", func() {
	Description("The runable service performs operations on runnables.")

	// Method describes a service method (endpoint)
	Method("list", func() {
		Description("Retrieve information about runnables registered in FuseML.")
		// Payload describes the method payload.
		// Here the payload is an object that consists of two fields.
		Payload(func() {
			// Field describes an object field given a field index, a field
			// name, a type and a description.
			Field(1, "kind", String, "The kind of runnables to list", func() {
				Example("builder")
			})
		})

		// Result describes the method result.
		// Here the result is a collection of runnables value.
		Result(ArrayOf(Runnable), "Return all registered runnables matching the query.")

		Error("NotFound", func() {
			Description("If the runnable is not found, should return 404 Not Found.")
		})

		// HTTP describes the HTTP transport mapping.
		HTTP(func() {
			// Requests to the service consist of HTTP GET requests.
			// The payload fields are encoded as path parameters.
			GET("/runnables")
			Param("kind", String, "Kind of a registered runnables", func() {
				Example("builder")
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
		Description("Register a runnable with the FuseML runnable store.")

		// Payload also accepts a Type object where you can list its attribute
		// as well as its required fields
		Payload(Runnable, "Runnable descriptor")

		Error("BadRequest", func() {
			Description("If the runnable does not have the required fields, should return 400 Bad Request.")
		})

		Result(Runnable)

		HTTP(func() {
			POST("/runnables")
			Response(StatusCreated)
			Response("BadRequest", StatusBadRequest)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("BadRequest", CodeInvalidArgument)
		})
	})

	Method("get", func() {
		Description("Retrieve an Runnable from FuseML.")

		Payload(func() {
			Field(1, "runnableNameOrId", String, "Runnable name or id", func() {
				Example("288BFD74-D973-18B5-FAA5-29ADF4569AC7")
			})
			Required("runnableNameOrId")
		})

		Error("BadRequest", func() {
			Description("If not name neither ID is given, should return 400 Bad Request.")
		})
		Error("NotFound", func() {
			Description("If there is no runnable with the given name/id, should return 404 Not Found.")
		})

		Result(Runnable)

		HTTP(func() {
			GET("/runnables/{runnableNameOrId}")
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
})

// RunnableImage describes an image of a runnable
var RunnableImage = Type("RunnableImage", func() {
	Field(1, "registryUrl", String, "The URL for the external registry where the image is stored (empty for internal images)",
		func() {
			Example("myregistry.io")
		})
	Field(2, "repository", String, "The image repository", func() {
		Example("example/builder")
	})
	Field(3, "tag", String, "The image tag", func() {
		Example("1.0")
	})

})

// RunnableRef describes runnable reference
var RunnableRef = Type("RunnableRef", func() {
	Field(1, "name", String, "Runnable name", func() {
		Example("BuilderRun1")
	})
	Field(2, "kind", String, "Runnable kind", func() {
		Example("builder")
	})
	Field(3, "labels", ArrayOf(String), "Runnable labels", func() {
		Example([2]string{"label1", "label2"})
	})
})

// RunnableInput is an input for a runnable
var RunnableInput = Type("RunnableInput", func() {
	Field(1, "name", String, "Input name", func() {
		Example("Input1")
	})
	Field(2, "kind", String, "Kind of input (e.g. runnable, dataset, model, parameter, etc.)", func() {
		Example("parameter")
	})
	Field(3, "runnable", RunnableRef, "Runnable reference")
	Field(4, "parameter", InputParameter, "Parameter description")
})

// InputParameter describes runnable input parameter
var InputParameter = Type("InputParameter", func() {
	Field(1, "datatype", String, "Parameter data type", func() {
		Example("file")
	})
	Field(2, "optional", Boolean, "Optional parameter", func() {
		Example(true)
	})
	Field(3, "default", String, "Default value", func() {
		Example("mydata.csv")
	})
})

// RunnabelOutput describes output of a runnable
var RunnabelOutput = Type("RunnableOutput", func() {
	Field(1, "name", String, "Output name", func() {
		Example("Output1")
	})
	Field(2, "kind", String, "Kind of output (e.g. runnable, dataset, model, metatada, etc.)", func() {
		Example("model")
	})
	Field(3, "runnable", RunnableRef, "Runnable reference")
	Field(4, "metadata", InputParameter, "Metadata description")
})

// Runnable description
var Runnable = Type("Runnable", func() {
	Field(1, "id", String, "The ID of the runnable", func() {
		Format(FormatUUID)
	})
	Field(2, "name", String, "The name of the runnable", func() {
		Example("MyTrainer")
	})
	Field(3, "kind", String, "The kind of runnable (builder, trainer, predictor etc.)", func() {
		Example("trainer")
	})
	Field(4, "image", RunnableImage, "The OCI container image associated with the runnable")
	Field(5, "inputs", ArrayOf(RunnableInput), "List of inputs (artifacts, values etc.) accepted by this runnable")
	Field(6, "outputs", ArrayOf(RunnabelOutput), "List of outputs (artifacts, values etc.) generated by this runnable")
	Field(7, "created", String, "The runnable's creation time", func() {
		Format(FormatDateTime)
		Example("2021-04-09T06:17:25Z")
	})
	Field(8, "labels", ArrayOf(String), "Labels associated with the runnable", func() {
		Example([1]string{"trainer"})
	})

	Required("name", "kind", "image", "inputs", "outputs", "labels")
})
