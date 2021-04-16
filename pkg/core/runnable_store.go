package fuseml

import (
	"time"

	runnable "github.com/fuseml/fuseml-core/gen/runnable"
	"github.com/google/uuid"
)

type RunnableStore struct {
	items map[uuid.UUID]*runnable.Runnable
}

var (
	runnableStore = RunnableStore{items: make(map[uuid.UUID]*runnable.Runnable)}
)

func (rs *RunnableStore) FindRunnable(id uuid.UUID) *runnable.Runnable {
	return rs.items[id]
}

func (rs *RunnableStore) GetAllRunnables(kind string) (result []*runnable.Runnable) {
	result = make([]*runnable.Runnable, 0, len(rs.items))
	for _, r := range rs.items {
		if kind == "all" || r.Kind == kind {
			result = append(result, r)
		}
	}
	return
}

func (rs *RunnableStore) AddRunnable(r *runnable.Runnable) (*runnable.Runnable, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	runnableId := id.String()
	runnableCreated := time.Now().Format(time.RFC3339)
	r.ID = &runnableId
	r.Created = &runnableCreated
	rs.items[id] = r
	return r, nil
}
