package fuseml

import (
	"context"
	"log"

	workflow "github.com/fuseml/fuseml-core/gen/workflow"
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
func (s *workflowsrvc) List(ctx context.Context, p *workflow.ListPayload) (res []*workflow.Workflow, err error) {
	s.logger.Print("workflow.list")
	return
}

// Register a workflow with the FuseML workflow store.
func (s *workflowsrvc) Register(ctx context.Context, p *workflow.Workflow) (res *workflow.Workflow, err error) {
	res = &workflow.Workflow{}
	s.logger.Print("workflow.register")
	return
}

// Retrieve an Workflow from FuseML.
func (s *workflowsrvc) Get(ctx context.Context, p *workflow.GetPayload) (res *workflow.Workflow, err error) {
	res = &workflow.Workflow{}
	s.logger.Print("workflow.get")
	return
}
