package manager

import (
	"context"
	"time"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

// createWorkflowListenerTimeout is the time (in minutes) that FuseML waits for the workflow listener
// to be available
const createWorkflowListenerTimeout = 1

type workflowManager struct {
	workflowBackend domain.WorkflowBackend
	workflowStore   domain.WorkflowStore
	codesetStore    domain.CodesetStore
}

// NewWorkflowManager initializes a Workflow Manager
func NewWorkflowManager(workflowBackend domain.WorkflowBackend, workflowStore domain.WorkflowStore, codesetStore domain.CodesetStore) domain.WorkflowManager {
	return &workflowManager{workflowBackend, workflowStore, codesetStore}
}

func (mgr *workflowManager) List(ctx context.Context, name *string) []*workflow.Workflow {
	return mgr.workflowStore.GetWorkflows(ctx, name)
}

func (mgr *workflowManager) Create(ctx context.Context, wf *workflow.Workflow) (*workflow.Workflow, error) {
	workflowDateCreated := time.Now().Format(time.RFC3339)
	wf.Created = &workflowDateCreated
	err := mgr.workflowBackend.CreateWorkflow(ctx, wf)
	if err != nil {
		return nil, err
	}
	return mgr.workflowStore.AddWorkflow(ctx, wf)
}

func (mgr *workflowManager) Get(ctx context.Context, name string) (*workflow.Workflow, error) {
	return mgr.workflowStore.GetWorkflow(ctx, name)
}

func (mgr *workflowManager) Delete(ctx context.Context, name string) error {
	return nil
}

func (mgr *workflowManager) AssignToCodeset(ctx context.Context, name, codesetProject, codesetName string) (wfListener *domain.WorkflowListener, webhookID *int64, err error) {
	_, err = mgr.workflowStore.GetWorkflow(ctx, name)
	if err != nil {
		return nil, nil, err
	}

	codeset, err := mgr.codesetStore.Find(ctx, codesetProject, codesetName)
	if err != nil {
		return nil, nil, err
	}

	wfListener, err = mgr.workflowBackend.CreateWorkflowListener(ctx, name, createWorkflowListenerTimeout*time.Minute)
	if err != nil {
		return nil, nil, err
	}

	webhookID, err = mgr.codesetStore.CreateWebhook(ctx, codeset, wfListener.URL)
	if err != nil {
		return nil, nil, err
	}

	mgr.workflowStore.AddCodesetAssignment(ctx, name, &domain.AssignedCodeset{Codeset: codeset, WebhookID: webhookID})
	mgr.workflowBackend.CreateWorkflowRun(ctx, name, codeset)
	return
}

func (mgr *workflowManager) UnassignFromCodeset(ctx context.Context, name, codesetProject, codesetName string) (err error) {
	codeset, err := mgr.codesetStore.Find(ctx, codesetProject, codesetName)
	if err != nil {
		return err
	}

	assignment, err := mgr.workflowStore.GetAssignedCodeset(ctx, name, codeset)
	if err != nil {
		return err
	}

	if assignment.WebhookID != nil {
		err = mgr.codesetStore.DeleteWebhook(ctx, codeset, assignment.WebhookID)
		if err != nil {
			return err
		}
	}

	if len(mgr.workflowStore.GetAssignedCodesets(ctx, name)) == 1 {
		err = mgr.workflowBackend.DeleteWorkflowListener(ctx, name)
		if err != nil {
			return err
		}
	}

	mgr.workflowStore.DeleteCodesetAssignment(ctx, name, codeset)
	return
}

func (mgr *workflowManager) ListAssignments(ctx context.Context, name *string) ([]*workflow.WorkflowAssignment, error) {
	return nil, nil
}

func (mgr *workflowManager) ListRuns(ctx context.Context, filter *domain.WorkflowRunFilter) ([]*workflow.WorkflowRun, error) {
	return nil, nil
}
