# Image URL to use all building/pushing image targets
IMG ?= fuseml-core:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

GO_LDFLAGS:=-ldflags '-s -w'

all: fuseml

# Run tests
test: generate lint
	go test ./... -coverprofile cover.out

# Build FuseML binaries
fuseml: generate lint
	go build ${GO_LDFLAGS} -o bin/fuseml_core ./cmd/fuseml_core
	go build ${GO_LDFLAGS} -o bin/fuseml_core-cli ./cmd/fuseml_core-cli

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
	golint ./...

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
