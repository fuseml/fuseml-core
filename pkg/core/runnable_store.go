package core

import (
	"context"
	"time"

	"github.com/fuseml/fuseml-core/pkg/domain"
	"github.com/jinzhu/copier"
)

// RunnableStore describes in memory store for runnables
type runnableStore struct {
	items map[string]*domain.Runnable
}

func NewRunnableStore() *runnableStore {
	return &runnableStore{
		items: make(map[string]*domain.Runnable),
	}
}

// Find returns a list of runnables identified by kind or labels
func (s *runnableStore) Find(ctx context.Context, kind string, labels map[string]string) (res []*domain.Runnable, err error) {
	res = make([]*domain.Runnable, 0, len(s.items))
	for _, r := range s.items {
		if kind == "all" || r.Kind == kind {
			rc := &domain.Runnable{}
			// return a deep copy of the internal runnable
			copier.Copy(&rc, r)
			res = append(res, rc)
		}
		// TODO: match labels
	}
	return
}

// Register adds a new runnable, based on the Runnable structure provided as argument
func (s *runnableStore) Register(ctx context.Context, r *domain.Runnable) (res *domain.Runnable, err error) {
	res = &domain.Runnable{}
	// return a deep copy of the internal runnable
	copier.Copy(&res, r)
	res.Created = time.Now()
	s.items[res.Id] = res
	return res, nil
}

// Get returns a runnable identified by id
func (s *runnableStore) Get(ctx context.Context, id string) (res *domain.Runnable, err error) {
	return s.items[id], nil
}
