package builder

import (
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TriggerTemplateBuilder holds the tekton trigger template definition.
type TriggerTemplateBuilder struct {
	TriggerTemplate v1alpha1.TriggerTemplate
}

// NewTriggerTemplateBuilder creates a TriggerTemplate with default values.
func NewTriggerTemplateBuilder(name, namespace string) *TriggerTemplateBuilder {
	b := &TriggerTemplateBuilder{}
	b.TriggerTemplate = v1alpha1.TriggerTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return b
}

// Meta sets the Meta structs of the TriggerTemplate.
// Any number of MetaOp modifiers can be passed.
func (b *TriggerTemplateBuilder) Meta(ops ...MetaOp) {
	for _, op := range ops {
		switch o := op.(type) {
		case ObjectMetaOp:
			o(&b.TriggerTemplate.ObjectMeta)
		case TypeMetaOp:
			o(&b.TriggerTemplate.TypeMeta)
		}
	}
}

// Param adds a ParamSpec to the TriggerTemplate spec.
func (b *TriggerTemplateBuilder) Param(name, description string) {
	b.TriggerTemplate.Spec.Params = append(b.TriggerTemplate.Spec.Params, v1alpha1.ParamSpec{
		Name:        name,
		Description: description,
	})
}

// ParamWithDefaultValue adds a ParamSpec with a default value to the TriggerTemplate spec.
func (b *TriggerTemplateBuilder) ParamWithDefaultValue(name, description, defaultValue string) {
	b.TriggerTemplate.Spec.Params = append(b.TriggerTemplate.Spec.Params, v1alpha1.ParamSpec{
		Name:        name,
		Description: description,
		Default:     &defaultValue,
	})
}

// ResourceTemplate adds a ResourceTemplate to the TriggerTemplate spec.
func (b *TriggerTemplateBuilder) ResourceTemplate(resoureceTemplate runtime.RawExtension) {
	b.TriggerTemplate.Spec.ResourceTemplates = append(b.TriggerTemplate.Spec.ResourceTemplates,
		v1alpha1.TriggerResourceTemplate{
			RawExtension: resoureceTemplate,
		})
}
