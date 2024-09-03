package chaos

import (
	"errors"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func NewBlockChaos(opts ...OptChaos) (*chaosmeshv1alpha1.BlockChaos, error) {
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
	if config.BlockChaos == nil {
		return nil, errors.New("httpChaos is required")
	}

	blockChaos := chaosmeshv1alpha1.BlockChaos{}
	blockChaos.Name = config.Name
	blockChaos.Namespace = config.Namespace
	config.BlockChaos.DeepCopyInto(&blockChaos.Spec)

	if config.Labels != nil {
		blockChaos.Labels = config.Labels
	}
	if config.Annotations != nil {
		blockChaos.Annotations = config.Annotations
	}

	return &blockChaos, nil
}
