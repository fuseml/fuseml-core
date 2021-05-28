package common

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	applicationc "github.com/fuseml/fuseml-core/gen/http/application/client"
	codesetc "github.com/fuseml/fuseml-core/gen/http/codeset/client"
	runnablec "github.com/fuseml/fuseml-core/gen/http/runnable/client"
	workflowc "github.com/fuseml/fuseml-core/gen/http/workflow/client"
	yaml "github.com/goccy/go-yaml"
	goahttp "goa.design/goa/v3/http"
)

// Clients holds a list of clients for all FuseML endpoints
type Clients struct {
	CodesetClient     *codesetc.Client
	ApplicationClient *applicationc.Client
	WorkflowClient    *workflowc.Client
	RunnableClient    *runnablec.Client
}

// InitializeClients initializes a list of fuseml clients based on global configuration parameters
func (c *Clients) InitializeClients(o *GlobalOptions) error {
	var (
		doer    goahttp.Doer                         = &http.Client{Timeout: time.Duration(o.Timeout) * time.Second}
		encoder func(*http.Request) goahttp.Encoder  = goahttp.RequestEncoder
		decoder func(*http.Response) goahttp.Decoder = responseDecoder
		scheme  string
		host    string
	)

	u, err := url.Parse(o.URL)
	if err != nil || u.Host == "" {
		// assume the scheme part is missing and default to https
		u, err = url.ParseRequestURI("https://" + o.URL)
		if err != nil || u.Host == "" {
			return fmt.Errorf("invalid URL %#v: %s", o.URL, err)
		}
	}

	scheme = u.Scheme
	host = u.Host

	if o.Verbose {
		doer = goahttp.NewDebugDoer(doer)
	}

	c.CodesetClient = codesetc.NewClient(scheme, host, doer, encoder, decoder, o.Verbose)
	c.ApplicationClient = applicationc.NewClient(scheme, host, doer, encoder, decoder, o.Verbose)
	c.WorkflowClient = workflowc.NewClient(scheme, host, doer, encoder, decoder, o.Verbose)
	c.RunnableClient = runnablec.NewClient(scheme, host, doer, encoder, decoder, o.Verbose)

	return nil
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
