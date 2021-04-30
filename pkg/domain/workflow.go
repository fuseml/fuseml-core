package domain

import (
	"context"

	"github.com/fuseml/fuseml-core/gen/workflow"
)

// WorkflowStore is an inteface to workflow stores
type WorkflowStore interface {
	Find(ctx context.Context, name string) *workflow.Workflow
	GetAll(ctx context.Context, name string) (result []*workflow.Workflow)
	Add(ctx context.Context, r *workflow.Workflow) (*workflow.Workflow, error)
}
