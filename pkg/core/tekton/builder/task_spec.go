package builder

import (
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
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
