package fuseml

import (
	"context"
	"errors"
	"log"

	codeset "github.com/fuseml/fuseml-core/gen/codeset"
)

// codeset service example implementation.
// The example methods log the requests and return zero values.
type codesetsrvc struct {
	logger *log.Logger
}

// NewCodeset returns the codeset service implementation.
func NewCodeset(logger *log.Logger) codeset.Service {
	return &codesetsrvc{logger}
}

// Retrieve information about codesets registered in FuseML.
func (s *codesetsrvc) List(ctx context.Context, p *codeset.ListPayload) (res []*codeset.Codeset, err error) {
	s.logger.Print("codeset.list")
	project := "all"
	if p.Project != nil {
		project = *p.Project
	}
	return codesetStore.GetAllCodesets(project, p.Label), nil
}

// Register a codeset with the FuseML codeset codesetStore.
func (s *codesetsrvc) Register(ctx context.Context, p *codeset.Codeset) (res *codeset.Codeset, err error) {
	s.logger.Print("codeset.register")
	return codesetStore.AddCodeset(p)
}

// Retrieve an Codeset from FuseML.
func (s *codesetsrvc) Get(ctx context.Context, p *codeset.GetPayload) (res *codeset.Codeset, err error) {
	s.logger.Print("codeset.get")
	/*
		if name == nil {
			return nil, codeset.MakeBadRequest(err)
		}
	*/
	r := codesetStore.FindCodeset(p.Project, p.Name)
	if r == nil {
		return nil, codeset.MakeNotFound(errors.New("could not find a codeset with the specified ID"))
	}
	return r, nil
}
