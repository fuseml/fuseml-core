package fuseml

import (
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
func (cs *CodesetStore) GetAllCodesets(project string, label *string) (result []*codeset.Codeset) {
	result = make([]*codeset.Codeset, 0, len(cs.items))
	for _, r := range cs.items {
		if project != "all" && r.Project != project {
			continue
		}
		if label != nil && *r.Label != *label {
			continue
		}
		result = append(result, r)
	}
	return
}

// 1. push into new repo
// 2. register in some other store ???
func (cs *CodesetStore) AddCodeset(c *codeset.Codeset) (*codeset.Codeset, error) {
	// Code itself was pushed from client, here we could do some additional registration
	cs.items[c.Name] = c
	return c, nil
}
