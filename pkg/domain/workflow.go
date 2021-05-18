package domain

import (
	"context"
	"log"

	"github.com/fuseml/fuseml-core/gen/workflow"
)

// WorkflowStore is an inteface to workflow stores
type WorkflowStore interface {
	GetWorkflow(ctx context.Context, name string) *workflow.Workflow
	GetAllWorkflows(ctx context.Context, name *string) (result []*workflow.Workflow)
	AddWorkflow(ctx context.Context, w *workflow.Workflow) (*workflow.Workflow, error)
	GetAssignedCodesets(ctx context.Context, workflowName string) []*Codeset
	GetAssignments(ctx context.Context, workflowName *string) map[string][]*Codeset
	AddCodesetAssignment(ctx context.Context, workflowName string, codeset *Codeset) []*Codeset
}

// WorkflowBackend is the interface for the FuseML workflows
type WorkflowBackend interface {
	CreateWorkflow(context.Context, *log.Logger, *workflow.Workflow) error
	CreateWorkflowRun(context.Context, string, *Codeset) error
	ListWorkflowRuns(context.Context, workflow.Workflow, WorkflowRunFilter) ([]*workflow.WorkflowRun, error)
	CreateWorkflowListener(context.Context, *log.Logger, string, bool) (*WorkflowListener, error)
	GetWorkflowListener(context.Context, *log.Logger, string) (*WorkflowListener, error)
}

// WorkflowRunFilter defines the available filter when listing workflow runs
type WorkflowRunFilter struct {
	ByLabel  []string
	ByStatus []string
}

// WorkflowListener defines a listener for a workflow
type WorkflowListener struct {
	Name         string
	Available    bool
	URL          string
	DashboardURL string
}
