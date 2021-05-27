package gitea

import (
	"log"
	"os"

	"code.gitea.io/sdk/gitea"
	"github.com/pkg/errors"

	config "github.com/fuseml/fuseml-core/pkg/core/config"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

// AdminClient describes the interface of Gitea Admin Client
type AdminClient interface {
	PrepareRepository(*domain.Codeset, *string) error
	CreateRepoWebhook(string, string, *string) error
	GetRepositories(org, label *string) ([]*domain.Codeset, error)
	GetRepository(org, name string) (*domain.Codeset, error)
	DeleteRepository(org, name string) error
}

// Client describes the interface of Gitea Client
type Client interface {
	GetOrg(orgname string) (*gitea.Organization, *gitea.Response, error)
	CreateOrg(gitea.CreateOrgOption) (*gitea.Organization, *gitea.Response, error)
	GetUserInfo(string) (*gitea.User, *gitea.Response, error)
	AdminCreateUser(gitea.CreateUserOption) (*gitea.User, *gitea.Response, error)
	ListOrgTeams(string, gitea.ListTeamsOptions) ([]*gitea.Team, *gitea.Response, error)
	AddTeamMember(int64, string) (*gitea.Response, error)
	GetRepo(string, string) (*gitea.Repository, *gitea.Response, error)
	CreateOrgRepo(string, gitea.CreateRepoOption) (*gitea.Repository, *gitea.Response, error)
	AddRepoTopic(string, string, string) (*gitea.Response, error)
	ListRepoHooks(string, string, gitea.ListHooksOptions) ([]*gitea.Hook, *gitea.Response, error)
	ListOrgRepos(string, gitea.ListOrgReposOptions) ([]*gitea.Repository, *gitea.Response, error)
	CreateRepoHook(string, string, gitea.CreateHookOption) (*gitea.Hook, *gitea.Response, error)
	ListRepoTopics(string, string, gitea.ListRepoTopicsOptions) ([]string, *gitea.Response, error)
	ListMyOrgs(gitea.ListOrgsOptions) ([]*gitea.Organization, *gitea.Response, error)
	DeleteRepo(string, string) (*gitea.Response, error)
}

// giteaAdminClient is the struct holding information about gitea client
type giteaAdminClient struct {
	giteaClient Client
	url         string
	logger      *log.Logger
}

var errGITEAURLMissing = "Value for gitea URL (GITEA_URL) was not provided."
var errGITEAUSERNAMEMissing = "Value for gitea user name (GITEA_USERNAME) was not provided."
var errGITEAPASSWORDMissing = "Value for gitea user password (GITEA_PASSWORD) was not provided."
var errRepoNotFound = "Repository by that name not found"

// NewAdminClient creates a new gitea client and performs authentication
// from the credentials provided as env variables
func NewAdminClient(logger *log.Logger) (AdminClient, error) {

	url, exists := os.LookupEnv("GITEA_URL")
	if !exists {
		return nil, errors.New(errGITEAURLMissing)
	}
	username, exists := os.LookupEnv("GITEA_USERNAME")
	if !exists {
		return nil, errors.New(errGITEAUSERNAMEMissing)
	}
	password, exists := os.LookupEnv("GITEA_PASSWORD")
	if !exists {
		return nil, errors.New(errGITEAPASSWORDMissing)
	}

	client, err := gitea.NewClient(url)
	if err != nil {
		return nil, errors.Wrap(err, "gitea client failed")
	}

	client.SetBasicAuth(username, password)

	return giteaAdminClient{
		giteaClient: client,
		url:         url,
		logger:      logger,
	}, nil
}

func generateUserName(org string) string {
	return config.DefaultUserName(org)
}

func (gac giteaAdminClient) GetGiteaURL() (string, error) {
	return gac.url, nil
}

// CreateOrg creates an Org in gitea
func (gac giteaAdminClient) CreateOrganization(org string) error {

	gac.logger.Println("creating org " + org)
	_, resp, err := gac.giteaClient.GetOrg(org)
	if resp == nil && err != nil {
		return errors.Wrap(err, "Failed to make get org request")
	}

	if resp != nil && resp.StatusCode == 200 {
		gac.logger.Printf("Organization already exists.")
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
	username := generateUserName(org)
	user, resp, err := gac.giteaClient.GetUserInfo(username)
	if resp == nil && err != nil {
		return errors.Wrap(err, "Failed to make get user request")
	}
	if user != nil && user.ID != 0 {
		gac.logger.Println("User already exists")
		return nil
	}

	gac.logger.Printf("Creating user '%s'", username)
	_, _, err = gac.giteaClient.AdminCreateUser(gitea.CreateUserOption{
		Username:           username,
		Email:              config.DefaultUserEmail(org),
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
func (gac giteaAdminClient) CreateRepo(c *domain.Codeset) error {
	repo, resp, err := gac.giteaClient.GetRepo(c.Project, c.Name)
	if resp == nil && err != nil {
		return errors.Wrap(err, "Failed to make get repo request")
	}

	if resp != nil && resp.StatusCode == 200 {
		gac.logger.Printf("Repository '%s' already exists under '%s'", c.Name, c.Project)
		c.URL = repo.CloneURL
		return nil
	}

	gac.logger.Printf("Creating repository '%s' under '%s'...", c.Name, c.Project)
	repo, _, err = gac.giteaClient.CreateOrgRepo(c.Project, gitea.CreateRepoOption{
		Name:          c.Name,
		AutoInit:      true,
		Private:       false,
		DefaultBranch: "main",
		Description:   c.Description,
	})

	if err != nil {
		return errors.Wrap(err, "Failed to create repository")
	}
	c.URL = repo.CloneURL

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
func (gac giteaAdminClient) CreateRepoWebhook(org, name string, listenerURL *string) error {
	if listenerURL == nil {
		gac.logger.Printf("Webhook listener URL not provided, skipping creation")
		return nil
	}
	hooks, _, err := gac.giteaClient.ListRepoHooks(org, name, gitea.ListHooksOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to list webhooks")
	}

	for _, hook := range hooks {
		url := hook.Config["url"]
		if url == *listenerURL {
			gac.logger.Printf("Webhook for '%s' already exists", name)
			return nil
		}
	}

	gac.logger.Printf("Creating Webhook for '%s' under '%s'...", name, org)
	gac.giteaClient.CreateRepoHook(org, name, gitea.CreateHookOption{
		Active:       true,
		BranchFilter: "*",
		Config: map[string]string{
			"secret":       config.HookSecret,
			"http_method":  "POST",
			"url":          *listenerURL,
			"content_type": "json",
		},
		Type: "gitea",
	})

	return nil
}

// Prepare the org, repository, and create user that clients can use for pushing
func (gac giteaAdminClient) PrepareRepository(code *domain.Codeset, listenerURL *string) error {

	err := gac.CreateOrganization(code.Project)
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

	err = gac.CreateRepoWebhook(code.Project, code.Name, listenerURL)
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
func (gac giteaAdminClient) GetReposForOrg(org string, label *string) ([]*domain.Codeset, error) {
	var codesets []*domain.Codeset
	gac.logger.Printf("Listing repos for org '%s'...", org)
	repos, _, err := gac.giteaClient.ListOrgRepos(org, gitea.ListOrgReposOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list project repos")
	}
	for _, repo := range repos {
		var labels []string
		labels, _, err = gac.giteaClient.ListRepoTopics(org, repo.Name, gitea.ListRepoTopicsOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "Failed to list repo topics")
		}
		if label != nil && !contains(labels, *label) {
			continue
		}

		codesets = append(codesets, &domain.Codeset{
			Name:        repo.Name,
			Project:     org,
			Labels:      labels,
			Description: repo.Description,
			URL:         repo.CloneURL,
		})
	}
	return codesets, nil
}

// Find all repositories, optionally filtered by project
func (gac giteaAdminClient) GetRepositories(org, label *string) ([]*domain.Codeset, error) {

	var allRepos []*domain.Codeset
	var orgs []*gitea.Organization

	if org == nil {
		gac.logger.Printf("Going through all orgs...")
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
func (gac giteaAdminClient) GetRepository(org, name string) (*domain.Codeset, error) {
	gac.logger.Printf("Get repo %s for org '%s'...", name, org)
	repo, _, err := gac.giteaClient.GetRepo(org, name)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read repository")
	}
	if repo == nil || repo.Name == "" {
		return nil, errors.New(errRepoNotFound)
	}

	ret := domain.Codeset{
		Name:        repo.Name,
		Project:     org,
		Description: repo.Description,
	}
	labels, _, err := gac.giteaClient.ListRepoTopics(org, name, gitea.ListRepoTopicsOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list repo topics")
	}
	ret.Labels = labels
	ret.URL = repo.CloneURL

	return &ret, nil
}

// Delete the repository
func (gac giteaAdminClient) DeleteRepository(org, name string) error {
	gac.logger.Printf("Going to delete repo %s for org '%s'...", name, org)

	_, resp, err := gac.giteaClient.GetRepo(org, name)

	if resp.StatusCode == 404 {
		gac.logger.Printf("Repo does not exist, no need to delete")
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "Failed to get repo")
	}

	_, err = gac.giteaClient.DeleteRepo(org, name)
	if err != nil {
		return errors.Wrap(err, "Failed to delete repository")
	}
	return nil
}
