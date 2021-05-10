package main

import (
	"mime"
	"net/http"
	"time"

	cli "github.com/fuseml/fuseml-core/gen/http/cli/fuseml_core"
	"github.com/goccy/go-yaml"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

func doHTTP(scheme, host string, timeout int, debug bool) (goa.Endpoint, interface{}, error) {
	var (
		doer goahttp.Doer
	)
	{
		doer = &http.Client{Timeout: time.Duration(timeout) * time.Second}
		if debug {
			doer = goahttp.NewDebugDoer(doer)
		}
	}

	return cli.ParseEndpoint(
		scheme,
		host,
		doer,
		goahttp.RequestEncoder,
		responseDecoder,
		debug,
	)
}

func httpUsageCommands() string {
	return cli.UsageCommands()
}

func httpUsageExamples() string {
	return cli.UsageExamples()
}

func responseDecoder(resp *http.Response) goahttp.Decoder {
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// default to YAML
		contentType = "application/x-yaml"
	} else {
		// sanitize
		if mediaType, _, err := mime.ParseMediaType(contentType); err == nil {
			contentType = mediaType
		}
	}
	if contentType == "application/x-yaml" {
		return yaml.NewDecoder(resp.Body)
	}
	return goahttp.ResponseDecoder(resp)
}
