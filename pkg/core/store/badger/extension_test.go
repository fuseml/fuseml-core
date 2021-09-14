package badger

import (
	"context"
	"os"
	"testing"

	"github.com/fuseml/fuseml-core/pkg/domain"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/tektoncd/pipeline/test/diff"
	"github.com/timshannon/badgerhold/v3"
)

func TestAddExtension(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{})
		assertNoError(t, err)

		got, err := store.GetExtension(ctx, ext.ID)
		assertNoError(t, err)

		if d := cmp.Diff(ext, got); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("existing", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{})
		assertNoError(t, err)

		_, err = store.AddExtension(ctx, ext)
		assertErrorMessage(t, domain.NewErrExtensionExists(ext.ID), err)
	})
}

func TestGetExtension(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{})
		assertNoError(t, err)

		got, err := store.GetExtension(ctx, ext.ID)
		assertNoError(t, err)

		if d := cmp.Diff(ext, got); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		_, err := store.GetExtension(ctx, "not found")
		assertErrorMessage(t, domain.NewErrExtensionNotFound("not found"), err)
	})
}

func TestGetExtensions(t *testing.T) {
	t.Run("all", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		exts := []*domain.Extension{
			{ID: "1"},
			{ID: "2"},
			{ID: "3"},
		}

		for _, ext := range exts {
			_, err := store.AddExtension(ctx, ext)
			assertNoError(t, err)
		}

		got := store.ListExtensions(ctx, nil)
		sortExtensionSlices := cmpopts.SortSlices(func(x, y *domain.Extension) bool { return x.ID < y.ID })
		if d := cmp.Diff(exts, got, sortExtensionSlices); d != "" {
			t.Errorf("Unexpected Extensions: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("with query", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		exts := []*domain.Extension{
			{ID: "1", Product: "p1"},
			{ID: "2", Product: "p1"},
			{ID: "3", Product: "p2"},
		}

		for _, ext := range exts {
			_, err := store.AddExtension(ctx, ext)
			assertNoError(t, err)
		}

		// by ID
		got := store.ListExtensions(ctx, &domain.ExtensionQuery{ExtensionID: "3"})
		if d := cmp.Diff(exts[2:], got); d != "" {
			t.Errorf("Unexpected Extensions: %s", diff.PrintWantGot(d))
		}

		// by product
		got = store.ListExtensions(ctx, &domain.ExtensionQuery{Product: "p1"})
		if d := cmp.Diff(exts[:2], got); d != "" {
			t.Errorf("Unexpected Extensions: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("empty", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		got := store.ListExtensions(ctx, nil)
		if d := cmp.Diff([]*domain.Extension{}, got); d != "" {
			t.Errorf("Unexpected Extensions: %s", diff.PrintWantGot(d))
		}

		got = store.ListExtensions(ctx, &domain.ExtensionQuery{Product: "p1"})
		if d := cmp.Diff([]*domain.Extension{}, got); d != "" {
			t.Errorf("Unexpected Extensions: %s", diff.PrintWantGot(d))
		}
	})
}

func TestUpdateExtension(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{})
		assertNoError(t, err)

		newExt := &domain.Extension{
			ID:       ext.ID,
			Product:  "p1",
			Version:  "v1",
			Services: map[string]*domain.ExtensionService{"test": {ID: "test"}},
		}
		err = store.UpdateExtension(ctx, newExt)
		assertNoError(t, err)

		got, err := store.GetExtension(ctx, ext.ID)
		assertNoError(t, err)

		if d := cmp.Diff(newExt, got); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		err := store.UpdateExtension(ctx, &domain.Extension{})
		assertErrorMessage(t, domain.NewErrExtensionNotFound(""), err)

		err = store.UpdateExtension(ctx, &domain.Extension{ID: "not found"})
		assertErrorMessage(t, domain.NewErrExtensionNotFound("not found"), err)
	})
}

func TestDeleteExtension(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{})
		assertNoError(t, err)

		err = store.DeleteExtension(ctx, ext.ID)
		assertNoError(t, err)

		_, err = store.GetExtension(ctx, ext.ID)
		assertErrorMessage(t, domain.NewErrExtensionNotFound(ext.ID), err)
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		err := store.DeleteExtension(ctx, "not found")
		assertErrorMessage(t, domain.NewErrExtensionNotFound("not found"), err)
	})
}

