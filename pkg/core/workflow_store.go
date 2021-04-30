package core

import (
	"context"
	"time"

	"github.com/fuseml/fuseml-core/gen/workflow"
)

// WorkflowStore describes in memory store for workflows
type WorkflowStore struct {
	items map[string]*workflow.Workflow
}

// NewWorkflowStore returns an in-memory workflow store instance
func NewWorkflowStore() *WorkflowStore {
	return &WorkflowStore{
		items: make(map[string]*workflow.Workflow),
	}
}

// Find returns a workflow identified by id
func (ws *WorkflowStore) Find(ctx context.Context, name string) *workflow.Workflow {
	return ws.items[name]
}

// GetAll returns all workflows that matches a given name.
func (ws *WorkflowStore) GetAll(ctx context.Context, name string) (result []*workflow.Workflow) {
	result = make([]*workflow.Workflow, 0, len(ws.items))
	for _, w := range ws.items {
		if name == "all" || w.Name == name {
			result = append(result, w)
		}
	}
	return
}

// Add adds a new workflow, based on the Workflow structure provided as argument
func (ws *WorkflowStore) Add(ctx context.Context, w *workflow.Workflow) (*workflow.Workflow, error) {
	workflowCreated := time.Now().Format(time.RFC3339)
	w.Created = &workflowCreated
	ws.items[w.Name] = w
	return w, nil
}
