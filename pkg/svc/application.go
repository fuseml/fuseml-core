package svc

import (
	"context"
	"github.com/pkg/errors"
	"log"

	"github.com/fuseml/fuseml-core/gen/application"
	"github.com/fuseml/fuseml-core/pkg/domain"
	"github.com/fuseml/fuseml-core/pkg/kubernetes"
)

// application service implementation.
type applicationsrvc struct {
	logger *log.Logger
	store  domain.ApplicationStore
}

// NewApplicationService returns the application service implementation.
func NewApplicationService(logger *log.Logger, store domain.ApplicationStore) application.Service {
	return &applicationsrvc{logger, store}
}

// Retrieve information about applications registered in FuseML.
func (s *applicationsrvc) List(ctx context.Context, p *application.ListPayload) (res []*application.Application, err error) {
	s.logger.Print("application.list")
	return s.store.GetAll(ctx, p.Type, p.Workflow)
}

// Register a application with the FuseML application store.
func (s *applicationsrvc) Register(ctx context.Context, a *application.Application) (res *application.Application, err error) {
	s.logger.Print("application.register")
	return s.store.Add(ctx, a)
}

// Retrieve an Application from FuseML.
func (s *applicationsrvc) Get(ctx context.Context, p *application.GetPayload) (res *application.Application, err error) {
	s.logger.Print("application.get")

	app := s.store.Find(ctx, p.Name)
	if app == nil {
		return nil, application.MakeNotFound(errors.New("Application with the specified name not found"))
	}
	return app, nil
}

// Delete an Application registered by FuseML.
func (s *applicationsrvc) Delete(ctx context.Context, p *application.DeletePayload) error {
	s.logger.Print("application.delete")
	app := s.store.Find(ctx, p.Name)
	if app == nil {
		s.logger.Print("application not found, nothing to delete")
		return nil
	}
	cluster, err := kubernetes.NewCluster()
	if err != nil {
		return errors.Wrap(err, "Failed initializing kubernetes cluster")
	}
	for _, r := range app.K8sResources {
		err := cluster.DeleteResource(ctx, r.Name, app.K8sNamespace, r.Kind)
		if err != nil {
			return errors.Wrap(err, "Failed deleting kubernetes resource "+r.Name)
		}
	}
	return s.store.Delete(ctx, p.Name)
}