func TestAddExtensionService(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{})
		assertNoError(t, err)

		svc, err := store.AddExtensionService(ctx, ext.ID, &domain.ExtensionService{})
		assertNoError(t, err)

		got, err := store.GetExtensionService(ctx, ext.ID, svc.ID)
		assertNoError(t, err)

		if d := cmp.Diff(svc, got); d != "" {
			t.Errorf("Unexpected ExtensionService: %s", diff.PrintWantGot(d))
		}

		svc, err = store.AddExtensionService(ctx, ext.ID, &domain.ExtensionService{ID: "test"})
		assertNoError(t, err)

		got, err = store.GetExtensionService(ctx, ext.ID, svc.ID)
		assertNoError(t, err)

		if d := cmp.Diff(svc, got); d != "" {
			t.Errorf("Unexpected ExtensionService: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("existing", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{})
		assertNoError(t, err)

		_, err = store.AddExtensionService(ctx, ext.ID, &domain.ExtensionService{ID: "test"})
		assertNoError(t, err)

		_, err = store.AddExtensionService(ctx, ext.ID, &domain.ExtensionService{ID: "test"})
		assertErrorMessage(t, domain.NewErrExtensionServiceExists(ext.ID, "test"), err)
	})
}
func TestGetExtensionService(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{})
		assertNoError(t, err)

		svc, err := store.AddExtensionService(ctx, ext.ID, &domain.ExtensionService{})
		assertNoError(t, err)

		got, err := store.GetExtensionService(ctx, ext.ID, svc.ID)
		assertNoError(t, err)

		if d := cmp.Diff(svc, got); d != "" {
			t.Errorf("Unexpected ExtensionService: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{})
		assertNoError(t, err)

		_, err = store.GetExtensionService(ctx, ext.ID, "not found")
		assertErrorMessage(t, domain.NewErrExtensionServiceNotFound(ext.ID, "not found"), err)
	})
}

func TestGetExtensionServices(t *testing.T) {
	t.Run("all", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{})
		assertNoError(t, err)

		svcs := []*domain.ExtensionService{
			{ID: "test1"},
			{ID: "test2"},
		}
		for _, svc := range svcs {
			_, err = store.AddExtensionService(ctx, ext.ID, svc)
			assertNoError(t, err)
		}

		got, err := store.ListExtensionServices(ctx, ext.ID)
		assertNoError(t, err)

		sortServiceSlices := cmpopts.SortSlices(func(x, y *domain.ExtensionService) bool { return x.ID < y.ID })
		if d := cmp.Diff(svcs, got, sortServiceSlices); d != "" {
			t.Errorf("Unexpected ExtensionServices: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("empty", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{})
		assertNoError(t, err)

		got, err := store.ListExtensionServices(ctx, ext.ID)
		assertNoError(t, err)

		if d := cmp.Diff([]*domain.ExtensionService{}, got); d != "" {
			t.Errorf("Unexpected ExtensionServices: %s", diff.PrintWantGot(d))
		}
	})
}

func TestUpdateExtensionService(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{})
		assertNoError(t, err)

		svc, err := store.AddExtensionService(ctx, ext.ID, &domain.ExtensionService{Resource: "test"})
		assertNoError(t, err)

		newSvc := &domain.ExtensionService{
			ID:       svc.ID,
			Resource: "test-updated"}
		err = store.UpdateExtensionService(ctx, ext.ID, newSvc)
		assertNoError(t, err)

		got, err := store.GetExtensionService(ctx, ext.ID, svc.ID)
		assertNoError(t, err)

		if d := cmp.Diff(newSvc, got); d != "" {
			t.Errorf("Unexpected ExtensionService: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{})
		assertNoError(t, err)

		err = store.UpdateExtensionService(ctx, ext.ID, &domain.ExtensionService{ID: "not found"})
		assertErrorMessage(t, domain.NewErrExtensionServiceNotFound(ext.ID, "not found"), err)
	})
}

