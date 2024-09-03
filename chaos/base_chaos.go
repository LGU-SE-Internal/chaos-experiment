package chaos

import chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"

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

func WithHttpChaosSpec(spec *chaosmeshv1alpha1.HTTPChaosSpec) OptChaos {
	return func(opt *ConfigChaos) {
		opt.HttpChaos = spec
	}
}
func WithWorkflowSpec(spec *chaosmeshv1alpha1.WorkflowSpec) OptChaos {
	return func(opt *ConfigChaos) {
		opt.Workflow = spec
	}
}
