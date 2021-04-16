package fuseml

import (
	"context"

	"github.com/pkg/errors"

	"github.com/fuseml/fuseml-core/gen/codeset"
)

type CodesetStore interface {
	Find(ctx context.Context, project, name string) (*codeset.Codeset, error)
	GetAll(ctx context.Context, project, label *string) ([]*codeset.Codeset, error)
	Add(ctx context.Context, c *codeset.Codeset) (*codeset.Codeset, error)
}

type GitAdmin interface {
	PrepareRepository(code *codeset.Codeset) error
	GetRepositories(org, label *string) ([]*codeset.Codeset, error)
	GetRepository(org, name string) (*codeset.Codeset, error)
}

type gitCodesetStore struct {
	gitAdmin GitAdmin
}

func NewGitCodesetStore(gitAdmin GitAdmin) *gitCodesetStore {
	return &gitCodesetStore{
		gitAdmin: gitAdmin,
	}
}

func (cs *gitCodesetStore) Find(ctx context.Context, project, name string) (*codeset.Codeset, error) {
	result, err := cs.gitAdmin.GetRepository(project, name)
	if err != nil {
		return nil, errors.Wrap(err, "Fetching Codeset failed")
	}
	return result, nil
}

// return codeset elements matching given project and label
func (cs *gitCodesetStore) GetAll(ctx context.Context, project, label *string) ([]*codeset.Codeset, error) {
	result, err := cs.gitAdmin.GetRepositories(project, label)
	if err != nil {
		return nil, errors.Wrap(err, "Fetching Codesets failed")
	}
	return result, nil
}

// 1. create org + new repo
// 2. TODO register in some other store ???
func (cs *gitCodesetStore) Add(ctx context.Context, c *codeset.Codeset) (*codeset.Codeset, error) {
	err := cs.gitAdmin.PrepareRepository(c)
	if err != nil {
		return nil, errors.Wrap(err, "Preparing Repository failed")
	}
	// Code itself needs to be pushed from client, here we could do some additional registration
	return c, nil
}
