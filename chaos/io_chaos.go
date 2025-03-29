package chaos

import (
	"errors"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func NewIOChaos(opts ...OptChaos) (*chaosmeshv1alpha1.IOChaos, error) {
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
	if config.IOChaos == nil {
		return nil, errors.New("IOChaos is required")
	}

	ioChaos := chaosmeshv1alpha1.IOChaos{}
	ioChaos.Name = config.Name
	ioChaos.Namespace = config.Namespace
	config.IOChaos.DeepCopyInto(&ioChaos.Spec)

	if config.Labels != nil {
		ioChaos.Labels = config.Labels
	}
	if config.Annotations != nil {
		ioChaos.Annotations = config.Annotations
	}

	return &ioChaos, nil
}

// OptIOChaos defines options for IO chaos
type OptIOChaos func(*chaosmeshv1alpha1.IOChaosSpec)

// WithIODelayAction configures the IO chaos to inject latency
func WithIODelayAction(delay string) OptIOChaos {
	return func(spec *chaosmeshv1alpha1.IOChaosSpec) {
		spec.Action = chaosmeshv1alpha1.IoLatency
		spec.Delay = delay
	}
}

// WithIOErrorAction configures the IO chaos to inject errors
func WithIOErrorAction(errno uint32) OptIOChaos {
	return func(spec *chaosmeshv1alpha1.IOChaosSpec) {
		spec.Action = chaosmeshv1alpha1.IoFaults
		spec.Errno = errno
	}
}

// WithIOMistakeAction configures the IO chaos to inject mistakes
func WithIOMistakeAction(filling chaosmeshv1alpha1.FillingType, maxOccurrences int64, maxLength int64) OptIOChaos {
	return func(spec *chaosmeshv1alpha1.IOChaosSpec) {
		spec.Action = chaosmeshv1alpha1.IoMistake
		spec.Mistake = &chaosmeshv1alpha1.MistakeSpec{
			Filling:        filling,
			MaxOccurrences: maxOccurrences,
			MaxLength:      maxLength,
		}
	}
}

// WithIOAttrOverrideAction configures the IO chaos to override attributes
func WithIOAttrOverrideAction(attr *chaosmeshv1alpha1.AttrOverrideSpec) OptIOChaos {
	return func(spec *chaosmeshv1alpha1.IOChaosSpec) {
		spec.Action = chaosmeshv1alpha1.IoAttrOverride
		spec.Attr = attr
	}
}

// WithIOPath sets the path for IO chaos
func WithIOPath(path string) OptIOChaos {
	return func(spec *chaosmeshv1alpha1.IOChaosSpec) {
		spec.Path = path
	}
}

// WithIOMethods sets the methods for IO chaos
func WithIOMethods(methods []chaosmeshv1alpha1.IoMethod) OptIOChaos {
	return func(spec *chaosmeshv1alpha1.IOChaosSpec) {
		spec.Methods = methods
	}
}

// WithIOPercent sets the percentage of injection for IO chaos
func WithIOPercent(percent int) OptIOChaos {
	return func(spec *chaosmeshv1alpha1.IOChaosSpec) {
		spec.Percent = percent
	}
}

// WithIOVolumePath sets the volume path for IO chaos
func WithIOVolumePath(volumePath string) OptIOChaos {
	return func(spec *chaosmeshv1alpha1.IOChaosSpec) {
		spec.VolumePath = volumePath
	}
}

// WithIOContainerNames sets specific container names for IO chaos
func WithIOContainerNames(containerNames []string) OptIOChaos {
	return func(spec *chaosmeshv1alpha1.IOChaosSpec) {
		spec.ContainerNames = containerNames
	}
}

// GenerateIOChaosSpec creates an IO chaos spec
func GenerateIOChaosSpec(namespace string, appName string, duration *string, volumePath string, opts ...OptIOChaos) *chaosmeshv1alpha1.IOChaosSpec {
	spec := &chaosmeshv1alpha1.IOChaosSpec{
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
		VolumePath: volumePath,
		Percent:    100, // Default to 100%
	}

	for _, opt := range opts {
		opt(spec)
	}

	if duration != nil && *duration != "" {
		spec.Duration = duration
	}

	return spec
}
