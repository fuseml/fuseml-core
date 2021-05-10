package builder

import (
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/pkg/apis/resource/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PipelineRunBuilder holds the tekton pipeline definition
type PipelineRunBuilder struct {
	PipelineRun v1beta1.PipelineRun
}

// NewPipelineRunBuilder creates a PipelineRun with default values.
func NewPipelineRunBuilder(generateName string) *PipelineRunBuilder {
	b := &PipelineRunBuilder{}
	b.PipelineRun = v1beta1.PipelineRun{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PipelineRun",
			APIVersion: "tekton.dev/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: generateName,
		},
	}
	return b
}

// Meta sets the Meta structs of the PipelineRun.
// Any number of MetaOp modifiers can be passed.
func (b *PipelineRunBuilder) Meta(ops ...MetaOp) {
	for _, op := range ops {
		switch o := op.(type) {
		case ObjectMetaOp:
			o(&b.PipelineRun.ObjectMeta)
		case TypeMetaOp:
			o(&b.PipelineRun.TypeMeta)
		}
	}
}

// GenerateName sets a GenerateName to the PipelineRun meta.
func (b *PipelineRunBuilder) GenerateName(generateName string) {
	b.PipelineRun.ObjectMeta.GenerateName = generateName
}

// ServiceAccount sets a ServiceAccountName to the PipelineRun spec.
func (b *PipelineRunBuilder) ServiceAccount(name string) {
	b.PipelineRun.Spec.ServiceAccountName = name
}

// PipelineRef sets a PipelineRef to the PipelineRun spec.
func (b *PipelineRunBuilder) PipelineRef(name string) {
	b.PipelineRun.Spec.PipelineRef = &v1beta1.PipelineRef{
		Name: name,
	}
}

// Workspace adds a WorkspaceBinding to the PipelineRun spec.
func (b *PipelineRunBuilder) Workspace(name string, accessMode string, size string) {
	b.PipelineRun.Spec.Workspaces = append(b.PipelineRun.Spec.Workspaces,
		v1beta1.WorkspaceBinding{
			Name: name,
			VolumeClaimTemplate: &v1.PersistentVolumeClaim{
				Spec: v1.PersistentVolumeClaimSpec{
					AccessModes: []v1.PersistentVolumeAccessMode{
						v1.PersistentVolumeAccessMode(accessMode)},
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							"storage": resource.MustParse(size),
						},
					},
				},
			},
		},
	)
}

// Param adds a Param to the PipelineRun spec.
func (b *PipelineRunBuilder) Param(name, value string) {
	b.PipelineRun.Spec.Params = append(b.PipelineRun.Spec.Params, v1beta1.Param{
		Name:  name,
		Value: *v1beta1.NewArrayOrString(value),
	})
}

// ResourceGit adds a PipelineResourceSpec of type 'git' to the PipelineRun spec.
func (b *PipelineRunBuilder) ResourceGit(name, url, revision string) {
	b.PipelineRun.Spec.Resources = append(b.PipelineRun.Spec.Resources, v1beta1.PipelineResourceBinding{
		Name: name,
		ResourceSpec: &v1alpha1.PipelineResourceSpec{
			Type: "git",
			Params: []v1alpha1.ResourceParam{{
				Name:  "url",
				Value: url,
			}, {
				Name:  "revision",
				Value: revision,
			}},
		},
	})
}
