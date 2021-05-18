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

func codesetRestToDomain(rp *codeset.RegisterPayload) (res *domain.Codeset, err error) {
	res = &domain.Codeset{
		Name:        rp.Name,
		Project:     rp.Project,
		Labels:      rp.Labels,
		Description: rp.Description,
	}

	return res, nil
}

func codesetDomainToRest(c *domain.Codeset) (res *codeset.Codeset) {
	res = &codeset.Codeset{
		Name:        c.Name,
		Project:     c.Project,
		Description: c.Description,
		Labels:      c.Labels,
		URL:         &c.URL,
	}

	return
}

// Retrieve information about codesets registered in FuseML.
func (s *codesetsrvc) List(ctx context.Context, p *codeset.ListPayload) (res []*codeset.Codeset, err error) {
	s.logger.Print("codeset.list")
	items, err := s.store.GetAll(ctx, p.Project, p.Label)
	res = make([]*codeset.Codeset, 0, len(items))
	for _, c := range items {
		res = append(res, codesetDomainToRest(c))
	}
	return res, err
}

// Register a codeset with the FuseML codeset codesetStore.
func (s *codesetsrvc) Register(ctx context.Context, p *codeset.RegisterPayload) (res *codeset.Codeset, err error) {
	s.logger.Print("codeset.register")
	c, err := codesetRestToDomain(p)
	if err != nil {
		return nil, codeset.MakeBadRequest(err)
	}
	c, err = s.store.Add(ctx, c)
	if err != nil {
		return nil, codeset.MakeBadRequest(err)
	}
	return codesetDomainToRest(c), nil
}

// Retrieve an Codeset from FuseML.
func (s *codesetsrvc) Get(ctx context.Context, p *codeset.GetPayload) (res *codeset.Codeset, err error) {
	s.logger.Print("codeset.get")
	c, err := s.store.Find(ctx, p.Project, p.Name)
	if err != nil {
		return nil, codeset.MakeBadRequest(err)
	}
	return codesetDomainToRest(c), nil
}
