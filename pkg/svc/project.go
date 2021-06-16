package svc

import (
	"context"
	"log"

	"github.com/fuseml/fuseml-core/gen/project"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

// project service implementation.
type projectsrvc struct {
	logger *log.Logger
	store  domain.ProjectStore
}

// NewProjectService returns the project service implementation.
func NewProjectService(logger *log.Logger, store domain.ProjectStore) project.Service {
	return &projectsrvc{logger, store}
}

func projectDomainToRest(p *domain.Project) (res *project.Project) {
	res = &project.Project{
		Name:        p.Name,
		Description: p.Description,
	}

	for _, u := range p.Users {
		res.Users = append(res.Users,
			&project.User{
				Name:  u.Name,
				Email: u.Email,
			},
		)
	}
	return
}

// Retrieve information about projects registered in FuseML.
func (s *projectsrvc) List(ctx context.Context) (res []*project.Project, err error) {
	s.logger.Print("project.list")
	items, err := s.store.GetAll(ctx)
	res = make([]*project.Project, 0, len(items))
	for _, c := range items {
		res = append(res, projectDomainToRest(c))
	}
	return res, err
}

// Retrieve an Project from FuseML.
func (s *projectsrvc) Get(ctx context.Context, p *project.GetPayload) (res *project.Project, err error) {
	s.logger.Print("project.get")
	c, err := s.store.Find(ctx, p.Name)
	if err != nil {
		return nil, project.MakeBadRequest(err)
	}
	return projectDomainToRest(c), nil
}

func (s *projectsrvc) Create(ctx context.Context, p *project.CreatePayload) (res *project.Project, err error) {
	s.logger.Print("project.create")
	c, err := s.store.Create(ctx, p.Name, p.Description)
	if err != nil {
		if err.Error() == "Project with that name already exists" {
			return nil, project.MakeConflict(err)
		}
		return nil, err
	}
	return projectDomainToRest(c), nil
}

func (s *projectsrvc) Delete(ctx context.Context, p *project.DeletePayload) error {
	s.logger.Print("project.delete")
	return s.store.Delete(ctx, p.Name)
}
