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

func codesetRestToDomain(restCodeset *codeset.Codeset) (res *domain.Codeset, err error) {
	res = &domain.Codeset{
		Name:        restCodeset.Name,
		Project:     restCodeset.Project,
		Labels:      restCodeset.Labels,
		Description: restCodeset.Description,
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
func (s *codesetsrvc) Register(ctx context.Context, p *codeset.RegisterPayload) (*codeset.RegisterResult, error) {
	s.logger.Print("codeset.register")
	c, err := codesetRestToDomain(&codeset.Codeset{
		Name:        p.Name,
		Project:     p.Project,
		Description: p.Description,
		Labels:      p.Labels,
	})
	if err != nil {
		return nil, codeset.MakeBadRequest(err)
	}
	c, username, password, err := s.store.Add(ctx, c)
	if err != nil {
		return nil, codeset.MakeBadRequest(err)
	}
	res := codeset.RegisterResult{
		Codeset:  codesetDomainToRest(c),
		Username: username,
		Password: password,
	}
	return &res, nil
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

func (s *codesetsrvc) Delete(ctx context.Context, p *codeset.DeletePayload) error {
	s.logger.Print("codeset.delete")
	return s.store.Delete(ctx, p.Project, p.Name)
}
