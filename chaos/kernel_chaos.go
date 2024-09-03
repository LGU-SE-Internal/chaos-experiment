package chaos

import (
	"errors"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func NewKernelChaos(opts ...OptChaos) (*chaosmeshv1alpha1.KernelChaos, error) {
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
	if config.HttpChaos == nil {
		return nil, errors.New("kernelChaos is required")
	}

	kernelChaos := chaosmeshv1alpha1.KernelChaos{}
	kernelChaos.Name = config.Name
	kernelChaos.Namespace = config.Namespace
	config.KernelChaos.DeepCopyInto(&kernelChaos.Spec)

	if config.Labels != nil {
		kernelChaos.Labels = config.Labels
	}
	if config.Annotations != nil {
		kernelChaos.Annotations = config.Annotations
	}

	return &kernelChaos, nil
}
