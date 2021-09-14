# Build the fuseml_core binary
FROM golang:1.16 as builder

ARG LDFLAGS="-w -s"

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY gen/ gen/
COPY cmd/ cmd/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GO111MODULE=on go build -a -ldflags "$LDFLAGS" -o fuseml_core ./cmd/fuseml_core

# Use docker scratch as minimal base image to package FuseML binaries
# Refer to https://docs.docker.com/develop/develop-images/baseimages/#create-a-simple-parent-image-using-scratch
# for more details
FROM scratch
WORKDIR /

COPY --from=builder /workspace/gen/http/openapi*.* ./gen/http/
COPY --from=builder /workspace/fuseml_core .

ENTRYPOINT ["/fuseml_core", "-host", "prod"]
