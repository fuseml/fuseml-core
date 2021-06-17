package gitea

import (
	"io"
	"log"
	"net/http"
	"os"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

type TestStore struct {
	projects       map[string]gitea.Organization
	projects2repos map[string]map[string]gitea.Repository
	teams          map[int64][]string
}

// Replace all methods that are caled from actual gitea client with the ones operating
// on local structures instead of git server
type testGiteaClient struct {
	testStore *TestStore
	logger    *log.Logger
}

func NewTestStore() *TestStore {
	return &TestStore{
		projects:       make(map[string]gitea.Organization),
		projects2repos: make(map[string]map[string]gitea.Repository),
		teams:          make(map[int64][]string),
	}
}

// Set a specific logger just for testing
func testLogger() *log.Logger {

	logger := log.New(os.Stderr, "[test] ", log.Ltime)
	// suppress the regular output from app
	logger.SetOutput(io.Discard)
	return logger
}

func newTestGiteaAdminClient(testStore *TestStore) *giteaAdminClient {
	return &giteaAdminClient{
		giteaClient: &testGiteaClient{testStore, testLogger()},
		logger:      testLogger(),
		url:         testURL,
	}
}

func (tc testGiteaClient) AddRepoTopic(string, string, string) (*gitea.Response, error) {
	return nil, nil
}
func (tc testGiteaClient) GetOrg(orgname string) (*gitea.Organization, *gitea.Response, error) {
	if org, ok := tc.testStore.projects[orgname]; ok {
		return &org, &gitea.Response{Response: &httpResp200}, nil
	}
	return nil, nil, nil
}
func (tc testGiteaClient) CreateOrg(opt gitea.CreateOrgOption) (*gitea.Organization, *gitea.Response, error) {
	org := gitea.Organization{UserName: opt.Name}
	tc.testStore.projects[opt.Name] = org
	tc.testStore.projects2repos[opt.Name] = make(map[string]gitea.Repository)
	return &org, nil, nil
}
func (tc testGiteaClient) GetUserInfo(string) (*gitea.User, *gitea.Response, error) {
	return &gitea.User{ID: 0}, nil, nil
}
func (tc testGiteaClient) AdminCreateUser(gitea.CreateUserOption) (*gitea.User, *gitea.Response, error) {
	return nil, nil, nil
}
func (tc testGiteaClient) AdminDeleteUser(user string) (*gitea.Response, error) {
	return nil, nil
}
func (tc testGiteaClient) ListOrgTeams(string, gitea.ListTeamsOptions) ([]*gitea.Team, *gitea.Response, error) {
	// return default team for any org
	teams := []*gitea.Team{{Name: "Owners", ID: 42}}
	return teams, nil, nil
}
func (tc testGiteaClient) AddTeamMember(id int64, username string) (*gitea.Response, error) {
	tc.testStore.teams[id] = append(tc.testStore.teams[id], username)
	return nil, nil
}
func (tc testGiteaClient) DeleteOrgMembership(org, user string) (*gitea.Response, error) {
	return &gitea.Response{Response: &httpResp200}, nil
}

func (tc testGiteaClient) ListTeamMembers(id int64, opts gitea.ListTeamMembersOptions) ([]*gitea.User, *gitea.Response, error) {
	users := make([]*gitea.User, 0)
	return users, &gitea.Response{Response: &httpResp200}, nil
}

func (tc testGiteaClient) GetRepo(owner, reponame string) (*gitea.Repository, *gitea.Response, error) {
	if repo, ok := tc.testStore.projects2repos[owner][reponame]; ok {
		return &repo, &gitea.Response{Response: &httpResp200}, nil
	}
	return nil, &gitea.Response{Response: &httpResp404}, nil
}
func (tc testGiteaClient) CreateOrgRepo(org string, repo gitea.CreateRepoOption) (*gitea.Repository, *gitea.Response, error) {
	r := gitea.Repository{Name: repo.Name}
	tc.testStore.projects2repos[org][repo.Name] = r
	return &r, nil, nil
}
func (tc testGiteaClient) ListRepoHooks(string, string, gitea.ListHooksOptions) ([]*gitea.Hook, *gitea.Response, error) {
	return nil, nil, nil
}
func (tc testGiteaClient) ListOrgRepos(org string, opt gitea.ListOrgReposOptions) ([]*gitea.Repository, *gitea.Response, error) {
	repos := make([]*gitea.Repository, 0)
	for _, repo := range tc.testStore.projects2repos[org] {
		r := repo
		repos = append(repos, &r)
	}
	return repos, nil, nil
}

func (tc testGiteaClient) ListUserOrgs(user string, opt gitea.ListOrgsOptions) ([]*gitea.Organization, *gitea.Response, error) {
	userOrgs := make([]*gitea.Organization, 0)
	return userOrgs, nil, nil
}

func (tc testGiteaClient) CreateRepoHook(string, string, gitea.CreateHookOption) (*gitea.Hook, *gitea.Response, error) {
	return &gitea.Hook{ID: int64(1)}, nil, nil
}
func (tc testGiteaClient) DeleteRepoHook(string, string, int64) (*gitea.Response, error) {
	return &gitea.Response{Response: &httpResp200}, nil
}
func (tc testGiteaClient) ListRepoTopics(org, repo string, opt gitea.ListRepoTopicsOptions) ([]string, *gitea.Response, error) {
	return nil, nil, nil
}
func (tc testGiteaClient) ListMyOrgs(gitea.ListOrgsOptions) ([]*gitea.Organization, *gitea.Response, error) {
	allOrgs := make([]*gitea.Organization, 0)
	for _, org := range tc.testStore.projects {
		o := org
		allOrgs = append(allOrgs, &o)
	}
	return allOrgs, nil, nil
}

func (tc testGiteaClient) DeleteRepo(owner, repo string) (*gitea.Response, error) {
	delete(tc.testStore.projects2repos[owner], repo)
	return nil, nil
}

func (tc testGiteaClient) DeleteOrg(orgname string) (*gitea.Response, error) {
	delete(tc.testStore.projects, orgname)
	return nil, nil
}

var (
	project1              = "test-project1"
	project2              = "test-project2"
	name                  = "test"
	testURL               = "http://gitea.example.io"
	testListenerStringURL = "tekton-listener"
	testListenerURL       = &testListenerStringURL
	httpResp200           = http.Response{StatusCode: 200}
	httpResp404           = http.Response{StatusCode: 404}
)

func getTestCodeset() *domain.Codeset {

	return &domain.Codeset{
		Project:     project1,
		Name:        name,
		Description: "Test description",
		Labels:      []string{"mlflow", "test"},
	}
}

func assertError(t testing.TB, got, want error) {
	t.Helper()

	if got != want {
		t.Errorf("got error %q, want %q", got, want)
	}
}

func TestPrepareRepository(t *testing.T) {

	testStore := NewTestStore()
	testGiteaAdminClient := newTestGiteaAdminClient(testStore)
	code := getTestCodeset()

	// checking initial state of owners team members
	if len(testStore.teams) > 0 {
		t.Errorf("Initial number of teams is not empty")
	}

	_, _, err := testGiteaAdminClient.PrepareRepository(code, testListenerURL)
	if err != nil {
		t.Errorf("Error preparing repository: %v", err)
	}

	// checking state of owners team member after adding repo
	if len(testStore.teams) < 1 {
		t.Errorf("No teams were created after adding repository")
	}

	// fetch the id of Owners team
	teams, _, _ := testGiteaAdminClient.giteaClient.ListOrgTeams(project1, gitea.ListTeamsOptions{})
	ownersID := teams[0].ID
	usersInTeam := testStore.teams[ownersID]

	if len(usersInTeam) < 1 {
		t.Errorf("No users present in Owners team after adding repository")
	}

	if !contains(usersInTeam, generateUserName(project1)) {
		t.Errorf("New user is not present in the Owners team")
	}
}

func TestGetRepository(t *testing.T) {

	testGiteaAdminClient := newTestGiteaAdminClient(NewTestStore())

	// Reading repo that was not added should throw error
	_, err := testGiteaAdminClient.GetRepository(project1, name)

	assertError(t, err, errRepoNotFound)

	// Prepare new repo
	testGiteaAdminClient.PrepareRepository(getTestCodeset(), testListenerURL)

	// Get the repo now
	c, err := testGiteaAdminClient.GetRepository(project1, name)
	if err != nil {
		t.Errorf("Error geting repository that was just created")
	}
	if c.Name != name || c.Project != project1 {
		t.Errorf("Wrong codeset returned: %v", c)
	}
}

func TestDeleteRepository(t *testing.T) {

	testGiteaAdminClient := newTestGiteaAdminClient(NewTestStore())

	// Reading repo that was not added should throw error
	_, err := testGiteaAdminClient.GetRepository(project1, name)

	assertError(t, err, errRepoNotFound)

	// Prepare new repo
	testGiteaAdminClient.PrepareRepository(getTestCodeset(), testListenerURL)

	// Get the repo now
	_, err = testGiteaAdminClient.GetRepository(project1, name)
	if err != nil {
		t.Errorf("Error geting repository that was just created")
	}

	err = testGiteaAdminClient.DeleteRepository(project1, name)
	if err != nil {
		t.Errorf("Error deleting repository")
	}

	c, _ := testGiteaAdminClient.GetRepository(project1, name)
	if c != nil {
		t.Errorf("Repository still present after deleting")
	}

	err = testGiteaAdminClient.DeleteRepository(project1, name)
	if err != nil {
		t.Errorf("Error: deleting non existent repository should not fail")
	}
}

func TestGetRepositories(t *testing.T) {

	testGiteaAdminClient := newTestGiteaAdminClient(NewTestStore())

	repos, err := testGiteaAdminClient.GetRepositories(&project1, nil)
	if len(repos) > 0 {
		t.Errorf("Initial set of repositories is not empty")
	}
	if err != nil {
		t.Errorf("Error reading list of repositories")
	}
	testGiteaAdminClient.PrepareRepository(getTestCodeset(), testListenerURL)

	repos, _ = testGiteaAdminClient.GetRepositories(&project1, nil)
	if len(repos) < 1 {
		t.Errorf("List of repositories is empty after adding")
	}

	// now add new project+repo and list all repos accross projects
	codeset2 := getTestCodeset()
	codeset2.Project = project2
	testGiteaAdminClient.PrepareRepository(codeset2, testListenerURL)

	repos, _ = testGiteaAdminClient.GetRepositories(nil, nil)
	if len(repos) != 2 {
		t.Errorf("There are not 2 repos in total")
	}
}

func TestAddDeleteOrgs(t *testing.T) {

	testGiteaAdminClient := newTestGiteaAdminClient(NewTestStore())

	testGiteaAdminClient.PrepareRepository(getTestCodeset(), testListenerURL)

	p1, err := testGiteaAdminClient.GetProject(project1)
	assertError(t, err, nil)

	if p1.Name != project1 {
		t.Errorf("wrong name of project: %v, not %s", p1.Name, project1)
	}

	p2, err := testGiteaAdminClient.CreateProject(project2, "description of "+project2, false)
	assertError(t, err, nil)

	if p2.Name != project2 {
		t.Errorf("wrong name of project: %v, not %s", p2.Name, project2)
	}

	// create same project, ignore if it exists
	_, err = testGiteaAdminClient.CreateProject(project2, "description of "+project2, true)
	assertError(t, err, nil)

	// create same project, fail if it exists
	_, err = testGiteaAdminClient.CreateProject(project2, "description of "+project2, false)
	assertError(t, err, errProjectExists)

	// list all projects, there should be 2
	projects, err := testGiteaAdminClient.GetProjects()
	assertError(t, err, nil)

	if len(projects) != 2 {
		t.Errorf("There are not 2 projects in total")
	}

	// project2 is empty, should not be a problem to delete
	err = testGiteaAdminClient.DeleteProject(project2)
	assertError(t, err, nil)

	// project1 is not empty, error on delete
	err = testGiteaAdminClient.DeleteProject(project1)
	assertError(t, err, errProjectNotEmpty)

	// list all projects after delete
	projects, err = testGiteaAdminClient.GetProjects()
	assertError(t, err, nil)

	if len(projects) != 1 {
		t.Errorf("There is not just 1 project in total (got %d)", len(projects))
	}
}

func TestNewGiteaAdminClient(t *testing.T) {

	os.Unsetenv("GITEA_URL")
	_, err := NewAdminClient(testLogger())

	assertError(t, err, errGITEAURLMissing)
}
