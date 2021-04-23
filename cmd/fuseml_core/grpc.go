package main

import (
	"context"
	"log"
	"net"
	"net/url"
	"sync"

	codeset "github.com/fuseml/fuseml-core/gen/codeset"
	codesetpb "github.com/fuseml/fuseml-core/gen/grpc/codeset/pb"
	codesetsvr "github.com/fuseml/fuseml-core/gen/grpc/codeset/server"
	runnablepb "github.com/fuseml/fuseml-core/gen/grpc/runnable/pb"
	runnablesvr "github.com/fuseml/fuseml-core/gen/grpc/runnable/server"
	workflowpb "github.com/fuseml/fuseml-core/gen/grpc/workflow/pb"
	workflowsvr "github.com/fuseml/fuseml-core/gen/grpc/workflow/server"
	runnable "github.com/fuseml/fuseml-core/gen/runnable"
	workflow "github.com/fuseml/fuseml-core/gen/workflow"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcmdlwr "goa.design/goa/v3/grpc/middleware"
	"goa.design/goa/v3/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// handleGRPCServer starts configures and starts a gRPC server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleGRPCServer(ctx context.Context, u *url.URL, runnableEndpoints *runnable.Endpoints, codesetEndpoints *codeset.Endpoints, workflowEndpoints *workflow.Endpoints, wg *sync.WaitGroup, errc chan error, logger *log.Logger, debug bool) {
	// Setup goa log adapter.
	var (
		adapter middleware.Logger
	)
	{
		adapter = middleware.NewLogger(logger)
	}

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to gRPC requests and
	// responses.
	var (
		runnableServer *runnablesvr.Server
		codesetServer  *codesetsvr.Server
		workflowServer *workflowsvr.Server
	)
	{
		runnableServer = runnablesvr.New(runnableEndpoints, nil)
		codesetServer = codesetsvr.New(codesetEndpoints, nil)
		workflowServer = workflowsvr.New(workflowEndpoints, nil)
	}

	// Initialize gRPC server with the middleware.
	srv := grpc.NewServer(
		grpcmiddleware.WithUnaryServerChain(
			grpcmdlwr.UnaryRequestID(),
			grpcmdlwr.UnaryServerLog(adapter),
		),
	)

	// Register the servers.
	runnablepb.RegisterRunnableServer(srv, runnableServer)
	codesetpb.RegisterCodesetServer(srv, codesetServer)
	workflowpb.RegisterWorkflowServer(srv, workflowServer)

	for svc, info := range srv.GetServiceInfo() {
		for _, m := range info.Methods {
			logger.Printf("serving gRPC method %s", svc+"/"+m.Name)
		}
	}

	// Register the server reflection service on the server.
	// See https://grpc.github.io/grpc/core/md_doc_server-reflection.html.
	reflection.Register(srv)

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		// Start gRPC server in a separate goroutine.
		go func() {
			lis, err := net.Listen("tcp", u.Host)
			if err != nil {
				errc <- err
			}
			logger.Printf("gRPC server listening on %q", u.Host)
			errc <- srv.Serve(lis)
		}()

		<-ctx.Done()
		logger.Printf("shutting down gRPC server at %q", u.Host)
		srv.Stop()
	}()
}
