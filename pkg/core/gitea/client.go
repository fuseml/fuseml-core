package gitea

import (
	"log"
	"math/rand"
	"os"

	"code.gitea.io/sdk/gitea"
	"github.com/pkg/errors"

	config "github.com/fuseml/fuseml-core/pkg/core/config"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

// AdminClient describes the interface of Gitea Admin Client
type AdminClient interface {
	PrepareRepository(*domain.Codeset, *string) (*string, *string, error)
	CreateRepoWebhook(string, string, *string) (*int64, error)
	DeleteRepoWebhook(string, string, *int64) error
	GetRepositories(org, label *string) ([]*domain.Codeset, error)
	GetRepository(org, name string) (*domain.Codeset, error)
	DeleteRepository(org, name string) error
	GetProjects() ([]*domain.Project, error)
	GetProject(org string) (*domain.Project, error)
	DeleteProject(org string) error
	CreateProject(string, string, bool) (*domain.Project, error)
}

// Client describes the interface of Gitea Client
type Client interface {
	GetOrg(orgname string) (*gitea.Organization, *gitea.Response, error)
	CreateOrg(gitea.CreateOrgOption) (*gitea.Organization, *gitea.Response, error)
	GetUserInfo(string) (*gitea.User, *gitea.Response, error)
	AdminCreateUser(gitea.CreateUserOption) (*gitea.User, *gitea.Response, error)
	AdminDeleteUser(string) (*gitea.Response, error)
	ListOrgTeams(string, gitea.ListTeamsOptions) ([]*gitea.Team, *gitea.Response, error)
	AddTeamMember(int64, string) (*gitea.Response, error)
	ListTeamMembers(int64, gitea.ListTeamMembersOptions) ([]*gitea.User, *gitea.Response, error)
	GetRepo(string, string) (*gitea.Repository, *gitea.Response, error)
	CreateOrgRepo(string, gitea.CreateRepoOption) (*gitea.Repository, *gitea.Response, error)
	AddRepoTopic(string, string, string) (*gitea.Response, error)
	ListRepoHooks(string, string, gitea.ListHooksOptions) ([]*gitea.Hook, *gitea.Response, error)
	ListOrgRepos(string, gitea.ListOrgReposOptions) ([]*gitea.Repository, *gitea.Response, error)
	CreateRepoHook(string, string, gitea.CreateHookOption) (*gitea.Hook, *gitea.Response, error)
	DeleteRepoHook(string, string, int64) (*gitea.Response, error)
	ListRepoTopics(string, string, gitea.ListRepoTopicsOptions) ([]string, *gitea.Response, error)
	ListMyOrgs(gitea.ListOrgsOptions) ([]*gitea.Organization, *gitea.Response, error)
	ListUserOrgs(string, gitea.ListOrgsOptions) ([]*gitea.Organization, *gitea.Response, error)
	DeleteRepo(string, string) (*gitea.Response, error)
	DeleteOrg(string) (*gitea.Response, error)
	DeleteOrgMembership(org, user string) (*gitea.Response, error)
}

// giteaAdminClient is the struct holding information about gitea client
type giteaAdminClient struct {
	giteaClient Client
	url         string
	logger      *log.Logger
}

const (
	errGITEAURLMissing           = giteaErr("Value for gitea URL (GITEA_URL) was not provided.")
	errGITEAADMINUSERNAMEMissing = giteaErr("Value for gitea admin user name (GITEA_ADMIN_USERNAME) was not provided.")
	errGITEAADMINPASSWORDMissing = giteaErr("Value for gitea admin user password (GITEA_ADMIN_PASSWORD) was not provided.")
	errRepoNotFound              = giteaErr("Repository by that name not found")
	errProjectNotEmpty           = giteaErr("Project has still codesets assigned. Delete them first")
)

type giteaErr string

func (e giteaErr) Error() string {
	return string(e)
}

var lettersForPassword = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var generatedPasswordLength = 16

// For now, turn off the password generation
// But we need to fine a way how user can alter this value (env variable, config file, client param...)
var generateUserPassword = false

// NewAdminClient creates a new gitea client and performs authentication
// from the credentials provided as env variables
func NewAdminClient(logger *log.Logger) (AdminClient, error) {

	url, exists := os.LookupEnv("GITEA_URL")
	if !exists {
		return nil, errGITEAURLMissing
	}
	username, exists := os.LookupEnv("GITEA_ADMIN_USERNAME")
	if !exists {
		return nil, errGITEAADMINUSERNAMEMissing
	}
	password, exists := os.LookupEnv("GITEA_ADMIN_PASSWORD")
	if !exists {
		return nil, errGITEAADMINPASSWORDMissing
	}

	client, err := gitea.NewClient(url)
	if err != nil {
		return nil, errors.Wrap(err, "gitea client failed")
	}

	client.SetBasicAuth(username, password)

	logger.Printf("Using GITEA from: %s", url)

	return giteaAdminClient{
		giteaClient: client,
		url:         url,
		logger:      logger,
	}, nil
}

func generateUserName(org string) string {
	return config.DefaultUserName(org)
}

func getUserPassword() string {
	if !generateUserPassword {
		return config.DefaultUserPassword
	}
	p := make([]rune, generatedPasswordLength)
	for i := range p {
		p[i] = lettersForPassword[rand.Intn(len(lettersForPassword))]
	}
	return string(p)
}

func (gac giteaAdminClient) GetGiteaURL() (string, error) {
	return gac.url, nil
}

// Create a Project (= implemented as Organization in git).
// If ignoreExisting argument is true, the call will not fail when a project with same name already exists.
func (gac giteaAdminClient) CreateProject(name, desc string, ignoreExisting bool) (*domain.Project, error) {
	gac.logger.Printf("Creating project %s....", name)

	_, resp, err := gac.giteaClient.GetOrg(name)
	if resp == nil && err != nil {
		return nil, errors.Wrap(err, "Failed to make get org request")
	}

	if resp != nil && resp.StatusCode == 200 {
		gac.logger.Printf("Project already exists.")
		if ignoreExisting {
			return nil, nil
		}
		return nil, domain.ErrProjectExists
	}

	_, _, err = gac.giteaClient.CreateOrg(gitea.CreateOrgOption{
		Name:        name,
		Description: desc,
	})

	if err != nil {
		return nil, errors.Wrap(err, "Failed to create project")
	}

	return &domain.Project{
		Name:        name,
		Description: desc,
	}, nil
}

// CreateOrg creates an Org in gitea. Does not return an error if it already exists
func (gac giteaAdminClient) createOrganizationIfNotPresent(org string) error {

	_, err := gac.CreateProject(org, "", true)
	return err
}

// create user assigned to current project
func (gac giteaAdminClient) CreateUser(org string) (*string, *string, error) {
	username := generateUserName(org)
	password := getUserPassword()
	user, resp, err := gac.giteaClient.GetUserInfo(username)
	if resp == nil && err != nil {
		return nil, nil, errors.Wrap(err, "Failed to make get user request")
	}
	if user != nil && user.ID != 0 {
		gac.logger.Println("User already exists")
		return nil, nil, nil
	}

	gac.logger.Printf("Creating user '%s'", username)
	_, _, err = gac.giteaClient.AdminCreateUser(gitea.CreateUserOption{
		Username:           username,
		Email:              config.DefaultUserEmail(org),
		Password:           password,
		MustChangePassword: gitea.OptionalBool(false),
		SendNotify:         false,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to create user")
	}

	teams, _, err := gac.giteaClient.ListOrgTeams(org, gitea.ListTeamsOptions{})
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to list org teams")
	}
	for _, team := range teams {
		if team.Name == "Owners" {
			_, err = gac.giteaClient.AddTeamMember(team.ID, username)
			if err != nil {
				return nil, nil, errors.Wrap(err, "Failed adding user to Owners")
			}
			break
		}
	}

	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to create application")
	}

	return &username, &password, nil
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
func (gac giteaAdminClient) CreateRepoWebhook(org, name string, listenerURL *string) (*int64, error) {
	if listenerURL == nil {
		gac.logger.Printf("Webhook listener URL not provided, skipping creation")
		return nil, nil
	}
	hooks, _, err := gac.giteaClient.ListRepoHooks(org, name, gitea.ListHooksOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list webhooks")
	}

	for _, hook := range hooks {
		url := hook.Config["url"]
		if url == *listenerURL {
			gac.logger.Printf("Webhook for '%s' already exists", name)
			return &hook.ID, nil
		}
	}

	gac.logger.Printf("Creating Webhook for '%s' under '%s'...", name, org)
	hook, _, _ := gac.giteaClient.CreateRepoHook(org, name, gitea.CreateHookOption{
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

	return &hook.ID, nil
}

// Delete a webhook for given repository
func (gac giteaAdminClient) DeleteRepoWebhook(org, name string, hookID *int64) error {
	gac.logger.Printf("Deleting Webhook for %q under %q...", name, org)
	resp, err := gac.giteaClient.DeleteRepoHook(org, name, *hookID)
	if err != nil {
		if resp.StatusCode == 404 {
			gac.logger.Printf("Webhook not found, skipping deletion")
			return nil
		}
		return errors.Wrap(err, "Failed to delete webhook")
	}
	return nil
}

// Prepare the org, repository, and create user that clients can use for pushing
func (gac giteaAdminClient) PrepareRepository(code *domain.Codeset, listenerURL *string) (*string, *string, error) {

	err := gac.createOrganizationIfNotPresent(code.Project)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Create org failed")
	}

	user, pass, err := gac.CreateUser(code.Project)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Create FuseML user failed")
	}

	err = gac.CreateRepo(code)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Create repo failed")
	}

	err = gac.AddRepoTopics(code.Project, code.Name, code.Labels)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to add topics to repository")
	}

	_, err = gac.CreateRepoWebhook(code.Project, code.Name, listenerURL)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Creating webhook failed")
	}
	return user, pass, nil
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
		return nil, errRepoNotFound
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

