package fuseml

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"

	runnable "github.com/fuseml/fuseml-core/gen/runnable"
)

// runnable service example implementation.
// The example methods log the requests and return zero values.
type runnablesrvc struct {
	logger *log.Logger
}

// NewRunnable returns the runnable service implementation.
func NewRunnable(logger *log.Logger) runnable.Service {
	return &runnablesrvc{logger}
}

// Retrieve information about runnables registered in FuseML.
func (s *runnablesrvc) List(ctx context.Context, p *runnable.ListPayload) (res []*runnable.Runnable, err error) {
	s.logger.Print("runnable.list")
	kind := "all"
	if p.Kind != nil {
		kind = *p.Kind
	}

	return runnableStore.GetAllRunnables(kind), nil
}

// Register a runnable with the FuseML runnable runnableStore.
func (s *runnablesrvc) Register(ctx context.Context, p *runnable.Runnable) (res *runnable.Runnable, err error) {
	s.logger.Print("runnable.register")
	return runnableStore.AddRunnable(p)
}

// Retrieve an Runnable from FuseML.
func (s *runnablesrvc) Get(ctx context.Context, p *runnable.GetPayload) (res *runnable.Runnable, err error) {
	s.logger.Print("runnable.get")
	id, err := uuid.Parse(p.RunnableNameOrID)
	if err != nil {
		return nil, runnable.MakeBadRequest(err)
	}
	r := runnableStore.FindRunnable(id)
	if r == nil {
		return nil, runnable.MakeNotFound(errors.New("could not find a runnable with the specified ID"))
	}
	return r, nil
}
