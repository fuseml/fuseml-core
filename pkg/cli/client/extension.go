package client

import (
	"context"
	"net/http"
	"net/url"

	goahttp "goa.design/goa/v3/http"

	"github.com/fuseml/fuseml-core/gen/extension"
	extensionc "github.com/fuseml/fuseml-core/gen/http/extension/client"
	"github.com/fuseml/fuseml-core/pkg/cli/common"
)

// ExtensionClient holds a client for the extension HTTP REST service
type ExtensionClient struct {
	c *extensionc.Client
}

// NewExtensionClient initializes a ExtensionClient
func NewExtensionClient(scheme string, host string, doer goahttp.Doer, encoder func(*http.Request) goahttp.Encoder,
	decoder func(*http.Response) goahttp.Decoder, verbose bool) *ExtensionClient {
	ec := &ExtensionClient{extensionc.NewClient(scheme, host, doer, encoder, decoder, verbose)}
	return ec
}

// ReadExtensionFromFile - register an extension from a file YAML or JSON descriptor.
func (ec *ExtensionClient) ReadExtensionFromFile(filepath string) (res *extension.Extension, err error) {

	var extDescriptor string
	err = common.LoadFileIntoVar(filepath, &extDescriptor)
	if err != nil {
		return nil, err
	}

	request, err := extensionc.BuildRegisterExtensionPayload(extDescriptor)
	if err != nil {
		return nil, err
	}

	return request, err
}

// RegisterExtension - register an extension.
func (ec *ExtensionClient) RegisterExtension(ext *extension.Extension) (res *extension.Extension, err error) {

	response, err := ec.c.RegisterExtension()(context.Background(), ext)
	if err != nil {
		return nil, err
	}

	return response.(*extension.Extension), nil
}

// GetExtension - get an extension.
func (ec *ExtensionClient) GetExtension(extensionID string) (*extension.Extension, error) {
	request, err := extensionc.BuildGetExtensionPayload(extensionID)
	if err != nil {
		return nil, err
	}

	response, err := ec.c.GetExtension()(context.Background(), request)
	if err != nil {
		return nil, err
	}

	return response.(*extension.Extension), nil
}

// ListExtension - list Extensions.
func (ec *ExtensionClient) ListExtension(query *extension.ExtensionQuery) ([]*extension.Extension, error) {

	wfs, err := ec.c.ListExtensions()(context.Background(), query)
	if err != nil {
		return nil, err
	}

	return wfs.([]*extension.Extension), nil
}

// DeleteExtension - delete an Extension.
func (ec *ExtensionClient) DeleteExtension(extensionID string) error {
	request, err := extensionc.BuildDeleteExtensionPayload(extensionID)
	if err != nil {
		return err
	}

	_, err = ec.c.DeleteExtension()(context.Background(), request)
	if err != nil {
		return err
	}

	return nil
}

// UpdateExtension - update the attributes of an extension.
func (ec *ExtensionClient) UpdateExtension(ext *extension.Extension) (res *extension.Extension, err error) {

	response, err := ec.c.UpdateExtension()(context.Background(), ext)
	if err != nil {
		return nil, err
	}

	return response.(*extension.Extension), nil
}

// AddService - add a service to an extension.
func (ec *ExtensionClient) AddService(svc *extension.ExtensionService) (res *extension.ExtensionService, err error) {

	response, err := ec.c.AddService()(context.Background(), svc)
	if err != nil {
		return nil, err
	}

	return response.(*extension.ExtensionService), nil
}

// DeleteService - delete a service from an extension.
func (ec *ExtensionClient) DeleteService(extensionID, serviceID string) error {
	request, err := extensionc.BuildDeleteServicePayload(extensionID, serviceID)
	if err != nil {
		return err
	}

	_, err = ec.c.DeleteService()(context.Background(), request)
	if err != nil {
		return err
	}

	return nil
}

// UpdateService - update the attributes of a service from an extension.
func (ec *ExtensionClient) UpdateService(service *extension.ExtensionService) (res *extension.ExtensionService, err error) {

	response, err := ec.c.UpdateService()(context.Background(), service)
	if err != nil {
		return nil, err
	}

	return response.(*extension.ExtensionService), nil
}

