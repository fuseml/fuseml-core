package runnable

import (
	"github.com/fuseml/fuseml-core/pkg/models"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
)

type Store struct {
	items map[uuid.UUID]*models.Runnable
}

var (
	store = Store{items: make(map[uuid.UUID]*models.Runnable)}
)

func (s *Store) FindRunnable(id uuid.UUID) *models.Runnable {
	return s.items[id]
}

func (s *Store) GetAllRunnables() (result []*models.Runnable) {
	result = make([]*models.Runnable, 0, len(s.items))
	for _, r := range s.items {
		result = append(result, r)
	}
	return
}

func (s *Store) AddRunnable(r *models.Runnable) *models.Runnable {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil
	}
	s.items[id] = r
	r.ID = strfmt.UUID(id.String())
	return r
}
