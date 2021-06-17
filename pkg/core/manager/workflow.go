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

// WorkflowManager implements the domain.WorkflowManager interface
type WorkflowManager struct {
	workflowBackend domain.WorkflowBackend
	workflowStore   domain.WorkflowStore
	codesetStore    domain.CodesetStore
}

// NewWorkflowManager initializes a Workflow Manager
// FIXME: instead of CodesetStore, receive a CodesetManager
func NewWorkflowManager(workflowBackend domain.WorkflowBackend, workflowStore domain.WorkflowStore, codesetStore domain.CodesetStore) *WorkflowManager {
	return &WorkflowManager{workflowBackend, workflowStore, codesetStore}
}

// List Workflows.
func (mgr *WorkflowManager) List(ctx context.Context, name *string) []*workflow.Workflow {
	return mgr.workflowStore.GetWorkflows(ctx, name)
}

// Create a new Workflow.
func (mgr *WorkflowManager) Create(ctx context.Context, wf *workflow.Workflow) (*workflow.Workflow, error) {
	workflowDateCreated := time.Now().Format(time.RFC3339)
	wf.Created = &workflowDateCreated
	err := mgr.workflowBackend.CreateWorkflow(ctx, wf)
	if err != nil {
		return nil, err
	}
	return mgr.workflowStore.AddWorkflow(ctx, wf)
}

// Get a Workflow.
func (mgr *WorkflowManager) Get(ctx context.Context, name string) (*workflow.Workflow, error) {
	return mgr.workflowStore.GetWorkflow(ctx, name)
}

// Delete a Workflow and its assignments.
func (mgr *WorkflowManager) Delete(ctx context.Context, name string) error {
	// unassign all assigned codesets, if there's any
	assignedCodesets := mgr.workflowStore.GetAssignedCodesets(ctx, name)
	for _, ac := range assignedCodesets {
		err := mgr.UnassignFromCodeset(ctx, name, ac.Codeset.Project, ac.Codeset.Name)
		if err != nil {
			return err
		}
	}

	// delete tekton pipeline
	err := mgr.workflowBackend.DeleteWorkflow(ctx, name)
	if err != nil {
		return err
	}

	// delete workflow
	err = mgr.workflowStore.DeleteWorkflow(ctx, name)
	if err != nil {
		return err
	}
	return nil
}

// AssignToCodeset assigns a Workflow to a Codeset.
func (mgr *WorkflowManager) AssignToCodeset(ctx context.Context, name, codesetProject, codesetName string) (wfListener *domain.WorkflowListener, webhookID *int64, err error) {
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

// UnassignFromCodeset unassign a Workflow from a Codeset
func (mgr *WorkflowManager) UnassignFromCodeset(ctx context.Context, name, codesetProject, codesetName string) (err error) {
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

// ListAssignments lists Workflow assignments.
func (mgr *WorkflowManager) ListAssignments(ctx context.Context, name *string) ([]*workflow.WorkflowAssignment, error) {
	assignments := []*workflow.WorkflowAssignment{}
	for wf, acs := range mgr.workflowStore.GetAssignments(ctx, name) {

		listener, err := mgr.workflowBackend.GetWorkflowListener(ctx, wf)
		if err != nil {
			return nil, err
		}

		assignment := newWorkflowAssignment(wf, acs, listener)
		assignments = append(assignments, assignment)
	}
	return assignments, nil
}

// ListRuns lists Workflow runs.
func (mgr *WorkflowManager) ListRuns(ctx context.Context, filter *domain.WorkflowRunFilter) ([]*workflow.WorkflowRun, error) {
	workflowRuns := []*workflow.WorkflowRun{}
	var wfName *string
	if filter != nil {
		wfName = filter.WorkflowName
	}
	workflows := mgr.workflowStore.GetWorkflows(ctx, wfName)

	for _, workflow := range workflows {
		runs, err := mgr.workflowBackend.ListWorkflowRuns(ctx, workflow, filter)
		if err != nil {
			return nil, err
		}
		workflowRuns = append(workflowRuns, runs...)
	}

	return workflowRuns, nil
}

func newWorkflowAssignment(workflowName string, codesets []*domain.AssignedCodeset, listener *domain.WorkflowListener) *workflow.WorkflowAssignment {
	assignment := &workflow.WorkflowAssignment{
		Workflow: &workflowName,
		Status: &workflow.WorkflowAssignmentStatus{
			Available: &listener.Available,
			URL:       &listener.DashboardURL,
		},
	}
	for _, c := range codesets {
		assignment.Codesets = append(assignment.Codesets, &workflow.Codeset{
			Name:        c.Codeset.Name,
			Project:     c.Codeset.Project,
			Description: c.Codeset.Description,
			Labels:      c.Codeset.Labels,
			URL:         &c.Codeset.URL,
		})
	}
	return assignment
}
