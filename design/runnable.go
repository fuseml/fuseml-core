package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("runnable", func() {
	Description("The runable service performs operations on runnables.")

	// List service endpoint
	Method("list", func() {
		Description("Retrieve information about runnables registered in FuseML. Only runnables matching all supplied criteria are returned.")
		Payload(func() {
			Field(1, "id", String,
				"Value or regular expression used to filter runnables by their ID",
				func() {
					Example("ml-trainer-123")
				})
			Field(2, "kind", String,
				"Value or regular expression used to filter runnables by their kind",
				func() {
					Example("trainer")
				})
			Field(3, "labels", MapOf(String, String),
				"List of values or regular expressions used filter results by their labels.",
				func() {
					Key(func() {
						Pattern(`^[A-Za-z0-9-_]+$`)
					})
					Example(map[string]string{
						"library":  "pytorch",
						"function": "predict|train",
					})
				})
			Required()
		})

		// Result is a collection of runnables
		Result(ArrayOf(Runnable), "Return all registered runnables matching the query.")

		Error("NotFound", func() {
			Description("If the runnable is not found, should return 404 Not Found.")
		})

		// HTTP describes the HTTP transport mapping.
		HTTP(func() {
			// Requests to the service consist of HTTP GET requests.
			// The payload fields are encoded as path parameters.
			GET("/runnables")
			Param("id")
			Param("kind")
			Param("labels")
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
		Description("Retrieve a Runnable from FuseML.")

		Payload(func() {
			Field(1, "id", String, "Unique runnable identifier", func() {
				Example("model-trainer-1234")
			})
			Required("id")
		})

		Error("NotFound", func() {
			Description("If there is no runnable with the given ID, should return 404 Not Found.")
		})

		Result(Runnable)

		HTTP(func() {
			GET("/runnables/{id}")
			Response(StatusOK)
			Response("NotFound", StatusNotFound)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
		})
	})
})

// RunnableContainer describes the container flavor of implementation of a runnable
var RunnableContainer = Type("RunnableContainer", func() {
	Field(1, "image", String, "Container image location",
		func() {
			Example("myregistry.io:tag")
		})
	Field(2, "entrypoint", String, "Container entrypoint. Expressions may be used to specify values.",
		func() {
			Example("/usr/local/bin/train-model.sh")
			Example("python {{inputs[codeset].path}}/data_transform.py")
		})
	Field(3, "env", MapOf(String, String),
		"List of environment variables and their values. Expressions may be used to specify values.",
		func() {
			Example(map[string]string{
				"EXPERIMENT_NAME":   "first-experiment",
				"INPUT_CODE_PATH":   "{{inputs[code].path}}",
				"OUTPUT_MODEL_PATH": "{{outputs[model].path}}",
			})
		})
	Field(4, "args", ArrayOf(String),
		"List of command line arguments. Expressions may be used to specify values.",
		func() {
			Example([6]string{
				"--experiment",
				"first-experiment",
				"--code-path",
				"{{inputs[code].path}}",
				"--model-out-file",
				"{{outputs[model].path}}",
			})
		})
	Required("image")
})

// InputPassByStrategy describes the strategy used to pass values from the framework to the container
var InputPassByStrategy = Type("InputPassByStrategy", func() {
	Field(1, "toPath", String, "Custom container path where the input value or reference is provided by the framework.", func() {
		Example("/workspace/myproject")
	})
})

// OutputPassByStrategy describes the strategy used to pass values from the container to the framework
var OutputPassByStrategy = Type("OutputPassByStrategy", func() {
	Field(1, "fromPath", String, "Custom container path where the output value or reference is provided by the container.", func() {
		Example("/workspace/outputs/model-url")
	})
})

// ArtifactArgSpec holds the attributes describing artifacts inputs and outputs
var ArtifactArgSpec = Type("ArtifactArgSpec", func() {
	Field(1, "store", String, "Artifact store name")
	Field(2, "name", String, "Artifact name or wildcard")
	Field(3, "project", String, "Artifact project name or wildcard")
	Field(4, "version", String, "Artifact version or wildcard")
	Field(5, "minCount", Int, "Minimum number of artifacts")
	Field(6, "maxCount", Int, "Maximum number of artifacts")
	Field(7, "storeType", String, "Artifact store type")
})

// RunnableInput describes a runnable input
var RunnableInput = Type("RunnableInput", func() {
	Field(1, "type", String, "The type of input", func() {
		Enum("parameter", "opaque", "codeset", "model", "dataset", "runnable")
		Example("parameter")
		Default("parameter")
	})
	Field(2, "description", String, "Input description", func() {
		MaxLength(1000)
		Default("")
	})
	Field(3, "optional", Boolean, "Optional input", func() {
		Default(false)
	})
	Field(4, "defaultValue", String, "Default value for optional input parameters")
	Field(5, "artifact", ArtifactArgSpec, "Artifact attributes. Valid only for artifact inputs.")
	Field(6, "passByValue", InputPassByStrategy, "Pass inputs by value to the container.")
	Field(7, "passByReference", InputPassByStrategy, "Pass inputs by reference to the container.")
	Field(8, "labels", MapOf(String, String),
		"List of custom labels used to determine how to connect this input to outputs of other runnables.",
		func() {
			Key(func() {
				Pattern(`^[A-Za-z0-9-_]+$`)
			})
			Example(map[string]string{
				"library":  "pytorch",
				"function": "predict",
			})
		})
})

// RunnableInput describes a runnable output
var RunnableOutput = Type("RunnableOutput", func() {
	Field(1, "type", String, "The type of output", func() {
		Enum("parameter", "opaque", "codeset", "model", "dataset", "runnable")
		Example("parameter")
		Default("parameter")
	})
	Field(2, "description", String, "Output description", func() {
		MaxLength(1000)
		Default("")
	})
	Field(3, "optional", Boolean, "Optional output", func() {
		Default(false)
	})
	Field(4, "defaultValue", String, "Default value for optional output parameter")
	Field(5, "artifact", ArtifactArgSpec, "Artifact attributes. Valid only for artifact outputs.")
	Field(6, "passByValue", OutputPassByStrategy, "Pass outputs by value from the container.")
	Field(7, "passByReference", OutputPassByStrategy, "Pass outputs by reference from the container.")
	Field(8, "labels", MapOf(String, String),
		"List of custom labels used to determine how to connect this output to inputs of other runnables.",
		func() {
			Key(func() {
				Pattern(`^[A-Za-z0-9-_]+$`)
			})
			Example(map[string]string{
				"library":  "pytorch",
				"function": "predict",
			})
		})
})

// Runnable description
var Runnable = Type("Runnable", func() {
	Field(1, "id", String, "The unique runnable identifier", func() {
		Pattern(`^[A-Za-z0-9-_]+$`)
		MinLength(1)
		MaxLength(100)
	})
	Field(2, "description", String, "Runnable description", func() {
		MaxLength(1000)
		Default("")
	})
	Field(3, "kind", String, "The kind of runnable (builder, trainer, predictor etc.)", func() {
		Enum("custom", "builder", "trainer", "predictor")
		Example("trainer")
		Default("custom")
	})
	Field(4, "container", RunnableContainer, "Runnable implementation")
	Field(5, "defaultInputPath", String,
		"The default container path where the container expects values of inputs passed by value to be provided as files or directories",
		func() {
			Example("/opt/inputs")
			Default("/workspace")
		})
	Field(6, "inputs", MapOf(String, RunnableInput), "Map of inputs (artifacts, parameters) accepted by this runnable, indexed by name", func() {
		Key(func() {
			Pattern(`^[A-Za-z0-9-_]+$`)
		})
	})
	Field(7, "defaultOutputPath", String,
		"The default container path where the container generates the values of outputs as files or directories",
		func() {
			Example("/opt/outputs")
			Default("/workspace/output")
		})
	Field(8, "outputs", MapOf(String, RunnableOutput), "Map of outputs (artifacts, parameters) generated by this runnable, indexed by name", func() {
		Key(func() {
			Pattern(`^[A-Za-z0-9-_]+$`)
		})
	})
	Field(9, "created", String, "The runnable's creation time", func() {
		Format(FormatDateTime)
		Example("2021-04-09T06:17:25Z")
	})
	Field(10, "labels", MapOf(String, String),
		"List of labels associated with the runnable.",
		func() {
			Key(func() {
				Pattern(`^[A-Za-z0-9-_]+$`)
			})
			Example(map[string]string{
				"vendor":       "acme",
				"extension":    "prediction-engine",
				"acceleration": "GPU",
			})
		})
	Required("container")
})
