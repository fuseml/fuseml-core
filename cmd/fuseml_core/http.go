package main

import (
	"context"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	application "github.com/fuseml/fuseml-core/gen/application"
	codeset "github.com/fuseml/fuseml-core/gen/codeset"
	applicationsvr "github.com/fuseml/fuseml-core/gen/http/application/server"
	codesetsvr "github.com/fuseml/fuseml-core/gen/http/codeset/server"
	openapisvr "github.com/fuseml/fuseml-core/gen/http/openapi/server"
	projectsvr "github.com/fuseml/fuseml-core/gen/http/project/server"
	runnablesvr "github.com/fuseml/fuseml-core/gen/http/runnable/server"
	versionsvr "github.com/fuseml/fuseml-core/gen/http/version/server"
	workflowsvr "github.com/fuseml/fuseml-core/gen/http/workflow/server"
	project "github.com/fuseml/fuseml-core/gen/project"
	runnable "github.com/fuseml/fuseml-core/gen/runnable"
	"github.com/fuseml/fuseml-core/gen/version"
	workflow "github.com/fuseml/fuseml-core/gen/workflow"

	"github.com/goccy/go-yaml"
	goahttp "goa.design/goa/v3/http"
	httpmdlwr "goa.design/goa/v3/http/middleware"
	"goa.design/goa/v3/middleware"
)

