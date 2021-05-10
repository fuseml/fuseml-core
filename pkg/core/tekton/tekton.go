package tekton

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"knative.dev/pkg/apis"

	"github.com/fuseml/fuseml-core/gen/codeset"
	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/core/tekton/builder"
)

const (
	// ErrWorkflowExists represents a error returned when creating a workflow with the same name as one
	// that already exists
	ErrWorkflowExists = WorkflowBackendErr("workflow already exists")
)

// WorkflowBackendErr are expected errors returned from the WorkflowBackend
type WorkflowBackendErr string

// WorkflowBackend implements the FuseML WorkflowBackend interface for tekton
type WorkflowBackend struct {
	namespace     string
	tektonClients *clients
}

// NewWorkflowBackend initializes Tekton backend
func NewWorkflowBackend(namespace string) (*WorkflowBackend, error) {
	clients, err := newClients(namespace)
	if err != nil {
		return nil, fmt.Errorf("error initializing tekton workflow backend: %w", err)
	}
	return &WorkflowBackend{namespace, clients}, nil
}

// CreateWorkflow receives a FuseML workflow and creates a Tekton pipeline from it
func (w *WorkflowBackend) CreateWorkflow(ctx context.Context, logger *log.Logger, workflow *workflow.Workflow) error {
	pipeline := generatePipeline(*workflow, w.namespace)
	logger.Printf("Creating tekton pipeline for workflow: %s...", workflow.Name)
	_, err := w.tektonClients.PipelineClient.Create(ctx, pipeline, metav1.CreateOptions{})
	if err != nil {
		if k8serr.IsAlreadyExists(err) {
			return ErrWorkflowExists
		}
		return fmt.Errorf("error creating tekton pipeline for workflow %q: %w", workflow.Name, err)
	}

	return nil
}

// CreateWorkflowRun creates a PipelineRun with its default values for received workflow and codeset
func (w *WorkflowBackend) CreateWorkflowRun(ctx context.Context, workflowName string, codeset codeset.Codeset) error {
	pipeline, err := w.tektonClients.PipelineClient.Get(ctx, workflowName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting tekton pipeline %q: %w", workflowName, err)
	}

	pipelineRun, err := generatePipelineRun(pipeline, codeset)
	if err != nil {
		return fmt.Errorf("error generating tekton pipeline run for workflow %q: %w", workflowName, err)
	}

	_, err = w.tektonClients.PipelineRunClient.Create(ctx, pipelineRun, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("error creating tekton pipeline run %q: %w", pipelineRun.Name, err)
	}
	return nil
}

// CreateListener creates tekton resources required to have a listener ready for triggering the pipeline
func (w *WorkflowBackend) CreateListener(ctx context.Context, logger *log.Logger, workflowName string, wait bool) (string, error) {

	pipeline, err := w.tektonClients.PipelineClient.Get(ctx, workflowName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("error getting tekton pipeline %q: %w", workflowName, err)
	}

	triggerTemplate := generateTriggerTemplate(pipeline)
	_, err = w.tektonClients.TriggerTemplateClient.Get(ctx, workflowName, metav1.GetOptions{})
	if err != nil {
		if !k8serr.IsNotFound(err) {
			return "", fmt.Errorf("error getting tekton trigger template %q: %w", workflowName, err)
		}
		logger.Printf("Creating tekton trigger template for workflow: %s...", workflowName)
		_, err := w.tektonClients.TriggerTemplateClient.Create(ctx, triggerTemplate, metav1.CreateOptions{})
		if err != nil {
			return "", fmt.Errorf("error creating tekton trigger template %q: %w", workflowName, err)
		}
	}

	triggerBinding := generateTriggerBinding(triggerTemplate)
	_, err = w.tektonClients.TriggerBindingClient.Get(ctx, workflowName, metav1.GetOptions{})
	if err != nil {
		if !k8serr.IsNotFound(err) {
			return "", fmt.Errorf("error getting tekton trigger binding %q: %w", workflowName, err)
		}
		logger.Printf("Creating tekton trigger binding for workflow: %s...", workflowName)
		_, err := w.tektonClients.TriggerBindingClient.Create(ctx, triggerBinding, metav1.CreateOptions{})
		if err != nil {
			return "", fmt.Errorf("error creating tekton trigger binding %q: %w", workflowName, err)
		}
	}

	eventListener := generateEventListener(triggerTemplate, triggerBinding)
	el, err := w.tektonClients.EventListenerClient.Get(ctx, workflowName, metav1.GetOptions{})
	if err != nil {
		if !k8serr.IsNotFound(err) {
			return "", fmt.Errorf("error getting tekton event listener %q: %w", workflowName, err)
		}
		logger.Printf("Creating tekton event listener for workflow: %s...", workflowName)
		el, err = w.tektonClients.EventListenerClient.Create(ctx, eventListener, metav1.CreateOptions{})
		if err != nil {
			return "", fmt.Errorf("error creating tekton event listener %q: %w", workflowName, err)
		}
	}

	if wait {
		interval := 1 * time.Second
		timeout := 1 * time.Minute
		if err := waitFor(w.eventListenerReady(ctx, el.Name), interval, timeout); err != nil {
			return "", fmt.Errorf("event listener %q did not get ready in the expected time: %w", el.Name, err)
		}

		el, _ = w.tektonClients.EventListenerClient.Get(ctx, workflowName, metav1.GetOptions{})
		return el.Status.Address.URL.String(), nil
	}
	return fmt.Sprintf("http://el-%s.%s.svc.cluster.local:8080", workflowName, w.namespace), nil
}

