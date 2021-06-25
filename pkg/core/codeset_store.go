package core

import (
	"context"

	"github.com/pkg/errors"

	"github.com/fuseml/fuseml-core/pkg/domain"
)

type codesetID struct {
	name    string
	project string
}

// GitCodesetStore describes a structure that accesses codeset store implemented in git
type GitCodesetStore struct {
	gitAdmin    domain.GitAdminClient
	subscribers map[codesetID][]domain.CodesetSubscriber
}

// NewGitCodesetStore returns codeset store instance
func NewGitCodesetStore(gitAdmin domain.GitAdminClient) *GitCodesetStore {
	subscribers := make(map[codesetID][]domain.CodesetSubscriber)
	return &GitCodesetStore{gitAdmin, subscribers}
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
	codeset, err := cs.Find(ctx, project, name)
	if err != nil {
		return nil
	}
	// notify codeset subscribers about a codeset being deleted
	for _, subscriber := range cs.subscribers[codesetID{name, project}] {
		subscriber.OnDeletingCodeset(ctx, codeset)
	}
	err = cs.gitAdmin.DeleteRepository(project, name)
	// TODO should we delete the project+user too? If it does not contain any repos?
	if err != nil {
		return errors.Wrap(err, "Deleting Codeset failed")
	}
	// upon a codeset deletion all subscribers associated to that codeset also needs to be removed
	cs.deleteSubscribers(codeset)
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

// Subscribe adds a subscriber interested on operations performed on a specific codeset
func (cs *GitCodesetStore) Subscribe(ctx context.Context, subscriber domain.CodesetSubscriber, codeset *domain.Codeset) error {
	if _, err := cs.Find(ctx, codeset.Project, codeset.Name); err != nil {
		return err
	}
	cs.subscribers[codesetID{codeset.Name, codeset.Project}] = append(cs.subscribers[codesetID{codeset.Name, codeset.Project}], subscriber)
	return nil
}

// Unsubscribe deletes a specific codeset subscriber
func (cs *GitCodesetStore) Unsubscribe(ctx context.Context, subscriber domain.CodesetSubscriber, codeset *domain.Codeset) error {
	cs.subscribers[codesetID{codeset.Name, codeset.Project}] = removeSubscriber(cs.subscribers[codesetID{codeset.Name, codeset.Project}], subscriber)
	return nil
}

func (cs *GitCodesetStore) deleteSubscribers(codeset *domain.Codeset) {
	delete(cs.subscribers, codesetID{codeset.Name, codeset.Project})
}

func removeSubscriber(subscribers []domain.CodesetSubscriber, subscriber domain.CodesetSubscriber) []domain.CodesetSubscriber {
	for i, s := range subscribers {
		if s == subscriber {
			subscribers[len(subscribers)-1], subscribers[i] = subscribers[i], subscribers[len(subscribers)-1]
			return subscribers[:len(subscribers)-1]
		}
	}
	return subscribers
}