// handleHTTPServer starts configures and starts a HTTP server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleHTTPServer(ctx context.Context, u *url.URL, versionEndpoints *version.Endpoints, runnableEndpoints *runnable.Endpoints, codesetEndpoints *codeset.Endpoints, projectEndpoints *project.Endpoints, workflowEndpoints *workflow.Endpoints, applicationEndpoints *application.Endpoints, wg *sync.WaitGroup, errc chan error, logger *log.Logger, debug bool) {
	// Setup goa log adapter.
	var (
		adapter middleware.Logger
	)
	{
		adapter = middleware.NewLogger(logger)
	}

	// Provide the transport specific request decoder and response encoder.
	// The goa http package has built-in support for JSON, XML and gob.
	// Other encodings can be used by providing the corresponding functions,
	// see goa.design/implement/encoding.
	var (
		dec = requestDecoder
		enc = responseEncoder
	)

	// Build the service HTTP request multiplexer and configure it to serve
	// HTTP requests to the service endpoints.
	var mux goahttp.Muxer
	{
		mux = goahttp.NewMuxer()
	}

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to HTTP requests and
	// responses.
	var (
		versionServer     *versionsvr.Server
		applicationServer *applicationsvr.Server
		runnableServer    *runnablesvr.Server
		codesetServer     *codesetsvr.Server
		projectServer     *projectsvr.Server
		openapiServer     *openapisvr.Server
		workflowServer    *workflowsvr.Server
	)
	{
		eh := errorHandler(logger)
		versionServer = versionsvr.New(versionEndpoints, mux, dec, enc, eh, nil)
		applicationServer = applicationsvr.New(applicationEndpoints, mux, dec, enc, eh, nil)
		runnableServer = runnablesvr.New(runnableEndpoints, mux, dec, enc, eh, nil)
		codesetServer = codesetsvr.New(codesetEndpoints, mux, dec, enc, eh, nil)
		projectServer = projectsvr.New(projectEndpoints, mux, dec, enc, eh, nil)
		openapiServer = openapisvr.New(nil, mux, dec, enc, eh, nil)
		workflowServer = workflowsvr.New(workflowEndpoints, mux, dec, enc, eh, nil)
		if debug {
			servers := goahttp.Servers{
				versionServer,
				applicationServer,
				runnableServer,
				codesetServer,
				projectServer,
				openapiServer,
				workflowServer,
			}
			servers.Use(httpmdlwr.Debug(mux, os.Stdout))
		}
	}
	// Configure the mux.
	versionsvr.Mount(mux, versionServer)
	applicationsvr.Mount(mux, applicationServer)
	runnablesvr.Mount(mux, runnableServer)
	codesetsvr.Mount(mux, codesetServer)
	projectsvr.Mount(mux, projectServer)
	openapisvr.Mount(mux)
	workflowsvr.Mount(mux, workflowServer)

	// Wrap the multiplexer with additional middlewares. Middlewares mounted
	// here apply to all the service endpoints.
	var handler http.Handler = mux
	{
		handler = httpmdlwr.Log(adapter)(handler)
		handler = httpmdlwr.RequestID()(handler)
	}

	// Start HTTP server using default configuration, change the code to
	// configure the server as required by your service.
	srv := &http.Server{Addr: u.Host, Handler: handler}
	for _, m := range versionServer.Mounts {
		logger.Printf("HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
	}
	for _, m := range applicationServer.Mounts {
		logger.Printf("HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
	}
	for _, m := range runnableServer.Mounts {
		logger.Printf("HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
	}
	for _, m := range codesetServer.Mounts {
		logger.Printf("HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
	}
	for _, m := range projectServer.Mounts {
		logger.Printf("HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
	}
	for _, m := range openapiServer.Mounts {
		logger.Printf("HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
	}
	for _, m := range workflowServer.Mounts {
		logger.Printf("HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
	}

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		// Start HTTP server in a separate goroutine.
		go func() {
			logger.Printf("HTTP server listening on %q", u.Host)
			errc <- srv.ListenAndServe()
		}()

		<-ctx.Done()
		logger.Printf("shutting down HTTP server at %q", u.Host)

		// Shutdown gracefully with a 30s timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_ = srv.Shutdown(ctx)
	}()
}

// errorHandler returns a function that writes and logs the given error.
// The function also writes and logs the error unique ID so that it's possible
// to correlate.
func errorHandler(logger *log.Logger) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		id := ctx.Value(middleware.RequestIDKey).(string)
		_, _ = w.Write([]byte("[" + id + "] encoding: " + err.Error()))
		logger.Printf("[%s] ERROR: %s", id, err.Error())
	}
}

// requestDecoder implements the goahttp.Decoder interface.
// Its return defaults to a YAML decoder, when a specific content type other
// than YAML is requested it returns the decoder from the Goa RequestDecoder
// function.
func requestDecoder(r *http.Request) goahttp.Decoder {
	ct := r.Header.Get("Content-Type")

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
		return yaml.NewDecoder(r.Body)
	default:
		return goahttp.RequestDecoder(r)
	}
}

// responseEncoder implements the goahttp.Encoder interface.
// Its return defaults to a YAML encoder, when a specific content type other
// than YAML is requested it returns the Encoder from the Goa ResponseEncoder
// function.
func responseEncoder(ctx context.Context, w http.ResponseWriter) goahttp.Encoder {
	var ct string
	if a := ctx.Value(goahttp.ContentTypeKey); a != nil {
		ct = a.(string)
	}
	var (
		enc goahttp.Encoder
		mt  string
		err error
	)

	negotiate := func(ct string) (goahttp.Encoder, string) {
		switch {
		case ct == "", ct == "*/*", ct == "application/x-yaml":
			return yaml.NewEncoder(w), "application/x-yaml"
		case ct == "text/x-yaml":
			return yaml.NewEncoder(w), ct
		case strings.HasSuffix(ct, "+yaml"):
			return yaml.NewEncoder(w), ct
		default:
			return goahttp.ResponseEncoder(ctx, w), ct
		}
	}

	if ct != "" {
		// If content type explicitly set in the DSL, infer the response encoder
		// from the content type context key.
		if mt, _, err = mime.ParseMediaType(ct); err == nil {
			enc, mt = negotiate(ct)
		}
		goahttp.SetContentType(w, mt)
		return enc
	}

	var accept string
	if a := ctx.Value(goahttp.AcceptTypeKey); a != nil {
		accept = a.(string)
	}

	// If Accept header exists in the request, infer the response encoder
	// from the header value.
	if enc, mt = negotiate(accept); enc == nil {
		// attempt to normalize
		if mt, _, err = mime.ParseMediaType(accept); err == nil {
			enc, mt = negotiate(mt)
		}
	}
	if enc == nil {
		enc, mt = negotiate("")
	}
	goahttp.SetContentType(w, mt)
	return enc
}