func (e WorkflowBackendErr) Error() string {
	return string(e)
}

func waitFor(waitFunc wait.ConditionFunc, interval, timeout time.Duration) error {
	return wait.PollImmediate(interval, timeout, waitFunc)
}

func (w *WorkflowBackend) eventListenerReady(ctx context.Context, name string) wait.ConditionFunc {
	return func() (bool, error) {
		el, err := w.tektonClients.EventListenerClient.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		// No conditions have been set yet
		if len(el.Status.Conditions) == 0 {
			return false, nil
		}
		if el.Status.GetCondition(apis.ConditionType(appsv1.DeploymentAvailable)) == nil {
			return false, nil
		}
		for _, cond := range el.Status.Conditions {
			if cond.Status != corev1.ConditionTrue {
				return false, nil
			}
		}
		if el.Status.Address.URL == nil {
			return false, nil
		}
		return true, nil
	}
}

func generatePipeline(w workflow.Workflow, namespace string) *v1beta1.Pipeline {
	resolver := newVariablesResolver()
	pb := builder.NewPipelineBuilder(w.Name, namespace)
	pb.Meta(builder.Label("fuseml/generated-from", w.Name))
	pb.Description(*w.Description)

	// process the FuseML workflow inputs
	for _, input := range w.Inputs {
		// an input of the 'codeset' type means a git repository for tekton.
		// adds a workspace, resource and the clone task to the pipeline, also
		// also add 'codeset-name' and 'codeset-version' parameters to the tekton
		// pipeline which represents codeset.name and codeset.version in a FuseML
		// workflow.
		if *input.Type == inputTypeCodeset {
			pb.Workspace(codesetWorkspaceName, false)
			pb.Resource("source-repo", "git", false)
			pb.Task("clone", cloneTaskName, nil, map[string]string{codesetWorkspaceName: codesetWorkspaceName},
				map[string]string{"source-repo": "source-repo"})
			pb.Param(codesetNameParam, "Reference to the codeset (git project)")
			resolver.addReference(fmt.Sprintf("inputs.%s.name", *input.Name), fmt.Sprintf("$(params.%s)", codesetNameParam))
			pb.ParamWithDefaultValue(codesetVersionParam, "Codeset version (git revision)", "main")
			resolver.addReference(fmt.Sprintf("inputs.%s.version", *input.Name), fmt.Sprintf("$(params.%s)", codesetVersionParam))
		} else {
			pb.ParamWithDefaultValue(*input.Name, *input.Description, *input.Default)
			resolver.addReference(fmt.Sprintf("inputs.%s", *input.Name), fmt.Sprintf("$(params.%s)", *input.Name))
		}
	}

	// process the FuseML workflow steps
STEPS:
	for _, step := range w.Steps {
		for _, output := range step.Outputs {
			// image type output parameters serve as input for the tekton builder-prep and builder tasks (kaniko).
			// Note that the builder represents two tasks in tekton, the first one (builder-prep) uses the image defined
			// on the step to provide a Dockerfile that will be used by the following task (kaniko) to build the image
			// expected by the step output.
			if output.Image != nil {
				var dockerfile string
				if output.Image.Dockerfile != nil {
					dockerfile = resolver.resolve(*output.Image.Dockerfile)
					codesetPath := getInputCodesetPath(step.Inputs)
					// in FuseML workflow the dockerfile path is a full path including the codeset.path, however
					// the kaniko task mounts the codeset at workingDir and expects the dockerfile path to be referenced
					// from it. So, remove the codeset.path from image.dockerfile input
					if codesetPath != "" {
						dockerfile = strings.Replace(dockerfile, fmt.Sprintf("%s/", codesetPath), "", 1)
					}
				}
				prepTaskName := fmt.Sprintf("%s-prep", *step.Name)
				pb.Task(prepTaskName, builderPrepTaskName, map[string]string{"IMAGE": resolver.resolve(*step.Image),
					"DOCKERFILE": dockerfile}, map[string]string{codesetWorkspaceName: codesetWorkspaceName}, nil)
				pb.Task(*step.Name, builderTaskName, map[string]string{"IMAGE": resolver.resolve(*output.Image.Name),
					"DOCKERFILE": fmt.Sprintf("$(tasks.%s.results.DOCKERFILE-PATH)", prepTaskName)},
					map[string]string{codesetWorkspaceName: codesetWorkspaceName}, nil)
				resolver.addReference(fmt.Sprintf("steps.%s.outputs.%s", *step.Name, *output.Name), *output.Image.Name)
				continue STEPS
			}
			// if the step output is the workflow output, map it to a PipelineResult in tekton
			if wo := stepOutputIsWorkflowOutput(output, w.Outputs); wo != nil {
				pb.Result(*wo.Name, *wo.Description, fmt.Sprintf("$(tasks.%s.results.%s)", *step.Name, *output.Name))
			}
		}

		// if the workflow step is not a pipeline task that references an existing TektonTask,
		// build the task spec from the FuseML workflow step.
		// generates a v1beta1.TaskSpec from a workflow.WorkflowStep
		taskSpec := toTektonTaskSpec(*step)
		taskWs := make(map[string]string)
		taskParams := make(map[string]string)
		for _, input := range step.Inputs {
			// if the step has a codeset as input add the workspace
			// TODO: for now it only supports 1 workspace
			if input.Codeset != nil {
				taskWs[taskSpec.Workspaces[0].Name] = codesetWorkspaceName
			} else {
				taskParams[*input.Name] = resolver.resolve(*input.Value)
			}
		}
		// if image is parametrized add 'IMAGE' param, resolving it
		if strings.Contains(*step.Image, "{{") {
			// The kubernetes nodes are unable to resolve the local FuseML registry
			// (registry.fuseml-registry), in that way, when the step uses an image
			// from the local FuseML registry, replace registry.fuseml-registry with
			// 127.0.0.1:30500
			image := resolver.resolve(*step.Image)
			if strings.HasPrefix(image, fuseMLRegistry) {
				image = strings.Replace(image, fuseMLRegistry, fuseMLRegistryLocal, 1)
			}
			taskParams[imageParamName] = image
		}
		pb.Task(*step.Name, taskSpec, taskParams, taskWs, nil)
	}
	return &pb.Pipeline
}

