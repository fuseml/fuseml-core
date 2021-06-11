package domain

import (
	"context"
	"time"

	"github.com/fuseml/fuseml-core/gen/workflow"
)

const (
	// ErrWorkflowExists describes the error message returned when trying to create a workflow with that already exists.
	ErrWorkflowExists = WorkflowErr("workflow already exists")
	// ErrWorkflowNotFound describes the error message returned when trying to get a workflow that does not exist.
	ErrWorkflowNotFound = WorkflowErr("could not find a workflow with the specified name")
	// ErrWorkflowNotAssignedToCodeset describes the error message returned when trying to unassign a workflow from a codeset
	// but it is not assigned to the codeset.
	ErrWorkflowNotAssignedToCodeset = WorkflowErr("workflow not assigned to codeset")
)

// WorkflowErr are expected errors returned when performing operations on workflows
type WorkflowErr string

func (e WorkflowErr) Error() string {
	return string(e)
}

// WorkflowManager describes the interface for a Workflow Manager
type WorkflowManager interface {
	Create(ctx context.Context, workflow *workflow.Workflow) (*workflow.Workflow, error)
	Get(ctx context.Context, name string) (*workflow.Workflow, error)
	Delete(ctx context.Context, name string) error
	List(ctx context.Context, name *string) []*workflow.Workflow
	AssignToCodeset(ctx context.Context, name, codesetProject, codesetName string) (*WorkflowListener, *int64, error)
	UnassignFromCodeset(ctx context.Context, name, codesetProject, codesetName string) error
	ListAssignments(ctx context.Context, name *string) ([]*workflow.WorkflowAssignment, error)
	ListRuns(ctx context.Context, filter *WorkflowRunFilter) ([]*workflow.WorkflowRun, error)
}

// WorkflowStore is an interface to workflow stores
type WorkflowStore interface {
	GetWorkflow(ctx context.Context, name string) (*workflow.Workflow, error)
	GetWorkflows(ctx context.Context, name *string) []*workflow.Workflow
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
	CreateWorkflow(ctx context.Context, workflow *workflow.Workflow) error
	DeleteWorkflow(ctx context.Context, workflowName string) error
	CreateWorkflowRun(ctx context.Context, workflowName string, codeset *Codeset) error
	ListWorkflowRuns(ctx context.Context, workflow *workflow.Workflow, filter *WorkflowRunFilter) ([]*workflow.WorkflowRun, error)
	CreateWorkflowListener(ctx context.Context, workflowName string, timeout time.Duration) (*WorkflowListener, error)
	DeleteWorkflowListener(ctx context.Context, workflowName string) error
	GetWorkflowListener(ctx context.Context, workflowName string) (*WorkflowListener, error)
}

// WorkflowRunFilter defines the available filter when listing workflow runs
type WorkflowRunFilter struct {
	WorkflowName   *string
	CodesetName    string
	CodesetProject string
	Status         []string
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
