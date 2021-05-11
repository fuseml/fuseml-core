package tekton

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	fakepipelineclient "github.com/tektoncd/pipeline/pkg/client/injection/client/fake"
	"github.com/tektoncd/pipeline/test/diff"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	faketriggersclient "github.com/tektoncd/triggers/pkg/client/injection/client/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	rtesting "knative.dev/pkg/reconciler/testing"

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/domain"
)

const (
	fuseMLWorkflow            = "testdata/workflow.yaml"
	wantTektonPipeline        = "testdata/tekton-pipeline.yaml"
	wantTektonPipelineRun     = "testdata/tekton-pipeline-run.yaml"
	wantTektonTriggerTemplate = "testdata/tekton-trigger-template.yaml"
	wantTektonTriggerBinding  = "testdata/tekton-trigger-binding.yaml"
	wantTektonEventListener   = "testdata/tekton-event-listener.yaml"
	testNamespace             = "test-namespace"
)

func TestCreateWorkflow(t *testing.T) {
	t.Run("new workflow", func(t *testing.T) {
		ctx, b, logs, logsOutput := initBackend(t)

		w := workflow.Workflow{}
		readYaml(t, fuseMLWorkflow, &w)

		err := b.CreateWorkflow(ctx, logs, &w)
		if err != nil {
			t.Fatal(err)
		}

		assertStrings(t, strings.TrimSuffix(logsOutput.String(), "\n"), "Creating tekton pipeline for workflow: mlflow-sklearn-e2e...")

		got, err := b.tektonClients.PipelineClient.Get(ctx, w.Name, metav1.GetOptions{})
		if err != nil {
			t.Fatalf("Failed to get Pipeline %q: %s", w.Name, err)
		}

		want := v1beta1.Pipeline{}
		readYaml(t, wantTektonPipeline, &want)

		ignoreTypeMetaField := cmpopts.IgnoreFields(v1beta1.Pipeline{}, "TypeMeta")
		if d := cmp.Diff(want, *got, cmpopts.SortSlices(func(x, y v1beta1.Param) bool { return x.Name < y.Name }), ignoreTypeMetaField); d != "" {
			t.Errorf("Unexpected Pipeline: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("existing workflow", func(t *testing.T) {
		ctx, b, logs, _ := initBackend(t)

		w := workflow.Workflow{}
		readYaml(t, fuseMLWorkflow, &w)

		err := b.CreateWorkflow(ctx, logs, &w)
		if err != nil {
			t.Fatal(err)
		}
		got := b.CreateWorkflow(ctx, logs, &w)
		assertError(t, got, errWorkflowExists)
	})
}

func TestCreateWorkflowRun(t *testing.T) {
	ctx, b, logs, _ := initBackend(t)

	w := workflow.Workflow{}
	readYaml(t, fuseMLWorkflow, &w)

	err := b.CreateWorkflow(ctx, logs, &w)
	if err != nil {
		t.Fatal(err)
	}

	cs := domain.Codeset{
		Name:    "mlflow-app-01",
		Project: "workspace",
		URL:     "http://gitea.10.160.5.140.nip.io/workspace/mlflow-app-01.git",
	}
	err = b.CreateWorkflowRun(ctx, w.Name, cs)
	if err != nil {
		t.Fatalf("Failed to create workflow run %q: %s", w.Name, err)
	}

	runs, err := b.tektonClients.PipelineRunClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list PipelineRuns: %s", err)
	}

	if len(runs.Items) > 1 {
		t.Errorf("Expected 1 PipelineRun, got %q", len(runs.Items))
	}

	got := runs.Items[0]
	want := v1beta1.PipelineRun{}
	readYaml(t, wantTektonPipelineRun, &want)

	ignoreStatusField := cmpopts.IgnoreFields(v1beta1.PipelineRunStatus{}, "Conditions", "PipelineRunStatusFields")
	if d := cmp.Diff(want, got, ignoreStatusField); d != "" {
		t.Errorf("Unexpected PipelineRun: %s", diff.PrintWantGot(d))
	}
}

func TestCreateListener(t *testing.T) {
	t.Run("new listener", func(t *testing.T) {

		ctx, b, logs, logsOutput := initBackend(t)

		discardOutput := bytes.Buffer{}
		discardLogs := log.New(&discardOutput, "[tekton-test] ", log.Ltime)
		w := workflow.Workflow{}
		readYaml(t, fuseMLWorkflow, &w)

		err := b.CreateWorkflow(ctx, discardLogs, &w)
		if err != nil {
			t.Fatal(err)
		}

		url, err := b.CreateListener(ctx, logs, w.Name, false)
		if err != nil {
			t.Fatalf("Failed to create listener for workflow %q: %s", w.Name, err)
		}

		assertStrings(t, url, fmt.Sprintf("http://el-%s.%s.svc.cluster.local:8080", w.Name, b.namespace))

		expectedLog := `Creating tekton trigger template for workflow: mlflow-sklearn-e2e...
Creating tekton trigger binding for workflow: mlflow-sklearn-e2e...
Creating tekton event listener for workflow: mlflow-sklearn-e2e...
`

		assertStrings(t, logsOutput.String(), expectedLog)

		gotTriggerTemplate, err := b.tektonClients.TriggerTemplateClient.Get(ctx, w.Name, metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}
		wantTriggerTemplate := v1alpha1.TriggerTemplate{}
		readYaml(t, wantTektonTriggerTemplate, &wantTriggerTemplate)

		// The resource template from a TriggerTemplate is stored as runtime.RawExtension (bytes)
		// so, ignore it here and compare it after unmarshalling it to a PipelineRun.
		ignoreTypeMetaField := cmpopts.IgnoreFields(v1alpha1.TriggerTemplate{}, "TypeMeta")
		ignoreRawTemplateField := cmpopts.IgnoreFields(runtime.RawExtension{}, "Raw")
		if d := cmp.Diff(wantTriggerTemplate, *gotTriggerTemplate, ignoreTypeMetaField, ignoreRawTemplateField); d != "" {
			t.Errorf("Unexpected TriggerTemplate: %s", diff.PrintWantGot(d))
		}

		gotPRTemplate := resourceTemplateToPipelineRun(t, gotTriggerTemplate.Spec.ResourceTemplates[0])
		wantPRTemplate := resourceTemplateToPipelineRun(t, wantTriggerTemplate.Spec.ResourceTemplates[0])
		if d := cmp.Diff(wantPRTemplate, gotPRTemplate); d != "" {
			t.Errorf("Unexpected TriggerTemplate ResourceTemplate: %s", diff.PrintWantGot(d))
		}

		gotTriggerBinding, err := b.tektonClients.TriggerBindingClient.Get(ctx, w.Name, metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}
		wantTriggerBinding := v1alpha1.TriggerBinding{}
		readYaml(t, wantTektonTriggerBinding, &wantTriggerBinding)

		ignoreTypeMetaField = cmpopts.IgnoreFields(v1alpha1.TriggerBinding{}, "TypeMeta")
		if d := cmp.Diff(wantTriggerBinding, *gotTriggerBinding, ignoreTypeMetaField); d != "" {
			t.Errorf("Unexpected TriggerBinding: %s", diff.PrintWantGot(d))
		}

		gotEventListener, err := b.tektonClients.EventListenerClient.Get(ctx, w.Name, metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}
		wantEventListener := v1alpha1.EventListener{}
		readYaml(t, wantTektonEventListener, &wantEventListener)

		ignoreTypeMetaField = cmpopts.IgnoreFields(v1alpha1.EventListener{}, "TypeMeta")
		if d := cmp.Diff(wantEventListener, *gotEventListener, ignoreTypeMetaField); d != "" {
			t.Errorf("Unexpected Event Listener: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("existing listener", func(t *testing.T) {
		ctx, b, logs, logsOutput := initBackend(t)

		discardOutput := bytes.Buffer{}
		discardLogs := log.New(&discardOutput, "[tekton-test] ", log.Ltime)
		w := workflow.Workflow{}
		readYaml(t, fuseMLWorkflow, &w)

		err := b.CreateWorkflow(ctx, discardLogs, &w)
		if err != nil {
			t.Fatal(err)
		}

		_, err = b.CreateListener(ctx, discardLogs, w.Name, false)
		if err != nil {
			t.Fatalf("Failed to create listener for workflow %q: %s", w.Name, err)
		}

		url, err := b.CreateListener(ctx, logs, w.Name, false)
		if err != nil {
			t.Fatalf("Failed to create listener for workflow %q: %s", w.Name, err)
		}

		assertStrings(t, url, fmt.Sprintf("http://el-%s.%s.svc.cluster.local:8080", w.Name, b.namespace))

		expectedLog := ""
		assertStrings(t, logsOutput.String(), expectedLog)
	})
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

func initBackend(t *testing.T) (context context.Context, backend *WorkflowBackend, logs *log.Logger, logsOutput *bytes.Buffer) {
	t.Helper()

	context, _ = rtesting.SetupFakeContext(t)
	backend = fakeNewWorkflowBackend(context, t, testNamespace)

	logsOutput = &bytes.Buffer{}
	logs = log.New(logsOutput, "", 0)
	return
}

func readYaml(t *testing.T, path string, obj interface{}) {
	t.Helper()

	wfFile, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read workflow file %s: %s", path, err)
	}
	err = yaml.Unmarshal(wfFile, &obj)
	if err != nil {
		t.Fatalf("Error unmarshiling workflow: %s", err)
	}
}

func resourceTemplateToPipelineRun(t *testing.T, resourceTemplate v1alpha1.TriggerResourceTemplate) v1beta1.PipelineRun {
	t.Helper()

	pr := v1beta1.PipelineRun{}
	err := json.Unmarshal(resourceTemplate.Raw, &pr)
	if err != nil {
		t.Fatal(err)
	}
	return pr
}

func newFakeClients(context context.Context, t *testing.T, namespace string) *clients {
	t.Helper()

	fc := &clients{}

	pcs := fakepipelineclient.Get(context)
	fc.PipelineClient = pcs.TektonV1beta1().Pipelines(namespace)
	fc.PipelineRunClient = pcs.TektonV1beta1().PipelineRuns(namespace)

	tcs := faketriggersclient.Get(context)
	fc.TriggerTemplateClient = tcs.TriggersV1alpha1().TriggerTemplates(namespace)
	fc.TriggerBindingClient = tcs.TriggersV1alpha1().TriggerBindings(namespace)
	fc.EventListenerClient = tcs.TriggersV1alpha1().EventListeners(namespace)
	return fc
}

func fakeNewWorkflowBackend(context context.Context, t *testing.T, namespace string) *WorkflowBackend {
	t.Helper()

	clients := newFakeClients(context, t, namespace)
	return &WorkflowBackend{namespace, clients}
}
