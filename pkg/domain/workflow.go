package domain

import (
	"context"
	"log"
	"time"

	"github.com/fuseml/fuseml-core/gen/workflow"
)

// WorkflowStore is an inteface to workflow stores
type WorkflowStore interface {
	GetWorkflow(ctx context.Context, name string) *workflow.Workflow
	GetWorkflows(ctx context.Context, name *string) (result []*workflow.Workflow)
	AddWorkflow(ctx context.Context, w *workflow.Workflow) (*workflow.Workflow, error)
	DeleteWorkflow(ctx context.Context, name string) error
	GetAssignedCodeset(ctx context.Context, workflowName string, codeset *Codeset) *AssignedCodeset
	GetAssignedCodesets(ctx context.Context, workflowName string) []*AssignedCodeset
	GetAssignments(ctx context.Context, workflowName *string) map[string][]*AssignedCodeset
	AddCodesetAssignment(ctx context.Context, workflowName string, assignedCodeset *AssignedCodeset) []*AssignedCodeset
	DeleteCodesetAssignment(ctx context.Context, workflowName string, codeset *Codeset) []*AssignedCodeset
}

// WorkflowBackend is the interface for the FuseML workflows
type WorkflowBackend interface {
	CreateWorkflow(ctx context.Context, logger *log.Logger, workflow *workflow.Workflow) error
	DeleteWorkflow(ctx context.Context, logger *log.Logger, workflowName string) error
	CreateWorkflowRun(ctx context.Context, logger *log.Logger, workflowName string, codeset *Codeset) error
	ListWorkflowRuns(ctx context.Context, workflow workflow.Workflow, filter WorkflowRunFilter) ([]*workflow.WorkflowRun, error)
	CreateWorkflowListener(ctx context.Context, logger *log.Logger, workflowName string, timeout time.Duration) (*WorkflowListener, error)
	DeleteWorkflowListener(ctx context.Context, logger *log.Logger, workflowName string) error
	GetWorkflowListener(ctx context.Context, workflowName string) (*WorkflowListener, error)
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

// AssignedCodeset describes a assigned codeset its webhook ID
type AssignedCodeset struct {
	Codeset   *Codeset
	WebhookID *int64
}
