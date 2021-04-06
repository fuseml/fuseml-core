
GO_LDFLAGS:=-ldflags '-s -w'
GO_GCFLAGS=

all: build

build: gen_core example tidy lint
	go build ${GO_GCFLAGS} -o bin/fuseml_core ${GO_LDFLAGS} ./cmd/fuseml_core
	go build ${GO_GCFLAGS} -o bin/fuseml_core-cli ${GO_LDFLAGS} ./cmd/fuseml_core-cli

test:
	ginkgo ./gen ./pkg

gen_core:
	go mod download
	go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go run goa.design/goa/v3/cmd/goa gen github.com/fuseml/fuseml-core/design

example:
	go run goa.design/goa/v3/cmd/goa example github.com/fuseml/fuseml-core/design

lint: fmt vet tidy

vet:
	go vet ./...

tidy:
	go mod tidy

fmt:
	go fmt ./...
