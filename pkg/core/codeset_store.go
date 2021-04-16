package fuseml

import (
	"context"

	"github.com/pkg/errors"

	"github.com/fuseml/fuseml-core/gen/codeset"
	giteaadmin "github.com/fuseml/fuseml-core/pkg/core/gitea"
)

type CodesetStore interface {
	Find(ctx context.Context, project, name string) (*codeset.Codeset, error)
	GetAll(ctx context.Context, project, label *string) ([]*codeset.Codeset, error)
	Add(ctx context.Context, c *codeset.Codeset) (*codeset.Codeset, error)
}

type inMemCodesetStore struct {
	// FIXME this is just internal representation, it should go away
	items map[string]*codeset.Codeset
}

func NewInMemCodesetStore() *inMemCodesetStore {
	return &inMemCodesetStore{
		items: make(map[string]*codeset.Codeset),
	}
}

func (cs *inMemCodesetStore) Find(ctx context.Context, project, name string) (*codeset.Codeset, error) {

	giteaAdmin, err := giteaadmin.NewGiteaAdminClient()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize Gitea admin client")
	}
	gitAdmin := NewGitAdmin(giteaAdmin)

	result, err := gitAdmin.GetRepo(project, name)
	if err != nil {
		return nil, errors.Wrap(err, "Fetching Codeset failed")
	}
	return result, nil
}

// return codeset elements matching given project and label
func (cs *inMemCodesetStore) GetAll(ctx context.Context, project, label *string) ([]*codeset.Codeset, error) {

	giteaAdmin, err := giteaadmin.NewGiteaAdminClient()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize Gitea admin client")
	}
	gitAdmin := NewGitAdmin(giteaAdmin)

	result, err := gitAdmin.GetRepos(project, label)
	if err != nil {
		return nil, errors.Wrap(err, "Fetching Codesets failed")
	}
	return result, nil
}

// 1. create org + new repo
// 2. TODO register in some other store ???
func (cs *inMemCodesetStore) Add(ctx context.Context, c *codeset.Codeset) (*codeset.Codeset, error) {

	giteaAdmin, err := giteaadmin.NewGiteaAdminClient()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize Gitea admin client")
	}
	gitAdmin := NewGitAdmin(giteaAdmin)

	err = gitAdmin.PrepareRepo(c)
	if err != nil {
		return nil, errors.Wrap(err, "Preparing Repository failed")
	}
	// Code itself needs to be pushed from client, here we could do some additional registration
	cs.items[c.Name] = c
	return c, nil
}
