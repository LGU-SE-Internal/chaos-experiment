package chaos

import (
	"errors"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func NewJvmChaos(opts ...OptChaos) (*chaosmeshv1alpha1.JVMChaos, error) {
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
		return nil, errors.New("jvmChaos is required")
	}

	jvmChaos := chaosmeshv1alpha1.JVMChaos{}
	jvmChaos.Name = config.Name
	jvmChaos.Namespace = config.Namespace
	config.JVMChaos.DeepCopyInto(&jvmChaos.Spec)

	if config.Labels != nil {
		jvmChaos.Labels = config.Labels
	}
	if config.Annotations != nil {
		jvmChaos.Annotations = config.Annotations
	}

	return &jvmChaos, nil
}
