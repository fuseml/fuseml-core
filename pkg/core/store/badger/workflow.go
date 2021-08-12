package badger

import (
	"context"

	"github.com/fuseml/fuseml-core/pkg/domain"
	"github.com/timshannon/badgerhold/v3"
)

// WorkflowStore is a wrapper around a badgerhold.Store that implements the domain.WorkflowStore interface.
type WorkflowStore struct {
	store *badgerhold.Store
}

// NewWorkflowStore creates a new WorkflowStore.
func NewWorkflowStore(store *badgerhold.Store) *WorkflowStore {
	return &WorkflowStore{store: store}
}

// GetWorkflow returns a workflow identified by its name.
func (ws *WorkflowStore) GetWorkflow(ctx context.Context, name string) (*domain.Workflow, error) {
	wf := &domain.Workflow{}
	err := ws.store.Get(name, wf)
	if err != nil {
		return nil, domain.ErrWorkflowNotFound
	}
	return wf, nil
}

// GetWorkflows returns all workflows or the one that matches a given name.
func (ws *WorkflowStore) GetWorkflows(ctx context.Context, name *string) []*domain.Workflow {
	result := []*domain.Workflow{}
	if name != nil {
		wf := &domain.Workflow{}
		err := ws.store.Get(name, wf)
		if err == nil {
			result = append(result, wf)
		}
		return result
	}

	err := ws.store.Find(&result, nil)
	if err != nil {
		return result
	}
	return result
}

// AddWorkflow adds a new workflow based on the Workflow structure provided as argument.
func (ws *WorkflowStore) AddWorkflow(ctx context.Context, w *domain.Workflow) (*domain.Workflow, error) {
	err := ws.store.Insert(w.Name, w)
	if err != nil {
		return nil, domain.ErrWorkflowExists
	}
	return w, nil
}

// DeleteWorkflow deletes the workflow from the store.
func (ws *WorkflowStore) DeleteWorkflow(ctx context.Context, name string) error {
	wf := domain.Workflow{}
	err := ws.store.Get(name, &wf)
	if err != nil {
		return nil
	}
	if len(wf.GetCodesetAssignments(ctx)) > 0 {
		return domain.ErrCannotDeleteAssignedWorkflow
	}

	return ws.store.Delete(name, wf)
}

// GetCodesetAssignment returns a list of codesets assigned to the specified workflow.
func (ws *WorkflowStore) GetCodesetAssignment(ctx context.Context, workflowName string, codeset *domain.Codeset) (*domain.CodesetAssignment, error) {
	wf := domain.Workflow{}
	err := ws.store.Get(workflowName, &wf)
	if err != nil {
		return nil, domain.ErrWorkflowNotFound
	}

	return wf.GetCodesetAssignment(ctx, codeset)
}

// GetCodesetAssignments returns a AssignedCodeset for the Workflow and Codeset.
func (ws *WorkflowStore) GetCodesetAssignments(ctx context.Context, workflowName string) []*domain.CodesetAssignment {
	wf := domain.Workflow{}
	err := ws.store.Get(workflowName, &wf)
	if err == nil {
		return wf.GetCodesetAssignments(ctx)
	}

	return []*domain.CodesetAssignment{}
}

// GetAllCodesetAssignments returns a map of workflows and its assigned codesets.
func (ws *WorkflowStore) GetAllCodesetAssignments(ctx context.Context, workflowName *string) (result map[string][]*domain.CodesetAssignment) {
	result = make(map[string][]*domain.CodesetAssignment)
	if workflowName != nil {
		wf := domain.Workflow{}
		err := ws.store.Get(*workflowName, &wf)
		assignments := wf.GetCodesetAssignments(ctx)
		if err != nil || len(assignments) == 0 {
			return
		}
		result[*workflowName] = assignments
		return
	}

	workflows := []*domain.Workflow{}
	ws.store.Find(&workflows, nil)
	for _, wf := range workflows {
		assignments := wf.GetCodesetAssignments(ctx)
		if len(assignments) > 0 {
			result[wf.Name] = assignments
		}
	}
	return
}

// AddCodesetAssignment adds a codeset to the list of assigned codesets of a workflow if it does not already exists.
func (ws *WorkflowStore) AddCodesetAssignment(ctx context.Context, workflowName string, codeset *domain.Codeset,
	webhookID *int64) ([]*domain.CodesetAssignment, error) {
	wf := domain.Workflow{}
	err := ws.store.Get(workflowName, &wf)
	if err != nil {
		return nil, domain.ErrWorkflowNotFound
	}

	err = wf.AssignToCodeset(ctx, codeset, webhookID)
	if err != nil {
		return nil, err
	}

	err = ws.store.Update(workflowName, &wf)
	if err != nil {
		return nil, err
	}
	return wf.GetCodesetAssignments(ctx), nil
}

// DeleteCodesetAssignment deletes a codeset from the list of assigned codesets of a workflow if it exists.
func (ws *WorkflowStore) DeleteCodesetAssignment(ctx context.Context, workflowName string, codeset *domain.Codeset) ([]*domain.CodesetAssignment, error) {
	wf := domain.Workflow{}
	err := ws.store.Get(workflowName, &wf)
	if err != nil {
		return nil, domain.ErrWorkflowNotFound
	}

	err = wf.UnassignFromCodeset(ctx, codeset)
	if err != nil {
		return nil, err
	}

	err = ws.store.Update(workflowName, &wf)
	if err != nil {
		return nil, err
	}

	return wf.GetCodesetAssignments(ctx), nil
}
