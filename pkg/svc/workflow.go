package svc

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/core/config"
	"github.com/fuseml/fuseml-core/pkg/core/tekton"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

// workflow service example implementation.
// The example methods log the requests and return zero values.
type workflowsrvc struct {
	logger       *log.Logger
	store        domain.WorkflowStore
	codesetStore domain.CodesetStore
	backend      domain.WorkflowBackend
}

// NewWorkflowService returns the workflow service implementation.
func NewWorkflowService(logger *log.Logger, store domain.WorkflowStore, codesetStore domain.CodesetStore) (workflow.Service, error) {
	backend, err := tekton.NewWorkflowBackend(config.FuseMLNamespace)
	if err != nil {
		return nil, err
	}
	return &workflowsrvc{logger, store, codesetStore, backend}, nil
}

// Retrieve information about workflows registered in FuseML.
func (s *workflowsrvc) List(ctx context.Context, w *workflow.ListPayload) (res []*workflow.Workflow, err error) {
	s.logger.Print("workflow.list")
	return s.listWorkflows(ctx, w.Name), nil
}

func (s *workflowsrvc) ListRuns(ctx context.Context, w *workflow.ListRunsPayload) ([]*workflow.WorkflowRun, error) {
	s.logger.Print("workflow.listWorkflowRuns")
	workflowRuns := []*workflow.WorkflowRun{}
	workflows := s.listWorkflows(ctx, w.WorkflowName)
	filters := domain.WorkflowRunFilter{}
	if w.CodesetName != nil {
		filters.ByLabel = append(filters.ByLabel, fmt.Sprintf("%s=%s", tekton.LabelCodesetName, *w.CodesetName))
	}
	if w.CodesetProject != nil {
		filters.ByLabel = append(filters.ByLabel, fmt.Sprintf("%s=%s", tekton.LabelCodesetProject, *w.CodesetProject))
	}
	if w.Status != nil {
		filters.ByStatus = append(filters.ByStatus, *w.Status)
	}

	for _, workflow := range workflows {
		runs, err := s.backend.ListWorkflowRuns(ctx, *workflow, filters)
		if err != nil {
			s.logger.Print(err)
			return nil, err
		}
		workflowRuns = append(workflowRuns, runs...)
	}

	return workflowRuns, nil
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
	if _, err = s.getWorkflow(ctx, w.Name); err != nil {
		s.logger.Print(err)
		return err
	}

	url, err := s.backend.CreateListener(ctx, s.logger, w.Name, true)
	if err != nil {
		s.logger.Print(err)
		return err
	}

	err = s.codesetStore.CreateWebhook(ctx, codeset, url)
	if err != nil {
		s.logger.Print(err)
		return err
	}

	err = s.backend.CreateWorkflowRun(ctx, w.Name, *codeset)
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

func (s *workflowsrvc) listWorkflows(ctx context.Context, workflowName *string) []*workflow.Workflow {
	name := "all"
	if workflowName != nil {
		name = *workflowName
	}
	return s.store.GetAll(ctx, name)
}
