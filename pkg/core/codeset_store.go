package fuseml

import (
	//	"time"

	codeset "github.com/fuseml/fuseml-core/gen/codeset"
	//	"github.com/google/uuid"
)

type CodesetStore struct {
	items map[string]*codeset.Codeset
}

var (
	codesetStore = CodesetStore{items: make(map[string]*codeset.Codeset)}
)

func (cs *CodesetStore) FindCodeset(project, name string) *codeset.Codeset {
	// FIXME use project
	return cs.items[name]
}

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

func (cs *CodesetStore) AddCodeset(r *codeset.Codeset) (*codeset.Codeset, error) {
	/*
		codesetCreated := time.Now().Format(time.RFC3339)
		r.Created = &codesetCreated
	*/
	cs.items[r.Name] = r
	return r, nil
}
