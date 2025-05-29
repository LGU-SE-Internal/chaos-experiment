package chaos

import (
	"errors"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func NewDnsChaos(opts ...OptChaos) (*chaosmeshv1alpha1.DNSChaos, error) {
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
	if config.DNSChaos == nil {
		return nil, errors.New("dnsChaos is required")
	}

	dnsChaos := chaosmeshv1alpha1.DNSChaos{}
	dnsChaos.Name = config.Name
	dnsChaos.Namespace = config.Namespace
	config.DNSChaos.DeepCopyInto(&dnsChaos.Spec)

	if config.Labels != nil {
		dnsChaos.Labels = config.Labels
	}
	if config.Annotations != nil {
		dnsChaos.Annotations = config.Annotations
	}

	return &dnsChaos, nil
}

// GenerateDnsChaosSpec creates a DNS chaos spec for the given namespace, app and patterns
func GenerateDnsChaosSpec(namespace string, appName string, duration *string, action chaosmeshv1alpha1.DNSChaosAction, patterns []string) *chaosmeshv1alpha1.DNSChaosSpec {
	spec := &chaosmeshv1alpha1.DNSChaosSpec{
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
		DomainNamePatterns: patterns,
	}

	if duration != nil && *duration != "" {
		spec.Duration = duration
	}

	return spec
}
