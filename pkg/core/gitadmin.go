/**
 * This is the administration interface to the git store
 * It is currently implemented using gitea API
 */
package fuseml

import (
	"github.com/pkg/errors"

	codeset "github.com/fuseml/fuseml-core/gen/codeset"
	giteaadmin "github.com/fuseml/fuseml-core/pkg/core/gitea"
)

type GitAdmin struct {
	admin *giteaadmin.GiteaAdminClient
}

func NewGitAdmin() (*GitAdmin, error) {
	admin, err := giteaadmin.NewGiteaAdminClient()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize Gitea admin client")
	}
	return &GitAdmin{admin}, nil
}

// Prepare the org, repository, and create user that clients can use for pushing
func (ga *GitAdmin) PrepareRepo(code *codeset.Codeset) error {
	return ga.admin.PrepareRepo(code)
}

// Find all repositories, optionally filtered by project
func (ga *GitAdmin) GetRepos(org *string) ([]*codeset.Codeset, error) {
	return ga.admin.GetRepos(org)
}

// Get the information about repository
func (ga *GitAdmin) GetRepo(org, name string) (*codeset.Codeset, error) {
	return ga.admin.GetRepo(org, name)
}
