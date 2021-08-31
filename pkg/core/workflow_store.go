package core

import (
	"context"

	"github.com/fuseml/fuseml-core/pkg/domain"
)

// WorkflowStore describes in memory store for workflows
type WorkflowStore struct {
	items map[string]*domain.Workflow
}

// NewWorkflowStore returns an in-memory workflow store instance
func NewWorkflowStore() *WorkflowStore {
	return &WorkflowStore{
		items: make(map[string]*domain.Workflow),
	}
}

// GetWorkflow returns a workflow identified by its name
func (ws *WorkflowStore) GetWorkflow(ctx context.Context, name string) (*domain.Workflow, error) {
	if _, exists := ws.items[name]; exists {
		return ws.items[name], nil
	}
	return nil, domain.ErrWorkflowNotFound
}

// GetWorkflows returns all workflows or the one that matches a given name.
func (ws *WorkflowStore) GetWorkflows(ctx context.Context, name *string) (result []*domain.Workflow) {
	result = []*domain.Workflow{}
	if name != nil {
		if wf, ok := ws.items[*name]; ok {
			result = append(result, wf)
		}
		return
	}
	for _, wf := range ws.items {
		result = append(result, wf)
	}
	return
}

// AddWorkflow adds a new workflow based on the Workflow structure provided as argument
func (ws *WorkflowStore) AddWorkflow(ctx context.Context, w *domain.Workflow) (*domain.Workflow, error) {
	if _, exists := ws.items[w.Name]; exists {
		return nil, domain.ErrWorkflowExists
	}
	ws.items[w.Name] = w
	return w, nil
}

// DeleteWorkflow deletes the workflow from the store
func (ws *WorkflowStore) DeleteWorkflow(ctx context.Context, name string) error {
	wf, found := ws.items[name]
	if !found {
		return nil
	}
	if len(wf.GetCodesetAssignments(ctx)) > 0 {
		return domain.ErrCannotDeleteAssignedWorkflow
	}
	delete(ws.items, name)
	return nil
}

// GetCodesetAssignments returns a list of codesets assigned to the specified workflow
func (ws *WorkflowStore) GetCodesetAssignments(ctx context.Context, workflowName string) []*domain.CodesetAssignment {
	if wf, exists := ws.items[workflowName]; exists {
		return wf.GetCodesetAssignments(ctx)
	}
	return []*domain.CodesetAssignment{}
}

// GetAllCodesetAssignments returns a map of workflows and its assigned codesets
func (ws *WorkflowStore) GetAllCodesetAssignments(ctx context.Context, workflowName *string) (result map[string][]*domain.CodesetAssignment) {
	if workflowName != nil {
		result = make(map[string][]*domain.CodesetAssignment, 1)
		if wf, exists := ws.items[*workflowName]; exists && len(wf.GetCodesetAssignments(ctx)) > 0 {
			result[*workflowName] = wf.GetCodesetAssignments(ctx)
		}
		return
	}
	result = make(map[string][]*domain.CodesetAssignment, len(ws.items))
	for _, wf := range ws.items {
		if csa := wf.GetCodesetAssignments(ctx); len(csa) > 0 {
			result[wf.Name] = csa
		}
	}
	return
}

// AddCodesetAssignment adds a codeset assignment to the list of assigned codesets of a workflow
func (ws *WorkflowStore) AddCodesetAssignment(ctx context.Context, workflowName string, codeset *domain.Codeset,
	webhookID *int64) ([]*domain.CodesetAssignment, error) {
	wf, ok := ws.items[workflowName]
	if !ok {
		return nil, domain.ErrWorkflowNotFound
	}

	err := wf.AssignToCodeset(ctx, codeset, webhookID)
	if err != nil {
		return nil, err
	}

	return wf.GetCodesetAssignments(ctx), nil
}

// DeleteCodesetAssignment deletes a codeset assignment from the list of assigned codesets of a workflow
func (ws *WorkflowStore) DeleteCodesetAssignment(ctx context.Context, workflowName string, codeset *domain.Codeset) ([]*domain.CodesetAssignment, error) {
	wf, ok := ws.items[workflowName]
	if !ok {
		return nil, domain.ErrWorkflowNotFound
	}

	err := wf.UnassignFromCodeset(ctx, codeset)
	if err != nil {
		return nil, err
	}

	return wf.GetCodesetAssignments(ctx), nil
}

// GetCodesetAssignment returns a codeset assignment for a given Workflow and Codeset
func (ws *WorkflowStore) GetCodesetAssignment(ctx context.Context, workflowName string, codeset *domain.Codeset) (*domain.CodesetAssignment, error) {
	wf, exists := ws.items[workflowName]
	if !exists {
		return nil, domain.ErrWorkflowNotFound
	}

	return wf.GetCodesetAssignment(ctx, codeset)
}
