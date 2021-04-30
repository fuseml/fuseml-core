package builder

import (
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventListenerBuilder holds the tekton event listener definition
type EventListenerBuilder struct {
	EventListener v1alpha1.EventListener
}

// NewEventListenerBuilder creates a EventListener with default values.
func NewEventListenerBuilder(name, namespace string) *EventListenerBuilder {
	b := &EventListenerBuilder{}
	b.EventListener = v1alpha1.EventListener{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return b
}

// ServiceAccount sets a ServiceAccountName to the EventListener spec.
func (b *EventListenerBuilder) ServiceAccount(name string) {
	b.EventListener.Spec.ServiceAccountName = name
}

// EventListenerTriggerBinding adds a EventListenerTrigger to the EventListener spec.
func (b *EventListenerBuilder) EventListenerTriggerBinding(templateName string, bindingsName ...string) {
	bindings := []*v1alpha1.TriggerSpecBinding{}
	for _, bName := range bindingsName {
		bindings = append(bindings, &v1alpha1.TriggerSpecBinding{
			Ref: bName,
		})
	}
	b.EventListener.Spec.Triggers = append(b.EventListener.Spec.Triggers, v1alpha1.EventListenerTrigger{
		Template: &v1alpha1.TriggerSpecTemplate{
			Ref: &templateName,
		},
		Bindings: bindings,
	})
}
