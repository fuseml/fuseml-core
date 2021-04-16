package fuseml

import (
	"time"

	workflow "github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/google/uuid"
)

type WorkflowStore struct {
	items map[uuid.UUID]*workflow.Workflow
}

var (
	workflowStore = WorkflowStore{items: make(map[uuid.UUID]*workflow.Workflow)}
)

func (ws *WorkflowStore) Find(id uuid.UUID) *workflow.Workflow {
	return ws.items[id]
}

func (ws *WorkflowStore) Get(name string) (result []*workflow.Workflow) {
	result = make([]*workflow.Workflow, 0, len(ws.items))
	for _, w := range ws.items {
		if name == "all" || w.Name == name {
			result = append(result, w)
		}
	}
	return
}

func (ws *WorkflowStore) Add(w *workflow.Workflow) (*workflow.Workflow, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	workflowId := id.String()
	workflowCreated := time.Now().Format(time.RFC3339)
	w.ID = &workflowId
	w.Created = &workflowCreated
	ws.items[id] = w
	return w, nil
}