func TestDeleteExtensionService(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{})
		assertNoError(t, err)

		svc, err := store.AddExtensionService(ctx, ext.ID, &domain.ExtensionService{})
		assertNoError(t, err)

		err = store.DeleteExtensionService(ctx, ext.ID, svc.ID)
		assertNoError(t, err)

		_, err = store.GetExtensionService(ctx, ext.ID, svc.ID)
		assertErrorMessage(t, domain.NewErrExtensionServiceNotFound(ext.ID, svc.ID), err)
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{})
		assertNoError(t, err)

		err = store.DeleteExtensionService(ctx, ext.ID, "not found")
		assertErrorMessage(t, domain.NewErrExtensionServiceNotFound(ext.ID, "not found"), err)
	})
}

func TestAddExtensionServiceEndpoint(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		endpoint, err := store.AddExtensionServiceEndpoint(ctx, ext.ID, "test-svc", &domain.ExtensionServiceEndpoint{URL: "http://test"})
		assertNoError(t, err)

		got, err := store.GetExtensionServiceEndpoint(ctx, ext.ID, "test-svc", endpoint.URL)
		assertNoError(t, err)

		if d := cmp.Diff(endpoint, got); d != "" {
			t.Errorf("Unexpected ExtensionServiceEndpoint: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("existing", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		endpoint, err := store.AddExtensionServiceEndpoint(ctx, ext.ID, "test-svc", &domain.ExtensionServiceEndpoint{URL: "http://test"})
		assertNoError(t, err)

		_, err = store.AddExtensionServiceEndpoint(ctx, ext.ID, "test-svc", &domain.ExtensionServiceEndpoint{URL: "http://test"})
		assertErrorMessage(t, domain.NewErrExtensionServiceEndpointExists("", "test-svc", endpoint.URL), err)
	})
}

func TestGetExtensionServiceEndpoint(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		endpoint, err := store.AddExtensionServiceEndpoint(ctx, ext.ID, "test-svc", &domain.ExtensionServiceEndpoint{URL: "http://test"})
		assertNoError(t, err)

		got, err := store.GetExtensionServiceEndpoint(ctx, ext.ID, "test-svc", endpoint.URL)
		assertNoError(t, err)

		if d := cmp.Diff(endpoint, got); d != "" {
			t.Errorf("Unexpected ExtensionServiceEndpoint: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		_, err = store.GetExtensionServiceEndpoint(ctx, ext.ID, "test-svc", "not found")
		assertErrorMessage(t, domain.NewErrExtensionServiceEndpointNotFound("", "test-svc", "not found"), err)
	})
}

func TestGetExtensionServiceEndpoints(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		endpoint, err := store.AddExtensionServiceEndpoint(ctx, ext.ID, "test-svc", &domain.ExtensionServiceEndpoint{URL: "http://test"})
		assertNoError(t, err)

		got, err := store.ListExtensionServiceEndpoints(ctx, ext.ID, "test-svc")
		assertNoError(t, err)

		if d := cmp.Diff([]*domain.ExtensionServiceEndpoint{endpoint}, got); d != "" {
			t.Errorf("Unexpected ExtensionServiceEndpoint: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		got, err := store.ListExtensionServiceEndpoints(ctx, ext.ID, "test-svc")
		assertNoError(t, err)

		if d := cmp.Diff([]*domain.ExtensionServiceEndpoint{}, got); d != "" {
			t.Errorf("Unexpected ExtensionServiceEndpoint: %s", diff.PrintWantGot(d))
		}
	})
}

func TestUpdateExtensionServiceEndpoint(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		endpoint, err := store.AddExtensionServiceEndpoint(ctx, ext.ID, "test-svc", &domain.ExtensionServiceEndpoint{URL: "http://test"})
		assertNoError(t, err)

		newEndpoint := &domain.ExtensionServiceEndpoint{URL: "http://test", Type: "test-updated"}

		err = store.UpdateExtensionServiceEndpoint(ctx, ext.ID, "test-svc", newEndpoint)
		assertNoError(t, err)

		got, err := store.GetExtensionServiceEndpoint(ctx, ext.ID, "test-svc", endpoint.URL)
		assertNoError(t, err)

		if d := cmp.Diff(newEndpoint, got); d != "" {
			t.Errorf("Unexpected ExtensionServiceEndpoint: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		err = store.UpdateExtensionServiceEndpoint(ctx, ext.ID, "test-svc", &domain.ExtensionServiceEndpoint{URL: "http://test"})
		assertErrorMessage(t, domain.NewErrExtensionServiceEndpointNotFound("", "test-svc", "http://test"), err)
	})
}

