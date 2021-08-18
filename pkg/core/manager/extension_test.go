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

func newExtension(extension *domain.Extension) (result *domain.ExtensionRecord) {
	return &domain.ExtensionRecord{
		Extension: *extension,
		Services:  make([]*domain.ExtensionServiceRecord, 0),
	}
}

func addService(extRecord *domain.ExtensionRecord, service *domain.ExtensionService) (result *domain.ExtensionServiceRecord) {
	result = &domain.ExtensionServiceRecord{
		ExtensionService: *service,
		Endpoints:        make([]*domain.ExtensionEndpoint, 0),
		Credentials:      make([]*domain.ExtensionCredentials, 0),
	}
	extRecord.Services = append(extRecord.Services, result)
	return
}

func addEndpoint(svcRecord *domain.ExtensionServiceRecord, endpoint *domain.ExtensionEndpoint) {
	svcRecord.Endpoints = append(svcRecord.Endpoints, endpoint)
}

func addCredentials(svcRecord *domain.ExtensionServiceRecord, credentials *domain.ExtensionCredentials) {
	svcRecord.Credentials = append(svcRecord.Credentials, credentials)
}

// Test registering an extension
func TestExtensionRegister(t *testing.T) {
	t.Run("explicit IDs", func(t *testing.T) {
		registry := newExtensionRegistry()
		e := &domain.Extension{
			ID: "testextension",
		}
		er := newExtension(e)
		s1 := &domain.ExtensionService{
			ExtensionServiceID: domain.ExtensionServiceID{
				ID: "testservice-001",
			},
		}
		sr1 := addService(er, s1)
		s2 := &domain.ExtensionService{
			ExtensionServiceID: domain.ExtensionServiceID{
				ID: "testservice-002",
			},
		}
		sr2 := addService(er, s2)
		ep1 := &domain.ExtensionEndpoint{
			ExtensionEndpointID: domain.ExtensionEndpointID{
				URL: "https://testendpoint-001.com",
			},
		}
		addEndpoint(sr1, ep1)
		ep2 := &domain.ExtensionEndpoint{
			ExtensionEndpointID: domain.ExtensionEndpointID{
				URL: "https://testendpoint-002.com",
			},
		}
		addEndpoint(sr2, ep2)
		c1 := &domain.ExtensionCredentials{
			ExtensionCredentialsID: domain.ExtensionCredentialsID{
				ID: "testcredentials-001",
			},
		}
		addCredentials(sr1, c1)
		c2 := &domain.ExtensionCredentials{
			ExtensionCredentialsID: domain.ExtensionCredentialsID{
				ID: "testcredentials-002",
			},
		}
		addCredentials(sr2, c2)

		erIn, err := registry.RegisterExtension(context.Background(), er)
		assertError(t, err, nil)
		if d := cmp.Diff(er, erIn); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}
		erOut, err := registry.GetExtension(context.Background(), "testextension", true)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, erOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}
		sr1Out, err := registry.GetService(context.Background(), domain.ExtensionServiceID{
			ExtensionID: "testextension",
			ID:          "testservice-001",
		}, true)
		assertError(t, err, nil)
		if d := cmp.Diff(sr1, sr1Out); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}
		sr2Out, err := registry.GetService(context.Background(), domain.ExtensionServiceID{
			ExtensionID: "testextension",
			ID:          "testservice-002",
		}, true)
		assertError(t, err, nil)
		if d := cmp.Diff(sr2, sr2Out); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}
		ep1Out, err := registry.GetEndpoint(
			context.Background(), domain.ExtensionEndpointID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-001",
				URL:         "https://testendpoint-001.com",
			})
		assertError(t, err, nil)
		if d := cmp.Diff(ep1, ep1Out); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}
		ep2Out, err := registry.GetEndpoint(
			context.Background(), domain.ExtensionEndpointID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-002",
				URL:         "https://testendpoint-002.com",
			})
		assertError(t, err, nil)
		if d := cmp.Diff(ep2, ep2Out); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}
		c1Out, err := registry.GetCredentials(
			context.Background(), domain.ExtensionCredentialsID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-001",
				ID:          "testcredentials-001",
			})
		assertError(t, err, nil)
		if d := cmp.Diff(c1, c1Out); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}
		c2Out, err := registry.GetCredentials(
			context.Background(), domain.ExtensionCredentialsID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-002",
				ID:          "testcredentials-002",
			})
		assertError(t, err, nil)
		if d := cmp.Diff(c2, c2Out); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("generated ID", func(t *testing.T) {
		registry := newExtensionRegistry()
		e := &domain.Extension{
			Product: "testproduct",
		}
		er := newExtension(e)
		s1 := &domain.ExtensionService{}
		sr1 := addService(er, s1)
		s2 := &domain.ExtensionService{Resource: "testresource"}
		sr2 := addService(er, s2)
		ep1 := &domain.ExtensionEndpoint{
			ExtensionEndpointID: domain.ExtensionEndpointID{
				URL: "https://testendpoint-001.com",
			},
		}
		addEndpoint(sr1, ep1)
		ep2 := &domain.ExtensionEndpoint{
			ExtensionEndpointID: domain.ExtensionEndpointID{
				URL: "https://testendpoint-002.com",
			},
		}
		addEndpoint(sr2, ep2)
		c1 := &domain.ExtensionCredentials{}
		addCredentials(sr1, c1)
		c2 := &domain.ExtensionCredentials{}
		addCredentials(sr2, c2)

		erIn, err := registry.RegisterExtension(context.Background(), er)
		assertError(t, err, nil)
		if d := cmp.Diff(er, erIn); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}
		if !strings.HasPrefix(er.ID, "testproduct-") {
			t.Errorf("Unexpected Extension ID: %s", er.ID)
		}
		if !strings.HasPrefix(sr1.ID, "testproduct-service-") {
			t.Errorf("Unexpected Service ID: %s", sr1.ID)
		}
		if !strings.HasPrefix(sr2.ID, "testresource-") {
			t.Errorf("Unexpected Service ID: %s", sr2.ID)
		}
		if !strings.HasPrefix(c1.ID, "creds-") {
			t.Errorf("Unexpected Credentials ID: %s", c1.ID)
		}
		if !strings.HasPrefix(c2.ID, "testresource-") {
			t.Errorf("Unexpected Credentials ID: %s", c2.ID)
		}

		erOut, err := registry.GetExtension(context.Background(), er.ID, true)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, erOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		sr1Out, err := registry.GetService(context.Background(), sr1.ExtensionServiceID, true)
		assertError(t, err, nil)
		if d := cmp.Diff(sr1, sr1Out); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}
		sr2Out, err := registry.GetService(context.Background(), sr2.ExtensionServiceID, true)
		assertError(t, err, nil)
		if d := cmp.Diff(sr2, sr2Out); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}
		ep1Out, err := registry.GetEndpoint(
			context.Background(), ep1.ExtensionEndpointID)
		assertError(t, err, nil)
		if d := cmp.Diff(ep1, ep1Out); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}
		ep2Out, err := registry.GetEndpoint(
			context.Background(), ep2.ExtensionEndpointID)
		assertError(t, err, nil)
		if d := cmp.Diff(ep2, ep2Out); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}
		c1Out, err := registry.GetCredentials(
			context.Background(), c1.ExtensionCredentialsID)
		assertError(t, err, nil)
		if d := cmp.Diff(c1, c1Out); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}
		c2Out, err := registry.GetCredentials(
			context.Background(), c2.ExtensionCredentialsID)
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
		e := &domain.Extension{
			ID: "testextension",
		}
		er := newExtension(e)
		erIn, err := registry.RegisterExtension(context.Background(), er)
		assertError(t, err, nil)
		if d := cmp.Diff(er, erIn); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}
		erOut, err := registry.GetExtension(context.Background(), "testextension", true)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, erOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		s1 := &domain.ExtensionService{
			ExtensionServiceID: domain.ExtensionServiceID{
				ExtensionID: "testextension",
				ID:          "testservice-001",
			},
		}
		sr1 := addService(er, s1)
		s2 := &domain.ExtensionService{
			ExtensionServiceID: domain.ExtensionServiceID{
				ExtensionID: "testextension",
				ID:          "testservice-002",
			},
		}
		sr2 := addService(er, s2)

		sr1In, err := registry.AddService(context.Background(), sr1)
		assertError(t, err, nil)
		if d := cmp.Diff(sr1, sr1In); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}
		sr1Out, err := registry.GetService(context.Background(), domain.ExtensionServiceID{
			ExtensionID: "testextension",
			ID:          "testservice-001",
		}, true)
		assertError(t, err, nil)
		if d := cmp.Diff(sr1, sr1Out); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}

		sr2In, err := registry.AddService(context.Background(), sr2)
		assertError(t, err, nil)
		if d := cmp.Diff(sr2, sr2In); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}
		sr2Out, err := registry.GetService(context.Background(), domain.ExtensionServiceID{
			ExtensionID: "testextension",
			ID:          "testservice-002",
		}, true)
		assertError(t, err, nil)
		if d := cmp.Diff(sr2, sr2Out); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}

		erOut, err = registry.GetExtension(context.Background(), "testextension", true)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, erOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		ep1 := &domain.ExtensionEndpoint{
			ExtensionEndpointID: domain.ExtensionEndpointID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-001",
				URL:         "https://testendpoint-001.com",
			},
		}
		addEndpoint(sr1, ep1)
		ep2 := &domain.ExtensionEndpoint{
			ExtensionEndpointID: domain.ExtensionEndpointID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-002",
				URL:         "https://testendpoint-002.com",
			},
		}
		addEndpoint(sr2, ep2)

		ep1In, err := registry.AddEndpoint(context.Background(), ep1)
		assertError(t, err, nil)
		if d := cmp.Diff(ep1, ep1In); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}
		ep1Out, err := registry.GetEndpoint(
			context.Background(), domain.ExtensionEndpointID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-001",
				URL:         "https://testendpoint-001.com",
			})
		assertError(t, err, nil)
		if d := cmp.Diff(ep1, ep1Out); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}

		ep2In, err := registry.AddEndpoint(context.Background(), ep2)
		assertError(t, err, nil)
		if d := cmp.Diff(ep2, ep2In); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}
		ep2Out, err := registry.GetEndpoint(
			context.Background(), domain.ExtensionEndpointID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-002",
				URL:         "https://testendpoint-002.com",
			})
		assertError(t, err, nil)
		if d := cmp.Diff(ep2, ep2Out); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}

		erOut, err = registry.GetExtension(context.Background(), "testextension", true)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, erOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		c1 := &domain.ExtensionCredentials{
			ExtensionCredentialsID: domain.ExtensionCredentialsID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-001",
				ID:          "testcredentials-001",
			},
		}
		addCredentials(sr1, c1)
		c2 := &domain.ExtensionCredentials{
			ExtensionCredentialsID: domain.ExtensionCredentialsID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-002",
				ID:          "testcredentials-002",
			},
		}
		addCredentials(sr2, c2)

		c1In, err := registry.AddCredentials(context.Background(), c1)
		assertError(t, err, nil)
		if d := cmp.Diff(c1, c1In); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}
		c1Out, err := registry.GetCredentials(
			context.Background(), domain.ExtensionCredentialsID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-001",
				ID:          "testcredentials-001",
			})
		assertError(t, err, nil)
		if d := cmp.Diff(c1, c1Out); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}

		c2In, err := registry.AddCredentials(context.Background(), c2)
		assertError(t, err, nil)
		if d := cmp.Diff(c2, c2In); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}
		c2Out, err := registry.GetCredentials(
			context.Background(), domain.ExtensionCredentialsID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-002",
				ID:          "testcredentials-002",
			})
		assertError(t, err, nil)
		if d := cmp.Diff(c2, c2Out); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}

		erOut, err = registry.GetExtension(context.Background(), "testextension", true)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, erOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

	})
}

