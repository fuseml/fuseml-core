/**
 * This is the administration interface to the git store
 */
package fuseml

import (
	codeset "github.com/fuseml/fuseml-core/gen/codeset"
)

type GitAdmin struct {
	admin AdminClient
}

type AdminClient interface {
	PrepareRepo(code *codeset.Codeset) error
	GetRepos(org, label *string) ([]*codeset.Codeset, error)
	GetRepo(org, name string) (*codeset.Codeset, error)
}

func NewGitAdmin(ac AdminClient) *GitAdmin {
	return &GitAdmin{ac}
}

// Prepare the org, repository, and create user that clients can use for pushing
func (ga *GitAdmin) PrepareRepo(code *codeset.Codeset) error {
	return ga.admin.PrepareRepo(code)
}

// Find all repositories, optionally filtered by project
func (ga *GitAdmin) GetRepos(org, label *string) ([]*codeset.Codeset, error) {
	return ga.admin.GetRepos(org, label)
}

// Get the information about repository
func (ga *GitAdmin) GetRepo(org, name string) (*codeset.Codeset, error) {
	return ga.admin.GetRepo(org, name)
}
