package svc

import (
	"context"
	"github.com/pkg/errors"
	"log"

	"github.com/fuseml/fuseml-core/gen/application"
	"github.com/fuseml/fuseml-core/pkg/domain"
	"github.com/fuseml/fuseml-core/pkg/kubernetes"
)

func appRestToDomain(ra *application.Application) (a *domain.Application, err error) {
	a = &domain.Application{
		Name:         ra.Name,
		Type:         ra.Type,
		URL:          ra.URL,
		Workflow:     ra.Workflow,
		K8sNamespace: ra.K8sNamespace,
	}
	if ra.Description != nil {
		a.Description = *ra.Description
	}
	for _, res := range ra.K8sResources {
		a.K8sResources = append(a.K8sResources,
			&domain.KubernetesResource{
				Name: res.Name,
				Kind: res.Kind,
			},
		)
	}
	return a, nil
}

func appDomainToRest(a *domain.Application) (ret *application.Application) {
	ret = &application.Application{
		Name:         a.Name,
		Type:         a.Type,
		URL:          a.URL,
		Workflow:     a.Workflow,
		K8sNamespace: a.K8sNamespace,
	}
	if a.Description != "" {
		ret.Description = &a.Description
	}
	for _, res := range a.K8sResources {
		ret.K8sResources = append(ret.K8sResources,
			&application.KubernetesResource{
				Name: res.Name,
				Kind: res.Kind,
			},
		)
	}
	return ret
}

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
	items, err := s.store.GetAll(ctx, p.Type, p.Workflow)
	res = make([]*application.Application, 0, len(items))
	for _, a := range items {
		res = append(res, appDomainToRest(a))
	}
	return res, err
}

// Register a application with the FuseML application store.
func (s *applicationsrvc) Register(ctx context.Context, a *application.Application) (res *application.Application, err error) {
	s.logger.Print("application.register")
	app, err := appRestToDomain(a)
	if err != nil {
		return nil, application.MakeBadRequest(err)
	}
	app, err = s.store.Add(ctx, app)
	return appDomainToRest(app), err
}

// Retrieve an Application from FuseML.
func (s *applicationsrvc) Get(ctx context.Context, p *application.GetPayload) (res *application.Application, err error) {
	s.logger.Print("application.get")

	app := s.store.Find(ctx, p.Name)
	if app == nil {
		return nil, application.MakeNotFound(errors.New("Application with the specified name not found"))
	}
	return appDomainToRest(app), nil
}

// Delete an Application registered by FuseML.
func (s *applicationsrvc) Delete(ctx context.Context, p *application.DeletePayload) error {
	s.logger.Print("application.delete")
	app := s.store.Find(ctx, p.Name)
	if app == nil {
		return application.MakeNotFound(errors.New("Application with the specified name not found"))
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
