package svc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/core/config"
	"github.com/fuseml/fuseml-core/pkg/core/tekton"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

// createWorkflowListenerTimeout is the time (in minutes) that FuseML waits for the workflow listener
// to be available
const createWorkflowListenerTimeout = 1

// workflow service example implementation.
// The example methods log the requests and return zero values.
type workflowsrvc struct {
	logger       *log.Logger
	store        domain.WorkflowStore
	codesetStore domain.CodesetStore
	backend      domain.WorkflowBackend
}

// NewWorkflowService returns the workflow service implementation.
func NewWorkflowService(logger *log.Logger, store domain.WorkflowStore, codesetStore domain.CodesetStore) (workflow.Service, error) {
	backend, err := tekton.NewWorkflowBackend(config.FuseMLNamespace)
	if err != nil {
		return nil, err
	}
	return &workflowsrvc{logger, store, codesetStore, backend}, nil
}

// List Workflows.
func (s *workflowsrvc) List(ctx context.Context, w *workflow.ListPayload) (res []*workflow.Workflow, err error) {
	s.logger.Print("workflow.list")
	return s.store.GetWorkflows(ctx, w.Name), nil
}

// Register a new Workflow.
func (s *workflowsrvc) Register(ctx context.Context, w *workflow.Workflow) (res *workflow.Workflow, err error) {
	s.logger.Print("workflow.register")
	err = s.backend.CreateWorkflow(ctx, s.logger, w)
	if err != nil {
		s.logger.Print(err)
		if err.Error() == "workflow already exists" {
			return nil, workflow.MakeConflict(err)
		}
		return nil, err
	}
	return s.store.AddWorkflow(ctx, w)
}

// Get a Workflow.
func (s *workflowsrvc) Get(ctx context.Context, w *workflow.GetPayload) (res *workflow.Workflow, err error) {
	s.logger.Print("workflow.get")
	return s.getWorkflow(ctx, w.Name)
}

// Delete a Workflow and its assignments.
func (s *workflowsrvc) Delete(ctx context.Context, d *workflow.DeletePayload) (err error) {
	s.logger.Print("workflow.delete")
	if _, err := s.getWorkflow(ctx, d.Name); err != nil {
		s.logger.Print(err)
		return err
	}

	// unassign all assigned codesets, if there's any
	assignedCodesets := s.store.GetAssignedCodesets(ctx, d.Name)
	for _, ac := range assignedCodesets {
		err := s.unassignCodesetFromWorkflow(ctx, d.Name, ac.Codeset)
		if err != nil {
			return err
		}
	}

	// delete tekton pipeline
	err = s.backend.DeleteWorkflow(ctx, s.logger, d.Name)
	if err != nil {
		s.logger.Print(err)
		return err
	}

	// delete workflow
	err = s.store.DeleteWorkflow(ctx, d.Name)
	if err != nil {
		s.logger.Print(err)
		return err
	}
	return nil
}

// Assign a Workflow to a Codeset.
func (s *workflowsrvc) Assign(ctx context.Context, w *workflow.AssignPayload) (err error) {
	s.logger.Print("workflow.assign")
	if _, err = s.getWorkflow(ctx, w.Name); err != nil {
		s.logger.Print(err)
		return err
	}

	codeset, err := s.codesetStore.Find(ctx, w.CodesetProject, w.CodesetName)
	if err != nil {
		s.logger.Print(err)
		return workflow.MakeNotFound(err)
	}

	wfListener, err := s.backend.CreateWorkflowListener(ctx, s.logger, w.Name, createWorkflowListenerTimeout*time.Minute)
	if err != nil {
		s.logger.Print(err)
		return err
	}

	webhookID, err := s.codesetStore.CreateWebhook(ctx, codeset, wfListener.URL)
	if err != nil {
		s.logger.Print(err)
		return err
	}

	s.store.AddCodesetAssignment(ctx, w.Name, &domain.AssignedCodeset{Codeset: codeset, WebhookID: webhookID})

	err = s.backend.CreateWorkflowRun(ctx, s.logger, w.Name, codeset)
	if err != nil {
		s.logger.Print(err)
		return err
	}
	return nil
}

