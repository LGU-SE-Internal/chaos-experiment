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

func GenerateStressChaosSpec(namespace string, appName string, duration *string, Stressors chaosmeshv1alpha1.Stressors) *chaosmeshv1alpha1.StressChaosSpec {

	spec := &chaosmeshv1alpha1.StressChaosSpec{
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
		Stressors: &Stressors,
	}
	if duration != nil && *duration != "" {
		spec.Duration = duration
	}
	return spec
}

// GenerateStressChaosSpecWithContainers creates a StressChaosSpec with specified container names
func GenerateStressChaosSpecWithContainers(namespace string, appName string, duration *string, Stressors chaosmeshv1alpha1.Stressors, containerNames []string) *chaosmeshv1alpha1.StressChaosSpec {
	spec := &chaosmeshv1alpha1.StressChaosSpec{
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
			ContainerNames: containerNames,
		},
		Stressors: &Stressors,
	}
	if duration != nil && *duration != "" {
		spec.Duration = duration
	}
	return spec
}
