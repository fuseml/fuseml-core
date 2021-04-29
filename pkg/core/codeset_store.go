package core

import (
	"context"

	"github.com/pkg/errors"

	"github.com/fuseml/fuseml-core/gen/codeset"

	config "github.com/fuseml/fuseml-core/pkg/core/config"
)

// GitAdmin is an inteface to git administration clients
type GitAdmin interface {
	PrepareRepository(*codeset.Codeset, *string) error
	CreateRepoWebhook(string, string, *string) error
	GetRepositories(org, label *string) ([]*codeset.Codeset, error)
	GetRepository(org, name string) (*codeset.Codeset, error)
}

// GitCodesetStore describes a stucture that accesses codeset store implemented in git
type GitCodesetStore struct {
	gitAdmin GitAdmin
}

// NewGitCodesetStore returns codeset store instance
func NewGitCodesetStore(gitAdmin GitAdmin) *GitCodesetStore {
	return &GitCodesetStore{
		gitAdmin: gitAdmin,
	}
}

// Find returns a codeset identified by project and name
func (cs *GitCodesetStore) Find(ctx context.Context, project, name string) (*codeset.Codeset, error) {
	result, err := cs.gitAdmin.GetRepository(project, name)
	if err != nil {
		return nil, errors.Wrap(err, "Fetching Codeset failed")
	}
	return result, nil
}

// GetAll returns all codesets matching given project and label
func (cs *GitCodesetStore) GetAll(ctx context.Context, project, label *string) ([]*codeset.Codeset, error) {
	result, err := cs.gitAdmin.GetRepositories(project, label)
	if err != nil {
		return nil, errors.Wrap(err, "Fetching Codesets failed")
	}
	return result, nil
}

// CreateWebhook adds a new webhook to a codeset
func (cs *GitCodesetStore) CreateWebhook(ctx context.Context, c *codeset.Codeset, listenerURL string) error {
	err := cs.gitAdmin.CreateRepoWebhook(c.Project, c.Name, &listenerURL)
	if err != nil {
		return errors.Wrap(err, "Creating webhook failed")
	}
	return nil
}

// Add creates new codeset
func (cs *GitCodesetStore) Add(ctx context.Context, c *codeset.Codeset) (*codeset.Codeset, error) {
	var listenerURL *string
	listenerURL = &config.StagingEventListenerURL
	err := cs.gitAdmin.PrepareRepository(c, listenerURL)
	if err != nil {
		return nil, errors.Wrap(err, "Preparing Repository failed")
	}
	// Code itself needs to be pushed from client, here we could do some additional registration
	return c, nil
}
