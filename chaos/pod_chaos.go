package chaos

import (
	"errors"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func NewPodChaos(opts ...OptChaos) (*chaosmeshv1alpha1.PodChaos, error) {
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
	if config.PodChaos == nil {
		return nil, errors.New("podChaos is required")
	}

	podChaos := chaosmeshv1alpha1.PodChaos{}
	podChaos.Name = config.Name
	podChaos.Namespace = config.Namespace
	config.PodChaos.DeepCopyInto(&podChaos.Spec)

	if config.Labels != nil {
		podChaos.Labels = config.Labels
	}
	if config.Annotations != nil {
		podChaos.Annotations = config.Annotations
	}

	return &podChaos, nil
}

func GeneratePodChaosSpec(namespace string, appName string, action chaosmeshv1alpha1.PodChaosAction) *chaosmeshv1alpha1.PodChaosSpec {

	spec := &chaosmeshv1alpha1.PodChaosSpec{
		Action: action,
		ContainerSelector: chaosmeshv1alpha1.ContainerSelector{
			PodSelector: chaosmeshv1alpha1.PodSelector{
				Selector: chaosmeshv1alpha1.PodSelectorSpec{
					GenericSelectorSpec: chaosmeshv1alpha1.GenericSelectorSpec{
						Namespaces: []string{namespace},
						LabelSelectors: map[string]string{
							"app": appName,
						},
					},
				},
				Mode: chaosmeshv1alpha1.AllMode,
			},
		},
	}
	return spec
}
