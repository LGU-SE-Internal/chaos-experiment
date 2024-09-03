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