// Test removing services, endpoints and credentials incrementally from an existing extension
func TestExtensionRemove(t *testing.T) {
	t.Run("incremental", func(t *testing.T) {
		registry := newExtensionRegistry()
		e := &domain.Extension{
			ID: "testextension",
		}
		er := newExtension(e)
		s1 := &domain.ExtensionService{
			ExtensionServiceID: domain.ExtensionServiceID{
				ID: "testservice-001",
			},
		}
		sr1 := addService(er, s1)
		s2 := &domain.ExtensionService{
			ExtensionServiceID: domain.ExtensionServiceID{
				ID: "testservice-002",
			},
		}
		sr2 := addService(er, s2)
		ep1 := &domain.ExtensionEndpoint{
			ExtensionEndpointID: domain.ExtensionEndpointID{
				URL: "https://testendpoint-001.com",
			},
		}
		addEndpoint(sr1, ep1)
		ep2 := &domain.ExtensionEndpoint{
			ExtensionEndpointID: domain.ExtensionEndpointID{
				URL: "https://testendpoint-002.com",
			},
		}
		addEndpoint(sr2, ep2)
		c1 := &domain.ExtensionCredentials{
			ExtensionCredentialsID: domain.ExtensionCredentialsID{
				ID: "testcredentials-001",
			},
		}
		addCredentials(sr1, c1)
		c2 := &domain.ExtensionCredentials{
			ExtensionCredentialsID: domain.ExtensionCredentialsID{
				ID: "testcredentials-002",
			},
		}
		addCredentials(sr2, c2)

		erIn, err := registry.RegisterExtension(context.Background(), er)
		assertError(t, err, nil)
		if d := cmp.Diff(er, erIn); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}
		erOut, err := registry.GetExtension(context.Background(), "testextension", true)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, erOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		sr1.Endpoints = sr1.Endpoints[:0]
		err = registry.RemoveEndpoint(
			context.Background(), domain.ExtensionEndpointID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-001",
				URL:         "https://testendpoint-001.com",
			})
		assertError(t, err, nil)
		_, err = registry.GetEndpoint(
			context.Background(), domain.ExtensionEndpointID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-001",
				URL:         "https://testendpoint-001.com",
			})
		assertErrorType(t, err, domain.NewErrExtensionEndpointNotFound("testextension", "testservice-001", "https://testendpoint-001.com"))
		erOut, err = registry.GetExtension(context.Background(), "testextension", true)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, erOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		sr2.Endpoints = sr2.Endpoints[:0]
		err = registry.RemoveEndpoint(
			context.Background(), domain.ExtensionEndpointID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-002",
				URL:         "https://testendpoint-002.com",
			})
		assertError(t, err, nil)
		_, err = registry.GetEndpoint(
			context.Background(), domain.ExtensionEndpointID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-002",
				URL:         "https://testendpoint-002.com",
			})
		assertErrorType(t, err, domain.NewErrExtensionEndpointNotFound("testextension", "testservice-002", "https://testendpoint-002.com"))
		erOut, err = registry.GetExtension(context.Background(), "testextension", true)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, erOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		sr1.Credentials = sr1.Credentials[:0]
		err = registry.RemoveCredentials(
			context.Background(), domain.ExtensionCredentialsID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-001",
				ID:          "testcredentials-001",
			})
		assertError(t, err, nil)
		_, err = registry.GetCredentials(
			context.Background(), domain.ExtensionCredentialsID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-001",
				ID:          "testcredentials-001",
			})
		assertErrorType(t, err, domain.NewErrExtensionCredentialsNotFound("testextension", "testservice-001", "testcredentials-001"))
		erOut, err = registry.GetExtension(context.Background(), "testextension", true)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, erOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		sr2.Credentials = sr2.Credentials[:0]
		err = registry.RemoveCredentials(
			context.Background(), domain.ExtensionCredentialsID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-002",
				ID:          "testcredentials-002",
			})
		assertError(t, err, nil)
		_, err = registry.GetCredentials(
			context.Background(), domain.ExtensionCredentialsID{
				ExtensionID: "testextension",
				ServiceID:   "testservice-002",
				ID:          "testcredentials-002",
			})
		assertErrorType(t, err, domain.NewErrExtensionCredentialsNotFound("testextension", "testservice-002", "testcredentials-002"))
		erOut, err = registry.GetExtension(context.Background(), "testextension", true)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, erOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		er.Services = er.Services[:1]
		err = registry.RemoveService(
			context.Background(), domain.ExtensionServiceID{
				ExtensionID: "testextension",
				ID:          "testservice-002",
			})
		assertError(t, err, nil)
		_, err = registry.GetService(context.Background(), domain.ExtensionServiceID{
			ExtensionID: "testextension",
			ID:          "testservice-002",
		}, true)
		assertErrorType(t, err, domain.NewErrExtensionServiceNotFound("testextension", "testservice-002"))
		erOut, err = registry.GetExtension(context.Background(), "testextension", true)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, erOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		er.Services = make([]*domain.ExtensionServiceRecord, 0)
		err = registry.RemoveService(
			context.Background(), domain.ExtensionServiceID{
				ExtensionID: "testextension",
				ID:          "testservice-001",
			})
		assertError(t, err, nil)
		_, err = registry.GetService(context.Background(), domain.ExtensionServiceID{
			ExtensionID: "testextension",
			ID:          "testservice-001",
		}, true)
		assertErrorType(t, err, domain.NewErrExtensionServiceNotFound("testextension", "testservice-001"))
		erOut, err = registry.GetExtension(context.Background(), "testextension", true)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, erOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		err = registry.RemoveExtension(context.Background(), "testextension")
		assertError(t, err, nil)
		_, err = registry.GetExtension(context.Background(), "testextension", true)
		assertErrorType(t, err, domain.NewErrExtensionNotFound("testextension"))

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
		er := newExtension(e)
		s1 := &domain.ExtensionService{
			ExtensionServiceID: domain.ExtensionServiceID{
				ID: "testservice-001",
			},
			Resource:     "testresource-one",
			Category:     "testcategory-one",
			Description:  "Test service 001",
			AuthRequired: false,
			Configuration: map[string]string{
				"svc-001-config-one": "svc-001-value-one",
				"svc-001-config-two": "svc-001-value-two",
			},
		}
		sr1 := addService(er, s1)
		s2 := &domain.ExtensionService{
			ExtensionServiceID: domain.ExtensionServiceID{
				ID: "testservice-002",
			},
			Resource:     "testresource-two",
			Category:     "testcategory-two",
			Description:  "Test service 002",
			AuthRequired: true,
			Configuration: map[string]string{
				"svc-002-config-one": "svc-002-value-one",
				"svc-002-config-two": "svc-002-value-two",
			},
		}
		sr2 := addService(er, s2)
		ep1 := &domain.ExtensionEndpoint{
			ExtensionEndpointID: domain.ExtensionEndpointID{
				URL: "https://testendpoint-001.com",
			},
			Type: domain.EETExternal,
			Configuration: map[string]string{
				"ep-001-config-one": "svc-001-value-one",
				"ep-001-config-two": "svc-001-value-two",
			},
		}
		addEndpoint(sr1, ep1)
		ep2 := &domain.ExtensionEndpoint{
			ExtensionEndpointID: domain.ExtensionEndpointID{
				URL: "https://testendpoint-002.com",
			},
			Type: domain.EETInternal,
			Configuration: map[string]string{
				"ep-002-config-one": "svc-002-value-one",
				"ep-002-config-two": "svc-002-value-two",
			},
		}
		addEndpoint(sr2, ep2)
		c1 := &domain.ExtensionCredentials{
			ExtensionCredentialsID: domain.ExtensionCredentialsID{
				ID: "testcredentials-001",
			},
			Scope:    domain.ECSGlobal,
			Default:  true,
			Projects: []string{},
			Users:    []string{},
			Configuration: map[string]string{
				"cred-001-config-one": "cred-001-value-one",
				"cred-001-config-two": "cred-001-value-two",
			},
		}
		addCredentials(sr1, c1)
		c2 := &domain.ExtensionCredentials{
			ExtensionCredentialsID: domain.ExtensionCredentialsID{
				ID: "testcredentials-002",
			},
			Scope:    domain.ECSUser,
			Default:  false,
			Projects: []string{"project-one", "project-two"},
			Users:    []string{"user-one", "user-two"},
			Configuration: map[string]string{
				"cred-002-config-one": "cred-002-value-one",
				"cred-002-config-two": "cred-002-value-two",
			},
		}
		addCredentials(sr2, c2)

		erIn, err := registry.RegisterExtension(context.Background(), er)
		assertError(t, err, nil)
		if d := cmp.Diff(er, erIn); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}
		erOut, err := registry.GetExtension(context.Background(), "testextension", true)
		assertError(t, err, nil)
		if d := cmp.Diff(erIn, erOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		er.Extension = domain.Extension{
			ID:          er.ID,
			Product:     "testproduct-update",
			Version:     "v2.0",
			Description: "Test extension v2.0",
			Zone:        "stalker",
			Configuration: map[string]string{
				"ext-config-one": "ext-value-one-updated",
				"ext-config-two": "ext-value-two-updated",
			},
		}
		err = registry.UpdateExtension(context.Background(), &er.Extension)
		assertError(t, err, nil)
		erOut, err = registry.GetExtension(context.Background(), "testextension", true)
		assertError(t, err, nil)
		if d := cmp.Diff(er, erOut); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		sr1.ExtensionService = domain.ExtensionService{
			ExtensionServiceID: sr1.ExtensionServiceID,
			Resource:           "testresource-one-updated",
			Category:           "testcategory-one-updated",
			Description:        "Test service 001 updated",
			AuthRequired:       true,
			Configuration: map[string]string{
				"svc-001-config-one": "svc-001-value-one-updated",
				"svc-001-config-two": "svc-001-value-two-updated",
			},
		}
		err = registry.UpdateService(context.Background(), &sr1.ExtensionService)
		assertError(t, err, nil)
		sr1Out, err := registry.GetService(context.Background(), sr1.ExtensionServiceID, true)
		assertError(t, err, nil)
		if d := cmp.Diff(sr1, sr1Out); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}

		sr2.ExtensionService = domain.ExtensionService{
			ExtensionServiceID: sr2.ExtensionServiceID,
			Resource:           "testresource-two-updated",
			Category:           "testcategory-two-updated",
			Description:        "Test service 002-updated",
			AuthRequired:       false,
			Configuration: map[string]string{
				"svc-002-config-one": "svc-002-value-one-updated",
				"svc-002-config-two": "svc-002-value-two-updated",
			},
		}
		err = registry.UpdateService(context.Background(), &sr2.ExtensionService)
		assertError(t, err, nil)
		sr2Out, err := registry.GetService(context.Background(), sr2.ExtensionServiceID, true)
		assertError(t, err, nil)
		if d := cmp.Diff(sr2, sr2Out); d != "" {
			t.Errorf("Unexpected Service: %s", diff.PrintWantGot(d))
		}

		*ep1 = domain.ExtensionEndpoint{
			ExtensionEndpointID: ep1.ExtensionEndpointID,
			Type:                domain.EETInternal,
			Configuration: map[string]string{
				"ep-001-config-one": "svc-001-value-one-updated",
				"ep-001-config-two": "svc-001-value-two-updated",
			},
		}
		err = registry.UpdateEndpoint(context.Background(), ep1)
		assertError(t, err, nil)
		ep1Out, err := registry.GetEndpoint(context.Background(), ep1.ExtensionEndpointID)
		assertError(t, err, nil)
		if d := cmp.Diff(ep1, ep1Out); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}

		*ep2 = domain.ExtensionEndpoint{
			ExtensionEndpointID: ep2.ExtensionEndpointID,
			Type:                domain.EETExternal,
			Configuration: map[string]string{
				"ep-002-config-one": "svc-002-value-one-updated",
				"ep-002-config-two": "svc-002-value-two-updated",
			},
		}
		err = registry.UpdateEndpoint(context.Background(), ep2)
		assertError(t, err, nil)
		ep2Out, err := registry.GetEndpoint(context.Background(), ep2.ExtensionEndpointID)
		assertError(t, err, nil)
		if d := cmp.Diff(ep2, ep2Out); d != "" {
			t.Errorf("Unexpected Endpoint: %s", diff.PrintWantGot(d))
		}

		*c1 = domain.ExtensionCredentials{
			ExtensionCredentialsID: c1.ExtensionCredentialsID,
			Scope:                  domain.ECSProject,
			Default:                true,
			Projects:               []string{"project-one", "project-two"},
			Users:                  []string{},
			Configuration: map[string]string{
				"cred-001-config-one": "cred-001-value-one-updated",
				"cred-001-config-two": "cred-001-value-two-updated",
			},
		}
		err = registry.UpdateCredentials(context.Background(), c1)
		assertError(t, err, nil)
		c1Out, err := registry.GetCredentials(context.Background(), c1.ExtensionCredentialsID)
		assertError(t, err, nil)
		if d := cmp.Diff(c1, c1Out); d != "" {
			t.Errorf("Unexpected Credentials: %s", diff.PrintWantGot(d))
		}

		*c2 = domain.ExtensionCredentials{
			ExtensionCredentialsID: c2.ExtensionCredentialsID,
			Scope:                  domain.ECSGlobal,
			Default:                true,
			Projects:               []string{},
			Users:                  []string{},
			Configuration: map[string]string{
				"cred-002-config-one": "cred-002-value-one-updated",
				"cred-002-config-two": "cred-002-value-two-updated",
			},
		}
		err = registry.UpdateCredentials(context.Background(), c2)
		assertError(t, err, nil)
		c2Out, err := registry.GetCredentials(context.Background(), c2.ExtensionCredentialsID)
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
		er1 := newExtension(e1)
		s1 := &domain.ExtensionService{
			ExtensionServiceID: domain.ExtensionServiceID{
				ID: "testservice-001",
			},
			Resource:     "testresource-one",
			Category:     "testcategory-one",
			Description:  "Test service 001",
			AuthRequired: false,
			Configuration: map[string]string{
				"svc-001-config-one": "svc-001-value-one",
				"svc-001-config-two": "svc-001-value-two",
			},
		}
		sr1 := addService(er1, s1)
		s2 := &domain.ExtensionService{
			ExtensionServiceID: domain.ExtensionServiceID{
				ID: "testservice-002",
			},
			Resource:     "testresource-two",
			Category:     "testcategory-two",
			Description:  "Test service 002",
			AuthRequired: true,
			Configuration: map[string]string{
				"svc-002-config-one": "svc-002-value-one",
				"svc-002-config-two": "svc-002-value-two",
			},
		}
		sr2 := addService(er1, s2)
		ep1 := &domain.ExtensionEndpoint{
			ExtensionEndpointID: domain.ExtensionEndpointID{
				URL: "https://testendpoint-001.com",
			},
			Type: domain.EETExternal,
			Configuration: map[string]string{
				"ep-001-config-one": "svc-001-value-one",
				"ep-001-config-two": "svc-001-value-two",
			},
		}
		addEndpoint(sr1, ep1)
		ep2 := &domain.ExtensionEndpoint{
			ExtensionEndpointID: domain.ExtensionEndpointID{
				URL: "https://testendpoint-002.com",
			},
			Type: domain.EETInternal,
			Configuration: map[string]string{
				"ep-002-config-one": "svc-002-value-one",
				"ep-002-config-two": "svc-002-value-two",
			},
		}
		addEndpoint(sr2, ep2)
		c1 := &domain.ExtensionCredentials{
			ExtensionCredentialsID: domain.ExtensionCredentialsID{
				ID: "testcredentials-001",
			},
			Scope:    domain.ECSGlobal,
			Default:  true,
			Projects: []string{},
			Users:    []string{},
			Configuration: map[string]string{
				"cred-001-config-one": "cred-001-value-one",
				"cred-001-config-two": "cred-001-value-two",
			},
		}
		addCredentials(sr1, c1)
		c2 := &domain.ExtensionCredentials{
			ExtensionCredentialsID: domain.ExtensionCredentialsID{
				ID: "testcredentials-002",
			},
			Scope:    domain.ECSUser,
			Default:  false,
			Projects: []string{"project-one", "project-two"},
			Users:    []string{"user-one", "user-two"},
			Configuration: map[string]string{
				"cred-002-config-one": "cred-002-value-one",
				"cred-002-config-two": "cred-002-value-two",
			},
		}
		addCredentials(sr2, c2)

		er1In, err := registry.RegisterExtension(context.Background(), er1)
		assertError(t, err, nil)
		if d := cmp.Diff(er1, er1In); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}
		er1Out, err := registry.GetExtension(context.Background(), "testextension-one", true)
		assertError(t, err, nil)
		if d := cmp.Diff(er1In, er1Out); d != "" {
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
		er2 := newExtension(e2)
		s3 := &domain.ExtensionService{
			ExtensionServiceID: domain.ExtensionServiceID{
				ID: "testservice-003",
			},
			Resource:     "testresource-one",
			Category:     "testcategory-one",
			Description:  "Test service 003",
			AuthRequired: false,
			Configuration: map[string]string{
				"svc-003-config-one": "svc-003-value-one",
				"svc-003-config-two": "svc-003-value-two",
			},
		}
		sr3 := addService(er2, s3)
		s4 := &domain.ExtensionService{
			ExtensionServiceID: domain.ExtensionServiceID{
				ID: "testservice-004",
			},
			Resource:     "testresource-two",
			Category:     "testcategory-two",
			Description:  "Test service 004",
			AuthRequired: true,
			Configuration: map[string]string{
				"svc-004-config-one": "svc-004-value-one",
				"svc-004-config-two": "svc-004-value-two",
			},
		}
		sr4 := addService(er2, s4)
		ep3 := &domain.ExtensionEndpoint{
			ExtensionEndpointID: domain.ExtensionEndpointID{
				URL: "https://testendpoint-003.com",
			},
			Type: domain.EETExternal,
			Configuration: map[string]string{
				"ep-003-config-one": "svc-003-value-one",
				"ep-003-config-two": "svc-003-value-two",
			},
		}
		addEndpoint(sr3, ep3)
		ep4 := &domain.ExtensionEndpoint{
			ExtensionEndpointID: domain.ExtensionEndpointID{
				URL: "https://testendpoint-004.com",
			},
			Type: domain.EETInternal,
			Configuration: map[string]string{
				"ep-004-config-one": "svc-004-value-one",
				"ep-004-config-two": "svc-004-value-two",
			},
		}
		addEndpoint(sr4, ep4)
		c3 := &domain.ExtensionCredentials{
			ExtensionCredentialsID: domain.ExtensionCredentialsID{
				ID: "testcredentials-003",
			},
			Scope:    domain.ECSGlobal,
			Default:  true,
			Projects: []string{},
			Users:    []string{},
			Configuration: map[string]string{
				"cred-003-config-one": "cred-003-value-one",
				"cred-003-config-two": "cred-003-value-two",
			},
		}
		addCredentials(sr3, c3)
		c4 := &domain.ExtensionCredentials{
			ExtensionCredentialsID: domain.ExtensionCredentialsID{
				ID: "testcredentials-004",
			},
			Scope:    domain.ECSUser,
			Default:  false,
			Projects: []string{"project-one", "project-two"},
			Users:    []string{"user-two", "user-three"},
			Configuration: map[string]string{
				"cred-004-config-one": "cred-004-value-one",
				"cred-004-config-two": "cred-004-value-two",
			},
		}
		addCredentials(sr4, c4)

		er2In, err := registry.RegisterExtension(context.Background(), er2)
		assertError(t, err, nil)
		if d := cmp.Diff(er2, er2In); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}
		er2Out, err := registry.GetExtension(context.Background(), "testextension-two", true)
		assertError(t, err, nil)
		if d := cmp.Diff(er2In, er2Out); d != "" {
			t.Errorf("Unexpected Extension: %s", diff.PrintWantGot(d))
		}

		var epType domain.ExtensionEndpointType = domain.EETExternal
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

		qRes := []*domain.ExtensionAccessDescriptor{{
			Extension:            er1.Extension,
			ExtensionService:     sr1.ExtensionService,
			ExtensionEndpoint:    *ep1,
			ExtensionCredentials: c1,
		}}
		qOut, err := registry.RunExtensionAccessQuery(context.Background(), q)
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

		qRes = []*domain.ExtensionAccessDescriptor{{
			Extension:            er1.Extension,
			ExtensionService:     sr2.ExtensionService,
			ExtensionEndpoint:    *ep2,
			ExtensionCredentials: c2,
		}}
		qOut, err = registry.RunExtensionAccessQuery(context.Background(), q)
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

		qRes = []*domain.ExtensionAccessDescriptor{
			{
				Extension:            er1.Extension,
				ExtensionService:     sr2.ExtensionService,
				ExtensionEndpoint:    *ep2,
				ExtensionCredentials: c2,
			},
			{
				Extension:            er2.Extension,
				ExtensionService:     sr4.ExtensionService,
				ExtensionEndpoint:    *ep4,
				ExtensionCredentials: c4,
			},
		}
		qOut, err = registry.RunExtensionAccessQuery(context.Background(), q)
		assertError(t, err, nil)
		if d := cmp.Diff(qRes, qOut); d != "" {
			t.Errorf("Unexpected Query Results: %s", diff.PrintWantGot(d))
		}

	})
}
