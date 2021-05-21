package svc

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/core/config"
	"github.com/fuseml/fuseml-core/pkg/core/tekton"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

// workflow service example implementation.
// The example methods log the requests and return zero values.
type workflowsrvc struct {
	logger       *log.Logger
	store        domain.WorkflowStore
	codesetStore domain.CodesetStore
	backend      domain.WorkflowBackend
}

// NewWorkflowService returns the workflow service implementation.
func NewWorkflowService(logger *log.Logger, store domain.WorkflowStore,
	codesetStore domain.CodesetStore) (workflow.Service, error) {
	backend, err := tekton.NewWorkflowBackend(config.FuseMLNamespace)
	if err != nil {
		return nil, err
	}
	return &workflowsrvc{logger, store, codesetStore, backend}, nil
}

// List Workflows.
func (s *workflowsrvc) List(ctx context.Context, w *workflow.ListPayload) (res []*workflow.Workflow, err error) {
	s.logger.Print("workflow.list")
	return s.store.GetAllWorkflows(ctx, w.Name), nil
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
	return s.getWorkflow(ctx, w.Name)
}

// Delete a Workflow and its assignments.
func (s *workflowsrvc) Delete(ctx context.Context, d *workflow.DeletePayload) (err error) {
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

	wfListener, err := s.backend.CreateWorkflowListener(ctx, s.logger, w.Name, true)
	if err != nil {
		s.logger.Print(err)
		return err
	}

	_, err = s.codesetStore.CreateWebhook(ctx, codeset, wfListener.URL)
	if err != nil {
		s.logger.Print(err)
		return err
	}

	s.store.AddCodesetAssignment(ctx, w.Name, codeset)

	err = s.backend.CreateWorkflowRun(ctx, w.Name, codeset)
	if err != nil {
		s.logger.Print(err)
		return err
	}
	return nil
}

// Unassign a Workflow from a Codeset.
func (s *workflowsrvc) Unassign(ctx context.Context, u *workflow.UnassignPayload) (err error) {
	return nil
}

// ListAssignments lists Workflow assignments.
func (s *workflowsrvc) ListAssignments(ctx context.Context, w *workflow.ListAssignmentsPayload) (res []*workflow.WorkflowAssignment, err error) {
	listeners := map[string]*domain.WorkflowListener{}
	for wf, codesets := range s.store.GetAssignments(ctx, w.Name) {
		var listener *domain.WorkflowListener
		if l, ok := listeners[wf]; ok {
			listener = l
		} else {
			listener, err = s.backend.GetWorkflowListener(ctx, s.logger, wf)
			if err != nil {
				s.logger.Print(err)
				return nil, err
			}
			listeners[wf] = listener
		}
		assignment := newRestWorkflowAssignment(wf, codesets, listener)
		res = append(res, assignment)
	}
	return
}

// List Workflow runs.
func (s *workflowsrvc) ListRuns(ctx context.Context, w *workflow.ListRunsPayload) ([]*workflow.WorkflowRun, error) {
	s.logger.Print("workflow.listRuns")
	workflowRuns := []*workflow.WorkflowRun{}
	workflows := s.store.GetAllWorkflows(ctx, w.WorkflowName)
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

func newRestWorkflowAssignment(workflowName string, codesets []*domain.Codeset, listener *domain.WorkflowListener) *workflow.WorkflowAssignment {
	assignment := &workflow.WorkflowAssignment{
		Workflow: &workflowName,
		Status: &workflow.WorkflowAssignmentStatus{
			Available: &listener.Available,
			URL:       &listener.DashboardURL,
		},
	}
	for _, c := range codesets {
		assignment.Codesets = append(assignment.Codesets, &workflow.Codeset{
			Name:        c.Name,
			Project:     c.Project,
			Description: c.Description,
			Labels:      c.Labels,
			URL:         &c.URL,
		})
	}
	return assignment
}
