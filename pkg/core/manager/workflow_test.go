package manager

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/core"
	"github.com/fuseml/fuseml-core/pkg/domain"
	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/pipeline/test/diff"
)

const errCodesetNotFound = codesetErr("codeset not found")

var (
	// workflowBackend stores WorkflowListener and WorkflowRuns for a workflow created by fakeWorkflowBackend
	workflowBackend domain.WorkflowBackend

	// workflowStore stores Workflow and Assignments
	workflowStore domain.WorkflowStore

	// codesetStore stores codesets that are created when initializing fakeWorkflowManager
	// The following codesets are created when calling newFakeWorkflowManager:
	// 1. name: cs0, project: csproject0
	// 2. name: cs1, project: csproject1
	// 3. name: cs2, project: csproject1
	codesetStore domain.CodesetStore

	// workflowRunStatuses are the possible Status for a WorkflowRun. The status of a WorkflowRun is set
	// accordingly to its order, cycling between the workflowRunStatuses. E.g. run0: Succeeded, run1: Failed,
	// run2: Succeeded, ...
	workflowRunStatuses = []string{"Succeeded", "Failed"}
)

type codesetErr string

func (e codesetErr) Error() string {
	return string(e)
}

func TestCreate(t *testing.T) {
	t.Run("new workflow", func(t *testing.T) {
		mgr := newFakeWorkflowManager(t)
		wf := workflow.Workflow{Name: "test"}
		got, err := mgr.Create(context.Background(), &wf)
		assertError(t, err, nil)

		want, _ := workflowStore.GetWorkflow(context.TODO(), wf.Name)
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Workflow: %s", diff.PrintWantGot(d))
		}

		codesets, _ := codesetStore.GetAll(context.TODO(), nil, nil)
		err = workflowBackend.CreateWorkflowRun(context.TODO(), wf.Name, codesets[0])
		assertError(t, err, nil)
	})

	t.Run("existing workflow", func(t *testing.T) {
		mgr := newFakeWorkflowManager(t)
		wf := workflow.Workflow{Name: "test"}
		_, err := mgr.Create(context.Background(), &wf)
		assertError(t, err, nil)

		_, err = mgr.Create(context.Background(), &wf)
		assertError(t, err, domain.ErrWorkflowExists)

		got := workflowStore.GetWorkflows(context.TODO(), nil)
		want := []*workflow.Workflow{&wf}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Workflow: %s", diff.PrintWantGot(d))
		}
	})

}

func TestList(t *testing.T) {

}

func TestGet(t *testing.T) {

}

func TestDelete(t *testing.T) {

}

func TestAssignToCodeset(t *testing.T) {

}

func TestUnassignFromCodeset(t *testing.T) {

}

func TestListAssignments(t *testing.T) {

}

func TestListRuns(t *testing.T) {

}

func assertError(t testing.TB, got, want error) {
	t.Helper()

	if got != want {
		t.Errorf("got error %q want %q", got, want)
	}
}

func newFakeWorkflowManager(t *testing.T) domain.WorkflowManager {
	t.Helper()

	workflowStore = core.NewWorkflowStore()
	workflowBackend = &fakeWorkflowBackend{t, make(map[string]*fakeStorableWorkflow)}
	codesetStore = &fakeCodesetStore{t, make(map[codesetID]fakeStorableCodeset)}

	// add codesets to the codeset store for the tests to use it:
	// 1. name: cs0, project: csproject0
	// 2. name: cs1, project: csproject1
	// 3. name: cs2, project: csproject1
	for i := 0; i < 3; i++ {
		projectIndex := i
		if i == 2 {
			projectIndex = i - 1
		}
		_, _, _, err := codesetStore.Add(context.Background(), &domain.Codeset{
			Name:    fmt.Sprintf("cs%d", i),
			Project: fmt.Sprintf("csproject%d", projectIndex),
			URL:     fmt.Sprintf("http://codeset/test-project%d/cs%d", projectIndex, i),
		})
		if err != nil {
			t.Fatalf("Error initializing fake codeset store")
		}
	}

	return NewWorkflowManager(workflowBackend, workflowStore, codesetStore)
}

type fakeStorableWorkflow struct {
	listener *domain.WorkflowListener
	runs     []*workflow.WorkflowRun
}

type fakeWorkflowBackend struct {
	t         *testing.T
	workflows map[string]*fakeStorableWorkflow
}

func (b *fakeWorkflowBackend) CreateWorkflow(ctx context.Context, w *workflow.Workflow) error {
	b.t.Helper()

	if _, exists := b.workflows[w.Name]; exists {
		return domain.ErrWorkflowExists
	}
	b.workflows[w.Name] = &fakeStorableWorkflow{nil, []*workflow.WorkflowRun{}}
	return nil
}

func (b *fakeWorkflowBackend) DeleteWorkflow(ctx context.Context, workflowName string) error {
	b.t.Helper()

	return nil
}

func (b *fakeWorkflowBackend) CreateWorkflowRun(ctx context.Context, workflowName string, codeset *domain.Codeset) error {
	b.t.Helper()

	return nil
}

func (b *fakeWorkflowBackend) ListWorkflowRuns(ctx context.Context, wf *workflow.Workflow, filter *domain.WorkflowRunFilter) ([]*workflow.WorkflowRun, error) {
	b.t.Helper()

	return nil, nil
}

func (b *fakeWorkflowBackend) CreateWorkflowListener(ctx context.Context, workflowName string, timeout time.Duration) (*domain.WorkflowListener, error) {
	b.t.Helper()

	return nil, nil
}

func (b *fakeWorkflowBackend) DeleteWorkflowListener(ctx context.Context, workflowName string) error {
	b.t.Helper()

	return nil
}

func (b *fakeWorkflowBackend) GetWorkflowListener(ctx context.Context, workflowName string) (*domain.WorkflowListener, error) {
	b.t.Helper()

	return nil, nil
}

type codesetID struct {
	name    string
	project string
}

type fakeStorableCodeset struct {
	codeset  *domain.Codeset
	webhooks map[int64]string
}

type fakeCodesetStore struct {
	t     *testing.T
	store map[codesetID]fakeStorableCodeset
}

func (fcs *fakeCodesetStore) Add(ctx context.Context, c *domain.Codeset) (*domain.Codeset, *string, *string, error) {
	fcs.t.Helper()

	fcs.store[codesetID{c.Name, c.Project}] = fakeStorableCodeset{codeset: c, webhooks: make(map[int64]string)}
	return c, nil, nil, nil
}

func (fcs *fakeCodesetStore) CreateWebhook(ctx context.Context, c *domain.Codeset, url string) (*int64, error) {
	fcs.t.Helper()

	return nil, nil
}

func (fcs *fakeCodesetStore) DeleteWebhook(ctx context.Context, c *domain.Codeset, id *int64) error {
	fcs.t.Helper()

	return nil
}

func (fcs *fakeCodesetStore) Delete(ctx context.Context, project, name string) error {
	return nil
}

func (fcs *fakeCodesetStore) Find(ctx context.Context, project, name string) (*domain.Codeset, error) {
	fcs.t.Helper()

	return nil, nil
}

func (fcs *fakeCodesetStore) GetAll(ctx context.Context, project, label *string) (res []*domain.Codeset, err error) {
	fcs.t.Helper()

	for _, c := range fcs.store {
		res = append(res, c.codeset)
	}
	return res, nil
}
