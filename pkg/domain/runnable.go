package domain

import (
	"context"

	"github.com/fuseml/fuseml-core/gen/runnable"
	"github.com/google/uuid"
)

// RunnableStore is an inteface to runnable stores
type RunnableStore interface {
	Find(ctx context.Context, id uuid.UUID) *runnable.Runnable
	GetAll(ctx context.Context, kind string) (result []*runnable.Runnable)
	Add(ctx context.Context, r *runnable.Runnable) (*runnable.Runnable, error)
}
