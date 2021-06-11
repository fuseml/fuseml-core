package core

import (
	"context"
	"fmt"
	"time"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

// storableWorkflow holds the workflow and the codesets assigned to it
type storableWorkflow struct {
	// Workflow assigned to the codeset
	workflow *workflow.Workflow
	// AssignedCodeset holds codesets assigned to the workflow and its hookID
	assignedCodesets []*domain.AssignedCodeset
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
func (ws *WorkflowStore) GetWorkflow(ctx context.Context, name string) (*workflow.Workflow, error) {
	if _, exists := ws.items[name]; exists {
		return ws.items[name].workflow, nil
	}
	return nil, domain.ErrWorkflowNotFound
}

// GetWorkflows returns all workflows or the one that matches a given name.
func (ws *WorkflowStore) GetWorkflows(ctx context.Context, name *string) (result []*workflow.Workflow) {
	result = make([]*workflow.Workflow, 0, len(ws.items))
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
func (ws *WorkflowStore) AddWorkflow(ctx context.Context, w *workflow.Workflow) (*workflow.Workflow, error) {
	if _, exists := ws.items[w.Name]; exists {
		return nil, fmt.Errorf("workflow %q already exists", w.Name)
	}
	workflowCreated := time.Now().Format(time.RFC3339)
	w.Created = &workflowCreated
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
func (ws *WorkflowStore) GetAssignedCodesets(ctx context.Context, workflowName string) []*domain.AssignedCodeset {
	if _, exists := ws.items[workflowName]; exists {
		return ws.items[workflowName].assignedCodesets
	}
	return nil
}

// GetAssignments returns a map of workflows and its assigned codesets
func (ws *WorkflowStore) GetAssignments(ctx context.Context, workflowName *string) (result map[string][]*domain.AssignedCodeset) {
	result = make(map[string][]*domain.AssignedCodeset, len(ws.items))
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
	assignedCodeset *domain.AssignedCodeset) []*domain.AssignedCodeset {
	assignedCodesets := ws.items[workflowName].assignedCodesets
	if assigned, _ := getAssignedCodeset(assignedCodesets, assignedCodeset.Codeset); assigned == nil {
		assignedCodesets = append(assignedCodesets, assignedCodeset)
		ws.items[workflowName].assignedCodesets = assignedCodesets
	}
	return assignedCodesets
}

// DeleteCodesetAssignment deletes a codeset from the list of assigned codesets of a workflow if it exists
func (ws *WorkflowStore) DeleteCodesetAssignment(ctx context.Context, workflowName string, codeset *domain.Codeset) []*domain.AssignedCodeset {
	assignedCodesets := ws.items[workflowName].assignedCodesets
	if _, i := getAssignedCodeset(assignedCodesets, codeset); i != -1 {
		assignedCodesets = removeAssignedCodeset(assignedCodesets, i)
		ws.items[workflowName].assignedCodesets = assignedCodesets
	}
	return assignedCodesets
}

// GetAssignedCodeset returns a AssignedCodeset for the Workflow and Codeset
func (ws *WorkflowStore) GetAssignedCodeset(ctx context.Context, workflowName string, codeset *domain.Codeset) *domain.AssignedCodeset {
	sw, exists := ws.items[workflowName]
	if !exists {
		return nil
	}
	ac, _ := getAssignedCodeset(sw.assignedCodesets, codeset)
	return ac
}

func getAssignedCodeset(assignedCodesets []*domain.AssignedCodeset, codeset *domain.Codeset) (*domain.AssignedCodeset, int) {
	for i, ac := range assignedCodesets {
		if ac.Codeset.Project == codeset.Project && ac.Codeset.Name == codeset.Name {
			return ac, i
		}
	}
	return nil, -1
}

func removeAssignedCodeset(codesets []*domain.AssignedCodeset, index int) []*domain.AssignedCodeset {
	codesets[index] = codesets[len(codesets)-1]
	return codesets[:len(codesets)-1]
}