func generatePipelineRun(p *v1beta1.Pipeline, codeset codeset.Codeset) (*v1beta1.PipelineRun, error) {
	codesetVersion := "main"
	prb := builder.NewPipelineRunBuilder(fmt.Sprintf("%s%s-%s-", pipelineRunPrefix, codeset.Project, codeset.Name))

	for _, param := range p.Spec.Params {
		if param.Default == nil {
			switch param.Name {
			case codesetNameParam:
				prb.Param(param.Name, codeset.Name)
			case codesetVersion:
				prb.Param(param.Name, codesetVersion)
			default:
				return nil, fmt.Errorf("pipeline run failed: could not set parameter value for %q", param.Name)
			}
		} else {
			prb.Param(param.Name, param.Default.StringVal)
		}
	}

	prb.Meta(builder.Label("fuseml/codeset-name", codeset.Name), builder.Label("fuseml/codeset-version", codesetVersion))
	prb.ServiceAccount(pipelineRunServiceAccount)
	prb.PipelineRef(p.Name)
	for _, ws := range p.Spec.Workspaces {
		prb.Workspace(ws.Name, workspaceAccessMode, workspaceSize)
	}

	for _, res := range p.Spec.Resources {
		if res.Type == "git" {
			prb.ResourceGit(res.Name, *codeset.URL, codesetVersion)
		}
	}
	return &prb.PipelineRun, nil
}

