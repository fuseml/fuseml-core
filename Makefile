
SWAGGER_SPEC:=api/swagger.yaml
GO_LDFLAGS:=-ldflags '-s -w'
GO_GCFLAGS=

all: fuseml_core fuseml_client

fuseml_core: gen_server lint
	go build ${GO_GCFLAGS} -o bin/fuseml_core ${GO_LDFLAGS} ./cmd/fuseml_core

fuseml_client: gen_client lint

test:
	ginkgo ./cmd ./pkg

gen_server:
	swagger generate server -t pkg -f $(SWAGGER_SPEC) -s api --exclude-main --regenerate-configureapi

gen_client:
	swagger generate client -t pkg -f $(SWAGGER_SPEC) -c client

lint: fmt vet tidy

vet:
	go vet ./...

tidy:
	go mod tidy

fmt:
	go fmt ./...
