package badger

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/pipeline/test/diff"
	"github.com/timshannon/badgerhold/v3"

	"github.com/fuseml/fuseml-core/pkg/domain"
)

func TestGetWorkflow(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		wf := domain.Workflow{Name: "test"}

		_, err := store.AddWorkflow(context.TODO(), &wf)
		assertNoError(t, err)

		got, err := store.GetWorkflow(context.TODO(), "test")
		assertNoError(t, err)
		if d := cmp.Diff(&wf, got); d != "" {
			t.Errorf("Unexpected Workflow: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		_, got := store.GetWorkflow(context.TODO(), "test")
		assertError(t, got, domain.ErrWorkflowNotFound)
	})
}

func TestGetWorkflows(t *testing.T) {
	t.Run("all", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		// empty
		want := []*domain.Workflow{}
		got := store.GetWorkflows(context.TODO(), nil)
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Workflow: %s", diff.PrintWantGot(d))
		}

		// add workflows
		for i := 0; i < 2; i++ {
			wf := domain.Workflow{Name: fmt.Sprintf("test-%d", i)}
			_, err := store.AddWorkflow(context.TODO(), &wf)
			assertNoError(t, err)

			want = append(want, &wf)
		}

		// should return all
		got = store.GetWorkflows(context.TODO(), nil)
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Workflow: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("by name", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		// empty
		want := []*domain.Workflow{}
		name := "test"
		got := store.GetWorkflows(context.TODO(), &name)
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Workflows: %s", diff.PrintWantGot(d))
		}

		// add workflows
		for i := 0; i < 2; i++ {
			wf := domain.Workflow{Name: fmt.Sprintf("test-%d", i)}
			_, err := store.AddWorkflow(context.TODO(), &wf)
			assertNoError(t, err)

			want = append(want, &wf)
		}

		// should return one workflow
		got = store.GetWorkflows(context.TODO(), &want[0].Name)
		if d := cmp.Diff(want[0], got[0]); d != "" {
			t.Errorf("Unexpected Workflows: %s", diff.PrintWantGot(d))
		}

		// should return no workflows
		name = "no-wf"
		got = store.GetWorkflows(context.TODO(), &name)
		if d := cmp.Diff([]*domain.Workflow{}, got); d != "" {
			t.Errorf("Unexpected Workflows: %s", diff.PrintWantGot(d))
		}
	})
}

func TestAddWorkflow(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		wf := domain.Workflow{Name: "test"}

		_, err := store.AddWorkflow(context.TODO(), &wf)
		assertNoError(t, err)

		got, err := store.GetWorkflow(context.TODO(), "test")
		assertNoError(t, err)
		if d := cmp.Diff(&wf, got); d != "" {
			t.Errorf("Unexpected Workflow: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("existing", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		wf := domain.Workflow{Name: "test"}

		_, err := store.AddWorkflow(context.TODO(), &wf)
		assertNoError(t, err)

		_, err = store.AddWorkflow(context.TODO(), &wf)
		assertError(t, err, domain.ErrWorkflowExists)
	})
}

func TestDeleteWorkflow(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		wf := domain.Workflow{Name: "test"}

		_, err := store.AddWorkflow(context.TODO(), &wf)
		assertNoError(t, err)

		err = store.DeleteWorkflow(context.TODO(), "test")
		assertNoError(t, err)

		_, err = store.GetWorkflow(context.TODO(), "test")
		assertError(t, err, domain.ErrWorkflowNotFound)
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		err := store.DeleteWorkflow(context.TODO(), "test")
		assertNoError(t, err)
	})

	t.Run("assigned", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		wfName := "test"
		wf := domain.Workflow{Name: wfName}

		_, err := store.AddWorkflow(context.TODO(), &wf)
		assertNoError(t, err)

		cs := &domain.Codeset{
			Name: "test-cs",
		}

		ca := domain.CodesetAssignment{Codeset: cs}
		store.AddCodesetAssignment(context.TODO(), wfName, &ca)

		err = store.DeleteWorkflow(context.TODO(), wfName)
		assertError(t, err, domain.ErrCannotDeleteAssignedWorkflow)
	})
}

func TestAddCodesetAssignment(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		wfName := "test-wf"
		wf := domain.Workflow{Name: wfName}

		_, err := store.AddWorkflow(context.TODO(), &wf)
		assertNoError(t, err)

		cs := &domain.Codeset{
			Name: "test-cs",
		}

		ca := domain.CodesetAssignment{Codeset: cs}
		got := store.AddCodesetAssignment(context.TODO(), wfName, &ca)

		want := []*domain.CodesetAssignment{&ca}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Assignments: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("existing", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		wfName := "test-wf"
		wf := domain.Workflow{Name: wfName}

		_, err := store.AddWorkflow(context.TODO(), &wf)
		assertNoError(t, err)

		cs := domain.Codeset{
			Name: "test-cs",
		}

		ca := domain.CodesetAssignment{Codeset: &cs}
		got := store.AddCodesetAssignment(context.TODO(), wfName, &ca)

		want := []*domain.CodesetAssignment{&ca}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Assignments: %s", diff.PrintWantGot(d))
		}

		got = store.AddCodesetAssignment(context.TODO(), wfName, &ca)
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Assignments: %s", diff.PrintWantGot(d))
		}
	})
}

