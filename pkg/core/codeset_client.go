package fuseml

import (
	"log"

	"code.gitea.io/sdk/gitea"
	"github.com/pkg/errors"

	codeset "github.com/fuseml/fuseml-core/gen/codeset"
	config "github.com/fuseml/fuseml-core/pkg/core/config"
	giteac "github.com/fuseml/fuseml-core/pkg/core/gitea"
)

var ()

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
		return errors.Wrap(err, "Failed to create org")
	}
	return nil
}

// create user assigned to current project
func (cc *CodesetClient) CreateUser(org string) error {
	username := config.DefaultUserName(org)
	user, resp, err := cc.giteaClient.GetUserInfo(username)
	if resp == nil && err != nil {
		return errors.Wrap(err, "Failed to make get user request")
	}
	if user.ID != 0 {
		log.Println("User already exists")
		return nil
	}

	log.Printf("Creating user '%s'", username)
	_, _, err = cc.giteaClient.AdminCreateUser(gitea.CreateUserOption{
		Username:           username,
		Email:              config.DefaultUserEmail,
		Password:           config.DefaultUserPassword,
		MustChangePassword: gitea.OptionalBool(false),
		SendNotify:         false,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to create user")
	}

	teams, _, err := cc.giteaClient.ListOrgTeams(org, gitea.ListTeamsOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to list org teams")
	}
	for _, team := range teams {
		if team.Name == "Owners" {
			_, err = cc.giteaClient.AddTeamMember(team.ID, username)
			if err != nil {
				return errors.Wrap(err, "Failed adding user to Owners")
			}
			break
		}
	}

	if err != nil {
		return errors.Wrap(err, "Failed to create application")
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
		return errors.Wrap(err, "Failed to create repository")
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
		if url == config.StagingEventListenerURL {
			log.Printf("Webhook for '%s' already exists", name)
			return nil
		}
	}

	log.Printf("Creating Webhook for '%s' under '%s'...", name, org)
	cc.giteaClient.CreateRepoHook(org, name, gitea.CreateHookOption{
		Active:       true,
		BranchFilter: "*",
		Config: map[string]string{
			"secret":       config.HookSecret,
			"http_method":  "POST",
			"url":          config.StagingEventListenerURL,
			"content_type": "json",
		},
		Type: "gitea",
	})

	return nil
}

// Prepare the org, repository, and create user that clients can use for pushing
func (cc *CodesetClient) PrepareRepo(code *codeset.Codeset) error {

	err := cc.CreateOrg(code.Project)
	if err != nil {
		return errors.Wrap(err, "Create org failed")
	}

	err = cc.CreateUser(code.Project)
	if err != nil {
		return errors.Wrap(err, "Create FuseML user failed")
	}

	err = cc.CreateRepo(code.Project, code.Name)
	if err != nil {
		return errors.Wrap(err, "Create repo failed")
	}

	err = cc.CreateRepoWebhook(code.Project, code.Name)
	if err != nil {
		return errors.Wrap(err, "Creating webhook failed")
	}
	return nil
}