// return all non-admin users that are Owners for given organization
func (gac giteaAdminClient) getProjectOwners(name string) ([]*domain.User, error) {

	var ret []*domain.User
	teams, _, err := gac.giteaClient.ListOrgTeams(name, gitea.ListTeamsOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list org teams")
	}
	for _, team := range teams {
		if team.Name != "Owners" {
			continue
		}
		users, _, err := gac.giteaClient.ListTeamMembers(team.ID, gitea.ListTeamMembersOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "Failed listing members of Owners team")
		}
		for _, u := range users {
			if u.IsAdmin {
				continue
			}
			ret = append(ret, &domain.User{
				Name:  u.UserName,
				Email: u.Email,
			})
		}
		break // no need to continue after checking Owners
	}
	return ret, nil
}

// get all projects
func (gac giteaAdminClient) GetProjects() ([]*domain.Project, error) {
	gac.logger.Printf("listing git orgs....")

	orgs, _, err := gac.giteaClient.ListMyOrgs(gitea.ListOrgsOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list orgs")
	}

	var ret []*domain.Project
	for _, o := range orgs {
		users, err := gac.getProjectOwners(o.UserName)
		if err != nil {
			return nil, err
		}
		ret = append(ret, &domain.Project{
			Name:        o.UserName,
			Description: o.Description,
			Users:       users,
		})
	}
	return ret, nil
}

