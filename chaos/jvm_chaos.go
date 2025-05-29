package chaos

import (
	"errors"
	"fmt"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func NewJvmChaos(opts ...OptChaos) (*chaosmeshv1alpha1.JVMChaos, error) {
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
	if config.JVMChaos == nil {
		return nil, errors.New("jvmChaos is required")
	}

	jvmChaos := chaosmeshv1alpha1.JVMChaos{}
	jvmChaos.Name = config.Name
	jvmChaos.Namespace = config.Namespace
	config.JVMChaos.DeepCopyInto(&jvmChaos.Spec)

	if config.Labels != nil {
		jvmChaos.Labels = config.Labels
	}
	if config.Annotations != nil {
		jvmChaos.Annotations = config.Annotations
	}

	return &jvmChaos, nil
}

type OptJVMChaos func(opt *chaosmeshv1alpha1.JVMChaosSpec)

func WithJVMAction(action chaosmeshv1alpha1.JVMChaosAction) OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.Action = action
	}
}

func WithJVMClass(class string) OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.JVMClassMethodSpec.Class = class
	}
}

func WithJVMMethod(method string) OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.JVMClassMethodSpec.Method = method
	}
}

func WithJVMLatencyDuration(latency int) OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.LatencyDuration = latency
	}
}

func WithJVMReturnValue(returnValue string) OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.ReturnValue = returnValue
	}
}

func WithJVMRandomIntReturn(min, max int) OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.ReturnValue = fmt.Sprintf("%d", min+rand.Intn(max-min+1))
	}
}

func WithJVMRandomStringReturn(length int) OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		randomStr := rand.String(length)
		opt.JVMParameter.ReturnValue = fmt.Sprintf("\"%s\"", randomStr)
	}
}

func WithJVMDefaultStringReturn() OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.ReturnValue = "\"chaos\""
	}
}

func WithJVMDefaultIntReturn() OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.ReturnValue = "42"
	}
}

func WithJVMException(exception string) OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.ThrowException = exception
	}
}

func WithJVMDefaultException() OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.ThrowException = "java.io.IOException(\"BOOM\")"
	}
}

func WithJVMName(name string) OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.Name = name
	}
}

func WithJVMRuleData(ruleData string) OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.RuleData = ruleData
	}
}

// JVM Stress helpers
func WithJVMStressCPUCount(count int) OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.CPUCount = count
	}
}

func WithJVMStressMemType(memType string) OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.MemoryType = memType
	}
}

// JVM MySQL helpers
func WithJVMMySQLConnector(version string) OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.MySQLConnectorVersion = version
	}
}

func WithJVMMySQLDatabase(database string) OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.Database = database
	}
}

func WithJVMMySQLTable(table string) OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.Table = table
	}
}

func WithJVMMySQLType(sqlType string) OptJVMChaos {
	return func(opt *chaosmeshv1alpha1.JVMChaosSpec) {
		opt.JVMParameter.SQLType = sqlType
	}
}

func GenerateJVMChaosSpec(namespace string, appName string, duration *string, opts ...OptJVMChaos) *chaosmeshv1alpha1.JVMChaosSpec {
	spec := &chaosmeshv1alpha1.JVMChaosSpec{
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
		JVMParameter: chaosmeshv1alpha1.JVMParameter{},
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
