package core

import (
	"context"

	"github.com/pkg/errors"

	"github.com/fuseml/fuseml-core/pkg/domain"
)

// GitProjectStore describes a stucture that accesses project store implemented in git
type GitProjectStore struct {
	gitAdmin GitAdmin
}

// NewGitProjectStore returns project store instance
func NewGitProjectStore(gitAdmin GitAdmin) *GitProjectStore {
	return &GitProjectStore{
		gitAdmin: gitAdmin,
	}
}

// Find returns a project identified by project and name
func (cs *GitProjectStore) Find(ctx context.Context, project string) (*domain.Project, error) {
	result, err := cs.gitAdmin.GetProject(project)
	if err != nil {
		return nil, errors.Wrap(err, "Fetching Project failed")
	}
	return result, nil
}

// GetAll returns all projects matching given project and label
func (cs *GitProjectStore) GetAll(ctx context.Context) ([]*domain.Project, error) {
	result, err := cs.gitAdmin.GetProjects()
	if err != nil {
		return nil, errors.Wrap(err, "Fetching Projects failed")
	}
	return result, nil
}

// Delete removes a project identified by project and name
func (cs *GitProjectStore) Delete(ctx context.Context, project string) error {
	err := cs.gitAdmin.DeleteProject(project)
	if err != nil {
		return errors.Wrap(err, "Deleting Project failed")
	}
	return nil
}
