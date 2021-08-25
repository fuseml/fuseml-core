# Image URL to use all building/pushing image targets
IMG ?= fuseml-core:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

GOOS:=$(shell go env GOOS)
GOARCH:=$(shell go env GOARCH)

LDFLAGS:= -w -s

PKG_PATH=github.com/fuseml/fuseml-core/pkg

GIT_COMMIT = $(shell git rev-parse --short=8 HEAD)
GIT_REF = $(shell git symbolic-ref -q --short HEAD || git describe --tags --abbrev=0 --exact-match)
BUILD_DATE = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

ifdef VERSION
	BINARY_VERSION = $(VERSION)
else
# Use `dev` as a default version value when compiling in the main branch
ifeq ($(GIT_REF),main)
	BINARY_VERSION = dev
# Use the reference name as a default version value when compiling in another branch/tag,
else
	BINARY_VERSION = $(GIT_REF)
endif
endif

LDFLAGS += -X $(PKG_PATH)/version.GitCommit=$(GIT_COMMIT)
LDFLAGS += -X $(PKG_PATH)/version.BuildDate=$(BUILD_DATE)
ifneq ($(BINARY_VERSION),)
LDFLAGS += -X $(PKG_PATH)/version.Version=$(BINARY_VERSION)
endif

GO_LDFLAGS:=-ldflags '$(LDFLAGS)'

all: fuseml

# Run tests
test: generate lint
	go test ./... -coverprofile cover.out -covermode=atomic

# Generate code, run linter and build FuseML binaries
fuseml: generate lint build

# Generate code, run linter, build FuseML release-ready archived binaries for all supported ARCHs and OSs
release: test release_all

build: build_server build_client

build_server:
	go build ${GO_LDFLAGS} -o bin/fuseml_core ./cmd/fuseml_core

build_client:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build ${GO_LDFLAGS} -o bin/fuseml ./cmd/fuseml_cli

release_all: server_release client_release client_release-darwin-amd64 client_release-darwin-arm64 client_release-linux-amd64 client_release-linux-arm client_release-linux-arm64 client_release-windows 

server_release: build_server
	tar zcf bin/fuseml_core.tar.gz -C bin/ --remove-files --transform="s#\.\/##" ./fuseml_core
	cd bin && sha256sum -b fuseml_core.tar.gz > fuseml_core.tar.gz.sha256

client_release: build_client
	tar zcf bin/fuseml-$(GOOS)-$(GOARCH).tar.gz -C bin/ --remove-files --transform="s#\.\/##" ./fuseml
	cd bin && sha256sum -b fuseml-$(GOOS)-$(GOARCH).tar.gz > fuseml-$(GOOS)-$(GOARCH).tar.gz.sha256

client_release-linux-amd64:
	$(MAKE) GOARCH="amd64" GOOS="linux" client_release

client_release-linux-arm:
	$(MAKE) GOARCH="arm" GOOS="linux" client_release

client_release-linux-arm64:
	$(MAKE) GOARCH="arm64" GOOS="linux" client_release

client_release-windows:
	$(MAKE) GOARCH="amd64" GOOS="windows" build_client
	cd bin; mv fuseml fuseml.exe && \
	zip -m fuseml-windows-amd64.zip fuseml.exe && \
	sha256sum -b fuseml-windows-amd64.zip > fuseml-windows-amd64.zip.sha256

client_release-darwin-amd64:
	$(MAKE) GOARCH="amd64" GOOS="darwin" client_release

client_release-darwin-arm64:
	$(MAKE) GOARCH="arm64" GOOS="darwin" client_release

# Run fuseml_core
runcore: generate lint
	go run ./cmd/fuseml_core

# Run fuseml_cli
runcli: generate lint
	go run ./cmd/fuseml_cli

# Generate code
generate:
	go mod download
	go run goa.design/goa/v3/cmd/goa gen github.com/fuseml/fuseml-core/design
	go get github.com/google/wire/cmd/wire@v0.5.0
	go generate cmd/fuseml_core/wire_gen.go

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
	docker build . -t ${IMG} --build-arg LDFLAGS="$(LDFLAGS)"

# Push the docker image
docker-push:
	docker push ${IMG}
