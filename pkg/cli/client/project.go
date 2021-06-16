package client

import (
	"context"
	"net/http"

	goahttp "goa.design/goa/v3/http"

	projectc "github.com/fuseml/fuseml-core/gen/http/project/client"
	"github.com/fuseml/fuseml-core/gen/project"
)

// ProjectClient holds a client for Project
type ProjectClient struct {
	c *projectc.Client
}

// NewProjectClient initializes a ProjectClient
func NewProjectClient(scheme string, host string, doer goahttp.Doer, encoder func(*http.Request) goahttp.Encoder,
	decoder func(*http.Response) goahttp.Decoder, verbose bool) *ProjectClient {
	pc := &ProjectClient{projectc.NewClient(scheme, host, doer, encoder, decoder, verbose)}
	return pc
}

// Create a new Project.
func (pc *ProjectClient) Create(name, desc string) (*project.Project, error) {
	request, err := projectc.BuildCreatePayload(name, desc)
	if err != nil {
		return nil, err
	}

	response, err := pc.c.Create()(context.Background(), request)
	if err != nil {
		return nil, err
	}

	return response.(*project.Project), nil
}

// Delete a Project and its assignments.
func (pc *ProjectClient) Delete(name string) (err error) {
	request, err := projectc.BuildDeletePayload(name)
	if err != nil {
		return
	}

	_, err = pc.c.Delete()(context.Background(), request)
	return
}

// Get a Project.
func (pc *ProjectClient) Get(name string) (*project.Project, error) {
	request, err := projectc.BuildGetPayload(name)
	if err != nil {
		return nil, err
	}

	response, err := pc.c.Get()(context.Background(), request)
	if err != nil {
		return nil, err
	}

	return response.(*project.Project), nil
}

// List Projects.
func (pc *ProjectClient) List() ([]*project.Project, error) {
	response, err := pc.c.List()(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return response.([]*project.Project), nil
}
