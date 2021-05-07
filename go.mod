module github.com/fuseml/fuseml-core

go 1.16

require (
	code.gitea.io/sdk/gitea v0.14.0
	github.com/dimfeld/httptreemux/v5 v5.3.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/goccy/go-yaml v1.8.9
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.2.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/otiai10/copy v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/tektoncd/pipeline v0.23.0
	github.com/tektoncd/triggers v0.13.0
	goa.design/goa/v3 v3.3.1
	google.golang.org/grpc v1.37.0
	google.golang.org/protobuf v1.26.0
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/client-go v0.19.7
	knative.dev/pkg v0.0.0-20210208131226-4b2ae073fa06
)
