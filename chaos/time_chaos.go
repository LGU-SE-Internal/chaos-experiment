package chaos

import (
	"errors"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func NewTimeChaos(opts ...OptChaos) (*chaosmeshv1alpha1.TimeChaos, error) {
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
	if config.TimeChaos == nil {
		return nil, errors.New("timeChaos is required")
	}

	timeChaos := chaosmeshv1alpha1.TimeChaos{}
	timeChaos.Name = config.Name
	timeChaos.Namespace = config.Namespace
	config.TimeChaos.DeepCopyInto(&timeChaos.Spec)

	if config.Labels != nil {
		timeChaos.Labels = config.Labels
	}
	if config.Annotations != nil {
		timeChaos.Annotations = config.Annotations
	}

	return &timeChaos, nil
}

func GenerateTimeChaosSpec(namespace string, appName string, duration *string, timeOffset string) *chaosmeshv1alpha1.TimeChaosSpec {
	spec := &chaosmeshv1alpha1.TimeChaosSpec{
		TimeOffset: timeOffset,
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

	if duration != nil && *duration != "" {
		spec.Duration = duration
	}

	return spec
}