func generateTriggerTemplate(p *v1beta1.Pipeline) *v1alpha1.TriggerTemplate {
	ttb := builder.NewTriggerTemplateBuilder(p.Name, p.Namespace)
	prb := builder.NewPipelineRunBuilder(pipelineRunPrefix)
	resolver := newVariablesResolver()
	for _, param := range p.Spec.Params {
		if param.Default != nil {
			ttb.ParamWithDefaultValue(param.Name, param.Description, param.Default.StringVal)
		} else {
			ttb.Param(param.Name, param.Description)
		}
		// if there is a codeset paramter we also need to add the codeset-url as paramter to
		// the template
		resolver.addReference(param.Name, fmt.Sprintf("$(tt.params.%s)", param.Name))
		if param.Name == codesetNameParam {
			ttb.Param(codesetURLParam, "The codeset URL (git repository URL)")
			resolver.addReference(codesetURLParam, fmt.Sprintf("$(tt.params.%s)", codesetURLParam))
			prb.GenerateName(fmt.Sprintf("%s%s-", pipelineRunPrefix, resolver.resolve(codesetNameParam)))
			prb.Meta(builder.Label("fuseml/codeset-name", resolver.resolve(codesetNameParam)))
		} else if param.Name == codesetVersionParam {
			prb.Meta(builder.Label("fuseml/codeset-version", resolver.resolve(codesetVersionParam)))
		}

		prb.Param(param.Name, resolver.resolve(param.Name))
	}

	for _, ws := range p.Spec.Workspaces {
		prb.Workspace(ws.Name, workspaceAccessMode, workspaceSize)
	}

	for _, res := range p.Spec.Resources {
		if res.Type == "git" {
			prb.ResourceGit(res.Name, resolver.resolve(codesetURLParam), resolver.resolve(codesetVersionParam))
		}
	}

	prb.ServiceAccount(pipelineRunServiceAccount)
	prb.PipelineRef(p.Name)

	prBytes, err := json.Marshal(prb.PipelineRun)
	if err != nil {
		log.Fatalf("Error marshalling PipelineRun: %s", err)
	}
	ttb.ResourceTemplate(runtime.RawExtension{Raw: prBytes})

	return &ttb.TriggerTemplate
}

func generateTriggerBinding(template *v1alpha1.TriggerTemplate) *v1alpha1.TriggerBinding {
	webhookParamsMap := map[string]string{
		codesetNameParam:    "$(body.repository.name)",
		codesetVersionParam: "$(body.commits[0].id)",
		codesetURLParam:     "$(body.repository.clone_url)",
	}

	tbb := builder.NewTriggerBindingBuilder(template.Name, template.Namespace)

	for _, param := range template.Spec.Params {
		if v, ok := webhookParamsMap[param.Name]; ok {
			tbb.Param(param.Name, v)
		}
	}
	return &tbb.TriggerBinding
}

func generateEventListener(template *v1alpha1.TriggerTemplate, binding *v1alpha1.TriggerBinding) *v1alpha1.EventListener {
	elb := builder.NewEventListenerBuilder(template.Name, template.Namespace)
	elb.ServiceAccount(pipelineRunServiceAccount)
	elb.TriggerBinding(template.Name, binding.Name)
	return &elb.EventListener
}

func toTektonTaskSpec(step workflow.WorkflowStep) v1beta1.TaskSpec {
	tb := builder.NewTaskSpecBuilder(*step.Name, *step.Image, stepDefaultCmd)

	for _, input := range step.Inputs {
		// if there is a codeset as input, add workspace to the task and
		// set its working directory to codeset.path
		if input.Codeset != nil {
			tb.WorkspaceWithMountPath(codesetWorkspaceName, *input.Codeset.Path)
			tb.WorkingDir(*input.Codeset.Path)
		} else {
			// else add it as a parameter to the tekton task
			tb.Param(*input.Name)
			// make the inputs also available as env variable (with the
			// FUSEML_ prefix) so that the container can use them
			tb.Env(fmt.Sprintf("%s%s", inputsVarPrefix, strings.ToUpper(*input.Name)),
				fmt.Sprintf("$(params.%s)", *input.Name))
		}
	}

	// if image is parameterized reference it as a task parameter to be able
	// to receive its value from a task output
	if strings.Contains(*step.Image, "{{") {
		tb.ParamWithDescription(imageParamName, "Name (reference) of the image to run")
		tb.Image(fmt.Sprintf("$(params.%s)", imageParamName))
	}

	// a workflow output that is not an image represents a result in a tekton task.
	// also adds an environment variable with the variable name so that the container
	// can set the task output
	for _, output := range step.Outputs {
		if output.Image == nil {
			tb.Result(*output.Name)
			tb.Env(stepOutputVarName, *output.Name)
		}
	}

	// load environment variables
	for _, stepEnv := range step.Env {
		tb.Env(*stepEnv.Name, *stepEnv.Value)
	}

	return tb.TaskSpec
}

func stepOutputIsWorkflowOutput(stepOutput *workflow.WorkflowStepOutput,
	workflowOutput []*workflow.WorkflowOutput) *workflow.WorkflowOutput {
	for _, wo := range workflowOutput {
		if *wo.Name == *stepOutput.Name {
			return wo
		}
	}
	return nil
}

func getInputCodesetPath(inputs []*workflow.WorkflowStepInput) string {
	for _, input := range inputs {
		if input.Codeset != nil {
			return *input.Codeset.Path
		}
	}
	return ""
}
