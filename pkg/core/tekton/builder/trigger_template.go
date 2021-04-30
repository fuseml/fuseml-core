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

// TriggerTemplateMeta sets the Meta structs of the TriggerTemplate.
// Any number of MetaOp modifiers can be passed.
func (b *TriggerTemplateBuilder) TriggerTemplateMeta(ops ...MetaOp) {
	for _, op := range ops {
		switch o := op.(type) {
		case ObjectMetaOp:
			o(&b.TriggerTemplate.ObjectMeta)
		case TypeMetaOp:
			o(&b.TriggerTemplate.TypeMeta)
		}
	}
}

// TriggerTemplateParam adds a ParamSpec to the TriggerTemplate spec.
func (b *TriggerTemplateBuilder) TriggerTemplateParam(name string, description string, defaultValue ...string) {
	param := v1alpha1.ParamSpec{
		Name:        name,
		Description: description,
	}
	if len(defaultValue) > 0 {
		param.Default = &defaultValue[0]
	}
	b.TriggerTemplate.Spec.Params = append(b.TriggerTemplate.Spec.Params, param)
}

// TriggerResourceTemplate adds a ResourceTemplate to the TriggerTemplate spec.
func (b *TriggerTemplateBuilder) TriggerResourceTemplate(resoureceTemplate runtime.RawExtension) {
	b.TriggerTemplate.Spec.ResourceTemplates = append(b.TriggerTemplate.Spec.ResourceTemplates,
		v1alpha1.TriggerResourceTemplate{
			RawExtension: resoureceTemplate,
		})
}
