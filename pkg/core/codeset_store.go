package core

import (
	"context"

	"github.com/pkg/errors"

	"github.com/fuseml/fuseml-core/pkg/domain"
)

// GitCodesetStore describes a structure that accesses codeset store implemented in git
type GitCodesetStore struct {
	gitAdmin domain.GitAdminClient
}

// NewGitCodesetStore returns codeset store instance
func NewGitCodesetStore(gitAdmin domain.GitAdminClient) *GitCodesetStore {
	return &GitCodesetStore{
		gitAdmin: gitAdmin,
	}
}

// Find returns a codeset identified by project and name
func (cs *GitCodesetStore) Find(ctx context.Context, project, name string) (*domain.Codeset, error) {
	result, err := cs.gitAdmin.GetRepository(project, name)
	if err != nil {
		return nil, errors.Wrap(err, "Fetching Codeset failed")
	}
	return result, nil
}

// Delete removes a codeset identified by project and name
func (cs *GitCodesetStore) Delete(ctx context.Context, project, name string) error {
	err := cs.gitAdmin.DeleteRepository(project, name)
	// TODO should we delete the project+user too? If it does not contain any repos?
	if err != nil {
		return errors.Wrap(err, "Deleting Codeset failed")
	}
	return nil
}

// GetAll returns all codesets matching given project and label
func (cs *GitCodesetStore) GetAll(ctx context.Context, project, label *string) ([]*domain.Codeset, error) {
	result, err := cs.gitAdmin.GetRepositories(project, label)
	if err != nil {
		return nil, errors.Wrap(err, "Fetching Codesets failed")
	}
	return result, nil
}

// CreateWebhook adds a new webhook to a codeset
func (cs *GitCodesetStore) CreateWebhook(ctx context.Context, c *domain.Codeset, listenerURL string) (*int64, error) {
	hookID, err := cs.gitAdmin.CreateRepoWebhook(c.Project, c.Name, &listenerURL)
	if err != nil {
		return nil, errors.Wrap(err, "Creating webhook failed")
	}
	return hookID, nil
}

// DeleteWebhook deletes a webhook from a codeset
func (cs *GitCodesetStore) DeleteWebhook(ctx context.Context, c *domain.Codeset, hookID *int64) error {
	err := cs.gitAdmin.DeleteRepoWebhook(c.Project, c.Name, hookID)
	if err != nil {
		return errors.Wrap(err, "Deleting webhook failed")
	}
	return nil
}

// Add creates new codeset
func (cs *GitCodesetStore) Add(ctx context.Context, c *domain.Codeset) (*domain.Codeset, *string, *string, error) {
	username, password, err := cs.gitAdmin.PrepareRepository(c, nil)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Preparing Repository failed")
	}
	// Code itself needs to be pushed from client, here we could do some additional registration
	return c, username, password, nil
}