func TestDeleteExtensionServiceEndpoint(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		endpoint, err := store.AddExtensionServiceEndpoint(ctx, ext.ID, "test-svc", &domain.ExtensionServiceEndpoint{URL: "http://test"})
		assertNoError(t, err)

		err = store.DeleteExtensionServiceEndpoint(ctx, ext.ID, "test-svc", endpoint.URL)
		assertNoError(t, err)

		_, err = store.GetExtensionServiceEndpoint(ctx, ext.ID, "test-svc", endpoint.URL)
		assertErrorMessage(t, domain.NewErrExtensionServiceEndpointNotFound("", "test-svc", endpoint.URL), err)
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		err = store.DeleteExtensionServiceEndpoint(ctx, ext.ID, "test-svc", "http://test")
		assertErrorMessage(t, domain.NewErrExtensionServiceEndpointNotFound("", "test-svc", "http://test"), err)
	})
}

func TestAddExtensionServiceCredential(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		cred, err := store.AddExtensionServiceCredentials(ctx, ext.ID, "test-svc", &domain.ExtensionServiceCredentials{})
		assertNoError(t, err)

		got, err := store.GetExtensionServiceCredentials(ctx, ext.ID, "test-svc", cred.ID)
		assertNoError(t, err)

		if d := cmp.Diff(cred, got); d != "" {
			t.Errorf("Unexpected ExtensionServiceCredential: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("existing", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		cred, err := store.AddExtensionServiceCredentials(ctx, ext.ID, "test-svc", &domain.ExtensionServiceCredentials{})
		assertNoError(t, err)

		_, err = store.AddExtensionServiceCredentials(ctx, ext.ID, "test-svc", &domain.ExtensionServiceCredentials{ID: cred.ID})
		assertErrorMessage(t, domain.NewErrExtensionServiceCredentialsExists("", "test-svc", cred.ID), err)

	})
}

func TestGetExtensionServiceCredential(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		cred, err := store.AddExtensionServiceCredentials(ctx, ext.ID, "test-svc", &domain.ExtensionServiceCredentials{})
		assertNoError(t, err)

		got, err := store.GetExtensionServiceCredentials(ctx, ext.ID, "test-svc", cred.ID)
		assertNoError(t, err)

		if d := cmp.Diff(cred, got); d != "" {
			t.Errorf("Unexpected ExtensionServiceCredential: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		_, err = store.GetExtensionServiceCredentials(ctx, ext.ID, "test-svc", "test")
		assertErrorMessage(t, domain.NewErrExtensionServiceCredentialsNotFound("", "test-svc", "test"), err)
	})
}

func TestGetExtensionServiceCredentials(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		cred, err := store.AddExtensionServiceCredentials(ctx, ext.ID, "test-svc", &domain.ExtensionServiceCredentials{})
		assertNoError(t, err)

		got, err := store.ListExtensionServiceCredentials(ctx, ext.ID, "test-svc")
		assertNoError(t, err)

		if d := cmp.Diff([]*domain.ExtensionServiceCredentials{cred}, got); d != "" {
			t.Errorf("Unexpected ExtensionServiceCredential: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		got, err := store.ListExtensionServiceCredentials(ctx, ext.ID, "test-svc")
		assertNoError(t, err)

		if d := cmp.Diff([]*domain.ExtensionServiceCredentials{}, got); d != "" {
			t.Errorf("Unexpected ExtensionServiceCredential: %s", diff.PrintWantGot(d))
		}
	})
}

func TestUpdateExtensionServiceCredential(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		cred, err := store.AddExtensionServiceCredentials(ctx, ext.ID, "test-svc", &domain.ExtensionServiceCredentials{})
		assertNoError(t, err)

		newCred := &domain.ExtensionServiceCredentials{
			ID:    cred.ID,
			Scope: domain.ECSGlobal,
		}

		err = store.UpdateExtensionServiceCredentials(ctx, ext.ID, "test-svc", newCred)
		assertNoError(t, err)

		got, err := store.GetExtensionServiceCredentials(ctx, ext.ID, "test-svc", cred.ID)
		assertNoError(t, err)

		if d := cmp.Diff(newCred, got); d != "" {
			t.Errorf("Unexpected ExtensionServiceCredential: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		err = store.UpdateExtensionServiceCredentials(ctx, ext.ID, "test-svc", &domain.ExtensionServiceCredentials{ID: "not found"})
		assertErrorMessage(t, domain.NewErrExtensionServiceCredentialsNotFound("", "test-svc", "not found"), err)
	})
}

