package chaos

import (
	"errors"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func NewStressChaos(opts ...OptChaos) (*chaosmeshv1alpha1.StressChaos, error) {
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
	if config.StressChaos == nil {
		return nil, errors.New("stressChaos is required")
	}

	stressChaos := chaosmeshv1alpha1.StressChaos{}
	stressChaos.Name = config.Name
	stressChaos.Namespace = config.Namespace
	config.StressChaos.DeepCopyInto(&stressChaos.Spec)

	if config.Labels != nil {
		stressChaos.Labels = config.Labels
	}
	if config.Annotations != nil {
		stressChaos.Annotations = config.Annotations
	}

	return &stressChaos, nil
}
