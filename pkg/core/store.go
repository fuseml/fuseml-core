package fuseml

import (
	"time"

	runnable "github.com/fuseml/fuseml-core/gen/runnable"
	"github.com/google/uuid"
)

type Store struct {
	items map[uuid.UUID]*runnable.Runnable
}

var (
	store = Store{items: make(map[uuid.UUID]*runnable.Runnable)}
)

func (s *Store) FindRunnable(id uuid.UUID) *runnable.Runnable {
	return s.items[id]
}

func (s *Store) GetAllRunnables() (result []*runnable.Runnable) {
	result = make([]*runnable.Runnable, 0, len(s.items))
	for _, r := range s.items {
		result = append(result, r)
	}
	return
}

func (s *Store) AddRunnable(r *runnable.Runnable) (*runnable.Runnable, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	runnableId := id.String()
	runnableCreated := time.Now().Format(time.RFC3339)
	r.ID = &runnableId
	r.Created = &runnableCreated
	s.items[id] = r
	return r, nil
}