func TestDeleteExtensionServiceCredential(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		cred, err := store.AddExtensionServiceCredentials(ctx, ext.ID, "test-svc", &domain.ExtensionServiceCredentials{})
		assertNoError(t, err)

		err = store.DeleteExtensionServiceCredentials(ctx, ext.ID, "test-svc", cred.ID)
		assertNoError(t, err)

		_, err = store.GetExtensionServiceCredentials(ctx, ext.ID, "test-svc", cred.ID)
		assertErrorMessage(t, domain.NewErrExtensionServiceCredentialsNotFound("", "test-svc", cred.ID), err)
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		err = store.DeleteExtensionServiceCredentials(ctx, ext.ID, "test-svc", "not found")
		assertErrorMessage(t, domain.NewErrExtensionServiceCredentialsNotFound("", "test-svc", "not found"), err)
	})
}

func TestGetExtensionAccessDescriptors(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		ext, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{
				"test-svc1": {
					ID:        "test-svc1",
					Endpoints: map[string]*domain.ExtensionServiceEndpoint{"http://test1": {URL: "http://test1"}},
				},
				"test-svc2": {
					ID:        "test-svc2",
					Endpoints: map[string]*domain.ExtensionServiceEndpoint{"http://test2": {URL: "http://test2"}},
				}}})
		assertNoError(t, err)

		got, err := store.GetExtensionAccessDescriptors(ctx, &domain.ExtensionQuery{ServiceID: "test-svc1"})
		assertNoError(t, err)

		endpoint1 := domain.ExtensionServiceEndpoint{
			URL:     "http://test1",
			Created: ext.Services["test-svc1"].Endpoints["http://test1"].Created,
			Updated: ext.Services["test-svc1"].Endpoints["http://test1"].Updated,
		}

		svc1 := domain.ExtensionService{
			ID:      "test-svc1",
			Created: ext.Services["test-svc1"].Created,
			Updated: ext.Services["test-svc1"].Updated,
			Endpoints: map[string]*domain.ExtensionServiceEndpoint{
				endpoint1.URL: &endpoint1,
			},
			Credentials: map[string]*domain.ExtensionServiceCredentials{},
		}

		want := []*domain.ExtensionAccessDescriptor{{
			Extension: domain.Extension{
				ID:      ext.ID,
				Created: ext.Created,
				Updated: ext.Updated,
				Services: map[string]*domain.ExtensionService{
					svc1.ID: &svc1,
				},
			},
			Service:  svc1,
			Endpoint: endpoint1,
		}}

		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected ExtensionAccessDescriptor: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("not found", func(t *testing.T) {
		store, done := newExtensionStore(t)
		defer done()
		ctx := context.Background()

		_, err := store.AddExtension(ctx, &domain.Extension{
			Services: map[string]*domain.ExtensionService{"test-svc": {ID: "test-svc"}}})
		assertNoError(t, err)

		got, err := store.GetExtensionAccessDescriptors(ctx, &domain.ExtensionQuery{ServiceID: "not found"})
		assertNoError(t, err)

		if d := cmp.Diff([]*domain.ExtensionAccessDescriptor{}, got); d != "" {
			t.Errorf("Unexpected ExtensionAccessDescriptor: %s", diff.PrintWantGot(d))
		}
	})
}

func newExtensionStore(t *testing.T) (*ExtensionStore, func()) {
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

	workflowStore := NewExtensionStore(store)

	return workflowStore, func() {
		store.Close()
		os.RemoveAll(dir)
	}
}

func assertErrorMessage(t *testing.T, want error, got error) {
	t.Helper()

	if got == nil {
		t.Fatalf("expected error, got nil")
	}

	if want.Error() != got.Error() {
		t.Errorf("expected error, got %q but want %q", got, want)
	}
}
