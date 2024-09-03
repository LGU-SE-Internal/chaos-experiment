package chaos

import (
	"errors"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"k8s.io/utils/pointer"
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

func GenerateSetsOfHttpChaosSpec(namespace string, podName string) []chaosmeshv1alpha1.HTTPChaosSpec {
	specs := make([]chaosmeshv1alpha1.HTTPChaosSpec, 0)

	basicSpec := &chaosmeshv1alpha1.HTTPChaosSpec{
		PodSelector: chaosmeshv1alpha1.PodSelector{
			Selector: chaosmeshv1alpha1.PodSelectorSpec{
				GenericSelectorSpec: chaosmeshv1alpha1.GenericSelectorSpec{
					Namespaces: []string{namespace},
				},
				Pods: map[string][]string{
					namespace: {
						podName,
					},
				},
			},
			Mode: chaosmeshv1alpha1.OneMode,
		},
		Target:              chaosmeshv1alpha1.PodHttpRequest, // can change
		PodHttpChaosActions: chaosmeshv1alpha1.PodHttpChaosActions{
			//Abort:   nil,
			//Delay: nil,
			//Replace: nil,
			//Patch:   nil,
		},
		Port:            8080,
		Path:            nil, // for filtering the request
		Method:          nil, // for filtering the request
		Code:            nil, // for filtering the request
		RequestHeaders:  nil, // for filtering the request
		ResponseHeaders: nil, // for filtering the request
		Duration:        pointer.String("5m"),
	}
	for _, target := range []chaosmeshv1alpha1.PodHttpChaosTarget{chaosmeshv1alpha1.PodHttpRequest, chaosmeshv1alpha1.PodHttpResponse} {
		cur := chaosmeshv1alpha1.HTTPChaosSpec{}
		basicSpec.DeepCopyInto(&cur)
		cur.Target = target

		for _, i := range []int{0, 1} {
			switch i {
			case 0:
				cur.PodHttpChaosActions.Abort = pointer.Bool(true)
			case 1:
				cur.PodHttpChaosActions.Abort = nil
				cur.PodHttpChaosActions.Delay = pointer.String("5m")
			case 2:
				//cur.PodHttpChaosActions.Replace =
			case 3:
				//cur.PodHttpChaosActions.Patch =
			}
		}

		specs = append(specs, cur)
	}
	return specs
}