// return project specified by name
func (gac giteaAdminClient) GetProject(name string) (*domain.Project, error) {
	gac.logger.Printf("Fetching git org %s....", name)

	org, _, err := gac.giteaClient.GetOrg(name)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to make get org request")
	}
	users, err := gac.getProjectOwners(name)
	if err != nil {
		return nil, err
	}
	ret := domain.Project{
		Name:        org.UserName,
		Description: org.Description,
		Users:       users,
	}
	return &ret, nil
}

func (gac giteaAdminClient) DeleteProject(org string) error {
	gac.logger.Printf("Deleting project %s....", org)
	// 1. check if they are no repos
	repos, _, err := gac.giteaClient.ListOrgRepos(org, gitea.ListOrgReposOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to list project repos")
	}

	if len(repos) > 0 {
		return errProjectNotEmpty
	}

	// 2. delete all members of the project, if they are not owning any other project
	owners, err := gac.getProjectOwners(org)
	if err != nil {
		return errors.Wrap(err, "Failed to list project owners")
	}
	usersOrgs := make(map[string]int)
	for _, owner := range owners {
		orgsForUser, _, err := gac.giteaClient.ListUserOrgs(owner.Name, gitea.ListOrgsOptions{})
		if err != nil {
			return errors.Wrap(err, "Failed to list orgs for user")
		}
		for range orgsForUser {
			usersOrgs[owner.Name]++
		}
	}
	for userName, orgNumber := range usersOrgs {
		if orgNumber == 1 {
			gac.logger.Printf("Removing user %s from %s ....", userName, org)
			if _, err := gac.giteaClient.DeleteOrgMembership(org, userName); err != nil {
				return errors.Wrap(err, "Failed to remove user from project")
			}

			gac.logger.Printf("Deleting user %s....", userName)
			if _, err := gac.giteaClient.AdminDeleteUser(userName); err != nil {
				return errors.Wrap(err, "Failed to delete user")
			}
		}
	}

	// 3. delete org now
	_, err = gac.giteaClient.DeleteOrg(org)
	if err != nil {
		return errors.Wrap(err, "Failed deleting project")
	}

	return nil
}
