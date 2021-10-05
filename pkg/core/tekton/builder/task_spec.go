package builder

import (
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// TaskSpecBuilder holds the tekton task spec definition.
type TaskSpecBuilder struct {
	TaskSpec v1beta1.TaskSpec
}

// NewTaskSpecBuilder creates a TaskSpec with default values.
func NewTaskSpecBuilder(name, image, command string) *TaskSpecBuilder {
	b := &TaskSpecBuilder{}
	b.TaskSpec = v1beta1.TaskSpec{
		Steps: []v1beta1.Step{{Container: corev1.Container{
			Name:    name,
			Image:   image,
			Command: []string{command},
		}}},
	}
	return b
}

// Param adds a ParamSpec to the TaskSpec.
func (b *TaskSpecBuilder) Param(name string) {
	b.TaskSpec.Params = append(b.TaskSpec.Params, v1beta1.ParamSpec{
		Name: name,
	})
}

// ParamWithDescription adds a ParamSpec with a description to the TaskSpec.
func (b *TaskSpecBuilder) ParamWithDescription(name, description string) {
	b.TaskSpec.Params = append(b.TaskSpec.Params, v1beta1.ParamSpec{
		Name:        name,
		Description: description,
	})
}

// Workspace adds a WorkspaceDeclaration to the TaskSpec.
func (b *TaskSpecBuilder) Workspace(name string) {
	b.TaskSpec.Workspaces = append(b.TaskSpec.Workspaces, v1beta1.WorkspaceDeclaration{
		Name: name,
	})
}

// WorkspaceWithMountPath adds a WorkspaceDeclaration with a mount path to the TaskSpec.
func (b *TaskSpecBuilder) WorkspaceWithMountPath(name, mountPath string) {
	b.TaskSpec.Workspaces = append(b.TaskSpec.Workspaces, v1beta1.WorkspaceDeclaration{
		Name:      name,
		MountPath: mountPath,
	})
}

// WorkingDir sets the WorkingDir on the TaskSpec step.
func (b *TaskSpecBuilder) WorkingDir(workingDir string) {
	b.TaskSpec.Steps[0].WorkingDir = workingDir
}

// Env adds a Env to the TaskSpec step.
func (b *TaskSpecBuilder) Env(name, value string) {
	b.TaskSpec.Steps[0].Env = append(b.TaskSpec.Steps[0].Env, corev1.EnvVar{
		Name:  name,
		Value: value,
	})
}

// Image sets the image on the TaskSpec step.
func (b *TaskSpecBuilder) Image(image string) {
	b.TaskSpec.Steps[0].Image = image
}

// Result adds a TaskResult to the TaskSpec.
func (b *TaskSpecBuilder) Result(name string) {
	b.TaskSpec.Results = append(b.TaskSpec.Results, v1beta1.TaskResult{
		Name: name,
	})
}

// Resources add a Resources to the TaskSpec.
func (b *TaskSpecBuilder) Resources(requests, limits map[string]string) {
	b.TaskSpec.Steps[0].Resources = corev1.ResourceRequirements{}
	if len(requests) > 0 {
		b.TaskSpec.Steps[0].Resources.Requests = toResourceList(requests)
	}

	if len(limits) > 0 {
		b.TaskSpec.Steps[0].Resources.Limits = toResourceList(limits)
	}
}

func toResourceList(resources map[string]string) corev1.ResourceList {
	res := corev1.ResourceList{}
	for k, v := range resources {
		res[corev1.ResourceName(k)] = resource.MustParse(v)
	}
	return res
}
