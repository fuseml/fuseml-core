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
	"time"

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
	"knative.dev/pkg/apis"
	kn "knative.dev/pkg/apis/duck/v1beta1"
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
		ctx, b, logsOutput := initBackend(t)

		w := workflow.Workflow{}
		readYaml(t, fuseMLWorkflow, &w)

		err := b.CreateWorkflow(ctx, &w)
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
		ctx, b, _ := initBackend(t)

		w := workflow.Workflow{}
		readYaml(t, fuseMLWorkflow, &w)

		err := b.CreateWorkflow(ctx, &w)
		if err != nil {
			t.Fatal(err)
		}
		got := b.CreateWorkflow(ctx, &w)
		assertError(t, got, errWorkflowExists)
	})
}

func TestCreateWorkflowRun(t *testing.T) {
	ctx, b, _ := initBackend(t)

	w := workflow.Workflow{}
	readYaml(t, fuseMLWorkflow, &w)

	err := b.CreateWorkflow(ctx, &w)
	if err != nil {
		t.Fatal(err)
	}

	cs := &domain.Codeset{
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

func TestListWorkflowRuns(t *testing.T) {
	t.Run("all", func(t *testing.T) {
		ctx, b, _ := initBackend(t)

		w := workflow.Workflow{}
		readYaml(t, fuseMLWorkflow, &w)

		err := b.CreateWorkflow(ctx, &w)
		if err != nil {
			t.Fatal(err)
		}

		runStatus := "Unknown"
		want := []*workflow.WorkflowRun{}

		for i := 1; i < 3; i++ {
			cs := createCodeset(t, i, i)
			runName := fmt.Sprintf("%s-%d", w.Name, i)
			runURL := "http://tekton.test/#/namespaces/test-namespace/pipelineruns/" + runName
			runStartTime := metav1.Now()
			b.createTestWorkflowRun(ctx, t, w.Name, &cs, runName, runStatus, runStartTime)
			runCsInputValue := fmt.Sprintf("%s:main", cs.URL)
			startTime := runStartTime.Format(time.RFC3339)
			completionTime := metav1.NewTime(runStartTime.Time.Add(time.Minute)).Format(time.RFC3339)
			want = append(want, &workflow.WorkflowRun{
				Name:           &runName,
				WorkflowRef:    &w.Name,
				Inputs:         []*workflow.WorkflowRunInput{{Input: w.Inputs[0], Value: &runCsInputValue}, {Input: w.Inputs[1], Value: w.Inputs[1].Default}},
				Outputs:        []*workflow.WorkflowRunOutput{{Output: w.Outputs[0]}},
				StartTime:      &startTime,
				CompletionTime: &completionTime,
				Status:         &runStatus,
				URL:            &runURL,
			})
		}

		filter := domain.WorkflowRunFilter{}
		got, err := b.ListWorkflowRuns(ctx, &w, filter)
		if err != nil {
			t.Fatalf("Failed to list PipelineRun: %s", err)
		}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected WorkflowRun: %s", diff.PrintWantGot(d))
		}
	})

	t.Run("filter by label", func(t *testing.T) {
		ctx, b, _ := initBackend(t)

		w := workflow.Workflow{}
		readYaml(t, fuseMLWorkflow, &w)

		err := b.CreateWorkflow(ctx, &w)
		if err != nil {
			t.Fatal(err)
		}

		runStatus := "Unknown"
		codesets := []domain.Codeset{}
		wants := []*workflow.WorkflowRun{}
		for i := 0; i < 2; i++ {
			cs := createCodeset(t, i, i)
			runName := fmt.Sprintf("%s-%d", w.Name, i)
			runURL := "http://tekton.test/#/namespaces/test-namespace/pipelineruns/" + runName
			runStartTime := metav1.Now()
			b.createTestWorkflowRun(ctx, t, w.Name, &cs, runName, runStatus, runStartTime)
			runCsInputValue := fmt.Sprintf("%s:main", cs.URL)
			startTime := runStartTime.Format(time.RFC3339)
			completionTime := metav1.NewTime(runStartTime.Time.Add(time.Minute)).Format(time.RFC3339)
			wants = append(wants, &workflow.WorkflowRun{
				Name:           &runName,
				WorkflowRef:    &w.Name,
				Inputs:         []*workflow.WorkflowRunInput{{Input: w.Inputs[0], Value: &runCsInputValue}, {Input: w.Inputs[1], Value: w.Inputs[1].Default}},
				Outputs:        []*workflow.WorkflowRunOutput{{Output: w.Outputs[0]}},
				StartTime:      &startTime,
				CompletionTime: &completionTime,
				Status:         &runStatus,
				URL:            &runURL,
			})
			codesets = append(codesets, cs)
		}

		filterNil := domain.WorkflowRunFilter{ByLabel: nil}
		want := wants
		got, err := b.ListWorkflowRuns(ctx, &w, filterNil)
		if err != nil {
			t.Fatalf("Failed to list WorkflowRun: %s", err)
		}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected WorkflowRun: %s", diff.PrintWantGot(d))
		}

		filterEmpty := domain.WorkflowRunFilter{ByLabel: []string{}}
		want = wants
		got, err = b.ListWorkflowRuns(ctx, &w, filterEmpty)
		if err != nil {
			t.Fatalf("Failed to list WorkflowRun: %s", err)
		}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected WorkflowRun: %s", diff.PrintWantGot(d))
		}

		filterNoResult := domain.WorkflowRunFilter{ByLabel: []string{fmt.Sprintf("%s=%s", LabelCodesetName, "do-no-exist")}}
		want = []*workflow.WorkflowRun{}
		got, err = b.ListWorkflowRuns(ctx, &w, filterNoResult)
		if err != nil {
			t.Fatalf("Failed to list WorkflowRun: %s", err)
		}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected WorkflowRun: %s", diff.PrintWantGot(d))
		}

		for i := 0; i < len(codesets); i++ {
			filterCodesetName := domain.WorkflowRunFilter{ByLabel: []string{fmt.Sprintf("%s=%s", LabelCodesetName, codesets[i].Name)}}
			want := []*workflow.WorkflowRun{wants[i]}
			got, err := b.ListWorkflowRuns(ctx, &w, filterCodesetName)
			if err != nil {
				t.Fatalf("Failed to list WorkflowRun: %s", err)
			}

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("Unexpected WorkflowRun: %s", diff.PrintWantGot(d))
			}
		}

		for i := 0; i < len(codesets); i++ {
			filterCodesetProject := domain.WorkflowRunFilter{ByLabel: []string{fmt.Sprintf("%s=%s", LabelCodesetProject, codesets[i].Project)}}
			want := []*workflow.WorkflowRun{wants[i]}
			got, err := b.ListWorkflowRuns(ctx, &w, filterCodesetProject)
			if err != nil {
				t.Fatalf("Failed to list WorkflowRun: %s", err)
			}

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("Unexpected WorkflowRun: %s", diff.PrintWantGot(d))
			}
		}

		for i := 0; i < len(codesets); i++ {
			filterCodesetNameProject := domain.WorkflowRunFilter{ByLabel: []string{fmt.Sprintf("%s=%s", LabelCodesetName, codesets[i].Name),
				fmt.Sprintf("%s=%s", LabelCodesetProject, codesets[i].Project)}}
			want := []*workflow.WorkflowRun{wants[i]}
			got, err := b.ListWorkflowRuns(ctx, &w, filterCodesetNameProject)
			if err != nil {
				t.Fatalf("Failed to list WorkflowRun: %s", err)
			}

			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("Unexpected WorkflowRun: %s", diff.PrintWantGot(d))
			}
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		ctx, b, _ := initBackend(t)

		w := workflow.Workflow{}
		readYaml(t, fuseMLWorkflow, &w)

		err := b.CreateWorkflow(ctx, &w)
		if err != nil {
			t.Fatal(err)
		}

		runsStatus := []string{"Unknown", "PipelineRunCancelled", "Succeeded", "Running"}
		wants := []*workflow.WorkflowRun{}
		for i := 0; i < len(runsStatus); i++ {
			cs := createCodeset(t, i, i)
			runName := fmt.Sprintf("%s-%d", w.Name, i)
			runURL := "http://tekton.test/#/namespaces/test-namespace/pipelineruns/" + runName
			runStartTime := metav1.Now()
			runStatus := runsStatus[i]
			b.createTestWorkflowRun(ctx, t, w.Name, &cs, runName, runStatus, runStartTime)
			startTime := runStartTime.Format(time.RFC3339)
			var completionTime *string
			if runStatus != "Running" {
				time := metav1.NewTime(runStartTime.Time.Add(time.Minute)).Format(time.RFC3339)
				completionTime = &time
			}
			runCsInputValue := fmt.Sprintf("%s:main", cs.URL)
			status := pipelineReasonToWorkflowStatus(runStatus)
			wants = append(wants, &workflow.WorkflowRun{
				Name:           &runName,
				WorkflowRef:    &w.Name,
				Inputs:         []*workflow.WorkflowRunInput{{Input: w.Inputs[0], Value: &runCsInputValue}, {Input: w.Inputs[1], Value: w.Inputs[1].Default}},
				Outputs:        []*workflow.WorkflowRunOutput{{Output: w.Outputs[0]}},
				StartTime:      &startTime,
				CompletionTime: completionTime,
				Status:         &status,
				URL:            &runURL,
			})
		}

		filterNil := domain.WorkflowRunFilter{ByStatus: nil}
		want := wants
		got, err := b.ListWorkflowRuns(ctx, &w, filterNil)
		if err != nil {
			t.Fatalf("Failed to list WorkflowRun: %s", err)
		}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected WorkflowRun: %s", diff.PrintWantGot(d))
		}

		filterEmpty := domain.WorkflowRunFilter{ByStatus: []string{}}
		want = wants
		got, err = b.ListWorkflowRuns(ctx, &w, filterEmpty)
		if err != nil {
			t.Fatalf("Failed to list WorkflowRun: %s", err)
		}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected WorkflowRun: %s", diff.PrintWantGot(d))
		}

		filterNoResult := domain.WorkflowRunFilter{ByStatus: []string{"Timeout"}}
		want = []*workflow.WorkflowRun{}
		got, err = b.ListWorkflowRuns(ctx, &w, filterNoResult)
		if err != nil {
			t.Fatalf("Failed to list WorkflowRun: %s", err)
		}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected WorkflowRun: %s", diff.PrintWantGot(d))
		}

		for i := 0; i < len(runsStatus); i++ {
			filterStatus := domain.WorkflowRunFilter{ByStatus: []string{pipelineReasonToWorkflowStatus(runsStatus[i])}}
			want := []*workflow.WorkflowRun{wants[i]}
			got, err := b.ListWorkflowRuns(ctx, &w, filterStatus)
			if err != nil {
				t.Fatalf("Failed to list WorkflowRun: %s", err)
			}
			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("Unexpected WorkflowRun: %s", diff.PrintWantGot(d))
			}
		}

		filterMultipleStatus := domain.WorkflowRunFilter{}
		for i := 0; i < len(runsStatus); i++ {
			filterMultipleStatus.ByStatus = append(filterMultipleStatus.ByStatus, pipelineReasonToWorkflowStatus(runsStatus[i]))
		}
		want = wants
		got, err = b.ListWorkflowRuns(ctx, &w, filterMultipleStatus)
		if err != nil {
			t.Fatalf("Failed to list WorkflowRun: %s", err)
		}
		if d := cmp.Diff(want, got); d != "" {
			t.Errorf("Unexpected WorkflowRun: %s", diff.PrintWantGot(d))
		}

	})

}

