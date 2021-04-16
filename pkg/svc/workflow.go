package svc

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

// workflow service example implementation.
// The example methods log the requests and return zero values.
type workflowsrvc struct {
	logger *log.Logger
	store  domain.WorkflowStore
}

// NewWorkflow returns the workflow service implementation.
func NewWorkflowService(logger *log.Logger, store domain.WorkflowStore) workflow.Service {
	return &workflowsrvc{logger, store}
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
	return s.store.Add(ctx, w)
}

// Retrieve an Workflow from FuseML.
func (s *workflowsrvc) Get(ctx context.Context, w *workflow.GetPayload) (res *workflow.Workflow, err error) {
	s.logger.Print("workflow.get")
	id, err := uuid.Parse(w.WorkflowNameOrID)
	if err != nil {
		return nil, workflow.MakeBadRequest(err)
	}
	wf := s.store.Find(ctx, id)
	if wf == nil {
		return nil, workflow.MakeNotFound(errors.New("could not find a workflow with the specified ID"))
	}
	return wf, nil
}
