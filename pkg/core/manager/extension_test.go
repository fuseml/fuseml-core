package manager

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/pipeline/test/diff"

	"github.com/fuseml/fuseml-core/pkg/core"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

func assertErrorType(t testing.TB, got, want error) {
	t.Helper()

	if d := cmp.Diff(got, want); d != "" {
		t.Errorf("Unexpected Error: %s", diff.PrintWantGot(d))
	}
}

func newExtensionRegistry() *ExtensionRegistry {
	return NewExtensionRegistry(core.NewExtensionStore())
}

// Test registering an extension
func TestExtensionRegister(t *testing.T) {
	t.Run("explicit IDs", func(t *testing.T) {
		registry := newExtensionRegistry()
		e := &domain.Extension{ID: "testextension"}
		ctx := context.Background()

		s1 := &domain.ExtensionService{ID: "testservice-001"}
		ep1 := &domain.ExtensionServiceEndpoint{URL: "https://testendpoint-001.com"}
		c1 := &domain.ExtensionServiceCredentials{ID: "testcredentials-001"}
		s1.AddEndpoint(ep1)
		s1.AddCredentials(c1)
		e.AddService(s1)

		s2 := &domain.ExtensionService{ID: "testservice-002"}
		ep2 := &domain.ExtensionServiceEndpoint{URL: "https://testendpoint-002.com"}
		c2 := &domain.ExtensionServiceCredentials{ID: "testcredentials-002"}
		s2.AddEndpoint(ep2)
		s2.AddCredentials(c2)
		e.AddService(s2)

		eIn, err := registry.RegisterExtension(ctx, e)
		assertError(t, err, nil)
		if d := cmp.Diff(e, eIn); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		eOut, err := registry.GetExtension(ctx, e.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(eIn, eOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		s1Out, err := registry.GetService(ctx, e.ID, s1.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(s1, s1Out); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}
		s2Out, err := registry.GetService(ctx, e.ID, s2.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(s2, s2Out); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}
		ep1Out, err := registry.GetEndpoint(ctx, e.ID, s1.ID, ep1.URL)
		assertError(t, err, nil)
		if d := cmp.Diff(ep1, ep1Out); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}
		ep2Out, err := registry.GetEndpoint(ctx, e.ID, s2.ID, ep2.URL)
		assertError(t, err, nil)
		if d := cmp.Diff(ep2, ep2Out); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}
		c1Out, err := registry.GetCredentials(ctx, e.ID, s1.ID, c1.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(c1, c1Out); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}
		c2Out, err := registry.GetCredentials(ctx, e.ID, s2.ID, c2.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(c2, c2Out); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("generated ID", func(t *testing.T) {
		registry := newExtensionRegistry()
		e := &domain.Extension{Product: "testproduct"}
		ctx := context.Background()

		s1 := &domain.ExtensionService{}
		ep1 := &domain.ExtensionServiceEndpoint{URL: "https://testendpoint-001.com"}
		c1 := &domain.ExtensionServiceCredentials{}
		s1.AddEndpoint(ep1)
		s1.AddCredentials(c1)
		e.AddService(s1)

		s2 := &domain.ExtensionService{Resource: "testresource"}
		ep2 := &domain.ExtensionServiceEndpoint{URL: "https://testendpoint-002.com"}
		c2 := &domain.ExtensionServiceCredentials{}
		s2.AddEndpoint(ep2)
		s2.AddCredentials(c2)
		e.AddService(s2)

		eIn, err := registry.RegisterExtension(ctx, e)
		assertError(t, err, nil)
		if d := cmp.Diff(e, eIn); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}
		if !strings.HasPrefix(e.ID, "testproduct-") {
			t.Errorf("Unexpected Extension ID: %s", eIn.ID)
		}
		if !strings.HasPrefix(s1.ID, "testproduct-service-") {
			t.Errorf("Unexpected Service ID: %s", s1.ID)
		}
		if !strings.HasPrefix(s2.ID, "testresource-") {
			t.Errorf("Unexpected Service ID: %s", s2.ID)
		}
		if !strings.HasPrefix(c1.ID, "creds-") {
			t.Errorf("Unexpected Credentials ID: %s", c1.ID)
		}
		if !strings.HasPrefix(c2.ID, "testresource-") {
			t.Errorf("Unexpected Credentials ID: %s", c2.ID)
		}

		eOut, err := registry.GetExtension(ctx, e.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(eIn, eOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		s1Out, err := registry.GetService(ctx, e.ID, s1.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(s1, s1Out); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}
		s2Out, err := registry.GetService(ctx, e.ID, s2.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(s2, s2Out); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}
		ep1Out, err := registry.GetEndpoint(ctx, e.ID, s1.ID, ep1.URL)
		assertError(t, err, nil)
		if d := cmp.Diff(ep1, ep1Out); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}
		ep2Out, err := registry.GetEndpoint(ctx, e.ID, s2.ID, ep2.URL)
		assertError(t, err, nil)
		if d := cmp.Diff(ep2, ep2Out); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}
		c1Out, err := registry.GetCredentials(ctx, e.ID, s1.ID, c1.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(c1, c1Out); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}
		c2Out, err := registry.GetCredentials(ctx, e.ID, s2.ID, c2.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(c2, c2Out); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}

	})
}

// Test adding services, endpoints and credentials to an existing extension
func TestExtensionAdd(t *testing.T) {
	t.Run("explicit IDs", func(t *testing.T) {
		registry := newExtensionRegistry()
		e := &domain.Extension{ID: "testextension"}
		ctx := context.Background()

		erIn, err := registry.RegisterExtension(ctx, e)
		assertError(t, err, nil)
		if d := cmp.Diff(e, erIn); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		eOut, err := registry.GetExtension(ctx, e.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, eOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		s1 := &domain.ExtensionService{ID: "testservice-001"}
		s1In, err := registry.AddService(ctx, e.ID, s1)
		assertError(t, err, nil)
		if d := cmp.Diff(s1, s1In); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}
		s1Out, err := registry.GetService(ctx, e.ID, s1.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(s1, s1Out); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}

		s2 := &domain.ExtensionService{ID: "testservice-002"}
		s2In, err := registry.AddService(ctx, e.ID, s2)
		assertError(t, err, nil)
		if d := cmp.Diff(s2, s2In); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}
		s2Out, err := registry.GetService(ctx, e.ID, s2.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(s2, s2Out); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}

		eOut, err = registry.GetExtension(ctx, e.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, eOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		ep1 := &domain.ExtensionServiceEndpoint{URL: "https://testendpoint-001.com"}
		ep1In, err := registry.AddEndpoint(ctx, e.ID, s1.ID, ep1)
		assertError(t, err, nil)
		if d := cmp.Diff(ep1, ep1In); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}
		ep1Out, err := registry.GetEndpoint(ctx, e.ID, s1.ID, ep1.URL)
		assertError(t, err, nil)
		if d := cmp.Diff(ep1, ep1Out); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}

		ep2 := &domain.ExtensionServiceEndpoint{URL: "https://testendpoint-002.com"}
		ep2In, err := registry.AddEndpoint(ctx, e.ID, s2.ID, ep2)
		assertError(t, err, nil)
		if d := cmp.Diff(ep2, ep2In); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}
		ep2Out, err := registry.GetEndpoint(ctx, e.ID, s2.ID, ep2.URL)
		assertError(t, err, nil)
		if d := cmp.Diff(ep2, ep2Out); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}

		eOut, err = registry.GetExtension(ctx, e.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, eOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		c1 := &domain.ExtensionServiceCredentials{ID: "testcredentials-001"}
		c1In, err := registry.AddCredentials(ctx, e.ID, s1.ID, c1)
		assertError(t, err, nil)
		if d := cmp.Diff(c1, c1In); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}
		c1Out, err := registry.GetCredentials(ctx, e.ID, s1.ID, c1.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(c1, c1Out); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}

		c2 := &domain.ExtensionServiceCredentials{ID: "testcredentials-002"}
		c2In, err := registry.AddCredentials(ctx, e.ID, s2.ID, c2)
		assertError(t, err, nil)
		if d := cmp.Diff(c2, c2In); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}
		c2Out, err := registry.GetCredentials(ctx, e.ID, s2.ID, c2.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(c2, c2Out); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}

		eOut, err = registry.GetExtension(ctx, e.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, eOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

	})
}

