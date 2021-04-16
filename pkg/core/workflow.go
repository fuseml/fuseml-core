package fuseml

import (
	"context"
	"errors"
	"log"

	workflow "github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/google/uuid"
)

// workflow service example implementation.
// The example methods log the requests and return zero values.
type workflowsrvc struct {
	logger *log.Logger
}

// NewWorkflow returns the workflow service implementation.
func NewWorkflow(logger *log.Logger) workflow.Service {
	return &workflowsrvc{logger}
}

// Retrieve information about workflows registered in FuseML.
func (s *workflowsrvc) List(ctx context.Context, w *workflow.ListPayload) (res []*workflow.Workflow, err error) {
	s.logger.Print("workflow.list")
	name := "all"
	if w.Name != nil {
		name = *w.Name
	}

	return workflowStore.Get(name), nil
}

// Register a workflow with the FuseML workflow store.
func (s *workflowsrvc) Register(ctx context.Context, w *workflow.Workflow) (res *workflow.Workflow, err error) {
	s.logger.Print("workflow.register")
	return workflowStore.Add(w)
}

// Retrieve an Workflow from FuseML.
func (s *workflowsrvc) Get(ctx context.Context, w *workflow.GetPayload) (res *workflow.Workflow, err error) {
	s.logger.Print("workflow.get")
	id, err := uuid.Parse(w.WorkflowNameOrID)
	if err != nil {
		return nil, workflow.MakeBadRequest(err)
	}
	wf := workflowStore.Find(id)
	if wf == nil {
		return nil, workflow.MakeNotFound(errors.New("could not find a workflow with the specified ID"))
	}
	return wf, nil
}
