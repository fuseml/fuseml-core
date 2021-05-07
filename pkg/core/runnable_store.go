package core

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/fuseml/fuseml-core/pkg/domain"
	"github.com/jinzhu/copier"
)

// RunnableStore describes in memory store for runnables
type RunnableStore struct {
	items map[string]*domain.Runnable
}

// NewRunnableStore creates and returns an in-memory runnable store instance
func NewRunnableStore() *RunnableStore {
	return &RunnableStore{
		items: make(map[string]*domain.Runnable),
	}
}

const (
	errRunnableExists = "a runnable with that ID already exists"
)

// Find returns a list of runnables matching the input query.
// Runnables may be matched by id, kind or labels. Only runnables that match all the
// supplied criteria will be returned.
func (s *RunnableStore) Find(ctx context.Context, id string, kind string, labels map[string]string) (res []*domain.Runnable, err error) {
	res = make([]*domain.Runnable, 0)

RUNNABLES:
	for _, r := range s.items {
		// match ID query
		if id != "" && r.ID != id {
			// try matching as a regexp
			if match, _ := regexp.Match(id, []byte(r.ID)); !match {
				continue RUNNABLES
			}
		}
		// match kind query
		if kind != "" || r.Kind != kind {
			// try matching as a regexp
			if match, _ := regexp.Match(kind, []byte(r.Kind)); !match {
				continue RUNNABLES
			}
		}
		// match label query
		for qLabelKey, qLabelValue := range labels {
			rLabelValue, hasLabel := r.Labels[qLabelKey]
			if !hasLabel {
				// runnable label key doesn't match query
				continue RUNNABLES
			}
			if qLabelValue == "" || qLabelValue == rLabelValue {
				// empty query label value or exact match
				continue
			}
			// try matching as a regexp
			if match, _ := regexp.Match(qLabelValue, []byte(rLabelValue)); !match {
				// runnable label value doesn't match query label regexp
				continue RUNNABLES
			}
		}

		rMatch := &domain.Runnable{}
		// return a deep copy of the internal runnable
		copier.Copy(&rMatch, r)
		res = append(res, rMatch)
	}
	return
}

// Register adds a new runnable, based on the Runnable structure provided as argument
func (s *RunnableStore) Register(ctx context.Context, r *domain.Runnable) (res *domain.Runnable, err error) {
	if _, found := s.items[r.ID]; found {
		return nil, errors.New(errRunnableExists)
	}
	res = &domain.Runnable{}
	// return a deep copy of the internal runnable
	copier.Copy(&res, r)
	res.Created = time.Now()
	s.items[res.ID] = res
	return res, nil
}

// Get returns a runnable identified by id
func (s *RunnableStore) Get(ctx context.Context, id string) (res *domain.Runnable, err error) {
	return s.items[id], nil
}
