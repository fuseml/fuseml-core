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
	// Codesets assigned to the workflow
	assignedCodesets []*domain.Codeset
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
func (ws *WorkflowStore) GetWorkflow(ctx context.Context, name string) *workflow.Workflow {
	if _, exists := ws.items[name]; exists {
		return ws.items[name].workflow
	}
	return nil
}

// GetAllWorkflows returns all workflows or the one that matches a given name.
func (ws *WorkflowStore) GetAllWorkflows(ctx context.Context, name *string) (result []*workflow.Workflow) {
	result = make([]*workflow.Workflow, 0, len(ws.items))
	if name != nil {
		result = append(result, ws.items[*name].workflow)
		return
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

// GetAssignedCodesets returns a list of codesets assigned to the specified workflow
func (ws *WorkflowStore) GetAssignedCodesets(ctx context.Context, workflowName string) []*domain.Codeset {
	if _, exists := ws.items[workflowName]; exists {
		return ws.items[workflowName].assignedCodesets
	}
	return nil
}

// GetAssignments returns a map of workflows and its assigned codesets
func (ws *WorkflowStore) GetAssignments(ctx context.Context, workflowName *string) (result map[string][]*domain.Codeset) {
	result = make(map[string][]*domain.Codeset, len(ws.items))
	if workflowName != nil {
		if _, exists := ws.items[*workflowName]; exists {
			result[*workflowName] = ws.items[*workflowName].assignedCodesets
		}
		return
	}
	for _, sw := range ws.items {
		result[sw.workflow.Name] = sw.assignedCodesets
	}
	return
}

// AddCodesetAssignment adds a codeset to the list of assigned codesets of a workflow if it does not already exists
func (ws *WorkflowStore) AddCodesetAssignment(ctx context.Context, workflowName string, codeset *domain.Codeset) []*domain.Codeset {
	assignedCodesets := ws.items[workflowName].assignedCodesets
	if !containsCodeset(assignedCodesets, codeset) {
		assignedCodesets = append(assignedCodesets, codeset)
		ws.items[workflowName].assignedCodesets = assignedCodesets
	}
	return assignedCodesets
}

func containsCodeset(slice []*domain.Codeset, codeset *domain.Codeset) bool {
	for _, c := range slice {
		if c.Project == codeset.Project && c.Name == codeset.Name {
			return true
		}
	}
	return false
}
