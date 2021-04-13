package fuseml

import (
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"path"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/pkg/errors"

	codeset "github.com/fuseml/fuseml-core/gen/codeset"
	giteac "github.com/fuseml/fuseml-core/pkg/core/gitea"
)

var (
	// FIXME: generate this and put it in a secret
	HookSecret = "generatedsecret"

	// StagingEventListenerURL should not exist
	// FIXME: detect this based on namespaces and services
	StagingEventListenerURL = "http://el-mlflow-listener.fuseml-workloads:8080"

	DefaultOrg = "workspace"
)

// CodesetClient provides functionality for talking to a
// Fuseml installation on Kubernetes
type CodesetClient struct {
	giteaClient   *gitea.Client
	giteaResolver *giteac.Resolver
}

func NewCodesetClient() (*CodesetClient, error) {
	cs := CodesetClient{}

	giteaResolver, err := giteac.NewGiteaResolver()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize Gitea client")
	}
	giteaClient, err := giteac.NewGiteaClient(giteaResolver)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize Gitea client")
	}
	cs.giteaResolver = giteaResolver
	cs.giteaClient = giteaClient
	return &cs, nil
}

// CreateOrg creates an Org in gitea
func (cc *CodesetClient) CreateOrg(org string) error {

	_, resp, err := cc.giteaClient.GetOrg(org)
	if resp == nil && err != nil {
		return errors.Wrap(err, "Failed to make get org request")
	}

	if resp.StatusCode == 200 {
		log.Printf("Organization already exists.")
		return nil
	}

	_, _, err = cc.giteaClient.CreateOrg(gitea.CreateOrgOption{
		Name: org,
	})

	if err != nil {
		return errors.Wrap(err, "failed to create org")
	}
	return nil
}

// create git repository with given name under given org
func (cc *CodesetClient) CreateRepo(org, name string) error {
	_, resp, err := cc.giteaClient.GetRepo(org, name)
	if resp == nil && err != nil {
		return errors.Wrap(err, "Failed to make get repo request")
	}

	if resp.StatusCode == 200 {
		log.Printf("Application '%s' already exists under '%s'", name, org)
		return nil
	}

	log.Printf("Creating repo '%s' under '%s'...", name, org)
	_, _, err = cc.giteaClient.CreateOrgRepo(org, gitea.CreateRepoOption{
		Name:          name,
		AutoInit:      true,
		Private:       true,
		DefaultBranch: "main",
	})

	if err != nil {
		return errors.Wrap(err, "Failed to create application")
	}

	return nil
}

// Create webhook for given repository and wire it to tekton listener
func (cc *CodesetClient) CreateRepoWebhook(org, name string) error {
	hooks, _, err := cc.giteaClient.ListRepoHooks(org, name, gitea.ListHooksOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to list webhooks")
	}

	for _, hook := range hooks {
		url := hook.Config["url"]
		if url == StagingEventListenerURL {
			log.Printf("Webhook for '%s' already exists", name)
			return nil
		}
	}

	log.Printf("Creating Webhook for '%s' under '%s'...", name, org)
	cc.giteaClient.CreateRepoHook(org, name, gitea.CreateHookOption{
		Active:       true,
		BranchFilter: "*",
		Config: map[string]string{
			"secret":       HookSecret,
			"http_method":  "POST",
			"url":          StagingEventListenerURL,
			"content_type": "json",
		},
		Type: "gitea",
	})

	return nil
}

// Push the code from local dir to remote repo
func (cc *CodesetClient) GitPush(org, name, location string) error {
	log.Printf("Pushing the code to the git repository...")

	giteaURL, err := cc.giteaResolver.GetGiteaURL()
	if err != nil {
		return errors.Wrap(err, "Failed to resolve gitea host")
	}

	u, err := url.Parse(giteaURL)
	if err != nil {
		return errors.Wrap(err, "Failed to parse gitea url")
	}

	username, password, err := cc.giteaResolver.GetGiteaCredentials()
	if err != nil {
		return errors.Wrap(err, "Failed to resolve gitea credentials")
	}

	u.User = url.UserPassword(username, password)
	u.Path = path.Join(u.Path, org, name)

	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf(`
cd "%s"
git init
git config user.name "Fuseml"
git config user.email ci@fuseml
git remote add fuseml "%s"
git fetch --all
git reset --soft fuseml/main
git add --all
git commit --no-gpg-sign -m "pushed at %s"
git push fuseml master:main
`, location, u.String(), time.Now().Format("20060102150405")))

	_, err = cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "Pushiong the code has failed")
	}
	return nil
}

// Prepare the repository and push the code from local path to remote repo
func (cc *CodesetClient) Push(code *codeset.Codeset, location string) error {

	err := cc.CreateOrg(code.Project)
	if err != nil {
		return errors.Wrap(err, "Create org failed")
	}

	err = cc.CreateRepo(code.Project, code.Name)
	if err != nil {
		return errors.Wrap(err, "Create repo failed")
	}

	err = cc.CreateRepoWebhook(code.Project, code.Name)
	if err != nil {
		return errors.Wrap(err, "webhook configuration failed")
	}

	err = cc.GitPush(code.Project, code.Name, location)
	if err != nil {
		return errors.Wrap(err, "failed to git push code")
	}

	return nil
}
