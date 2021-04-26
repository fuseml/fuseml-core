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
	PrepareRepository(code *codeset.Codeset) error
	GetRepositories(org, label *string) ([]*codeset.Codeset, error)
	GetRepository(org, name string) (*codeset.Codeset, error)
}

type GiteaClient interface {
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
}

type giteaAdminClient struct {
	giteaClient GiteaClient
	url         string
	logger      *log.Logger
}

var ErrGITEA_URLMissing = "Value for gitea URL (GITEA_URL) was not provided."
var ErrGITEA_USERNAMEMissing = "Value for gitea user name (GITEA_USERNAME) was not provided."
var ErrGITEA_PASSWORDMissing = "Value for gitea user password (GITEA_PASSWORD) was not provided."
var ErrRepoNotFound = "Repository by that name not found"

// NewGiteaAdminClient creates a new gitea client and performs authentication
// from the credentials provided as env variables
func NewGiteaAdminClient(logger *log.Logger) (GiteaAdminClient, error) {

	url, exists := os.LookupEnv("GITEA_URL")
	if !exists {
		return nil, errors.New(ErrGITEA_URLMissing)
	}
	username, exists := os.LookupEnv("GITEA_USERNAME")
	if !exists {
		return nil, errors.New(ErrGITEA_USERNAMEMissing)
	}
	password, exists := os.LookupEnv("GITEA_PASSWORD")
	if !exists {
		return nil, errors.New(ErrGITEA_PASSWORDMissing)
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

func GenerateUserName(org string) string {
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
	username := GenerateUserName(org)
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

	if resp != nil && resp.StatusCode == 200 {
		gac.logger.Printf("Repository '%s' already exists under '%s'", c.Name, c.Project)
		return nil
	}

	gac.logger.Printf("Creating repository '%s' under '%s'...", c.Name, c.Project)
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
			"url":          config.StagingEventListenerURL,
			"content_type": "json",
		},
		Type: "gitea",
	})

	return nil
}

// Prepare the org, repository, and create user that clients can use for pushing
func (gac giteaAdminClient) PrepareRepository(code *codeset.Codeset) error {

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
	gac.logger.Printf("Listing repos for org '%s'...", org)
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
func (gac giteaAdminClient) GetRepositories(org, label *string) ([]*codeset.Codeset, error) {

	var allRepos []*codeset.Codeset
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
func (gac giteaAdminClient) GetRepository(org, name string) (*codeset.Codeset, error) {
	gac.logger.Printf("Get repo %s for org '%s'...", name, org)
	repo, _, err := gac.giteaClient.GetRepo(org, name)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read repository")
	}
	if repo == nil || repo.Name == "" {
		return nil, errors.New(ErrRepoNotFound)
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
	ret.URL = &repo.CloneURL

	return &ret, nil
}
