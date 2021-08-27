package core

import (
	"context"
	"fmt"

	"github.com/fuseml/fuseml-core/pkg/domain"
)

// storableWorkflow holds the workflow and the codesets assigned to it
type storableWorkflow struct {
	// Workflow assigned to the codeset
	workflow *domain.Workflow
	// AssignedCodeset holds codesets assigned to the workflow and its hookID
	assignedCodesets []*domain.CodesetAssignment
}

// WorkflowStore describes in memory store for workflows
type WorkflowStore struct {
	items map[string]*storableWorkflow
}

// NewWorkflowStore returns an in-memory workflow store instance
func NewWorkflowStore() *WorkflowStore {
	return &WorkflowStore{
		items: make(map[string]*storableWorkflow),
	}
}

// GetWorkflow returns a workflow identified by its name
func (ws *WorkflowStore) GetWorkflow(ctx context.Context, name string) (*domain.Workflow, error) {
	if _, exists := ws.items[name]; exists {
		return ws.items[name].workflow, nil
	}
	return nil, domain.ErrWorkflowNotFound
}

// GetWorkflows returns all workflows or the one that matches a given name.
func (ws *WorkflowStore) GetWorkflows(ctx context.Context, name *string) (result []*domain.Workflow) {
	result = make([]*domain.Workflow, 0, len(ws.items))
	if name != nil {
		if sw, ok := ws.items[*name]; ok {
			result = append(result, sw.workflow)
			return
		}
		return result

	}
	for _, sw := range ws.items {
		result = append(result, sw.workflow)
	}
	return
}

// AddWorkflow adds a new workflow based on the Workflow structure provided as argument
func (ws *WorkflowStore) AddWorkflow(ctx context.Context, w *domain.Workflow) (*domain.Workflow, error) {
	if _, exists := ws.items[w.Name]; exists {
		return nil, domain.ErrWorkflowExists
	}
	sw := storableWorkflow{workflow: w}
	ws.items[w.Name] = &sw
	return w, nil
}

// DeleteWorkflow deletes the workflow from the store
func (ws *WorkflowStore) DeleteWorkflow(ctx context.Context, name string) error {
	sw, found := ws.items[name]
	if !found {
		return nil
	}
	if len(sw.assignedCodesets) > 0 {
		return fmt.Errorf("cannot delete workflow, there are codesets assigned to it")
	}
	delete(ws.items, name)
	return nil
}

// GetAssignedCodesets returns a list of codesets assigned to the specified workflow
func (ws *WorkflowStore) GetAssignedCodesets(ctx context.Context, workflowName string) []*domain.CodesetAssignment {
	if _, exists := ws.items[workflowName]; exists {
		return ws.items[workflowName].assignedCodesets
	}
	return nil
}

// GetAssignments returns a map of workflows and its assigned codesets
func (ws *WorkflowStore) GetAssignments(ctx context.Context, workflowName *string) (result map[string][]*domain.CodesetAssignment) {
	result = make(map[string][]*domain.CodesetAssignment, len(ws.items))
	if workflowName != nil {
		if sw, exists := ws.items[*workflowName]; exists && len(sw.assignedCodesets) > 0 {
			result[*workflowName] = sw.assignedCodesets
		}
		return
	}
	for _, sw := range ws.items {
		if len(sw.assignedCodesets) > 0 {
			result[sw.workflow.Name] = sw.assignedCodesets
		}
	}
	return
}

// AddCodesetAssignment adds a codeset to the list of assigned codesets of a workflow if it does not already exists
func (ws *WorkflowStore) AddCodesetAssignment(ctx context.Context, workflowName string,
	assignedCodeset *domain.CodesetAssignment) []*domain.CodesetAssignment {
	assignedCodesets := ws.items[workflowName].assignedCodesets
	if assigned, _ := getAssignedCodeset(assignedCodesets, assignedCodeset.Codeset); assigned == nil {
		assignedCodesets = append(assignedCodesets, assignedCodeset)
		ws.items[workflowName].assignedCodesets = assignedCodesets
	}
	return assignedCodesets
}

// DeleteCodesetAssignment deletes a codeset from the list of assigned codesets of a workflow if it exists
func (ws *WorkflowStore) DeleteCodesetAssignment(ctx context.Context, workflowName string, codeset *domain.Codeset) []*domain.CodesetAssignment {
	assignedCodesets := ws.items[workflowName].assignedCodesets
	if _, i := getAssignedCodeset(assignedCodesets, codeset); i != -1 {
		assignedCodesets = removeAssignedCodeset(assignedCodesets, i)
		ws.items[workflowName].assignedCodesets = assignedCodesets
	}
	return assignedCodesets
}

// GetAssignedCodeset returns a AssignedCodeset for the Workflow and Codeset
func (ws *WorkflowStore) GetAssignedCodeset(ctx context.Context, workflowName string, codeset *domain.Codeset) (*domain.CodesetAssignment, error) {
	sw, exists := ws.items[workflowName]
	if !exists {
		return nil, domain.ErrWorkflowNotFound
	}
	ac, _ := getAssignedCodeset(sw.assignedCodesets, codeset)
	if ac == nil {
		return nil, domain.ErrWorkflowNotAssignedToCodeset
	}
	return ac, nil
}

func getAssignedCodeset(assignedCodesets []*domain.CodesetAssignment, codeset *domain.Codeset) (*domain.CodesetAssignment, int) {
	for i, ac := range assignedCodesets {
		if ac.Codeset.Project == codeset.Project && ac.Codeset.Name == codeset.Name {
			return ac, i
		}
	}
	return nil, -1
}

func removeAssignedCodeset(codesets []*domain.CodesetAssignment, index int) []*domain.CodesetAssignment {
	codesets[index] = codesets[len(codesets)-1]
	return codesets[:len(codesets)-1]
}