// Unassign a Workflow from a Codeset.
func (s *workflowsrvc) Unassign(ctx context.Context, u *workflow.UnassignPayload) (err error) {
	s.logger.Print("workflow.unassign")
	codeset, err := s.codesetStore.Find(ctx, u.CodesetProject, u.CodesetName)
	if err != nil {
		s.logger.Print(err)
		return workflow.MakeNotFound(err)
	}
	return s.unassignCodesetFromWorkflow(ctx, u.Name, codeset)
}

// ListAssignments lists Workflow assignments.
func (s *workflowsrvc) ListAssignments(ctx context.Context, w *workflow.ListAssignmentsPayload) (res []*workflow.WorkflowAssignment, err error) {
	s.logger.Print("workflow.listAssignments")
	listeners := map[string]*domain.WorkflowListener{}
	for wf, acs := range s.store.GetAssignments(ctx, w.Name) {
		var listener *domain.WorkflowListener
		if l, ok := listeners[wf]; ok {
			listener = l
		} else {
			listener, err = s.backend.GetWorkflowListener(ctx, wf)
			if err != nil {
				s.logger.Print(err)
				return nil, err
			}
			listeners[wf] = listener
		}
		assignment := newRestWorkflowAssignment(wf, acs, listener)
		res = append(res, assignment)
	}
	return
}

// List Workflow runs.
func (s *workflowsrvc) ListRuns(ctx context.Context, w *workflow.ListRunsPayload) ([]*workflow.WorkflowRun, error) {
	s.logger.Print("workflow.listRuns")
	workflowRuns := []*workflow.WorkflowRun{}
	workflows := s.store.GetWorkflows(ctx, w.Name)
	filters := domain.WorkflowRunFilter{}
	if w.CodesetName != nil {
		filters.ByLabel = append(filters.ByLabel, fmt.Sprintf("%s=%s", tekton.LabelCodesetName, *w.CodesetName))
	}
	if w.CodesetProject != nil {
		filters.ByLabel = append(filters.ByLabel, fmt.Sprintf("%s=%s", tekton.LabelCodesetProject, *w.CodesetProject))
	}
	if w.Status != nil {
		filters.ByStatus = append(filters.ByStatus, *w.Status)
	}

	for _, workflow := range workflows {
		runs, err := s.backend.ListWorkflowRuns(ctx, *workflow, filters)
		if err != nil {
			s.logger.Print(err)
			return nil, err
		}
		workflowRuns = append(workflowRuns, runs...)
	}

	return workflowRuns, nil
}

func (s *workflowsrvc) getWorkflow(ctx context.Context, name string) (*workflow.Workflow, error) {
	if name == "" {
		return nil, workflow.MakeBadRequest(errors.New("empty workflow name"))
	}
	wf := s.store.GetWorkflow(ctx, name)
	if wf == nil {
		return nil, workflow.MakeNotFound(errors.New("could not find a workflow with the specified name"))
	}
	return wf, nil
}

func newRestWorkflowAssignment(workflowName string, codesets []*domain.AssignedCodeset, listener *domain.WorkflowListener) *workflow.WorkflowAssignment {
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

func (s *workflowsrvc) unassignCodesetFromWorkflow(ctx context.Context, workflowName string, codeset *domain.Codeset) (err error) {
	assignment := s.store.GetAssignedCodeset(ctx, workflowName, codeset)
	if assignment == nil {
		err = fmt.Errorf("workflow not assigned to codeset")
		s.logger.Print(err)
		return workflow.MakeNotFound(err)
	}

	err = s.codesetStore.DeleteWebhook(ctx, codeset, assignment.WebhookID)
	if err != nil {
		s.logger.Print(err)
		return err
	}

	if len(s.store.GetAssignedCodesets(ctx, workflowName)) == 1 {
		err = s.backend.DeleteWorkflowListener(ctx, s.logger, workflowName)
		if err != nil {
			s.logger.Print(err)
			return err
		}
	}
	s.store.DeleteCodesetAssignment(ctx, workflowName, codeset)
	return
}
