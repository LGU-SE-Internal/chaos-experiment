package chaos

import (
	"errors"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func NewNetworkChaos(opts ...OptChaos) (*chaosmeshv1alpha1.NetworkChaos, error) {
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
	if config.NetworkChaos == nil {
		return nil, errors.New("networkChaos is required")
	}

	networkChaos := chaosmeshv1alpha1.NetworkChaos{}
	networkChaos.Name = config.Name
	networkChaos.Namespace = config.Namespace
	config.NetworkChaos.DeepCopyInto(&networkChaos.Spec)

	if config.Labels != nil {
		networkChaos.Labels = config.Labels
	}
	if config.Annotations != nil {
		networkChaos.Annotations = config.Annotations
	}

	return &networkChaos, nil
}
