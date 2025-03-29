package chaos

import (
	"errors"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func NewNetworkChaos(opts ...OptChaos) (*chaosmeshv1alpha1.NetworkChaos, error) {
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
	if config.NetworkChaos == nil {
		return nil, errors.New("networkChaos is required")
	}

	networkChaos := chaosmeshv1alpha1.NetworkChaos{}
	networkChaos.Name = config.Name
	networkChaos.Namespace = config.Namespace
	config.NetworkChaos.DeepCopyInto(&networkChaos.Spec)

	if config.Labels != nil {
		networkChaos.Labels = config.Labels
	}
	if config.Annotations != nil {
		networkChaos.Annotations = config.Annotations
	}

	return &networkChaos, nil
}

type OptNetworkChaos func(opt *chaosmeshv1alpha1.NetworkChaosSpec)

func WithNetworkAction(action chaosmeshv1alpha1.NetworkChaosAction) OptNetworkChaos {
	return func(opt *chaosmeshv1alpha1.NetworkChaosSpec) {
		opt.Action = action
	}
}

func WithNetworkDirection(direction chaosmeshv1alpha1.Direction) OptNetworkChaos {
	return func(opt *chaosmeshv1alpha1.NetworkChaosSpec) {
		opt.Direction = direction
	}
}

func WithNetworkDevice(device string) OptNetworkChaos {
	return func(opt *chaosmeshv1alpha1.NetworkChaosSpec) {
		opt.Device = device
	}
}

func WithNetworkDuration(duration *string) OptNetworkChaos {
	return func(opt *chaosmeshv1alpha1.NetworkChaosSpec) {
		opt.Duration = duration
	}
}


func WithNetworkTargetDevice(device string) OptNetworkChaos {
	return func(opt *chaosmeshv1alpha1.NetworkChaosSpec) {
		opt.TargetDevice = device
	}
}

func WithNetworkExternalTargets(targets []string) OptNetworkChaos {
	return func(opt *chaosmeshv1alpha1.NetworkChaosSpec) {
		opt.ExternalTargets = targets
	}
}

// Specific TC parameters options

func WithNetworkDelay(latency string, correlation string, jitter string) OptNetworkChaos {
	return func(opt *chaosmeshv1alpha1.NetworkChaosSpec) {
		opt.TcParameter.Delay = &chaosmeshv1alpha1.DelaySpec{
			Latency:     latency,
			Correlation: correlation,
			Jitter:      jitter,
		}
	}
}

func WithNetworkLoss(loss string, correlation string) OptNetworkChaos {
	return func(opt *chaosmeshv1alpha1.NetworkChaosSpec) {
		opt.TcParameter.Loss = &chaosmeshv1alpha1.LossSpec{
			Loss:        loss,
			Correlation: correlation,
		}
	}
}

func WithNetworkDuplicate(duplicate string, correlation string) OptNetworkChaos {
	return func(opt *chaosmeshv1alpha1.NetworkChaosSpec) {
		opt.TcParameter.Duplicate = &chaosmeshv1alpha1.DuplicateSpec{
			Duplicate:   duplicate,
			Correlation: correlation,
		}
	}
}

func WithNetworkCorrupt(corrupt string, correlation string) OptNetworkChaos {
	return func(opt *chaosmeshv1alpha1.NetworkChaosSpec) {
		opt.TcParameter.Corrupt = &chaosmeshv1alpha1.CorruptSpec{
			Corrupt:     corrupt,
			Correlation: correlation,
		}
	}
}

func WithNetworkBandwidth(rate string, limit uint32, buffer uint32) OptNetworkChaos {
	return func(opt *chaosmeshv1alpha1.NetworkChaosSpec) {
		opt.TcParameter.Bandwidth = &chaosmeshv1alpha1.BandwidthSpec{
			Rate:   rate,
			Limit:  limit,
			Buffer: buffer,
		}
	}
}

func GenerateNetworkChaosSpec(namespace string, appName string, duration *string, action chaosmeshv1alpha1.NetworkChaosAction, opts ...OptNetworkChaos) *chaosmeshv1alpha1.NetworkChaosSpec {
	spec := &chaosmeshv1alpha1.NetworkChaosSpec{
		Action: action,
		PodSelector: chaosmeshv1alpha1.PodSelector{
			Selector: chaosmeshv1alpha1.PodSelectorSpec{
				GenericSelectorSpec: chaosmeshv1alpha1.GenericSelectorSpec{
					Namespaces:     []string{namespace},
					LabelSelectors: map[string]string{"app": appName},
				},
			},
			Mode: chaosmeshv1alpha1.AllMode,
		},
		Direction: chaosmeshv1alpha1.To, // Default direction
	}

	if duration != nil && *duration != "" {
		spec.Duration = duration
	}

	for _, opt := range opts {
		if opt != nil {
			opt(spec)
		}
	}

	return spec
}