func TestCreateListener(t *testing.T) {
	t.Run("new listener", func(t *testing.T) {

		ctx, b, logsOutput := initBackend(t)

		originalLogger := b.logger
		discardOutput := bytes.Buffer{}
		discardLogs := log.New(&discardOutput, "[tekton-test] ", log.Ltime)
		w := workflow.Workflow{}
		readYaml(t, fuseMLWorkflow, &w)

		b.logger = discardLogs
		err := b.CreateWorkflow(ctx, &w)
		b.logger = originalLogger
		if err != nil {
			t.Fatal(err)
		}

		url, err := b.CreateListener(ctx, w.Name, false)
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
		ctx, b, logsOutput := initBackend(t)

		originalLogger := b.logger
		discardOutput := bytes.Buffer{}
		discardLogs := log.New(&discardOutput, "[tekton-test] ", log.Ltime)
		w := workflow.Workflow{}
		readYaml(t, fuseMLWorkflow, &w)

		b.logger = discardLogs
		err := b.CreateWorkflow(ctx, &w)
		if err != nil {
			t.Fatal(err)
		}

		_, err = b.CreateListener(ctx, w.Name, false)
		b.logger = originalLogger
		if err != nil {
			t.Fatalf("Failed to create listener for workflow %q: %s", w.Name, err)
		}

		url, err := b.CreateListener(ctx, w.Name, false)
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

func initBackend(t *testing.T) (context context.Context, backend *WorkflowBackend, logsOutput *bytes.Buffer) {
	t.Helper()

	context, _ = rtesting.SetupFakeContext(t)
	logsOutput = &bytes.Buffer{}
	logs := log.New(logsOutput, "", 0)

	backend = fakeNewWorkflowBackend(context, logs, t, testNamespace)

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

func fakeNewWorkflowBackend(context context.Context, logger *log.Logger, t *testing.T, namespace string) *WorkflowBackend {
	t.Helper()

	clients := newFakeClients(context, t, namespace)
	return &WorkflowBackend{logger, "http://tekton.test", namespace, clients}
}

func createCodeset(t *testing.T, nameID, projectID int) domain.Codeset {
	t.Helper()

	name := fmt.Sprintf("mlflow-app-%d", nameID)
	project := fmt.Sprintf("workspace-%d", projectID)
	url := fmt.Sprintf("http://gitea.10.160.5.140.nip.io/%s/%s.git", name, project)
	return domain.Codeset{Name: name, Project: project, URL: url}
}

func (b WorkflowBackend) createTestWorkflowRun(ctx context.Context, t *testing.T, workflow string, cs *domain.Codeset,
	runName string, status string, startTime metav1.Time) {
	t.Helper()
	err := b.CreateWorkflowRun(ctx, workflow, cs)
	if err != nil {
		t.Fatalf("Failed to create workflow run %q: %s", workflow, err)
	}

	// the fake pipeline run client does not generate a name for the pipeline run, in that
	// way it is not possible to create multiple pipeline runs as they conflict on their name ("").
	// To get around that, create the workflow run then change its name by recreating it with a
	// another name
	prun, err := b.tektonClients.PipelineRunClient.Get(ctx, "", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get pipeline run: %s", err)
	}
	prun.ObjectMeta.Name = runName
	prun.Status.Conditions = kn.Conditions{apis.Condition{Reason: status}}
	completionTime := metav1.NewTime(startTime.Time.Add(time.Minute))
	prun.Status.StartTime = &startTime
	if status != "Running" {
		prun.Status.CompletionTime = &completionTime
	}
	err = b.tektonClients.PipelineRunClient.Delete(ctx, "", metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Failed to get delete run: %s", err)
	}
	_, err = b.tektonClients.PipelineRunClient.Create(ctx, prun, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create pipeline run: %s", err)
	}
}
