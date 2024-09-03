package chaos

import (
	"errors"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func NewHttpChaos(opts ...OptChaos) (*chaosmeshv1alpha1.HTTPChaos, error) {
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
		return nil, errors.New("httpChaos is required")
	}

	httpChaos := chaosmeshv1alpha1.HTTPChaos{}
	httpChaos.Name = config.Name
	httpChaos.Namespace = config.Namespace
	config.HttpChaos.DeepCopyInto(&httpChaos.Spec)

	if config.Labels != nil {
		httpChaos.Labels = config.Labels
	}
	if config.Annotations != nil {
		httpChaos.Annotations = config.Annotations
	}

	return &httpChaos, nil
}