func TestGetAssignedCodesets(t *testing.T) {
	t.Run("no assignments", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		wfName := "test-wf"
		wf := domain.Workflow{Name: wfName}

		_, err := store.AddWorkflow(context.TODO(), &wf)
		assertNoError(t, err)

		got := store.GetAssignedCodesets(context.TODO(), wfName)
		want := []*domain.CodesetAssignment{}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Assignments: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("with assignments", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		wfName := "test-wf"
		wf := domain.Workflow{Name: wfName}

		_, err := store.AddWorkflow(context.TODO(), &wf)
		assertNoError(t, err)

		cs := domain.Codeset{
			Name: "test-cs",
		}

		ac := domain.CodesetAssignment{Codeset: &cs}
		store.AddCodesetAssignment(context.TODO(), wfName, &ac)

		got := store.GetAssignedCodesets(context.TODO(), wfName)
		want := []*domain.CodesetAssignment{&ac}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Assignments: %s", diff.PrintWantGot(d))
		}
	})
}

func TestGetAssignments(t *testing.T) {
	t.Run("no assignments", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		wfName := "test-wf"
		wf := domain.Workflow{Name: wfName}

		_, err := store.AddWorkflow(context.TODO(), &wf)
		assertNoError(t, err)

		got := store.GetAssignments(context.TODO(), &wfName)
		want := map[string][]*domain.CodesetAssignment{}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Assignments: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("with assignments", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		wfName := "test-wf"
		wf := domain.Workflow{Name: wfName}

		_, err := store.AddWorkflow(context.TODO(), &wf)
		assertNoError(t, err)

		cs := domain.Codeset{
			Name: "test-cs",
		}

		ca := domain.CodesetAssignment{Codeset: &cs}
		store.AddCodesetAssignment(context.TODO(), wfName, &ca)

		got := store.GetAssignments(context.TODO(), &wfName)
		want := map[string][]*domain.CodesetAssignment{wfName: {&ca}}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Assignments: %s", diff.PrintWantGot(d))
		}
	})
}

func TestDeleteCodesetAssignment(t *testing.T) {
	t.Run("no assignments", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		wfName := "test-wf"
		wf := domain.Workflow{Name: wfName}

		_, err := store.AddWorkflow(context.TODO(), &wf)
		assertNoError(t, err)

		cs := &domain.Codeset{
			Name: "test-cs",
		}

		got := store.DeleteCodesetAssignment(context.TODO(), wfName, cs)
		want := []*domain.CodesetAssignment{}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Assignments: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("with assignments", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		wfName := "test-wf"
		wf := domain.Workflow{Name: wfName}

		_, err := store.AddWorkflow(context.TODO(), &wf)
		assertNoError(t, err)

		cs1 := domain.Codeset{
			Name: "test-cs",
		}

		cs2 := domain.Codeset{
			Name: "test-cs2",
		}

		ac := domain.CodesetAssignment{Codeset: &cs1}
		store.AddCodesetAssignment(context.TODO(), wfName, &ac)

		ca := domain.CodesetAssignment{Codeset: &cs2}
		store.AddCodesetAssignment(context.TODO(), wfName, &ca)

		got := store.DeleteCodesetAssignment(context.TODO(), wfName, &cs1)
		want := []*domain.CodesetAssignment{{Codeset: &cs2}}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Assignments: %s", diff.PrintWantGot(d))
		}

		got = store.DeleteCodesetAssignment(context.TODO(), wfName, &cs2)
		want = []*domain.CodesetAssignment{}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Assignments: %s", diff.PrintWantGot(d))
		}
	})
}

func TestGetAssignedCodeset(t *testing.T) {
	t.Run("no workflow", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		cs := domain.Codeset{
			Name: "test-cs",
		}
		_, got := store.GetAssignedCodeset(context.TODO(), "no-wf", &cs)
		assertError(t, got, domain.ErrWorkflowNotFound)
	})

	t.Run("with workflow, no assignments", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		wfName := "test-wf"
		wf := domain.Workflow{Name: wfName}

		_, err := store.AddWorkflow(context.TODO(), &wf)
		assertNoError(t, err)

		cs := domain.Codeset{
			Name: "test-cs",
		}

		_, err = store.GetAssignedCodeset(context.TODO(), wfName, &cs)
		assertError(t, err, domain.ErrWorkflowNotAssignedToCodeset)
	})

	t.Run("with assignment", func(t *testing.T) {
		store, done := newWorkflowStore(t)
		defer done()

		wfName := "test-wf"
		wf := domain.Workflow{Name: wfName}

		_, err := store.AddWorkflow(context.TODO(), &wf)
		assertNoError(t, err)

		cs := domain.Codeset{
			Name: "test-cs",
		}

		ca := domain.CodesetAssignment{Codeset: &cs}
		store.AddCodesetAssignment(context.TODO(), wfName, &ca)

		got, err := store.GetAssignedCodeset(context.TODO(), wfName, &cs)
		assertNoError(t, err)
		want := &ca
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Assignments: %s", diff.PrintWantGot(d))
		}

	})
}

func assertError(t testing.TB, got, want error) {
	t.Helper()

	if got != want {
		t.Errorf("got error %q want %q", got, want)
	}
}

func assertNoError(t testing.TB, got error) {
	t.Helper()

	if got != nil {
		t.Errorf("got error %q wants no error", got)
	}
}

func newWorkflowStore(t *testing.T) (*WorkflowStore, func()) {
	t.Helper()

	dir := tmpDir(t)
	opt := badgerhold.DefaultOptions
	opt.Logger = nil
	opt.Dir = dir
	opt.ValueDir = dir

	store, err := badgerhold.Open(opt)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}

	workflowStore := NewWorkflowStore(store)

	return workflowStore, func() {
		store.Close()
		os.RemoveAll(dir)
	}
}

// tmpDir returns a temporary dir path.
func tmpDir(t *testing.T) string {
	t.Helper()

	name, err := ioutil.TempDir("", "fuseml-storage-test")
	if err != nil {
		t.Errorf("failed to create temp dir: %v", err)
	}
	return name
}
