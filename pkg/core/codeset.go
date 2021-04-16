package fuseml

import (
	"context"
	"log"

	codeset "github.com/fuseml/fuseml-core/gen/codeset"
)

// codeset service implementation.
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
	return codesetStore.GetAllCodesets(p.Project, p.Label)
}

// Register a codeset with the FuseML codeset codesetStore.
func (s *codesetsrvc) Register(ctx context.Context, p *codeset.RegisterPayload) (res *codeset.Codeset, err error) {
	s.logger.Print("codeset.register")
	return codesetStore.AddCodeset(p.Codeset)
}

// Retrieve an Codeset from FuseML.
func (s *codesetsrvc) Get(ctx context.Context, p *codeset.GetPayload) (res *codeset.Codeset, err error) {
	s.logger.Print("codeset.get")
	return codesetStore.FindCodeset(p.Project, p.Name)
}
