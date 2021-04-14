package fuseml

import (
	"github.com/pkg/errors"

	codeset "github.com/fuseml/fuseml-core/gen/codeset"
)

type CodesetStore struct {
	// FIXME this is just internal representation, it should go away
	items map[string]*codeset.Codeset
}

var (
	codesetStore = CodesetStore{items: make(map[string]*codeset.Codeset)}
)

// FIXME check git content, not internal map
func (cs *CodesetStore) FindCodeset(project, name string) *codeset.Codeset {
	return cs.items[name]
}

// FIXME for all projects and labels, return codeset element
func (cs *CodesetStore) GetAllCodesets(project *string, label *string) ([]*codeset.Codeset, error) {
	// FIXME move this to init
	csc, err := NewCodesetClient()
	if err != nil {
		return nil, errors.Wrap(err, "Creating codeset client failed")
	}

	result, err := csc.GetRepos(project)
	if err != nil {
		return nil, errors.Wrap(err, "Fetching Codesets failed")
	}
        // FIXME check for label too
	return result, nil
}

// 1. create org + new repo
// 2. register in some other store ???
func (cs *CodesetStore) AddCodeset(c *codeset.Codeset) (*codeset.Codeset, error) {
	csc, err := NewCodesetClient()
	if err != nil {
		return nil, errors.Wrap(err, "Creating codeset client failed")
	}
	err = csc.PrepareRepo(c)
	if err != nil {
		return nil, errors.Wrap(err, "Preparing Repository failed")
	}
	// Code itself needs to be pushed from client, here we could do some additional registration
	cs.items[c.Name] = c
	return c, nil
}
