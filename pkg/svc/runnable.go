package svc

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"

	"github.com/fuseml/fuseml-core/gen/runnable"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

// runnable service example implementation.
// The example methods log the requests and return zero values.
type runnablesrvc struct {
	logger *log.Logger
	store  domain.RunnableStore
}

// NewRunnableService returns the runnable service implementation.
func NewRunnableService(logger *log.Logger, store domain.RunnableStore) runnable.Service {
	return &runnablesrvc{logger, store}
}

// Retrieve information about runnables registered in FuseML.
func (s *runnablesrvc) List(ctx context.Context, p *runnable.ListPayload) (res []*runnable.Runnable, err error) {
	s.logger.Print("runnable.list")
	kind := "all"
	if p.Kind != nil {
		kind = *p.Kind
	}

	return s.store.GetAll(ctx, kind), nil
}

// Register a runnable with the FuseML runnable runnableStore.
func (s *runnablesrvc) Register(ctx context.Context, p *runnable.Runnable) (res *runnable.Runnable, err error) {
	s.logger.Print("runnable.register")
	return s.store.Add(ctx, p)
}

// Retrieve a Runnable from FuseML.
func (s *runnablesrvc) Get(ctx context.Context, p *runnable.GetPayload) (res *runnable.Runnable, err error) {
	s.logger.Print("runnable.get")
	id, err := uuid.Parse(p.RunnableNameOrID)
	if err != nil {
		return nil, runnable.MakeBadRequest(err)
	}
	r := s.store.Find(ctx, id)
	if r == nil {
		return nil, runnable.MakeNotFound(errors.New("could not find a runnable with the specified ID"))
	}
	return r, nil
}