// Test removing services, endpoints and credentials incrementally from an existing extension
func TestExtensionRemove(t *testing.T) {
	t.Run("incremental", func(t *testing.T) {
		registry := newExtensionRegistry()
		e := &domain.Extension{ID: "testextension"}
		ctx := context.Background()

		s1 := &domain.ExtensionService{ID: "testservice-001"}
		ep1 := &domain.ExtensionServiceEndpoint{URL: "https://testendpoint-001.com"}
		c1 := &domain.ExtensionServiceCredentials{ID: "testcredentials-001"}
		s1.AddEndpoint(ep1)
		s1.AddCredentials(c1)
		e.AddService(s1)

		s2 := &domain.ExtensionService{ID: "testservice-002"}
		ep2 := &domain.ExtensionServiceEndpoint{URL: "https://testendpoint-002.com"}
		c2 := &domain.ExtensionServiceCredentials{ID: "testcredentials-002"}
		s2.AddEndpoint(ep2)
		s2.AddCredentials(c2)
		e.AddService(s2)

		eIn, err := registry.RegisterExtension(ctx, e)
		assertError(t, err, nil)
		if d := cmp.Diff(e, eIn); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}
		eOut, err := registry.GetExtension(ctx, e.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(eIn, eOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		err = registry.RemoveEndpoint(ctx, e.ID, s1.ID, ep1.URL)
		assertError(t, err, nil)
		_, err = registry.GetEndpoint(ctx, e.ID, s1.ID, ep1.URL)
		assertErrorType(t, err, domain.NewErrExtensionServiceEndpointNotFound(e.ID, s1.ID, ep1.URL))

		eOut, err = registry.GetExtension(ctx, e.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(eIn, eOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		err = registry.RemoveEndpoint(ctx, e.ID, s2.ID, ep2.URL)
		assertError(t, err, nil)
		_, err = registry.GetEndpoint(ctx, e.ID, s2.ID, ep2.URL)
		assertErrorType(t, err, domain.NewErrExtensionServiceEndpointNotFound(e.ID, s2.ID, ep2.URL))

		eOut, err = registry.GetExtension(ctx, e.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(eIn, eOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		err = registry.RemoveCredentials(ctx, e.ID, s1.ID, c1.ID)
		assertError(t, err, nil)
		_, err = registry.GetCredentials(ctx, e.ID, s1.ID, c1.ID)
		assertErrorType(t, err, domain.NewErrExtensionServiceCredentialsNotFound(e.ID, s1.ID, c1.ID))

		eOut, err = registry.GetExtension(ctx, e.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(eIn, eOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		err = registry.RemoveCredentials(ctx, e.ID, s2.ID, c2.ID)
		assertError(t, err, nil)
		_, err = registry.GetCredentials(ctx, e.ID, s2.ID, c2.ID)
		assertErrorType(t, err, domain.NewErrExtensionServiceCredentialsNotFound(e.ID, s2.ID, c2.ID))

		eOut, err = registry.GetExtension(ctx, e.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(eIn, eOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		err = registry.RemoveService(ctx, e.ID, s1.ID)
		assertError(t, err, nil)
		_, err = registry.GetService(ctx, e.ID, s1.ID)
		assertErrorType(t, err, domain.NewErrExtensionServiceNotFound(e.ID, s1.ID))

		eOut, err = registry.GetExtension(ctx, e.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(eIn, eOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		err = registry.RemoveService(ctx, e.ID, s2.ID)
		assertError(t, err, nil)
		_, err = registry.GetService(ctx, e.ID, s2.ID)
		assertErrorType(t, err, domain.NewErrExtensionServiceNotFound(e.ID, s2.ID))

		eOut, err = registry.GetExtension(ctx, e.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(eIn, eOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		err = registry.RemoveExtension(ctx, e.ID)
		assertError(t, err, nil)
		_, err = registry.GetExtension(ctx, e.ID)
		assertErrorType(t, err, domain.NewErrExtensionNotFound(e.ID))

	})
}

// Test updating an existing extension, services, endpoints and set of credentials
func TestExtensionUpdate(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		registry := newExtensionRegistry()
		e := &domain.Extension{
			ID:          "testextension",
			Product:     "testproduct",
			Version:     "v1.0",
			Description: "Test extension v1.0",
			Zone:        "twilight",
			Configuration: map[string]string{
				"ext-config-one": "ext-value-one",
				"ext-config-two": "ext-value-two",
			},
		}
		ctx := context.Background()

		s1 := &domain.ExtensionService{
			ID:           "testservice-001",
			Resource:     "testresource-one",
			Category:     "testcategory-one",
			Description:  "Test service 001",
			AuthRequired: false,
			Configuration: map[string]string{
				"svc-001-config-one": "svc-001-value-one",
				"svc-001-config-two": "svc-001-value-two",
			},
		}
		ep1 := &domain.ExtensionServiceEndpoint{
			URL:  "https://testendpoint-001.com",
			Type: domain.EETExternal,
			Configuration: map[string]string{
				"ep-001-config-one": "svc-001-value-one",
				"ep-001-config-two": "svc-001-value-two",
			},
		}
		c1 := &domain.ExtensionServiceCredentials{
			ID:       "testcredentials-001",
			Scope:    domain.ECSGlobal,
			Default:  true,
			Projects: []string{},
			Users:    []string{},
			Configuration: map[string]string{
				"cred-001-config-one": "cred-001-value-one",
				"cred-001-config-two": "cred-001-value-two",
			},
		}
		s1.AddEndpoint(ep1)
		s1.AddCredentials(c1)
		e.AddService(s1)

		s2 := &domain.ExtensionService{
			ID:           "testservice-002",
			Resource:     "testresource-two",
			Category:     "testcategory-two",
			Description:  "Test service 002",
			AuthRequired: true,
			Configuration: map[string]string{
				"svc-002-config-one": "svc-002-value-one",
				"svc-002-config-two": "svc-002-value-two",
			},
		}
		ep2 := &domain.ExtensionServiceEndpoint{
			URL:  "https://testendpoint-002.com",
			Type: domain.EETInternal,
			Configuration: map[string]string{
				"ep-002-config-one": "svc-002-value-one",
				"ep-002-config-two": "svc-002-value-two",
			},
		}
		c2 := &domain.ExtensionServiceCredentials{
			ID: "testcredentials-002",

			Scope:    domain.ECSUser,
			Default:  false,
			Projects: []string{"project-one", "project-two"},
			Users:    []string{"user-one", "user-two"},
			Configuration: map[string]string{
				"cred-002-config-one": "cred-002-value-one",
				"cred-002-config-two": "cred-002-value-two",
			},
		}
		s2.AddEndpoint(ep2)
		s2.AddCredentials(c2)
		e.AddService(s2)

		eIn, err := registry.RegisterExtension(ctx, e)
		assertError(t, err, nil)
		if d := cmp.Diff(e, eIn); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}
		eOut, err := registry.GetExtension(ctx, e.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(eIn, eOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		e = &domain.Extension{
			ID:          e.ID,
			Product:     "testproduct-update",
			Version:     "v2.0",
			Description: "Test extension v2.0",
			Zone:        "stalker",
			Configuration: map[string]string{
				"ext-config-one": "ext-value-one-updated",
				"ext-config-two": "ext-value-two-updated",
			},
			Services: e.Services,
		}
		err = registry.UpdateExtension(ctx, e)
		assertError(t, err, nil)
		eOut, err = registry.GetExtension(ctx, e.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(e, eOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		s1 = &domain.ExtensionService{
			ID:           s1.ID,
			Resource:     "testresource-one-updated",
			Category:     "testcategory-one-updated",
			Description:  "Test service 001 updated",
			AuthRequired: true,
			Configuration: map[string]string{
				"svc-001-config-one": "svc-001-value-one-updated",
				"svc-001-config-two": "svc-001-value-two-updated",
			},
			Endpoints:   s1.Endpoints,
			Credentials: s1.Credentials,
		}
		err = registry.UpdateService(ctx, e.ID, s1)
		assertError(t, err, nil)
		s1Out, err := registry.GetService(ctx, e.ID, s1.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(s1, s1Out); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}

		s2 = &domain.ExtensionService{
			ID:           s2.ID,
			Resource:     "testresource-two-updated",
			Category:     "testcategory-two-updated",
			Description:  "Test service 002-updated",
			AuthRequired: false,
			Configuration: map[string]string{
				"svc-002-config-one": "svc-002-value-one-updated",
				"svc-002-config-two": "svc-002-value-two-updated",
			},
			Endpoints:   s2.Endpoints,
			Credentials: s2.Credentials,
		}
		err = registry.UpdateService(ctx, e.ID, s2)
		assertError(t, err, nil)
		s2Out, err := registry.GetService(ctx, e.ID, s2.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(s2, s2Out); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}

		ep1 = &domain.ExtensionServiceEndpoint{
			URL:  ep1.URL,
			Type: domain.EETInternal,
			Configuration: map[string]string{
				"ep-001-config-one": "svc-001-value-one-updated",
				"ep-001-config-two": "svc-001-value-two-updated",
			},
		}
		err = registry.UpdateEndpoint(ctx, e.ID, s1.ID, ep1)
		assertError(t, err, nil)
		ep1Out, err := registry.GetEndpoint(ctx, e.ID, s1.ID, ep1.URL)
		assertError(t, err, nil)
		if d := cmp.Diff(ep1, ep1Out); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}

		ep2 = &domain.ExtensionServiceEndpoint{
			URL:  ep2.URL,
			Type: domain.EETExternal,
			Configuration: map[string]string{
				"ep-002-config-one": "svc-002-value-one-updated",
				"ep-002-config-two": "svc-002-value-two-updated",
			},
		}
		err = registry.UpdateEndpoint(ctx, e.ID, s2.ID, ep2)
		assertError(t, err, nil)
		ep2Out, err := registry.GetEndpoint(ctx, e.ID, s2.ID, ep2.URL)
		assertError(t, err, nil)
		if d := cmp.Diff(ep2, ep2Out); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}

		c1 = &domain.ExtensionServiceCredentials{
			ID:       c1.ID,
			Scope:    domain.ECSProject,
			Default:  true,
			Projects: []string{"project-one", "project-two"},
			Users:    []string{},
			Configuration: map[string]string{
				"cred-001-config-one": "cred-001-value-one-updated",
				"cred-001-config-two": "cred-001-value-two-updated",
			},
		}
		err = registry.UpdateCredentials(ctx, e.ID, s1.ID, c1)
		assertError(t, err, nil)
		c1Out, err := registry.GetCredentials(ctx, e.ID, s1.ID, c1.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(c1, c1Out); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}

		c2 = &domain.ExtensionServiceCredentials{
			ID:       c2.ID,
			Scope:    domain.ECSGlobal,
			Default:  true,
			Projects: []string{},
			Users:    []string{},
			Configuration: map[string]string{
				"cred-002-config-one": "cred-002-value-one-updated",
				"cred-002-config-two": "cred-002-value-two-updated",
			},
		}
		err = registry.UpdateCredentials(ctx, e.ID, s2.ID, c2)
		assertError(t, err, nil)
		c2Out, err := registry.GetCredentials(ctx, e.ID, s2.ID, c2.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(c2, c2Out); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}

	})
}

// Test running queries on the extension registry
func TestExtensionQuery(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		registry := newExtensionRegistry()
		ctx := context.Background()
		e1 := &domain.Extension{
			ID:          "testextension-one",
			Product:     "testproduct",
			Version:     "v1.0",
			Description: "Test extension v1.0",
			Zone:        "twilight",
			Configuration: map[string]string{
				"ext-config-one": "ext-value-one",
				"ext-config-two": "ext-value-two",
			},
		}

		s1 := &domain.ExtensionService{
			ID:           "testservice-001",
			Resource:     "testresource-one",
			Category:     "testcategory-one",
			Description:  "Test service 001",
			AuthRequired: false,
			Configuration: map[string]string{
				"svc-001-config-one": "svc-001-value-one",
				"svc-001-config-two": "svc-001-value-two",
			},
		}
		ep1 := &domain.ExtensionServiceEndpoint{
			URL:  "https://testendpoint-001.com",
			Type: domain.EETExternal,
			Configuration: map[string]string{
				"ep-001-config-one": "svc-001-value-one",
				"ep-001-config-two": "svc-001-value-two",
			},
		}
		c1 := &domain.ExtensionServiceCredentials{
			ID:       "testcredentials-001",
			Scope:    domain.ECSGlobal,
			Default:  true,
			Projects: []string{},
			Users:    []string{},
			Configuration: map[string]string{
				"cred-001-config-one": "cred-001-value-one",
				"cred-001-config-two": "cred-001-value-two",
			},
		}
		s1.AddEndpoint(ep1)
		s1.AddCredentials(c1)
		e1.AddService(s1)

		s2 := &domain.ExtensionService{
			ID:           "testservice-002",
			Resource:     "testresource-two",
			Category:     "testcategory-two",
			Description:  "Test service 002",
			AuthRequired: true,
			Configuration: map[string]string{
				"svc-002-config-one": "svc-002-value-one",
				"svc-002-config-two": "svc-002-value-two",
			},
		}
		ep2 := &domain.ExtensionServiceEndpoint{
			URL:  "https://testendpoint-002.com",
			Type: domain.EETInternal,
			Configuration: map[string]string{
				"ep-002-config-one": "svc-002-value-one",
				"ep-002-config-two": "svc-002-value-two",
			},
		}
		c2 := &domain.ExtensionServiceCredentials{
			ID:       "testcredentials-002",
			Scope:    domain.ECSUser,
			Default:  false,
			Projects: []string{"project-one", "project-two"},
			Users:    []string{"user-one", "user-two"},
			Configuration: map[string]string{
				"cred-002-config-one": "cred-002-value-one",
				"cred-002-config-two": "cred-002-value-two",
			},
		}
		s2.AddEndpoint(ep2)
		s2.AddCredentials(c2)
		e1.AddService(s2)

		e1In, err := registry.RegisterExtension(ctx, e1)
		assertError(t, err, nil)
		if d := cmp.Diff(e1, e1In); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}
		e1Out, err := registry.GetExtension(ctx, e1.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(e1In, e1Out); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		e2 := &domain.Extension{
			ID:          "testextension-two",
			Product:     "testproduct",
			Version:     "v1.2.0",
			Description: "Test extension v1.2",
			Zone:        "twilight",
			Configuration: map[string]string{
				"ext-config-one": "ext-value-one",
				"ext-config-two": "ext-value-two",
			},
		}
		s3 := &domain.ExtensionService{
			ID:           "testservice-003",
			Resource:     "testresource-one",
			Category:     "testcategory-one",
			Description:  "Test service 003",
			AuthRequired: false,
			Configuration: map[string]string{
				"svc-003-config-one": "svc-003-value-one",
				"svc-003-config-two": "svc-003-value-two",
			},
		}
		ep3 := &domain.ExtensionServiceEndpoint{
			URL:  "https://testendpoint-003.com",
			Type: domain.EETExternal,
			Configuration: map[string]string{
				"ep-003-config-one": "svc-003-value-one",
				"ep-003-config-two": "svc-003-value-two",
			},
		}
		c3 := &domain.ExtensionServiceCredentials{
			ID:       "testcredentials-003",
			Scope:    domain.ECSGlobal,
			Default:  true,
			Projects: []string{},
			Users:    []string{},
			Configuration: map[string]string{
				"cred-003-config-one": "cred-003-value-one",
				"cred-003-config-two": "cred-003-value-two",
			},
		}
		s3.AddEndpoint(ep3)
		s3.AddCredentials(c3)
		e2.AddService(s3)

		s4 := &domain.ExtensionService{
			ID:           "testservice-004",
			Resource:     "testresource-two",
			Category:     "testcategory-two",
			Description:  "Test service 004",
			AuthRequired: true,
			Configuration: map[string]string{
				"svc-004-config-one": "svc-004-value-one",
				"svc-004-config-two": "svc-004-value-two",
			},
		}
		ep4 := &domain.ExtensionServiceEndpoint{
			URL:  "https://testendpoint-004.com",
			Type: domain.EETInternal,
			Configuration: map[string]string{
				"ep-004-config-one": "svc-004-value-one",
				"ep-004-config-two": "svc-004-value-two",
			},
		}
		c4 := &domain.ExtensionServiceCredentials{
			ID:       "testcredentials-004",
			Scope:    domain.ECSUser,
			Default:  false,
			Projects: []string{"project-one", "project-two"},
			Users:    []string{"user-two", "user-three"},
			Configuration: map[string]string{
				"cred-004-config-one": "cred-004-value-one",
				"cred-004-config-two": "cred-004-value-two",
			},
		}
		s4.AddEndpoint(ep4)
		s4.AddCredentials(c4)
		e2.AddService(s4)

		e2In, err := registry.RegisterExtension(ctx, e2)
		assertError(t, err, nil)
		if d := cmp.Diff(e2, e2In); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}
		e2Out, err := registry.GetExtension(ctx, e2.ID)
		assertError(t, err, nil)
		if d := cmp.Diff(e2In, e2Out); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		var epType domain.ExtensionServiceEndpointType = domain.EETExternal
		q := &domain.ExtensionQuery{
			ExtensionID:        "testextension-one",
			Product:            "testproduct",
			VersionConstraints: "1.0",
			Zone:               "twilight",
			StrictZoneMatch:    true,
			ServiceID:          "testservice-001",
			ServiceResource:    "testresource-one",
			ServiceCategory:    "testcategory-one",
			EndpointURL:        "https://testendpoint-001.com",
			Type:               &epType,
			CredentialsID:      "testcredentials-001",
			CredentialsScope:   domain.ECSGlobal,
			User:               "",
			Project:            "",
		}

		e1Copy := *e1
		e1Copy.Services = map[string]*domain.ExtensionService{s1.ID: s1}
		qRes := []*domain.ExtensionAccessDescriptor{{
			Extension:   e1Copy,
			Service:     *s1,
			Endpoint:    *ep1,
			Credentials: c1,
		}}
		qOut, err := registry.GetExtensionAccessDescriptors(ctx, q)
		assertError(t, err, nil)
		if d := cmp.Diff(qRes, qOut); d != "" {
			t.Errorf("Unexpected Query Results: %s", diff.PrintWantGot(d))
		}

		q = &domain.ExtensionQuery{
			Product:            "testproduct",
			VersionConstraints: ">=1.0,<1.1",
			Zone:               "twilight",
			StrictZoneMatch:    true,
			ServiceResource:    "testresource-two",
			CredentialsScope:   domain.ECSUser,
			User:               "user-one",
		}

		e1Copy.Services = map[string]*domain.ExtensionService{s2.ID: s2}
		qRes = []*domain.ExtensionAccessDescriptor{{
			Extension:   e1Copy,
			Service:     *s2,
			Endpoint:    *ep2,
			Credentials: c2,
		}}
		qOut, err = registry.GetExtensionAccessDescriptors(ctx, q)
		assertError(t, err, nil)
		if d := cmp.Diff(qRes, qOut); d != "" {
			t.Errorf("Unexpected Query Results: %s", diff.PrintWantGot(d))
		}

		q = &domain.ExtensionQuery{
			Product:         "testproduct",
			Zone:            "twilight",
			StrictZoneMatch: true,
			ServiceResource: "testresource-two",
		}

		e2Copy := *e2
		e2Copy.Services = map[string]*domain.ExtensionService{s4.ID: s4}
		qRes = []*domain.ExtensionAccessDescriptor{
			{
				Extension:   e1Copy,
				Service:     *s2,
				Endpoint:    *ep2,
				Credentials: c2,
			},
			{
				Extension:   e2Copy,
				Service:     *s4,
				Endpoint:    *ep4,
				Credentials: c4,
			},
		}
		qOut, err = registry.GetExtensionAccessDescriptors(ctx, q)
		assertError(t, err, nil)
		if d := cmp.Diff(qRes, qOut); d != "" {
			t.Errorf("Unexpected Query Results: %s", diff.PrintWantGot(d))
		}

	})
}
