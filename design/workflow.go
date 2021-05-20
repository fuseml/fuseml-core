package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("workflow", func() {
	Description("The workflow service performs operations on workflows.")

	Method("list", func() {
		Description("List workflows registered in FuseML.")
		Payload(func() {
			Field(1, "name", String, "List workflows with the specified name", func() {
				Example("workflowA")
			})
		})
		Result(ArrayOf(Workflow), "Return all registered workflows matching the query.")

		Error("NotFound", func() {
			Description("If the workflow is not found, should return 404 Not Found.")
		})

		HTTP(func() {
			GET("/workflows")
			Param("name", String, "List workflows with the specified name", func() {
				Example("workflowA")
			})
			Response(StatusOK)
			Response("NotFound", StatusNotFound)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
		})

	})

	Method("register", func() {
		Description("Register a workflow within the FuseML workflow store.")
		Payload(Workflow, "Workflow descriptor")
		Error("BadRequest", func() {
			Description("If the workflow does not have the required fields, should return 400 Bad Request.")
		})
		Result(Workflow)

		HTTP(func() {
			POST("/workflows")
			Response(StatusCreated)
			Response("BadRequest", StatusBadRequest)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("BadRequest", CodeInvalidArgument)
		})
	})

	Method("get", func() {
		Description("Retrieve Workflow(s) from FuseML.")

		Payload(func() {
			Field(1, "name", String, "Workflow name", func() {
				Example("mlflow-sklearn-e2e")
			})
			Required("name")
		})

		Error("BadRequest", func() {
			Description("If name is not given, should return 400 Bad Request.")
		})
		Error("NotFound", func() {
			Description("If there is no workflow with the given name, should return 404 Not Found.")
		})

		Result(Workflow)

		HTTP(func() {
			GET("/workflows/{name}")
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

	Method("assign", func() {
		Description("Assigins a Workflow to a Codeset")

		Payload(func() {
			Field(1, "name", String, "Name of the Workflow to be associated with the codeset", func() {
				Example("mlflow-sklearn-e2e")
			})
			Field(2, "codesetProject", String, "Project that hosts the codeset to assign the workflow to", func() {
				Example("workspace")
			})
			Field(3, "codesetName", String, "Codeset to assign the workflow to", func() {
				Example("mlflow-project-001")
			})
			Required("name", "codesetProject", "codesetName")
		})

		Error("BadRequest", func() {
			Description("If no workflowName or codeset is given, should return 400 Bad Request.")
		})
		Error("NotFound", func() {
			Description("If there is no workflow with the given name or codeset, should return 404 Not Found.")
		})

		HTTP(func() {
			POST("/workflows/assign")
			Param("name")
			Param("codesetName")
			Param("codesetProject")
			Response(StatusCreated)
			Response("BadRequest", StatusBadRequest)
			Response("NotFound", StatusNotFound)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("BadRequest", CodeInvalidArgument)
			Response("NotFound", CodeNotFound)
		})
	})

	Method("listRuns", func() {
		Description("List workflow runs")

		Payload(func() {
			Field(1, "workflowName", String, "Name of the Workflow to list runs from", func() {
				Example("mlflow-sklearn-e2e")
			})
			Field(2, "codesetName", String, "Name of the codeset to list runs from", func() {
				Example("mlflow-project-001")
			})
			Field(3, "codesetProject", String, "Name of the codeset project to list runs from", func() {
				Example("workspace")
			})
			Field(4, "status", String, "status of the workflow runs to list", func() {
				Enum("Started", "Running", "Cancelled", "Succeeded", "Failed", "Completed", "Timeout")
				Example("Succeeded")

			})
		})

		Error("NotFound", func() {
			Description("If there is no workflow with the given name, should return 404 Not Found.")
		})

		Result(ArrayOf(WorkflowRun), "Return all runs for a workflow.")

		HTTP(func() {
			GET("/workflows/runs")
			Param("workflowName")
			Param("codesetName")
			Param("codesetProject")
			Param("status")
			Response(StatusOK)
			Response("NotFound", StatusNotFound)
		})

		GRPC(func() {
			Response(CodeOK)
			Response("NotFound", CodeNotFound)
		})

	})

	Method("listAssignments", func() {
		Description("List workflow assignments to codesets")

		Payload(func() {
			Field(1, "name", String, "Name of the workflow to list assignments", func() {
				Example("mlflow-sklearn-e2e")
			})
		})

		Result(ArrayOf(WorkflowAssignment), "Return a list of workflow assignments.")

		HTTP(func() {
			GET("/workflows/assignments")
			Param("name")
			Response(StatusOK)
		})

		GRPC(func() {
			Response(CodeOK)
		})
	})
})

// Workflow describes a FuseML workflow
var Workflow = Type("Workflow", func() {
	Field(1, "created", String, "The workflow creation time", func() {
		Format(FormatDateTime)
		Example("2021-04-09T06:17:25Z")
	})
	Field(2, "name", String, "Name of the workflow", func() {
		Example("TrainAndServe")
	})
	Field(3, "description", String, "Description for the workflow", func() {
		Example("This workflow is just trains a model and serve it")
	})
	Field(4, "inputs", ArrayOf(WorkflowInput), "Inputs for the workflow")
	Field(5, "outputs", ArrayOf(WorkflowOutput), "Outputs from the workflow")
	Field(6, "steps", ArrayOf(WorkflowStep), "Steps to be executed by the workflow")

	Required("name", "steps")
})

// WorkflowInput defines the input for a FuseML workflow
var WorkflowInput = Type("WorkflowInput", func() {
	Field(1, "name", String, "Name of the input", func() {
		Example("mlflow-codeset")
	})
	Field(2, "description", String, "Description of the input", func() {
		Example("An MLFlow project codeset")
	})
	Field(3, "type", String, "The type of the input (codeset, string, ...)", func() {
		Example("codeset")
	})
	Field(4, "default", String, "Default value for the input", func() {
		Example("mlflow-example")
	})
	Field(5, "labels", ArrayOf(String), "Labels associated with the input", func() {
		Example([]string{"mlflow-project"})
	})
})

// WorkflowOutput defines the output from a FuseML workflow
var WorkflowOutput = Type("WorkflowOutput", func() {
	Field(1, "name", String, "Name of the output", func() {
		Example("prediction-url")
	})
	Field(2, "description", String, "Description of the output", func() {
		Example("The URL where the exposed prediction service endpoint can be contacted to run predictions.")
	})
	Field(3, "type", String, "The data type of the output", func() {
		Example("string")
	})
})

// WorkflowStep defines a step for a FuseML workflow
var WorkflowStep = Type("WorkflowStep", func() {
	Field(1, "name", String, "The name of the step", func() {
		Example("predictor")
	})
	Field(2, "image", String, "The image used to execute the step", func() {
		Example("ghcr.io/fuseml/kfserving-predictor:1.0")
	})
	Field(3, "inputs", ArrayOf(WorkflowStepInput), "List of inputs for the step")
	Field(4, "outputs", ArrayOf(WorkflowStepOutput), "List of output from the step")
	Field(5, "env", ArrayOf(StepEnv), "List of environment variables available for the container running the step")
})

// WorkflowStepInput defines the input for a FuseML workflow step
var WorkflowStepInput = Type("WorkflowStepInput", func() {
	Field(1, "name", String, "Name of the input", func() {
		Example("model-uri")
	})
	Field(2, "value", String, "Value of the input", func() {
		Example("s3://mlflow-artifacts/3/c7ae3b0e6fd44b4b96f7066c66672551/artifacts/model")
	})
	Field(3, "codeset", StepInputCodeset, "Codeset associated with the input")
})

// StepInputCodeset defines the Codeset type of input for a FuseML workflow step
var StepInputCodeset = Type("StepInputCodeset", func() {
	Field(1, "name", String, "Name or ID of the codeset", func() {
		Example("mlflow-project")
	})
	Field(2, "path", String, "Path where the codeset will be mounted inside the container running the step", func() {
		Example("/project")
	})
})

// WorkflowStepOutput defines the output from a FuseML workflow step
var WorkflowStepOutput = Type("WorkflowStepOutput", func() {
	Field(1, "name", String, "Name of the variable to hold the step output value", func() {
		Example("model-uri")
	})
	Field(2, "image", StepOutputImage, "If the step builds a container image as output it will be referenced as 'image'")
})

// StepOutputImage defines the output from a FuseML workflow when it builds a container image
var StepOutputImage = Type("StepOutputImage", func() {
	Field(1, "dockerfile", String, "Path to the Dockerfile used to build the image", func() {
		Example("/project/.fuseml/Dockerfile")
	})
	Field(2, "name", String, "Name of the image, including the repository where the image will be stored", func() {
		Example("registry.fuseml-registry/mlflow-project/mlflow-codeset:0.1")
	})
})

// StepEnv defines the environment variables that are loaded inside the container running a FuseML workflow step
var StepEnv = Type("StepEnv", func() {
	Field(1, "name", String, "Name of the environment variable", func() {
		Example("PATH")
	})
	Field(2, "value", String, "Value to set for the enviroment variable", func() {
		Example("/project")
	})
})

// WorkflowRun describes a workflow run returned when listed
var WorkflowRun = Type("WorkflowRun", func() {
	Field(1, "name", String, "Name of the run")
	Field(2, "workflowRef", String, "Reference to the Workflow")
	Field(3, "inputs", ArrayOf(WorkflowRunInput), "Workflow run inputs")
	Field(4, "outputs", ArrayOf(WorkflowRunOutput), "Outputs from the workflow run")
	Field(5, "startTime", String, "The time when the workflow run started", func() {
		Format(FormatDateTime)
		Example("2021-04-09T06:17:25Z")
	})
	Field(6, "completionTime", String, "The time when the workflow run completed", func() {
		Format(FormatDateTime)
		Example("2021-04-09T06:20:35Z")
	})
	Field(7, "status", String, "The current status of the workflow run", func() {
		Enum("Started", "Running", "Cancelled", "Succeeded", "Failed", "Completed", "Timeout", "Unknown")
		Example("Succeeded")
	})
	Field(8, "URL", String, "Dashboard URL to the workflow run")
})

// WorkflowRunInput describes a input from a WorkflowRun including its value
var WorkflowRunInput = Type("WorkflowRunInput", func() {
	Field(1, "input", WorkflowInput, "The workflow input")
	Field(2, "value", String, "The input value set by the Workflow run")
})

// WorkflowRunInput describes the output from a WorkflowRun including its value
var WorkflowRunOutput = Type("WorkflowRunOutput", func() {
	Field(1, "output", WorkflowOutput, "The workflow output")
	Field(2, "value", String, "The output value set by the Workflow run")
})

// WorkflowAssignment describes the assignment between a workflow and codesets
var WorkflowAssignment = Type("WorkflowAssignment", func() {
	Field(1, "workflow", String, "Workflow assigned to the codeset")
	Field(2, "status", WorkflowAssignmentStatus, "The status of the assignment")
	Field(3, "codesets", ArrayOf(Codeset), "Codesets assigned to the workflow")
})

// WorkflowAssignmentStatus describes the status of the resource responsible for the
// assignment between a workflow and codesets
var WorkflowAssignmentStatus = Type("WorkflowAssignmentStatus", func() {
	Field(1, "available", Boolean, "The state of the assignment")
	Field(2, "URL", String, "Dashboard URL to the resource responsible for the assignment")
})
