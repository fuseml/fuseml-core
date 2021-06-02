package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/fuseml/fuseml-core/gen/application"
	"github.com/fuseml/fuseml-core/gen/codeset"
	"github.com/fuseml/fuseml-core/gen/runnable"
	"github.com/fuseml/fuseml-core/gen/version"
	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/core"
	"github.com/fuseml/fuseml-core/pkg/core/gitea"
	"github.com/fuseml/fuseml-core/pkg/svc"
	ver "github.com/fuseml/fuseml-core/pkg/version"
)

func main() {
	// Define command line flags, add any other flag required to configure the
	// service.
	var (
		hostF     = flag.String("host", "dev", "Server host (valid values: dev, prod)")
		domainF   = flag.String("domain", "", "Host domain name (overrides host domain specified in service design)")
		httpPortF = flag.String("http-port", "", "HTTP port (overrides host HTTP port specified in service design)")
		grpcPortF = flag.String("grpc-port", "", "gRPC port (overrides host gRPC port specified in service design)")
		secureF   = flag.Bool("secure", false, "Use secure scheme (https or grpcs)")
		dbgF      = flag.Bool("debug", false, "Log request and response bodies")
	)
	flag.Parse()

	// Setup logger. Replace logger with your own log package of choice.
	var (
		logger *log.Logger
	)
	{
		logger = log.New(os.Stderr, "[fuseml] ", log.Ltime)
	}

	logger.Printf("version: %s", ver.GetInfoStr())

	gitAdmin, err := gitea.NewAdminClient(logger)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to initialize Gitea admin client: ", err.Error())
		os.Exit(1)
	}

	// Initialize the services.
	var (
		versionSvc     version.Service
		applicationSvc application.Service
		runnableSvc    runnable.Service
		codesetSvc     codeset.Service
		workflowSvc    workflow.Service
	)
	{
		versionSvc = svc.NewVersionService(logger)
		codesetStore := core.NewGitCodesetStore(gitAdmin)
		applicationSvc = svc.NewApplicationService(logger, core.NewApplicationStore())
		runnableSvc = svc.NewRunnableService(logger, core.NewRunnableStore())
		codesetSvc = svc.NewCodesetService(logger, codesetStore)
		workflowSvc, err = svc.NewWorkflowService(logger, core.NewWorkflowStore(), codesetStore)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to initialize Workflow service:", err.Error())
			os.Exit(1)
		}
	}

	// Wrap the services in endpoints that can be invoked from other services
	// potentially running in different processes.
	var (
		versionEndpoints     *version.Endpoints
		applicationEndpoints *application.Endpoints
		runnableEndpoints    *runnable.Endpoints
		codesetEndpoints     *codeset.Endpoints
		workflowEndpoints    *workflow.Endpoints
	)
	{
		versionEndpoints = version.NewEndpoints(versionSvc)
		applicationEndpoints = application.NewEndpoints(applicationSvc)
		runnableEndpoints = runnable.NewEndpoints(runnableSvc)
		codesetEndpoints = codeset.NewEndpoints(codesetSvc)
		workflowEndpoints = workflow.NewEndpoints(workflowSvc)
	}

	// Create channel used by both the signal handler and server goroutines
	// to notify the main goroutine when to stop the server.
	errc := make(chan error)

	// Setup interrupt handler. This optional step configures the process so
	// that SIGINT and SIGTERM signals cause the services to stop gracefully.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// Start the servers and send errors (if any) to the error channel.
	switch *hostF {
	case "dev":
		{
			addr := "http://localhost:8000"
			u, err := url.Parse(addr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid URL %#v: %s\n", addr, err)
				os.Exit(1)
			}
			if *secureF {
				u.Scheme = "https"
			}
			if *domainF != "" {
				u.Host = *domainF
			}
			if *httpPortF != "" {
				h, _, err := net.SplitHostPort(u.Host)
				if err != nil {
					fmt.Fprintf(os.Stderr, "invalid URL %#v: %s\n", u.Host, err)
					os.Exit(1)
				}
				u.Host = net.JoinHostPort(h, *httpPortF)
			} else if u.Port() == "" {
				u.Host = net.JoinHostPort(u.Host, "80")
			}
			handleHTTPServer(ctx, u, versionEndpoints, runnableEndpoints, codesetEndpoints, workflowEndpoints, applicationEndpoints,
				&wg, errc, logger, *dbgF)
		}

		{
			addr := "grpc://localhost:8080"
			u, err := url.Parse(addr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid URL %#v: %s\n", addr, err)
				os.Exit(1)
			}
			if *secureF {
				u.Scheme = "grpcs"
			}
			if *domainF != "" {
				u.Host = *domainF
			}
			if *grpcPortF != "" {
				h, _, err := net.SplitHostPort(u.Host)
				if err != nil {
					fmt.Fprintf(os.Stderr, "invalid URL %#v: %s\n", u.Host, err)
					os.Exit(1)
				}
				u.Host = net.JoinHostPort(h, *grpcPortF)
			} else if u.Port() == "" {
				u.Host = net.JoinHostPort(u.Host, "8080")
			}
			handleGRPCServer(ctx, u, runnableEndpoints, codesetEndpoints, workflowEndpoints, applicationEndpoints, &wg, errc, logger, *dbgF)
		}

	case "prod":
		{
			addr := "http://0.0.0.0"
			u, err := url.Parse(addr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid URL %#v: %s\n", addr, err)
				os.Exit(1)
			}
			if *secureF {
				u.Scheme = "https"
			}
			if *domainF != "" {
				u.Host = *domainF
			}
			if *httpPortF != "" {
				h, _, err := net.SplitHostPort(u.Host)
				if err != nil {
					fmt.Fprintf(os.Stderr, "invalid URL %#v: %s\n", u.Host, err)
					os.Exit(1)
				}
				u.Host = net.JoinHostPort(h, *httpPortF)
			} else if u.Port() == "" {
				u.Host = net.JoinHostPort(u.Host, "80")
			}
			handleHTTPServer(ctx, u, versionEndpoints, runnableEndpoints, codesetEndpoints, workflowEndpoints, applicationEndpoints,
				&wg, errc, logger, *dbgF)
		}

		{
			addr := "grpc://0.0.0.0"
			u, err := url.Parse(addr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid URL %#v: %s\n", addr, err)
				os.Exit(1)
			}
			if *secureF {
				u.Scheme = "grpcs"
			}
			if *domainF != "" {
				u.Host = *domainF
			}
			if *grpcPortF != "" {
				h, _, err := net.SplitHostPort(u.Host)
				if err != nil {
					fmt.Fprintf(os.Stderr, "invalid URL %#v: %s\n", u.Host, err)
					os.Exit(1)
				}
				u.Host = net.JoinHostPort(h, *grpcPortF)
			} else if u.Port() == "" {
				u.Host = net.JoinHostPort(u.Host, "8080")
			}
			handleGRPCServer(ctx, u, runnableEndpoints, codesetEndpoints, workflowEndpoints, applicationEndpoints, &wg, errc, logger, *dbgF)
		}

	default:
		fmt.Fprintf(os.Stderr, "invalid host argument: %q (valid hosts: dev|prod)\n", *hostF)
	}

	// Wait for signal.
	logger.Printf("exiting (%v)", <-errc)

	// Send cancellation signal to the goroutines.
	cancel()

	wg.Wait()
	logger.Println("exited")
}
