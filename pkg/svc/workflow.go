package svc

import (
	"context"
	"log"
	"strings"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

// workflow service example implementation.
// The example methods log the requests and return zero values.
type workflowsrvc struct {
	logger *log.Logger
	mgr    domain.WorkflowManager
}

// NewWorkflowService returns the workflow service implementation.
func NewWorkflowService(logger *log.Logger, workflowManager domain.WorkflowManager) workflow.Service {
	return &workflowsrvc{logger, workflowManager}
}

// List Workflows.
func (s *workflowsrvc) List(ctx context.Context, w *workflow.ListPayload) (res []*workflow.Workflow, err error) {
	s.logger.Print("workflow.list")
	return s.mgr.List(ctx, w.Name), nil
}

// Create a new Workflow.
func (s *workflowsrvc) Create(ctx context.Context, w *workflow.Workflow) (res *workflow.Workflow, err error) {
	s.logger.Print("workflow.create")
	res, err = s.mgr.Create(ctx, w)
	if err != nil {
		s.logger.Print(err)
		if err == domain.ErrWorkflowExists {
			return nil, workflow.MakeConflict(err)
		}
		return nil, err
	}
	return
}

// Get a Workflow.
func (s *workflowsrvc) Get(ctx context.Context, w *workflow.GetPayload) (res *workflow.Workflow, err error) {
	s.logger.Print("workflow.get")
	res, err = s.mgr.Get(ctx, w.Name)
	if err != nil {
		s.logger.Print(err)
		if err == domain.ErrWorkflowNotFound {
			return nil, workflow.MakeNotFound(err)
		}
		return nil, err
	}
	return
}

// Delete a Workflow and its assignments.
func (s *workflowsrvc) Delete(ctx context.Context, d *workflow.DeletePayload) (err error) {
	s.logger.Print("workflow.delete")
	err = s.mgr.Delete(ctx, d.Name)
	if err != nil {
		s.logger.Print(err)
		return
	}
	return
}

// Assign a Workflow to a Codeset.
func (s *workflowsrvc) Assign(ctx context.Context, w *workflow.AssignPayload) (err error) {
	s.logger.Print("workflow.assign")
	_, _, err = s.mgr.AssignToCodeset(ctx, w.Name, w.CodesetProject, w.CodesetName)
	if err != nil {
		s.logger.Print(err)
		// FIXME: codeset needs to thrown a known error when trying to get a codeset that does not exist
		// to properly compare the returned error.
		if err == domain.ErrWorkflowNotFound || strings.Contains(err.Error(), "Fetching Codeset failed") {
			return workflow.MakeNotFound(err)
		}
	}
	return
}

// Unassign a Workflow from a Codeset.
func (s *workflowsrvc) Unassign(ctx context.Context, u *workflow.UnassignPayload) (err error) {
	s.logger.Print("workflow.unassign")
	err = s.mgr.UnassignFromCodeset(ctx, u.Name, u.CodesetProject, u.CodesetName)
	if err != nil {
		s.logger.Print(err)
		if err == domain.ErrWorkflowNotFound || strings.Contains(err.Error(), "Fetching Codeset failed") || err == domain.ErrWorkflowNotAssignedToCodeset {
			return workflow.MakeNotFound(err)
		}
	}
	return
}

// ListAssignments lists Workflow assignments.
func (s *workflowsrvc) ListAssignments(ctx context.Context, w *workflow.ListAssignmentsPayload) (assignments []*workflow.WorkflowAssignment, err error) {
	s.logger.Print("workflow.listAssignments")
	assignments, err = s.mgr.ListAssignments(ctx, w.Name)
	if err != nil {
		return nil, err
	}
	return
}

// List Workflow runs.
func (s *workflowsrvc) ListRuns(ctx context.Context, w *workflow.ListRunsPayload) (runs []*workflow.WorkflowRun, err error) {
	s.logger.Print("workflow.listRuns")
	filter := domain.WorkflowRunFilter{WorkflowName: w.Name}
	if w.CodesetName != nil {
		filter.CodesetName = *w.CodesetName
	}
	if w.CodesetProject != nil {
		filter.CodesetProject = *w.CodesetProject
	}
	if w.Status != nil {
		filter.Status = []string{*w.Status}
	}
	runs, err = s.mgr.ListRuns(ctx, &filter)
	if err != nil {
		return nil, err
	}
	return
}
