package badger

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/pipeline/test/diff"
	"github.com/timshannon/badgerhold/v3"

	"github.com/fuseml/fuseml-core/pkg/domain"
)

func TestApplicationAdd(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		store, done := newApplicationStore(t)
		defer done()

		app := domain.Application{
			Name: "test-app",
		}

		got, err := store.Add(context.TODO(), &app)
		assertNoError(t, err)

		if d := cmp.Diff(&app, got); d != "" {
			t.Errorf("Unexpected Application: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("existing", func(t *testing.T) {
		store, done := newApplicationStore(t)
		defer done()

		app := domain.Application{
			Name: "test-app",
			Type: "testType1",
		}

		store.Add(context.TODO(), &app)

		app.Type = "testType2"
		got, err := store.Add(context.TODO(), &app)
		assertNoError(t, err)

		if d := cmp.Diff(&app, got); d != "" {
			t.Errorf("Unexpected Application: %s", diff.PrintWantGot(d))
		}
	})
}

func TestApplicationFind(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		store, done := newApplicationStore(t)
		defer done()

		app := domain.Application{
			Name: "test-app",
		}

		store.Add(context.TODO(), &app)

		got := store.Find(context.TODO(), app.Name)
		if d := cmp.Diff(&app, got); d != "" {
			t.Errorf("Unexpected Application: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("non-existing", func(t *testing.T) {
		store, done := newApplicationStore(t)
		defer done()

		got := store.Find(context.TODO(), "non-existing")
		if got != nil {
			t.Errorf("Expected nil, got %v", got)
		}
	})
}

func TestApplicationGetAll(t *testing.T) {
	t.Run("all", func(t *testing.T) {
		store, done := newApplicationStore(t)
		defer done()

		app1 := domain.Application{
			Name: "test-app1",
		}
		app2 := domain.Application{
			Name: "test-app2",
		}
		app3 := domain.Application{
			Name: "test-app3",
		}

		store.Add(context.TODO(), &app1)
		store.Add(context.TODO(), &app2)
		store.Add(context.TODO(), &app3)

		got, err := store.GetAll(context.TODO(), nil, nil)
		assertNoError(t, err)
		want := []*domain.Application{&app1, &app2, &app3}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Applications: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("by type", func(t *testing.T) {
		store, done := newApplicationStore(t)
		defer done()

		app1 := domain.Application{
			Name: "test-app1",
			Type: "testType1",
		}

		app2 := domain.Application{
			Name: "test-app2",
			Type: "testType2",
		}

		store.Add(context.TODO(), &app1)
		store.Add(context.TODO(), &app2)

		got, err := store.GetAll(context.TODO(), &app1.Type, nil)
		assertNoError(t, err)

		want := []*domain.Application{&app1}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Applications: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("by workflow", func(t *testing.T) {
		store, done := newApplicationStore(t)
		defer done()

		app1 := domain.Application{
			Name:     "test-app1",
			Type:     "testType1",
			Workflow: "wf-1",
		}

		app2 := domain.Application{
			Name:     "test-app2",
			Type:     "testType2",
			Workflow: "wf-2",
		}

		store.Add(context.TODO(), &app1)
		store.Add(context.TODO(), &app2)

		got, err := store.GetAll(context.TODO(), nil, &app1.Workflow)
		assertNoError(t, err)

		want := []*domain.Application{&app1}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Applications: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("by type and workflow", func(t *testing.T) {
		store, done := newApplicationStore(t)
		defer done()

		app1 := domain.Application{
			Name:     "test-app1",
			Type:     "testType1",
			Workflow: "wf-1",
		}

		app2 := domain.Application{
			Name:     "test-app2",
			Type:     "testType2",
			Workflow: "wf-2",
		}

		app3 := domain.Application{
			Name:     "test-app3",
			Type:     "testType1",
			Workflow: "wf-2",
		}

		app4 := domain.Application{
			Name:     "test-app4",
			Type:     "testType2",
			Workflow: "wf-1",
		}

		store.Add(context.TODO(), &app1)
		store.Add(context.TODO(), &app2)
		store.Add(context.TODO(), &app3)
		store.Add(context.TODO(), &app4)

		got, err := store.GetAll(context.TODO(), &app1.Type, &app1.Workflow)
		assertNoError(t, err)

		want := []*domain.Application{&app1}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected Applications: %s", diff.PrintWantGot(d))
		}
	})
}

func TestApplicationDelete(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		store, done := newApplicationStore(t)
		defer done()

		app := domain.Application{
			Name: "test-app",
		}

		store.Add(context.TODO(), &app)

		store.Delete(context.TODO(), app.Name)

		got := store.Find(context.TODO(), app.Name)
		if got != nil {
			t.Errorf("Expected nil, got %v", got)
		}
	})

	t.Run("non-existing", func(t *testing.T) {
		store, done := newApplicationStore(t)
		defer done()

		store.Delete(context.TODO(), "non-existing")

		got := store.Find(context.TODO(), "non-existing")
		if got != nil {
			t.Errorf("Expected nil, got %v", got)
		}
	})
}

func newApplicationStore(t *testing.T) (*ApplicationStore, func()) {
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

	applicationStore := NewApplicationStore(store)

	return applicationStore, func() {
		store.Close()
		os.RemoveAll(dir)
	}
}
