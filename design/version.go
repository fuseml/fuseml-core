package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("version", func() {
	Description("The version service displays version information.")

	Method("get", func() {
		Description("Retrieve version information.")

		Result(VersionInfo)

		HTTP(func() {
			GET("/version")
			Response(StatusOK)
		})

		GRPC(func() {
			Response(CodeOK)
		})
	})
})

// VersionInfo describes server version information
var VersionInfo = Type("VersionInfo", func() {

	Field(1, "version", String, "The server version", func() {
		Example("v1.0")
	})
	Field(2, "gitCommit", String, "The git commit corresponding to the running server version", func() {
		Example("4833d673")
	})
	Field(3, "buildDate", String, "The date the server binary was built", func() {
		Example("2021-06-02T10:21:03Z")
	})
	Field(4, "golangVersion", String, "The GO version used to build the binary", func() {
		Example("go1.16.0")
	})
	Field(5, "golangCompiler", String, "The GO compiler used to build the binary", func() {
		Example("gc")
	})
	Field(6, "platform", String, "The platform where the server is running", func() {
		Example("linux/amd64")
	})
})
