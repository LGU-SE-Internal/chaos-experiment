package chaos

import (
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type ConfigChaos struct {
	Name        string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string

	HttpChaos            *chaosmeshv1alpha1.HTTPChaosSpec
	BlockChaos           *chaosmeshv1alpha1.BlockChaosSpec
	DNSChaos             *chaosmeshv1alpha1.DNSChaosSpec
	IOChaos              *chaosmeshv1alpha1.IOChaosSpec
	JVMChaos             *chaosmeshv1alpha1.JVMChaosSpec
	KernelChaos          *chaosmeshv1alpha1.KernelChaosSpec
	NetworkChaos         *chaosmeshv1alpha1.NetworkChaosSpec
	PhysicalMachineChaos *chaosmeshv1alpha1.PhysicalMachineChaosSpec
	PodChaos             *chaosmeshv1alpha1.PodChaosSpec
	StressChaos          *chaosmeshv1alpha1.StressChaosSpec
	TimeChaos            *chaosmeshv1alpha1.TimeChaosSpec
	Workflow             *chaosmeshv1alpha1.WorkflowSpec
}

type OptChaos func(opt *ConfigChaos)

func WithName(name string) OptChaos {
	return func(opt *ConfigChaos) {
		opt.Name = name
	}
}
func WithNamespace(namespace string) OptChaos {
	return func(opt *ConfigChaos) {
		opt.Namespace = namespace
	}
}

func WithLabels(labels map[string]string) OptChaos {
	return func(opt *ConfigChaos) {
		opt.Labels = labels
	}
}

func WithAnnotations(annotations map[string]string) OptChaos {
	return func(opt *ConfigChaos) {
		opt.Annotations = annotations
	}
}

func WithPodChaosSpec(spec *chaosmeshv1alpha1.PodChaosSpec) OptChaos {
	return func(opt *ConfigChaos) {
		opt.PodChaos = spec
	}
}

func WithStressChaosSpec(spec *chaosmeshv1alpha1.StressChaosSpec) OptChaos {
	return func(opt *ConfigChaos) {
		opt.StressChaos = spec
	}
}

func WithHttpChaosSpec(spec *chaosmeshv1alpha1.HTTPChaosSpec) OptChaos {
	return func(opt *ConfigChaos) {
		opt.HttpChaos = spec
	}
}

func WithIOChaosSpec(spec *chaosmeshv1alpha1.IOChaosSpec) OptChaos {
	return func(config *ConfigChaos) {
		config.IOChaos = spec
	}
}

func WithTimeChaosSpec(spec *v1alpha1.TimeChaosSpec) OptChaos {
	return func(config *ConfigChaos) {
		config.TimeChaos = spec
	}
}

func WithNetworkChaosSpec(spec *chaosmeshv1alpha1.NetworkChaosSpec) OptChaos {
	return func(config *ConfigChaos) {
		config.NetworkChaos = spec
	}
}

func WithDnsChaosSpec(spec *chaosmeshv1alpha1.DNSChaosSpec) OptChaos {
	return func(config *ConfigChaos) {
		config.DNSChaos = spec
	}
}

func WithJVMChaosSpec(spec *chaosmeshv1alpha1.JVMChaosSpec) OptChaos {
	return func(config *ConfigChaos) {
		config.JVMChaos = spec
	}
}

func WithWorkflowSpec(spec *chaosmeshv1alpha1.WorkflowSpec) OptChaos {
	return func(opt *ConfigChaos) {
		opt.Workflow = spec
	}
}
