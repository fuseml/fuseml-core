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

// TaskParam adds a ParamSpec to the TaskSpec.
func (b *TaskSpecBuilder) TaskParam(name string, description ...string) {
	p := v1beta1.ParamSpec{
		Name: name,
	}
	if len(description) > 0 {
		p.Description = description[0]
	}
	b.TaskSpec.Params = append(b.TaskSpec.Params, p)
}

// TaskWorkspace adds a WorkspaceDeclaration to the TaskSpec.
func (b *TaskSpecBuilder) TaskWorkspace(name string, mountPath ...string) {
	ws := v1beta1.WorkspaceDeclaration{
		Name: name,
	}
	if len(mountPath) > 0 {
		ws.MountPath = mountPath[0]
	}
	b.TaskSpec.Workspaces = append(b.TaskSpec.Workspaces, ws)
}

// TaskWorkingDir sets the WorkingDir on the TaskSpec step.
func (b *TaskSpecBuilder) TaskWorkingDir(workingDir string) {
	b.TaskSpec.Steps[0].WorkingDir = workingDir
}

// TaskEnvVar adds a Env to the TaskSpec step.
func (b *TaskSpecBuilder) TaskEnvVar(name, value string) {
	b.TaskSpec.Steps[0].Env = append(b.TaskSpec.Steps[0].Env, corev1.EnvVar{
		Name:  name,
		Value: value,
	})
}

// TaskImage sets the image on the TaskSpec step.
func (b *TaskSpecBuilder) TaskImage(image string) {
	b.TaskSpec.Steps[0].Image = image
}

// TaskResult adds a TaskResult to the TaskSpec.
func (b *TaskSpecBuilder) TaskResult(name string) {
	b.TaskSpec.Results = append(b.TaskSpec.Results, v1beta1.TaskResult{
		Name: name,
	})
}
