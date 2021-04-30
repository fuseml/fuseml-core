package builder

import (
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TriggerBindingBuilder holds the tekton trigger binding definition.
type TriggerBindingBuilder struct {
	TriggerBinding v1alpha1.TriggerBinding
}

// NewTriggerBindingBuilder creates a TriggerBinding with default values.
func NewTriggerBindingBuilder(name, namespace string) *TriggerBindingBuilder {
	b := &TriggerBindingBuilder{}
	b.TriggerBinding = v1alpha1.TriggerBinding{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return b
}

// TriggerBindingParam adds a Param to the TriggerBinding spec.
func (b *TriggerBindingBuilder) TriggerBindingParam(name, value string) {
	b.TriggerBinding.Spec.Params = append(b.TriggerBinding.Spec.Params, v1alpha1.Param{
		Name:  name,
		Value: value,
	})
}
