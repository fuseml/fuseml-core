package badger

import (
	"context"

	"github.com/fuseml/fuseml-core/pkg/domain"
	"github.com/timshannon/badgerhold/v3"
)

// storableWorkflow holds the workflow and the codesets assigned to it
type storableWorkflow struct {
	// Workflow assigned to the codeset
	Workflow *domain.Workflow
	// AssignedCodeset holds codesets assigned to the workflow and its hookID
	AssignedCodesets []*domain.CodesetAssignment
}

type WorkflowStore struct {
	store *badgerhold.Store
}

func NewWorkflowStore(store *badgerhold.Store) *WorkflowStore {
	return &WorkflowStore{store: store}
}

// GetWorkflow returns a workflow identified by its name
func (ws *WorkflowStore) GetWorkflow(ctx context.Context, name string) (*domain.Workflow, error) {
	sw := storableWorkflow{}
	err := ws.store.Get(name, &sw)
	if err != nil {
		return nil, domain.ErrWorkflowNotFound
	}
	return sw.Workflow, nil
}

// GetWorkflows returns all workflows or the one that matches a given name.
func (ws *WorkflowStore) GetWorkflows(ctx context.Context, name *string) []*domain.Workflow {
	result := []*domain.Workflow{}
	if name != nil {
		sw := storableWorkflow{}
		err := ws.store.Get(name, &sw)
		if err == nil {
			result = append(result, sw.Workflow)
		}
		return result
	}

	sw := []storableWorkflow{}
	err := ws.store.Find(&sw, nil)
	if err != nil {
		return result
	}

	for _, swf := range sw {
		result = append(result, swf.Workflow)
	}
	return result
}

// AddWorkflow adds a new workflow based on the Workflow structure provided as argument
func (ws *WorkflowStore) AddWorkflow(ctx context.Context, w *domain.Workflow) (*domain.Workflow, error) {
	err := ws.store.Insert(w.Name, storableWorkflow{Workflow: w})
	if err != nil {
		return nil, domain.ErrWorkflowExists
	}
	return w, nil
}

// DeleteWorkflow deletes the workflow from the store
func (ws *WorkflowStore) DeleteWorkflow(ctx context.Context, name string) error {
	sw := storableWorkflow{}
	err := ws.store.Get(name, &sw)
	if err != nil {
		return nil
	}
	if sw.AssignedCodesets != nil {
		return domain.ErrCannotDeleteAssignedWorkflow
	}
	ws.store.Delete(name, sw)
	return nil
}

// AddCodesetAssignment adds a codeset to the list of assigned codesets of a workflow if it does not already exists
func (ws *WorkflowStore) AddCodesetAssignment(ctx context.Context, workflowName string,
	assignment *domain.CodesetAssignment) []*domain.CodesetAssignment {
	sw := storableWorkflow{}
	err := ws.store.Get(workflowName, &sw)
	if err != nil {
		return nil
	}
	if sw.AssignedCodesets != nil {
		if assigned, _ := getAssignedCodeset(sw.AssignedCodesets, assignment.Codeset); assigned != nil {
			return sw.AssignedCodesets
		}
		sw.AssignedCodesets = append(sw.AssignedCodesets, assignment)
	} else {
		assignedCodesets := []*domain.CodesetAssignment{assignment}
		sw.AssignedCodesets = assignedCodesets
	}
	err = ws.store.Update(workflowName, &sw)
	if err != nil {
		return nil
	}
	return sw.AssignedCodesets
}

// DeleteCodesetAssignment deletes a codeset from the list of assigned codesets of a workflow if it exists
func (ws *WorkflowStore) DeleteCodesetAssignment(ctx context.Context, workflowName string, codeset *domain.Codeset) []*domain.CodesetAssignment {
	sw := storableWorkflow{}
	err := ws.store.Get(workflowName, &sw)
	if err != nil || sw.AssignedCodesets == nil {
		return []*domain.CodesetAssignment{}
	}

	if _, i := getAssignedCodeset(sw.AssignedCodesets, codeset); i != -1 {
		sw.AssignedCodesets = removeAssignedCodeset(sw.AssignedCodesets, i)
		ws.store.Update(workflowName, &sw)
	}
	return sw.AssignedCodesets
}

// GetAssignedCodesets returns a list of codesets assigned to the specified workflow
func (ws *WorkflowStore) GetAssignedCodesets(ctx context.Context, workflowName string) []*domain.CodesetAssignment {
	sw := storableWorkflow{}
	err := ws.store.Get(workflowName, &sw)
	if err != nil || sw.AssignedCodesets == nil {
		return []*domain.CodesetAssignment{}
	}
	return sw.AssignedCodesets
}

// GetAssignments returns a map of workflows and its assigned codesets
func (ws *WorkflowStore) GetAssignments(ctx context.Context, workflowName *string) (result map[string][]*domain.CodesetAssignment) {
	result = make(map[string][]*domain.CodesetAssignment)
	if workflowName != nil {
		sw := storableWorkflow{}
		err := ws.store.Get(*workflowName, &sw)
		if err != nil || sw.AssignedCodesets == nil {
			return
		}
		result[*workflowName] = sw.AssignedCodesets
		return
	}

	sws := []*storableWorkflow{}
	ws.store.Find(&sws, nil)
	for _, sw := range sws {
		if sw.AssignedCodesets != nil {
			result[sw.Workflow.Name] = sw.AssignedCodesets
		}
	}
	return
}

// GetAssignedCodeset returns a AssignedCodeset for the Workflow and Codeset
func (ws *WorkflowStore) GetAssignedCodeset(ctx context.Context, workflowName string, codeset *domain.Codeset) (*domain.CodesetAssignment, error) {
	sw := storableWorkflow{}
	err := ws.store.Get(workflowName, &sw)
	if err != nil {
		return nil, domain.ErrWorkflowNotFound
	}

	if sw.AssignedCodesets != nil {
		ac, _ := getAssignedCodeset(sw.AssignedCodesets, codeset)
		if ac != nil {
			return ac, nil
		}
	}

	return nil, domain.ErrWorkflowNotAssignedToCodeset
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
