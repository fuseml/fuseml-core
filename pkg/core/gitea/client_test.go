package gitea

import (
	"code.gitea.io/sdk/gitea"
	codeset "github.com/fuseml/fuseml-core/gen/codeset"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
)

type TestStore struct {
	repositories   map[string]gitea.Repository
	projects       map[string]gitea.Organization
	projects2repos map[string][]*gitea.Repository
	teams          map[int64][]string
}

// Replace all methods that are caled from actual gitea client with the ones operating
// on local structures instead of git server
type testGiteaClient struct {
	testStore *TestStore
}

func NewTestStore() *TestStore {
	return &TestStore{
		repositories:   make(map[string]gitea.Repository),
		projects:       make(map[string]gitea.Organization),
		projects2repos: make(map[string][]*gitea.Repository),
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

func NewTestGiteaAdminClient(testStore *TestStore) *giteaAdminClient {
	return &giteaAdminClient{
		giteaClient: &testGiteaClient{testStore},
		logger:      testLogger(),
		url:         testURL,
	}
}

func (tc testGiteaClient) AddRepoTopic(string, string, string) (*gitea.Response, error) {
	return nil, nil
}
func (tc testGiteaClient) GetOrg(orgname string) (*gitea.Organization, *gitea.Response, error) {
	if org, ok := tc.testStore.projects[orgname]; ok {
		httpResp := http.Response{StatusCode: 200}
		return &org, &gitea.Response{Response: &httpResp}, nil
	}
	return nil, nil, nil
}
func (tc testGiteaClient) CreateOrg(opt gitea.CreateOrgOption) (*gitea.Organization, *gitea.Response, error) {
	org := gitea.Organization{UserName: opt.Name}
	tc.testStore.projects[opt.Name] = org
	tc.testStore.projects2repos[opt.Name] = make([]*gitea.Repository, 0)
	return &org, nil, nil
}
func (tc testGiteaClient) GetUserInfo(string) (*gitea.User, *gitea.Response, error) {
	return &gitea.User{ID: 0}, nil, nil
}
func (tc testGiteaClient) AdminCreateUser(gitea.CreateUserOption) (*gitea.User, *gitea.Response, error) {
	return nil, nil, nil
}
func (tc testGiteaClient) ListOrgTeams(string, gitea.ListTeamsOptions) ([]*gitea.Team, *gitea.Response, error) {
	// return default team for any org
	teams := []*gitea.Team{&gitea.Team{Name: "Owners", ID: 42}}
	return teams, nil, nil
}
func (tc testGiteaClient) AddTeamMember(id int64, username string) (*gitea.Response, error) {
	tc.testStore.teams[id] = append(tc.testStore.teams[id], username)
	return nil, nil
}
func (tc testGiteaClient) GetRepo(owner, reponame string) (*gitea.Repository, *gitea.Response, error) {
	if repo, ok := tc.testStore.repositories[reponame]; ok {
		return &repo, nil, nil
	}
	return nil, nil, nil

}
func (tc testGiteaClient) CreateOrgRepo(org string, repo gitea.CreateRepoOption) (*gitea.Repository, *gitea.Response, error) {
	r := gitea.Repository{Name: repo.Name}
	tc.testStore.repositories[repo.Name] = r
	tc.testStore.projects2repos[org] = append(tc.testStore.projects2repos[org], &r)
	return &r, nil, nil
}
func (tc testGiteaClient) ListRepoHooks(string, string, gitea.ListHooksOptions) ([]*gitea.Hook, *gitea.Response, error) {
	return nil, nil, nil
}
func (tc testGiteaClient) ListOrgRepos(org string, opt gitea.ListOrgReposOptions) ([]*gitea.Repository, *gitea.Response, error) {
	if repos, ok := tc.testStore.projects2repos[org]; ok {
		return repos, nil, nil
	}
	return nil, nil, nil
}

func (tc testGiteaClient) CreateRepoHook(string, string, gitea.CreateHookOption) (*gitea.Hook, *gitea.Response, error) {
	return nil, nil, nil
}
func (tc testGiteaClient) ListRepoTopics(org, repo string, opt gitea.ListRepoTopicsOptions) ([]string, *gitea.Response, error) {
	return nil, nil, nil
}
func (tc testGiteaClient) ListMyOrgs(gitea.ListOrgsOptions) ([]*gitea.Organization, *gitea.Response, error) {
	allOrgs := make([]*gitea.Organization, 0)
	for _, org := range tc.testStore.projects {
		allOrgs = append(allOrgs, &org)
	}
	return allOrgs, nil, nil
}

var (
	project1              = "test-project1"
	project2              = "test-project2"
	name                  = "test"
	testURL               = "http://gitea.example.io"
	testListenerStringURL = "tekton-listener"
	testListenerURL       = &testListenerStringURL
)

func getTestCodeset() *codeset.Codeset {

	description := "Test description"
	return &codeset.Codeset{
		Project:     project1,
		Name:        name,
		Description: &description,
		Labels:      []string{"mlflow", "test"},
	}
}

func assertError(t testing.TB, got error, want string) {
	t.Helper()
	if got == nil {
		t.Fatal("didn't get an error but wanted one")
	}

	if got.Error() != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPrepareRepository(t *testing.T) {

	testStore := NewTestStore()
	testGiteaAdminClient := NewTestGiteaAdminClient(testStore)
	code := getTestCodeset()

	// checking initial state of owners team members
	if len(testStore.teams) > 0 {
		t.Errorf("Initial number of teams is not empty")
	}

	err := testGiteaAdminClient.PrepareRepository(code, testListenerURL)
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

	testGiteaAdminClient := NewTestGiteaAdminClient(NewTestStore())

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

func TestGetRepositories(t *testing.T) {

	testGiteaAdminClient := NewTestGiteaAdminClient(NewTestStore())

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

func TestNewGiteaAdminClient(t *testing.T) {

	os.Unsetenv("GITEA_URL")
	_, err := NewAdminClient(testLogger())

	assertError(t, err, errGITEAURLMissing)
}
