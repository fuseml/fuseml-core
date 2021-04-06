package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("openapi", func() {
	Description("The swagger service serves the API swagger definition.")
	HTTP(func() {
		Path("/api")
	})
	Files("/openapi.json", "gen/http/openapi.json", func() {
		Description("JSON document containing the API swagger definition")
	})
	Files("/openapi3.json", "gen/http/openapi3.json", func() {
		Description("JSON document containing the API swagger definition")
	})
	Files("/openapi.yaml", "gen/http/openapi.yaml", func() {
		Description("JSON document containing the API swagger definition")
	})
	Files("/openapi3.yaml", "gen/http/openapi3.yaml", func() {
		Description("JSON document containing the API swagger definition")
	})
})