// ListServices - list all services from an extension.
func (ec *ExtensionClient) ListServices(extensionID string) (res []*extension.ExtensionService, err error) {
	request, err := extensionc.BuildListServicesPayload(extensionID)
	if err != nil {
		return nil, err
	}

	response, err := ec.c.ListServices()(context.Background(), request)
	if err != nil {
		return nil, err
	}

	return response.([]*extension.ExtensionService), nil
}

// AddEndpoint - add an endpoint to an extension service.
func (ec *ExtensionClient) AddEndpoint(ep *extension.ExtensionEndpoint) (res *extension.ExtensionEndpoint, err error) {

	response, err := ec.c.AddEndpoint()(context.Background(), ep)
	if err != nil {
		return nil, err
	}

	return response.(*extension.ExtensionEndpoint), nil
}

// DeleteEndpoint - delete an endpoint from an extension service.
func (ec *ExtensionClient) DeleteEndpoint(extensionID, serviceID, URL string) error {

	url := url.QueryEscape(URL)
	request, err := extensionc.BuildDeleteEndpointPayload(extensionID, serviceID, url)
	if err != nil {
		return err
	}

	_, err = ec.c.DeleteEndpoint()(context.Background(), request)
	if err != nil {
		return err
	}

	return nil
}

// UpdateEndpoint - update the attributes of an endpoint from an extension service.
func (ec *ExtensionClient) UpdateEndpoint(endpoint *extension.ExtensionEndpoint) (res *extension.ExtensionEndpoint, err error) {

	url := url.QueryEscape(*endpoint.URL)
	endpoint.URL = &url
	response, err := ec.c.UpdateEndpoint()(context.Background(), endpoint)
	if err != nil {
		return nil, err
	}

	return response.(*extension.ExtensionEndpoint), nil
}

// ListEndpoints - list all endpoints from an extension service.
func (ec *ExtensionClient) ListEndpoints(extensionID, serviceID string) (res []*extension.ExtensionEndpoint, err error) {
	request, err := extensionc.BuildListEndpointsPayload(extensionID, serviceID)
	if err != nil {
		return nil, err
	}

	response, err := ec.c.ListEndpoints()(context.Background(), request)
	if err != nil {
		return nil, err
	}

	return response.([]*extension.ExtensionEndpoint), nil
}

// AddCredentials - add a set of credentials to an extension service.
func (ec *ExtensionClient) AddCredentials(ep *extension.ExtensionCredentials) (res *extension.ExtensionCredentials, err error) {

	response, err := ec.c.AddCredentials()(context.Background(), ep)
	if err != nil {
		return nil, err
	}

	return response.(*extension.ExtensionCredentials), nil
}

// DeleteCredentials - delete a set of credentials from an extension service.
func (ec *ExtensionClient) DeleteCredentials(extensionID, serviceID, credentialsID string) error {
	request, err := extensionc.BuildDeleteCredentialsPayload(extensionID, serviceID, credentialsID)
	if err != nil {
		return err
	}

	_, err = ec.c.DeleteCredentials()(context.Background(), request)
	if err != nil {
		return err
	}

	return nil
}

// UpdateCredentials - update the attributes of a set of credentials from an extension service.
func (ec *ExtensionClient) UpdateCredentials(credentials *extension.ExtensionCredentials) (res *extension.ExtensionCredentials, err error) {

	response, err := ec.c.UpdateCredentials()(context.Background(), credentials)
	if err != nil {
		return nil, err
	}

	return response.(*extension.ExtensionCredentials), nil
}

// ListCredentials - list all credentials from an extension service.
func (ec *ExtensionClient) ListCredentials(extensionID, serviceID string) (res []*extension.ExtensionCredentials, err error) {
	request, err := extensionc.BuildListCredentialsPayload(extensionID, serviceID)
	if err != nil {
		return nil, err
	}

	response, err := ec.c.ListCredentials()(context.Background(), request)
	if err != nil {
		return nil, err
	}

	return response.([]*extension.ExtensionCredentials), nil
}
