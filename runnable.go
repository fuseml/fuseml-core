package fuseml

import (
	"context"
	"log"

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
	return
}

// Register a runnable with the FuseML runnable store.
func (s *runnablesrvc) Register(ctx context.Context, p *runnable.Runnable) (res *runnable.Runnable, err error) {
	res = &runnable.Runnable{}
	s.logger.Print("runnable.register")
	return
}

// Retrieve an Runnable from FuseML.
func (s *runnablesrvc) Get(ctx context.Context, p *runnable.GetPayload) (res *runnable.Runnable, err error) {
	res = &runnable.Runnable{}
	s.logger.Print("runnable.get")
	return
}
