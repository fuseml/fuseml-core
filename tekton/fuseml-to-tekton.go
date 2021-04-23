package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/fuseml/fuseml-core/gen/codeset"
	workflow "github.com/fuseml/fuseml-core/gen/workflow"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespace          = "fuseml-workloads"
	input_type_codeset = "codeset"
)

var (
	tektonClients *clients
	cs            codeset.Codeset
	resolver      *variablesResolver
)

func main() {
	data := []byte(`
id: 4850f99a-9df7-11eb-bad0-1e00e226385a
created: "2021-04-15T16:31:36+02:00"
name: mlflow-kfserving-e2e
description: |
  End-to-end pipeline template that takes in an MLFlow compatible codeset,
  runs the MLFlow project to train a model, then creates a KFServing prediction
  service that can be used to run predictions against the model."
inputs:
  - name: mlflow-codeset
    description: an MLFlow compatible codeset
    type: codeset
  - name: predictor
    description: type of predictor engine
    type: string
    default: auto
outputs:
  - name: prediction-url
    description: "The URL where the exposed prediction service endpoint can be contacted to run predictions."
    type: string
steps:
  - name: builder
    image: ghcr.io/fuseml/mlflow-builder:1.0
    inputs:
      - name: null
        value: null
        codeset:
          name: '{{ inputs.mlflow-codeset }}'
          path: /project
    outputs:
      - name: mlflow-env
        image:
          dockerfile: '/project/.fuseml/Dockerfile'
          name: 'registry.fuseml-registry/mlflow-builder/{{ inputs.mlflow-codeset.name }}:{{ inputs.mlflow-codeset.version }}'
        fromfile: null
  - name: trainer
    image: '{{ steps.builder.outputs.mlflow-env }}'
    inputs:
      - name: null
        value: null
        codeset:
          name: '{{ inputs.mlflow-codeset }}'
          path: '/project'
    outputs:
      - name: mlflow-model-url
        image: null
        fromfile: null
    env:
      - name: MLFLOW_TRACKING_URI
        value: "http://mlflow"
      - name: MLFLOW_S3_ENDPOINT_URL
        value: "http://mlflow-minio:9000"
      - name: AWS_ACCESS_KEY_ID
        value: CdkokHmIRZZ4s8PAQ1RI
      - name: AWS_SECRET_ACCESS_KEY
        value: LyXpfxpJ58B0nkQEopknq0h3lwV1jivv2MlWFM8R
  - name: predictor
    image: 'ghcr.io/flaviodsr/kfserving-predictor:1.0@sha256:2678dcbeed446263a1d37d2ad66fd8dcb0da0f9834ad96a828d7711263414584'
    inputs:
      - name: model
        value: '{{ steps.trainer.outputs.mlflow-model-url }}'
        codeset: null
      - name: predictor
        value: '{{ inputs.predictor }}'
        codeset: null
      - codeset:
          name: '{{ inputs.mlflow-codeset }}'
          path: '/project'
    outputs:
      - name: prediction-url
        image: null
        fromfile: null
    env:
      - name: AWS_ACCESS_KEY_ID
        value: CdkokHmIRZZ4s8PAQ1RI
      - name: AWS_SECRET_ACCESS_KEY
        value: LyXpfxpJ58B0nkQEopknq0h3lwV1jivv2MlWFM8R
`)

	w := workflow.Workflow{}

	err := yaml.Unmarshal([]byte(data), &w)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	cs = codeset.Codeset{
		Name: "codeset-name",
	}

	tektonClients = newClients(namespace)
	resolver = NewVariablesResolver()

	pipeline := createTektonPipeline(w)
	_, err = tektonClients.PipelineClient.Create(context.TODO(), pipeline, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
func createTektonPipeline(w workflow.Workflow) *v1beta1.Pipeline {
	labels := map[string]string{"fuseml/generated-from": *w.ID}
	meta := metav1.ObjectMeta{Name: w.Name, Namespace: namespace, Labels: labels}
	workspaces := []v1beta1.PipelineWorkspaceDeclaration{}
	resources := []v1beta1.PipelineDeclaredResource{}
	params := []v1beta1.ParamSpec{}
	tasks := []v1beta1.PipelineTask{}
	pipelineResults := []v1beta1.PipelineResult{}

	// process the pipeline inputs
	for _, input := range w.Inputs {

		// an input of the 'codeset' type means a git repository for tekton.
		// adds a workspace, resource and the clone task to the pipeline, also
		// also add 'codeset-name' parameter to the tekton pipeline.
		if *input.Type == input_type_codeset {
			workspaces = append(workspaces, v1beta1.WorkspacePipelineDeclaration{
				Name: "source",
			})
			resources = append(resources, v1beta1.PipelineDeclaredResource{
				Name: "source-repo",
				Type: "git",
			})
			tasks = append(tasks, v1beta1.PipelineTask{
				Name:    "clone",
				TaskRef: &v1beta1.TaskRef{Name: "clone"},
				Resources: &v1beta1.PipelineTaskResources{
					Inputs: []v1beta1.PipelineTaskInputResource{
						{Name: "source-repo", Resource: "source-repo"}},
				},
				Workspaces: []v1beta1.WorkspacePipelineTaskBinding{
					{Name: "source", Workspace: "source"}},
			})
			params = append(params, v1beta1.ParamSpec{
				Name:        "codeset",
				Description: "Reference to the codeset (project on git)",
			})
			resolver.addReference(fmt.Sprintf("inputs.%s.name", *input.Name), "$(params.codeset)")
			// TODO: not sure where to get this information yet
			resolver.addReference(fmt.Sprintf("inputs.%s.version", *input.Name), "0.1")
		} else {
			params = append(params, v1beta1.ParamSpec{
				Name:        *input.Name,
				Description: *input.Description,
				Default:     v1beta1.NewArrayOrString(*input.Default),
			})
			resolver.addReference(fmt.Sprintf("inputs.%s", *input.Name), fmt.Sprintf("$(params.%s)", *input.Name))
		}
	}

	// process the workflow steps
	for _, step := range w.Steps {
		params := []v1beta1.Param{}
		workspaces := []v1beta1.WorkspacePipelineTaskBinding{}
		var pipelineTask v1beta1.PipelineTask

		for _, output := range step.Outputs {
			// image type output parameters serve as input for the tekton builder task
			if output.Image != nil {
				dockerfile := resolver.resolve(*output.Image.Dockerfile)
				codesetPath := getInputCodesetPath(step.Inputs)
				if codesetPath != "" {
					dockerfile = strings.Replace(dockerfile, fmt.Sprintf("%s/", codesetPath), "", 1)
				}
				params = append(params, v1beta1.Param{
					Name: "IMAGE",
					Value: v1beta1.ArrayOrString{
						Type:      v1beta1.ParamTypeString,
						StringVal: resolver.resolve(*output.Image.Name),
					}}, v1beta1.Param{
					Name: "DOCKERFILE",
					Value: v1beta1.ArrayOrString{
						Type:      v1beta1.ParamTypeString,
						StringVal: dockerfile,
					},
				})
				// builder refers to the kaniko task in tekton, add a reference to it
				pipelineTask = v1beta1.PipelineTask{
					Name:    *step.Name,
					TaskRef: &v1beta1.TaskRef{Name: "kaniko"},
					Params:  params,
					Workspaces: []v1beta1.WorkspacePipelineTaskBinding{
						{Name: "source", Workspace: "source"}},
				}
				resolver.addReference(fmt.Sprintf("steps.%s.outputs.%s", *step.Name, *output.Name), *output.Image.Name)
			}
			// if the step output is the workflow output, map it to a PipelineResult in tekton
			if po := stepOutputIsPipelineOutput(output, w.Outputs); po != nil {
				pipelineResults = append(pipelineResults, v1beta1.PipelineResult{
					Name:        *po.Name,
					Description: *po.Description,
					Value:       fmt.Sprintf("$(tasks.%s.results.%s)", *step.Name, *output.Name),
				})
			}
		}

		// if pipelineTask is not defined yet it means that the task is not referencing
		// an existing TektonTask. In that case, build the task spec from the fuseml step.
		if pipelineTask.Name == "" {
			// generates a v1beta1.TaskSpec from a pipeline.PipelineStep
			taskSpec := stepToTektonTaskSpec(*step)
			for _, input := range step.Inputs {
				// if the step has a codeset as input add the workspace
				// TODO: for now it only supports 1 workspace
				if input.Codeset != nil {
					workspaces = append(workspaces, v1beta1.WorkspacePipelineTaskBinding{
						Name:      taskSpec.Workspaces[0].Name,
						Workspace: "source",
					})
				} else {
					params = append(params, v1beta1.Param{
						Name:  *input.Name,
						Value: *v1beta1.NewArrayOrString(resolver.resolve(*input.Value)),
					})
				}
			}

			// if image is parametrized add 'IMAGE' param, resolving it
			if strings.Contains(*step.Image, "{{") {
				// The kubernetes nodes are unable to resolve the local FuseML registry
				// (registry.fuseml-registry), in that way, when the step uses an image
				// from the local FuseML registry, replace registry.fuseml-registry with
				// 127.0.0.1:30500
				image := resolver.resolve(*step.Image)
				if strings.HasPrefix(image, "registry.fuseml-registry") {
					image = strings.Replace(image, "registry.fuseml-registry", "127.0.0.1:30500", 1)
				}
				params = append(params, v1beta1.Param{
					Name:  "IMAGE",
					Value: *v1beta1.NewArrayOrString(image),
				})
			}
			pipelineTask = v1beta1.PipelineTask{
				Name:       *step.Name,
				Params:     params,
				Workspaces: workspaces,
				TaskSpec: &v1beta1.EmbeddedTask{
					TaskSpec: taskSpec,
				},
			}
		}
		// for now assume that all tasks runs serially
		numTasks := len(tasks)
		if numTasks > 0 {
			pipelineTask.RunAfter = append(pipelineTask.RunAfter, tasks[numTasks-1].Name)
		}
		tasks = append(tasks, pipelineTask)
	}

	pipeline := &v1beta1.Pipeline{
		ObjectMeta: meta,
		Spec: v1beta1.PipelineSpec{
			Workspaces: workspaces,
			Resources:  resources,
			Params:     params,
			Tasks:      tasks,
			Results:    pipelineResults,
		},
	}
	return pipeline
}

func stepToTektonTaskSpec(step workflow.WorkflowStep) v1beta1.TaskSpec {
	image := *step.Image
	params := []v1beta1.ParamSpec{}
	workspaces := []v1beta1.WorkspaceDeclaration{}
	var workingDir string
	env := []corev1.EnvVar{}
	results := []v1beta1.TaskResult{}

	for _, input := range step.Inputs {
		// if there is a codeset as input, add workspace to the task
		if input.Codeset != nil {
			workspaces = append(workspaces, v1beta1.WorkspaceDeclaration{
				Name:      "source",
				MountPath: *input.Codeset.Path,
			})
			workingDir = *input.Codeset.Path
		} else {
			// else add it as a parameter to the tekton task
			params = append(params, v1beta1.ParamSpec{
				Name: *input.Name,
			})
			// make the inputs also available as env variable with the
			// FUSEML_ prefix
			env = append(env, corev1.EnvVar{
				Name:  fmt.Sprintf("FUSEML_%s", strings.ToUpper(*input.Name)),
				Value: fmt.Sprintf("$(params.%s)", *input.Name),
			})
		}
	}

	// if image is parameterized reference it as a task parameter to be able
	// to receive its value from a task output
	if strings.Contains(*step.Image, "{{") {
		params = append(params, v1beta1.ParamSpec{
			Name:        "IMAGE",
			Description: "Name (reference) of the image to run",
		})
		image = "$(params.IMAGE)"
	}

	// an output that is not an image represents a result in tekton
	for _, output := range step.Outputs {
		if output.Image == nil {
			results = append(results, v1beta1.TaskResult{
				Name: *output.Name,
			})
			env = append(env, corev1.EnvVar{
				Name:  "TASK_RESULT",
				Value: *output.Name})
		}
	}

	// load environment variables
	for _, stepEnv := range step.Env {
		env = append(env, corev1.EnvVar{
			Name:  *stepEnv.Name,
			Value: *stepEnv.Value,
		})
	}

	return v1beta1.TaskSpec{
		Params:     params,
		Workspaces: workspaces,
		Results:    results,
		Steps: []v1beta1.Step{{Container: corev1.Container{
			Name:       *step.Name,
			Image:      image,
			WorkingDir: workingDir,
			Env:        env,
			Command:    []string{"run"},
		}}},
	}
}

func stepOutputIsPipelineOutput(stepOutput *workflow.WorkflowStepOutput,
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
