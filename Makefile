# Image URL to use all building/pushing image targets
IMG ?= fuseml-core:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

GO_LDFLAGS:=-ldflags '-s -w'

all: fuseml_all

# Run tests
test: generate lint
	go test ./... -coverprofile cover.out

# Generate code, run linter and build FuseML binaries
fuseml: generate lint build

fuseml_all: generate lint build_all 

build: build_server build_client_local

build_all: build_server build_client-amd64 build_client-windows build_client-darwin-amd64

build_server:
	go build ${GO_LDFLAGS} -o bin/fuseml_core ./cmd/fuseml_core

build_client_local:
	go build ${GO_LDFLAGS} -o bin/fuseml_core-cli ./cmd/fuseml_core-cli

build_client-amd64:
	GOARCH="amd64" GOOS="linux" go build ${GO_LDFLAGS} -o bin/fuseml_core-cli-linux-amd64 ./cmd/fuseml_core-cli

build_client-windows:
	GOARCH="amd64" GOOS="windows" go build ${GO_LDFLAGS} -o bin/fuseml_core-cli-windows-amd64 ./cmd/fuseml_core-cli

build_client-darwin-amd64:
	GOARCH="amd64" GOOS="darwin" go build ${GO_LDFLAGS} -o bin/fuseml_core-cli-darwin-amd64 ./cmd/fuseml_core-cli

# Run fuseml_core
runcore: generate lint
	go run ./cmd/fuseml_core

# Run fuseml_core-cli
runcli: generate lint
	go run ./cmd/fuseml_core-cli

# Generate code
generate:
	go mod download
	go run goa.design/goa/v3/cmd/goa gen github.com/fuseml/fuseml-core/design

# Lint code
lint: fmt vet tidy
	golint `go list ./... | grep -v "/design"`

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Run go mod tidy against code
tidy:
	go mod tidy

# Download dependecies needed to generate code by goa
deps:
	go get goa.design/goa/v3/cmd/goa@v3.3.1
	go get goa.design/goa/v3/http/codegen/openapi/v2@v3.3.1
	go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	go get golang.org/x/lint/golint

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}
