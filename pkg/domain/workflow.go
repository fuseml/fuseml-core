package domain

import (
	"context"
	"log"

	"github.com/fuseml/fuseml-core/gen/codeset"
	"github.com/fuseml/fuseml-core/gen/workflow"
)

// WorkflowStore is an inteface to workflow stores
type WorkflowStore interface {
	Find(ctx context.Context, name string) *workflow.Workflow
	GetAll(ctx context.Context, name string) (result []*workflow.Workflow)
	Add(ctx context.Context, r *workflow.Workflow) (*workflow.Workflow, error)
}

// WorkflowBackend is the interface for the FuseML workflows
type WorkflowBackend interface {
	CreateListener(context.Context, *log.Logger, string, bool) (string, error)
	CreateWorkflow(context.Context, *log.Logger, *workflow.Workflow) error
	CreateWorkflowRun(context.Context, string, codeset.Codeset) error
	ListWorkflowRuns(context.Context, workflow.Workflow, WorkflowRunFilter) ([]*workflow.WorkflowRun, error)
}

// WorkflowRunFilter defines the available filter when listing workflow runs
type WorkflowRunFilter struct {
	ByLabel  []string
	ByStatus []string
}
