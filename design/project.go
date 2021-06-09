package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("project", func() {
	Description("The project service performs operations on Projects.")

	Method("list", func() {
		Description("Retrieve information about FuseML Projects.")

		Result(ArrayOf(Project), "Return all Projects.")

		HTTP(func() {
			GET("/projects")
			Response(StatusOK)
		})

		GRPC(func() {
			Response(CodeOK)
		})

	})

	Method("get", func() {
		Description("Retrieve a Project from FuseML.")

		Payload(func() {
			Field(1, "name", String, "Project name", func() {
				Example("mlflow-project-01")
			})
			Required("name")
		})

		Error("BadRequest", func() {
			Description("If name is not given, should return 400 Bad Request.")
		})
		Error("NotFound", func() {
			Description("If there is no project with the given name, should return 404 Not Found.")
		})

		Result(Project)

		HTTP(func() {
			GET("/projects/{name}")
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
		Description("Delete a FuseML Project.")

		Payload(func() {
			Field(1, "name", String, "Project name", func() {
				Example("mlflow-project-01")
			})
			Required("name")
		})

		Error("BadRequest", func() {
			Description("If name is not given, should return 400 Bad Request.")
		})

		HTTP(func() {
			DELETE("/projects/{name}")
			Response(StatusNoContent)
			Response("BadRequest", StatusBadRequest)
		})
		GRPC(func() {
			Response(CodeOK)
			Response("BadRequest", CodeInvalidArgument)
		})
	})
})

// Project describes the Project
var Project = Type("Project", func() {

	Field(1, "name", String, "The name of the Project", func() {
		Example("mlflow-project-01")
	})
	Field(2, "users", ArrayOf(User), "Users assigned to the Project")
	Field(3, "description", String, "Project description", func() {
		Example("Set of MLFlow applications")
		Default("")
	})
	Required("name")
})

// User describes the user assigned to the project
var User = Type("User", func() {
	Field(1, "name", String, "User name", func() {
		Example("fuseml-mlflow-project-01")
	})
	Field(2, "email", String, "User email", func() {
		Example("fuseml-mlflow-project-01@fuseml.org")
	})
	Required("name", "email")
})
