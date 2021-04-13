/**
 * This is a simplified version of fuseml/paas/gitea/resolver.
 * This one does not access Kubernetes cluster, but ruther fills the values from env variables.
 */
package gitea

import (
	"github.com/pkg/errors"
	"os"
)

type Resolver struct {
	url      string
	username string
	password string
}

func NewGiteaResolver() (*Resolver, error) {
	url, exists := os.LookupEnv("GITEA_URL")
	if !exists {
		return nil, errors.New("Value for gitea URL (GITEA_URL) was not provided.")
	}
	username, exists := os.LookupEnv("GITEA_USERNAME")
	if !exists {
		return nil, errors.New("Value for gitea user name (GITEA_USERNAME) was not provided.")
	}
	password, exists := os.LookupEnv("GITEA_PASSWORD")
	if !exists {
		return nil, errors.New("Value for gitea user password (GITEA_PASSWORD) was not provided.")
	}
	return &Resolver{
		url:      url,
		username: username,
		password: password,
	}, nil
}

func (r *Resolver) GetGiteaURL() (string, error) {
	return r.url, nil
}

func (r *Resolver) GetGiteaCredentials() (string, string, error) {
	return r.username, r.password, nil
}
