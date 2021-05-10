package builder

import (
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PipelineBuilder holds the tekton pipeline definition.
type PipelineBuilder struct {
	Pipeline v1beta1.Pipeline
}

// NewPipelineBuilder creates a Pipeline with default values.
func NewPipelineBuilder(name, namespace string) *PipelineBuilder {
	b := &PipelineBuilder{}
	b.Pipeline = v1beta1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return b
}

// Meta sets the Meta structs of the Pipeline.
// Any number of MetaOp modifiers can be passed.
func (b *PipelineBuilder) Meta(ops ...MetaOp) {
	for _, op := range ops {
		switch o := op.(type) {
		case ObjectMetaOp:
			o(&b.Pipeline.ObjectMeta)
		case TypeMetaOp:
			o(&b.Pipeline.TypeMeta)
		}
	}
}

// Description sets the Description on the Pipeline spec.
func (b *PipelineBuilder) Description(description string) {
	b.Pipeline.Spec.Description = description
}

// Param adds a ParamSpec to the Pipeline spec.
func (b *PipelineBuilder) Param(name string, description string) {
	b.Pipeline.Spec.Params = append(b.Pipeline.Spec.Params, v1beta1.ParamSpec{
		Name:        name,
		Description: description,
	})
}

// ParamWithDefaultValue adds a ParamSpec with a default value to the Pipeline spec.
func (b *PipelineBuilder) ParamWithDefaultValue(name, description, defaultValue string) {
	b.Pipeline.Spec.Params = append(b.Pipeline.Spec.Params, v1beta1.ParamSpec{
		Name:        name,
		Description: description,
		Default:     v1beta1.NewArrayOrString(defaultValue),
	})
}

// Workspace adds a WorkspacePipelineDeclaration to the Pipeline spec.
func (b *PipelineBuilder) Workspace(name string, optional bool) {

	b.Pipeline.Spec.Workspaces = append(b.Pipeline.Spec.Workspaces, v1beta1.WorkspacePipelineDeclaration{
		Name:     name,
		Optional: optional,
	})
}

// Resource adds a PipelineDeclaredResource to the Pipeline spec.
func (b *PipelineBuilder) Resource(name string, tp string, optional bool) {
	b.Pipeline.Spec.Resources = append(b.Pipeline.Spec.Resources, v1beta1.PipelineDeclaredResource{
		Name:     name,
		Type:     tp,
		Optional: optional,
	})
}

// Task adds a PipelineTask to the Pipeline spec.
// The PipelineTask can reference an existing task by name (string task paramter), or embeds the task
// as a TaskSpec (v1beta1.TaskSpec task paramter)
func (b *PipelineBuilder) Task(name string, task interface{}, params map[string]string,
	workspaces map[string]string, resources map[string]string) {
	pt := v1beta1.PipelineTask{
		Name: name,
	}
	switch t := task.(type) {
	case string:
		pt.TaskRef = &v1beta1.TaskRef{Name: t}
	case v1beta1.TaskSpec:
		pt.TaskSpec = &v1beta1.EmbeddedTask{TaskSpec: t}
	}
	for name, value := range params {
		pt.Params = append(pt.Params, v1beta1.Param{
			Name:  name,
			Value: *v1beta1.NewArrayOrString(value),
		})
	}
	for name, workspace := range workspaces {
		pt.Workspaces = append(pt.Workspaces, v1beta1.WorkspacePipelineTaskBinding{
			Name:      name,
			Workspace: workspace,
		})
	}
	if resources != nil {
		ptir := []v1beta1.PipelineTaskInputResource{}
		for name, resource := range resources {
			ptir = append(ptir, v1beta1.PipelineTaskInputResource{
				Name:     name,
				Resource: resource,
			})
		}
		pt.Resources = &v1beta1.PipelineTaskResources{
			Inputs: ptir,
		}
	}
	// for now assume that all tasks runs serially
	numTasks := len(b.Pipeline.Spec.Tasks)
	if numTasks > 0 {
		pt.RunAfter = append(pt.RunAfter, b.Pipeline.Spec.Tasks[numTasks-1].Name)
	}
	b.Pipeline.Spec.Tasks = append(b.Pipeline.Spec.Tasks, pt)
}

// Result adds a Result to the Pipeline spec.
func (b *PipelineBuilder) Result(name, description, value string) {
	b.Pipeline.Spec.Results = append(b.Pipeline.Spec.Results, v1beta1.PipelineResult{
		Name:        name,
		Description: description,
		Value:       value,
	})
}
