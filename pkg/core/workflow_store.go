package core

import (
	"context"
	"time"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

// WorkflowStore describes in memory store for workflows
type WorkflowStore struct {
	items        map[string]*workflow.Workflow
	backend      domain.WorkflowBackend
	codesetStore domain.CodesetStore
}

// NewWorkflowStore returns an in-memory workflow store instance
func NewWorkflowStore(backend domain.WorkflowBackend, cs domain.CodesetStore) *WorkflowStore {
	return &WorkflowStore{
		items:        make(map[string]*workflow.Workflow),
		backend:      backend,
		codesetStore: cs,
	}
}

// Find returns a workflow identified by id
func (ws *WorkflowStore) Find(ctx context.Context, name string) *workflow.Workflow {
	return ws.items[name]
}

// GetAll returns all workflows that matches a given name.
func (ws *WorkflowStore) GetAll(ctx context.Context, name string) (result []*workflow.Workflow) {
	result = make([]*workflow.Workflow, 0)
	for _, w := range ws.items {
		if name == "all" || w.Name == name {
			result = append(result, w)
		}
	}
	return
}

// Add adds a new workflow, based on the Workflow structure provided as argument
func (ws *WorkflowStore) Add(ctx context.Context, w *workflow.Workflow) (*workflow.Workflow, error) {
	err := ws.backend.CreateWorkflow(ctx, w)
	if err != nil {
		return nil, err
	}

	workflowCreated := time.Now().Format(time.RFC3339)
	w.Created = &workflowCreated
	ws.items[w.Name] = w
	return w, nil
}

// AssignCodeset assigns a codeset to a workflow
func (ws *WorkflowStore) AssignCodeset(ctx context.Context, w *workflow.Workflow, c *domain.Codeset) error {

	url, err := ws.backend.CreateListener(ctx, w.Name, true)
	if err != nil {
		return err
	}

	err = ws.codesetStore.CreateWebhook(ctx, c, url)
	if err != nil {
		// FIXME: delete the listener
		return err
	}

	err = ws.backend.CreateWorkflowRun(ctx, w.Name, c)
	if err != nil {
		return err
	}

	return nil
}

// GetAllRuns lists all the runs of a workflow meeting the filter criteria
func (ws *WorkflowStore) GetAllRuns(ctx context.Context, w *workflow.Workflow, filters domain.WorkflowRunFilter) (res []*workflow.WorkflowRun, err error) {

	res, err = ws.backend.ListWorkflowRuns(ctx, w, filters)
	if err != nil {
		return nil, err
	}

	return
}
