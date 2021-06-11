package manager

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/core"
	"github.com/fuseml/fuseml-core/pkg/domain"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
	t.Run("all", func(t *testing.T) {
		mgr := newFakeWorkflowManager(t)
		want := []*workflow.Workflow{}

		// no workflows
		got := mgr.List(context.TODO(), nil)
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Workflow list: %s", diff.PrintWantGot(d))
		}

		// create 3 workflows (wf0, wf1, wf2)
		for i := 0; i < 3; i++ {
			wf, err := mgr.Create(context.Background(), &workflow.Workflow{Name: fmt.Sprintf("wf%d", i)})
			assertError(t, err, nil)
			want = append(want, wf)
		}

		got = mgr.List(context.TODO(), nil)
		if d := cmp.Diff(want, got, cmpopts.SortSlices(func(x, y *workflow.Workflow) bool { return x.Name < y.Name })); d != "" {
			t.Errorf("Unexpected Workflow list: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("by workflow name", func(t *testing.T) {
		mgr := newFakeWorkflowManager(t)
		want := []*workflow.Workflow{}

		// no workflows
		wfName := "does-not-exist"
		got := mgr.List(context.TODO(), &wfName)
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Workflow list: %s", diff.PrintWantGot(d))
		}

		// create 3 workflows (wf0, wf1, wf2)
		for i := 0; i < 3; i++ {
			wf, err := mgr.Create(context.Background(), &workflow.Workflow{Name: fmt.Sprintf("wf%d", i)})
			assertError(t, err, nil)
			want = append(want, wf)
		}

		for i := 0; i < len(want); i++ {
			name := fmt.Sprintf("wf%d", i)
			got := mgr.List(context.TODO(), &name)
			if d := cmp.Diff([]*workflow.Workflow{want[i]}, got, cmpopts.SortSlices(func(x, y *workflow.Workflow) bool { return x.Name < y.Name })); d != "" {
				t.Errorf("Unexpected Workflow list: %s", diff.PrintWantGot(d))
			}
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		mgr := newFakeWorkflowManager(t)
		want, err := mgr.Create(context.Background(), &workflow.Workflow{Name: "wf"})
		assertError(t, err, nil)

		got, err := mgr.Get(context.Background(), want.Name)
		assertError(t, err, nil)

		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Workflow: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("not found", func(t *testing.T) {
		mgr := newFakeWorkflowManager(t)
		_, err := mgr.Get(context.Background(), "wf")
		assertError(t, err, domain.ErrWorkflowNotFound)
	})
}

func TestDelete(t *testing.T) {

}

func TestAssignToCodeset(t *testing.T) {
	t.Run("assign", func(t *testing.T) {
		mgr := newFakeWorkflowManager(t)

		wf, err := mgr.Create(context.Background(), &workflow.Workflow{Name: "wf"})
		assertError(t, err, nil)

		codesets, _ := codesetStore.GetAll(context.TODO(), nil, nil)
		wantListener, webhookID, err := mgr.AssignToCodeset(context.Background(), wf.Name, codesets[0].Project, codesets[0].Name)
		assertError(t, err, nil)

		got := workflowStore.GetAssignments(context.TODO(), &wf.Name)
		want := map[string][]*domain.AssignedCodeset{wf.Name: {{Codeset: codesets[0], WebhookID: webhookID}}}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Assignment: %s", diff.PrintWantGot(d))
		}

		gotListener, err := workflowBackend.GetWorkflowListener(context.TODO(), wf.Name)
		assertError(t, err, nil)
		if d := cmp.Diff(wantListener, gotListener); d != "" {
			t.Errorf("Unexpected Listener: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("workflow not found", func(t *testing.T) {
		mgr := newFakeWorkflowManager(t)

		wfName := "unknownWf"
		codesets, _ := codesetStore.GetAll(context.TODO(), nil, nil)
		_, _, got := mgr.AssignToCodeset(context.Background(), wfName, codesets[0].Project, codesets[0].Name)
		assertError(t, got, domain.ErrWorkflowNotFound)

		gotAss := workflowStore.GetAssignments(context.TODO(), nil)
		wantAss := map[string][]*domain.AssignedCodeset{}
		if d := cmp.Diff(wantAss, gotAss); d != "" {
			t.Errorf("Unexpected Assignment: %s", diff.PrintWantGot(d))
		}

		_, err := workflowBackend.GetWorkflowListener(context.TODO(), wfName)
		assertStrings(t, err.Error(), "listener not found")

	})

	t.Run("codeset not found", func(t *testing.T) {
		mgr := newFakeWorkflowManager(t)

		wf, err := mgr.Create(context.Background(), &workflow.Workflow{Name: "wf"})
		assertError(t, err, nil)

		_, _, got := mgr.AssignToCodeset(context.Background(), wf.Name, "unknownProj", "unknownCs")
		assertError(t, got, errCodesetNotFound)

		gotAss := workflowStore.GetAssignments(context.TODO(), nil)
		wantAss := map[string][]*domain.AssignedCodeset{}
		if d := cmp.Diff(wantAss, gotAss); d != "" {
			t.Errorf("Unexpected Assignment: %s", diff.PrintWantGot(d))
		}

		_, err = workflowBackend.GetWorkflowListener(context.TODO(), wf.Name)
		assertStrings(t, err.Error(), "listener not found")
	})
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

func assertStrings(t testing.TB, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("got %q want %q", got, want)
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

	if _, exists := b.workflows[workflowName]; !exists {
		return fmt.Errorf("workflow not found")
	}

	runs := b.workflows[workflowName].runs
	name := fmt.Sprintf("%s-run%d", workflowName, len(runs))
	codesetInputName := "codeset-name"
	codesetInputType := "codeset"
	codesetInputValue := fmt.Sprintf("%s/%s", codeset.Project, codeset.Name)
	stringInputName := "predictor"
	stringInputType := "string"
	stringInputValue := "sklearn"

	run := &workflow.WorkflowRun{
		Name:        &name,
		WorkflowRef: &workflowName,
		Inputs: []*workflow.WorkflowRunInput{
			{Input: &workflow.WorkflowInput{Name: &codesetInputName, Type: &codesetInputType}, Value: &codesetInputValue},
			{Input: &workflow.WorkflowInput{Name: &stringInputName, Type: &stringInputType}, Value: &stringInputValue}},
		Status: &workflowRunStatuses[len(runs)%len(workflowRunStatuses)]}

	b.workflows[workflowName].runs = append(b.workflows[workflowName].runs, run)
	return nil
}

func (b *fakeWorkflowBackend) ListWorkflowRuns(ctx context.Context, wf *workflow.Workflow, filter *domain.WorkflowRunFilter) ([]*workflow.WorkflowRun, error) {
	b.t.Helper()

	return nil, nil
}

func (b *fakeWorkflowBackend) CreateWorkflowListener(ctx context.Context, workflowName string, timeout time.Duration) (*domain.WorkflowListener, error) {
	b.t.Helper()

	listener := b.workflows[workflowName].listener
	if listener == nil {
		listener = &domain.WorkflowListener{Name: workflowName, Available: true, URL: fmt.Sprintf("http://%s.listener.test", workflowName),
			DashboardURL: fmt.Sprintf("http://dashboard.test/%s", workflowName)}
		b.workflows[workflowName].listener = listener
	}
	return listener, nil
}

func (b *fakeWorkflowBackend) DeleteWorkflowListener(ctx context.Context, workflowName string) error {
	b.t.Helper()

	return nil
}

func (b *fakeWorkflowBackend) GetWorkflowListener(ctx context.Context, workflowName string) (*domain.WorkflowListener, error) {
	b.t.Helper()

	if wf, exists := b.workflows[workflowName]; exists {
		if wf.listener != nil {
			return wf.listener, nil
		}
	}
	return nil, fmt.Errorf("listener not found")
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

	id := rand.Int63()
	fcs.store[codesetID{c.Name, c.Project}].webhooks[id] = url
	return &id, nil
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

	if sc, exists := fcs.store[codesetID{name, project}]; exists {
		return sc.codeset, nil
	}
	return nil, errCodesetNotFound
}

func (fcs *fakeCodesetStore) GetAll(ctx context.Context, project, label *string) (res []*domain.Codeset, err error) {
	fcs.t.Helper()

	for _, c := range fcs.store {
		res = append(res, c.codeset)
	}
	return res, nil
}
