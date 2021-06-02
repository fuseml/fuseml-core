package client

import (
	"context"
	"net/http"

	goahttp "goa.design/goa/v3/http"

	versionc "github.com/fuseml/fuseml-core/gen/http/version/client"
	"github.com/fuseml/fuseml-core/gen/version"
)

// VersionClient holds a client for Version
type VersionClient struct {
	c *versionc.Client
}

// NewVersionClient initializes a VersionClient
func NewVersionClient(scheme string, host string, doer goahttp.Doer, encoder func(*http.Request) goahttp.Encoder,
	decoder func(*http.Response) goahttp.Decoder, verbose bool) *VersionClient {
	vc := &VersionClient{versionc.NewClient(scheme, host, doer, encoder, decoder, verbose)}
	return vc
}

// Get version information.
func (vc *VersionClient) Get() (*version.VersionInfo, error) {
	response, err := vc.c.Get()(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	return response.(*version.VersionInfo), nil
}
