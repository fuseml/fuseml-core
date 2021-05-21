package main

import (
	"mime"
	"net/http"
	"strings"
	"time"

	cli "github.com/fuseml/fuseml-core/gen/http/cli/fuseml_core"
	yaml "github.com/goccy/go-yaml"
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
	ct := resp.Header.Get("Content-Type")
	if ct != "" {
		// sanitize
		if mediaType, _, err := mime.ParseMediaType(ct); err == nil {
			ct = mediaType
		}
	}

	switch {
	case ct == "", ct == "application/x-yaml", ct == "text/x-yaml":
		fallthrough
	case strings.HasSuffix(ct, "+yaml"):
		return yaml.NewDecoder(resp.Body, yaml.Strict())
	default:
		return goahttp.ResponseDecoder(resp)
	}
}
