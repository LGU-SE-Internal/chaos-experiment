package chaos

import (
	"errors"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func NewWorkflowChaos(opts ...OptChaos) (*chaosmeshv1alpha1.Workflow, error) {
	config := ConfigChaos{}
	for _, opt := range opts {
		if opt != nil {
			opt(&config)
		}
	}

	if config.Name == "" {
		return nil, errors.New("the resource name is required")
	}
	if config.Namespace == "" {
		return nil, errors.New("the namespace is required")
	}
	if config.Workflow == nil {
		return nil, errors.New("workflow is required")
	}

	workflow := chaosmeshv1alpha1.Workflow{}
	workflow.Name = config.Name
	workflow.Namespace = config.Namespace
	config.Workflow.DeepCopyInto(&workflow.Spec)

	if config.Labels != nil {
		workflow.Labels = config.Labels
	}
	if config.Annotations != nil {
		workflow.Annotations = config.Annotations
	}

	return &workflow, nil
}
