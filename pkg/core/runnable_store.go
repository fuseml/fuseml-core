package fuseml

import (
	"time"

	runnable "github.com/fuseml/fuseml-core/gen/runnable"
	"github.com/google/uuid"
)

// RunnableStore describes in memory store for runnables
type RunnableStore struct {
	items map[uuid.UUID]*runnable.Runnable
}

var (
	runnableStore = RunnableStore{items: make(map[uuid.UUID]*runnable.Runnable)}
)

// FindRunnable returns a runnable identified by id
func (rs *RunnableStore) FindRunnable(id uuid.UUID) *runnable.Runnable {
	return rs.items[id]
}

// GetAllRunnables returns all runnables of a given type.
// Type can be "all" for returning runnables of all types.
func (rs *RunnableStore) GetAllRunnables(kind string) (result []*runnable.Runnable) {
	result = make([]*runnable.Runnable, 0, len(rs.items))
	for _, r := range rs.items {
		if kind == "all" || r.Kind == kind {
			result = append(result, r)
		}
	}
	return
}

// AddRunnable adds a new runnable, based on the Runnable structure provided as argument
func (rs *RunnableStore) AddRunnable(r *runnable.Runnable) (*runnable.Runnable, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	runnableID := id.String()
	runnableCreated := time.Now().Format(time.RFC3339)
	r.ID = &runnableID
	r.Created = &runnableCreated
	rs.items[id] = r
	return r, nil
}
