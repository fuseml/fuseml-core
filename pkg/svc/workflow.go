package svc

import (
	"context"
	"errors"
	"log"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/core/config"
	"github.com/fuseml/fuseml-core/pkg/core/tekton"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

// WorkflowBackend is the interface for the FuseML workflows
type WorkflowBackend interface {
	CreateListener(context.Context, *log.Logger, string, bool) (string, error)
	CreateWorkflow(context.Context, *log.Logger, *workflow.Workflow) error
	CreateWorkflowRun(context.Context, string, domain.Codeset) error
}

// workflow service example implementation.
// The example methods log the requests and return zero values.
type workflowsrvc struct {
	logger       *log.Logger
	store        domain.WorkflowStore
	codesetStore domain.CodesetStore
	backend      WorkflowBackend
}

// NewWorkflowService returns the workflow service implementation.
func NewWorkflowService(logger *log.Logger, store domain.WorkflowStore, codesetStore domain.CodesetStore) workflow.Service {
	backend, err := tekton.NewWorkflowBackend(config.FuseMLNamespace)
	if err != nil {
		logger.Print(err)
	}
	return &workflowsrvc{logger, store, codesetStore, backend}
}

// Retrieve information about workflows registered in FuseML.
func (s *workflowsrvc) List(ctx context.Context, w *workflow.ListPayload) (res []*workflow.Workflow, err error) {
	s.logger.Print("workflow.list")
	name := "all"
	if w.Name != nil {
		name = *w.Name
	}

	return s.store.GetAll(ctx, name), nil
}

// Register a workflow with the FuseML workflow store.
func (s *workflowsrvc) Register(ctx context.Context, w *workflow.Workflow) (res *workflow.Workflow, err error) {
	s.logger.Print("workflow.register")
	err = s.backend.CreateWorkflow(ctx, s.logger, w)
	if err != nil {
		s.logger.Print(err)
		return nil, err
	}
	return s.store.Add(ctx, w)
}

// Retrieve a Workflow from FuseML.
func (s *workflowsrvc) Get(ctx context.Context, w *workflow.GetPayload) (res *workflow.Workflow, err error) {
	return s.getWorkflow(ctx, w.Name)
}

// Assign a Workflow to a Codeset
func (s *workflowsrvc) Assign(ctx context.Context, w *workflow.AssignPayload) (err error) {
	s.logger.Print("workflow.assign")
	codeset, err := s.codesetStore.Find(ctx, w.CodesetProject, w.CodesetName)
	if err != nil {
		s.logger.Print(err)
		return workflow.MakeNotFound(err)
	}
	if _, err = s.getWorkflow(ctx, w.WorkflowName); err != nil {
		s.logger.Print(err)
		return err
	}

	url, err := s.backend.CreateListener(ctx, s.logger, w.WorkflowName, true)
	if err != nil {
		s.logger.Print(err)
		return err
	}

	err = s.codesetStore.CreateWebhook(ctx, codeset, url)
	if err != nil {
		s.logger.Print(err)
		return err
	}

	err = s.backend.CreateWorkflowRun(ctx, w.WorkflowName, *codeset)
	if err != nil {
		s.logger.Print(err)
		return err
	}
	return nil
}

func (s *workflowsrvc) getWorkflow(ctx context.Context, name string) (*workflow.Workflow, error) {
	if name == "" {
		return nil, workflow.MakeBadRequest(errors.New("empty workflow name"))
	}
	wf := s.store.Find(ctx, name)
	if wf == nil {
		return nil, workflow.MakeNotFound(errors.New("could not find a workflow with the specified ID"))
	}
	return wf, nil
}
