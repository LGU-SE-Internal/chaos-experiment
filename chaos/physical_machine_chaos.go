package chaos

import (
	"errors"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func NewPhysicalMachineChaos(opts ...OptChaos) (*chaosmeshv1alpha1.PhysicalMachineChaos, error) {
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
	if config.PhysicalMachineChaos == nil {
		return nil, errors.New("physicalMachineChaos is required")
	}

	physicalMachineChaos := chaosmeshv1alpha1.PhysicalMachineChaos{}
	physicalMachineChaos.Name = config.Name
	physicalMachineChaos.Namespace = config.Namespace
	config.PhysicalMachineChaos.DeepCopyInto(&physicalMachineChaos.Spec)

	if config.Labels != nil {
		physicalMachineChaos.Labels = config.Labels
	}
	if config.Annotations != nil {
		physicalMachineChaos.Annotations = config.Annotations
	}

	return &physicalMachineChaos, nil
}
