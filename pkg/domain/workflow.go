package domain

import (
	"context"

	"github.com/fuseml/fuseml-core/gen/workflow"
)

// WorkflowStore is an inteface to workflow stores
type WorkflowStore interface {
	Find(ctx context.Context, name string) *workflow.Workflow
	GetAll(ctx context.Context, name string) (result []*workflow.Workflow)
	Add(ctx context.Context, w *workflow.Workflow) (*workflow.Workflow, error)
	AssignCodeset(ctx context.Context, w *workflow.Workflow, c *Codeset) error
	GetAllRuns(ctx context.Context, w *workflow.Workflow, filters WorkflowRunFilter) ([]*workflow.WorkflowRun, error)
}

// WorkflowBackend is the interface for the FuseML workflows
type WorkflowBackend interface {
	CreateListener(context.Context, string, bool) (string, error)
	CreateWorkflow(context.Context, *workflow.Workflow) error
	CreateWorkflowRun(context.Context, string, *Codeset) error
	ListWorkflowRuns(context.Context, *workflow.Workflow, WorkflowRunFilter) ([]*workflow.WorkflowRun, error)
}

// WorkflowRunFilter defines the available filter when listing workflow runs
type WorkflowRunFilter struct {
	ByLabel  []string
	ByStatus []string
}
