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

	"github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/fuseml/fuseml-core/pkg/core/config"
	"github.com/fuseml/fuseml-core/pkg/core/tekton/builder"
)

// CreatePipeline receives a FuseML workflow and creates a Tekton pipeline from it
func CreatePipeline(ctx context.Context, logger *log.Logger, workflow *workflow.Workflow) (*v1beta1.Pipeline, error) {
	tektonClients := newClients(config.FuseMlNamespace)

	pipeline := generatePipeline(*workflow, config.FuseMlNamespace)
	logger.Printf("Creating tekton pipeline for workflow: %s...", workflow.Name)
	return tektonClients.PipelineClient.Create(ctx, pipeline, metav1.CreateOptions{})
}

// CreateListener creates tekton resources required to have a listener ready for triggering the pipeline
func CreateListener(ctx context.Context, logger *log.Logger, name string) (string, error) {
	tektonClients := newClients(config.FuseMlNamespace)

	pipeline, err := tektonClients.PipelineClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	triggerTemplate := generateTriggerTemplate(pipeline)
	_, err = tektonClients.TriggerTemplateClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !k8serr.IsNotFound(err) {
			return "", err
		}
		logger.Printf("Creating tekton trigger template for workflow: %s...", name)
		_, err := tektonClients.TriggerTemplateClient.Create(ctx, triggerTemplate, metav1.CreateOptions{})
		if err != nil {
			return "", err
		}
	}

	triggerBinding := generateTriggerBinding(triggerTemplate)
	_, err = tektonClients.TriggerBindingClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !k8serr.IsNotFound(err) {
			return "", err
		}
		logger.Printf("Creating tekton trigger binding for workflow: %s...", name)
		_, err := tektonClients.TriggerBindingClient.Create(ctx, triggerBinding, metav1.CreateOptions{})
		if err != nil {
			return "", err
		}
	}

	eventListener := generateEventListener(triggerTemplate, triggerBinding)
	el, err := tektonClients.EventListenerClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !k8serr.IsNotFound(err) {
			return "", err
		}
		logger.Printf("Creating tekton event listener for workflow: %s...", name)
		el, err = tektonClients.EventListenerClient.Create(ctx, eventListener, metav1.CreateOptions{})
		if err != nil {
			return "", err
		}
	}

	if err := waitFor(eventListenerReady(ctx, el.Name, el.Namespace), 1*time.Second, 1*time.Minute); err != nil {
		return "", fmt.Errorf("event listener '%s' did not get ready in the expected time", el.Name)
	}

	el, _ = tektonClients.EventListenerClient.Get(ctx, name, metav1.GetOptions{})
	return el.Status.Address.URL.String(), nil
}

func waitFor(waitFunc wait.ConditionFunc, interval, timeout time.Duration) error {
	return wait.PollImmediate(interval, timeout, waitFunc)
}

func eventListenerReady(ctx context.Context, name, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		tektonClients := newClients(config.FuseMlNamespace)
		el, err := tektonClients.EventListenerClient.Get(ctx, name, metav1.GetOptions{})
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
			// image type output parameters serve as input for the tekton builder task (kaniko)
			if output.Image != nil {
				dockerfile := resolver.resolve(*output.Image.Dockerfile)
				codesetPath := getInputCodesetPath(step.Inputs)
				// in FuseML workflow the dockerfile path is a full path including the codeset.path, however
				// the kaniko task mounts the codeset at workingDir and expects the dockerfile path to be referenced
				// from it. So, remove the codeset.path from image.dockerfile input
				if codesetPath != "" {
					dockerfile = strings.Replace(dockerfile, fmt.Sprintf("%s/", codesetPath), "", 1)
				}
				pb.Task(*step.Name, builderTaskName, map[string]string{"IMAGE": resolver.resolve(*output.Image.Name),
					"DOCKERFILE": dockerfile}, map[string]string{codesetWorkspaceName: codesetWorkspaceName}, nil)
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
