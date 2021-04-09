
GO_LDFLAGS:=-ldflags '-s -w'
GO_GCFLAGS=

all: gen-core build

build: tidy lint
	go build ${GO_GCFLAGS} -o bin/fuseml_core ${GO_LDFLAGS} ./cmd/fuseml_core
	go build ${GO_GCFLAGS} -o bin/fuseml_core-cli ${GO_LDFLAGS} ./cmd/fuseml_core-cli

test:
	ginkgo ./gen ./pkg

gen-core:
	go mod download
	go run goa.design/goa/v3/cmd/goa gen github.com/fuseml/fuseml-core/design

example: gen
	go run goa.design/goa/v3/cmd/goa example github.com/fuseml/fuseml-core/design

lint: fmt vet tidy

vet:
	go vet ./...

tidy:
	go mod tidy

fmt:
	go fmt ./...

deps:
	go get goa.design/goa/v3/cmd/goa@v3.3.1
	go get goa.design/goa/v3/http/codegen/openapi/v2@v3.3.1
	go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

