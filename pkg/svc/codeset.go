package svc

import (
	"context"
	"log"

	"github.com/fuseml/fuseml-core/gen/codeset"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

// codeset service implementation.
type codesetsrvc struct {
	logger *log.Logger
	store  domain.CodesetStore
}

// NewCodesetService returns the codeset service implementation.
func NewCodesetService(logger *log.Logger, store domain.CodesetStore) codeset.Service {
	return &codesetsrvc{logger, store}
}

// Retrieve information about codesets registered in FuseML.
func (s *codesetsrvc) List(ctx context.Context, p *codeset.ListPayload) (res []*codeset.Codeset, err error) {
	s.logger.Print("codeset.list")
	return s.store.GetAll(ctx, p.Project, p.Label)
}

// Register a codeset with the FuseML codeset codesetStore.
func (s *codesetsrvc) Register(ctx context.Context, p *codeset.RegisterPayload) (res *codeset.Codeset, err error) {
	s.logger.Print("codeset.register")
	return s.store.Add(ctx, p.Codeset)
}

// Retrieve an Codeset from FuseML.
func (s *codesetsrvc) Get(ctx context.Context, p *codeset.GetPayload) (res *codeset.Codeset, err error) {
	s.logger.Print("codeset.get")
	return s.store.Find(ctx, p.Project, p.Name)
}
