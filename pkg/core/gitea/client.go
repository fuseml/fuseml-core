package gitea

import (
	"log"
	"os"

	"code.gitea.io/sdk/gitea"
	"github.com/pkg/errors"

	codeset "github.com/fuseml/fuseml-core/gen/codeset"
	config "github.com/fuseml/fuseml-core/pkg/core/config"
)

type GiteaAdminClient interface {
	PrepareRepo(code *codeset.Codeset) error
	GetRepos(org, label *string) ([]*codeset.Codeset, error)
	GetRepo(org, name string) (*codeset.Codeset, error)
}

type giteaAdminClient struct {
	giteaClient *gitea.Client
	url         string
}

// NewGiteaAdminClient creates a new gitea client and performs authentication
// from the credentials provided as env variables
func NewGiteaAdminClient() (GiteaAdminClient, error) {

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

	client, err := gitea.NewClient(url)
	if err != nil {
		return nil, errors.Wrap(err, "gitea client failed")
	}

	client.SetBasicAuth(username, password)

	return giteaAdminClient{
		giteaClient: client,
		url:         url,
	}, nil
}

func (gac giteaAdminClient) GetGiteaURL() (string, error) {
	return gac.url, nil
}

// CreateOrg creates an Org in gitea
func (gac giteaAdminClient) CreateOrg(org string) error {

	_, resp, err := gac.giteaClient.GetOrg(org)
	if resp == nil && err != nil {
		return errors.Wrap(err, "Failed to make get org request")
	}

	if resp.StatusCode == 200 {
		log.Printf("Organization already exists.")
		return nil
	}

	_, _, err = gac.giteaClient.CreateOrg(gitea.CreateOrgOption{
		Name: org,
	})

	if err != nil {
		return errors.Wrap(err, "Failed to create org")
	}
	return nil
}

// create user assigned to current project
func (gac giteaAdminClient) CreateUser(org string) error {
	username := config.DefaultUserName(org)
	user, resp, err := gac.giteaClient.GetUserInfo(username)
	if resp == nil && err != nil {
		return errors.Wrap(err, "Failed to make get user request")
	}
	if user.ID != 0 {
		log.Println("User already exists")
		return nil
	}

	log.Printf("Creating user '%s'", username)
	_, _, err = gac.giteaClient.AdminCreateUser(gitea.CreateUserOption{
		Username:           username,
		Email:              config.DefaultUserEmail,
		Password:           config.DefaultUserPassword,
		MustChangePassword: gitea.OptionalBool(false),
		SendNotify:         false,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to create user")
	}

	teams, _, err := gac.giteaClient.ListOrgTeams(org, gitea.ListTeamsOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to list org teams")
	}
	for _, team := range teams {
		if team.Name == "Owners" {
			_, err = gac.giteaClient.AddTeamMember(team.ID, username)
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
func (gac giteaAdminClient) CreateRepo(c *codeset.Codeset) error {
	_, resp, err := gac.giteaClient.GetRepo(c.Project, c.Name)
	if resp == nil && err != nil {
		return errors.Wrap(err, "Failed to make get repo request")
	}

	if resp.StatusCode == 200 {
		log.Printf("Repository '%s' already exists under '%s'", c.Name, c.Project)
		return nil
	}

	log.Printf("Creating repository '%s' under '%s'...", c.Name, c.Project)
	_, _, err = gac.giteaClient.CreateOrgRepo(c.Project, gitea.CreateRepoOption{
		Name:          c.Name,
		AutoInit:      true,
		Private:       true,
		DefaultBranch: "main",
		Description:   *c.Description,
	})

	if err != nil {
		return errors.Wrap(err, "Failed to create repository")
	}

	return nil
}

// Add topics to given repository
func (gac giteaAdminClient) AddRepoTopics(org, name string, labels []string) error {
	for _, label := range labels {
		_, err := gac.giteaClient.AddRepoTopic(org, name, label)
		if err != nil {
			return err
		}
	}
	return nil
}

// Create webhook for given repository and wire it to tekton listener
func (gac giteaAdminClient) CreateRepoWebhook(org, name string) error {
	hooks, _, err := gac.giteaClient.ListRepoHooks(org, name, gitea.ListHooksOptions{})
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
	gac.giteaClient.CreateRepoHook(org, name, gitea.CreateHookOption{
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
func (gac giteaAdminClient) PrepareRepo(code *codeset.Codeset) error {

	err := gac.CreateOrg(code.Project)
	if err != nil {
		return errors.Wrap(err, "Create org failed")
	}

	err = gac.CreateUser(code.Project)
	if err != nil {
		return errors.Wrap(err, "Create FuseML user failed")
	}

	err = gac.CreateRepo(code)
	if err != nil {
		return errors.Wrap(err, "Create repo failed")
	}

	err = gac.AddRepoTopics(code.Project, code.Name, code.Labels)
	if err != nil {
		return errors.Wrap(err, "Failed to add topics to repository")
	}

	err = gac.CreateRepoWebhook(code.Project, code.Name)
	if err != nil {
		return errors.Wrap(err, "Creating webhook failed")
	}
	return nil
}

// simple check if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// Get all repositories for given project, filter them for label (if given)
func (gac giteaAdminClient) GetReposForOrg(org string, label *string) ([]*codeset.Codeset, error) {
	var codesets []*codeset.Codeset
	log.Printf("Listing repos for org '%s'...", org)
	repos, _, err := gac.giteaClient.ListOrgRepos(org, gitea.ListOrgReposOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list project repos")
	}
	for _, repo := range repos {
		var labels []string
		if label != nil {
			labels, _, err = gac.giteaClient.ListRepoTopics(org, repo.Name, gitea.ListRepoTopicsOptions{})
			if err != nil {
				return nil, errors.Wrap(err, "Failed to list repo topics")
			}
			if !contains(labels, *label) {
				continue
			}
		}
		codesets = append(codesets, &codeset.Codeset{
			Name:    repo.Name,
			Project: org,
			Labels:  labels,
		})
	}
	return codesets, nil
}

// Find all repositories, optionally filtered by project
func (gac giteaAdminClient) GetRepos(org, label *string) ([]*codeset.Codeset, error) {

	var allRepos []*codeset.Codeset
	var orgs []*gitea.Organization

	if org == nil {
		log.Printf("Going through all orgs...")
		var err error
		orgs, _, err = gac.giteaClient.ListMyOrgs(gitea.ListOrgsOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "Failed to list orgs")
		}
	} else {
		orgs = append(orgs, &gitea.Organization{UserName: *org})
	}

	for _, o := range orgs {
		repos, err := gac.GetReposForOrg(o.UserName, label)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to list repos for org "+o.UserName)
		}
		allRepos = append(allRepos, repos...)
	}
	return allRepos, nil
}

// Get the information about repository
func (gac giteaAdminClient) GetRepo(org, name string) (*codeset.Codeset, error) {
	log.Printf("Get repo %s for org '%s'...", name, org)
	repo, _, err := gac.giteaClient.GetRepo(org, name)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read repository")
	}

	ret := codeset.Codeset{
		Name:    repo.Name,
		Project: org,
	}
	if repo.Description == "" {
		ret.Description = nil
	} else {
		ret.Description = &repo.Description
	}
	labels, _, err := gac.giteaClient.ListRepoTopics(org, name, gitea.ListRepoTopicsOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list repo topics")
	}
	ret.Labels = labels

	return &ret, nil
}
