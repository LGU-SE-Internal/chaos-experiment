package handler

import (
	"fmt"
	"strconv"

	chaos "github.com/CUHK-SE-Group/chaos-experiment/chaos"
	"github.com/CUHK-SE-Group/chaos-experiment/client"
	controllers "github.com/CUHK-SE-Group/chaos-experiment/controllers"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/utils/pointer"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

type ChaosType int

const (
	// Default indicates an unknown Type.
	Default ChaosType = iota

	// PodChaos
	PodKill
	PodFailure
	ContainerKill

	// StressChaos
	MemoryStress
	CPUStress

	// HTTPChaos
	HTTPAbort
	HTTPDelay
	HTTPReplace
	HTTPPatch

	// ...
)

// 定义 ChaosType 对应的 map
var ChaosTypeMap = map[ChaosType]string{
	PodKill:       "PodKill",
	PodFailure:    "PodFailure",
	ContainerKill: "ContainerKill",
	MemoryStress:  "MemoryStress",
	CPUStress:     "CPUStress",
	HTTPAbort:     "HTTPAbort",
	HTTPDelay:     "HTTPDelay",
	HTTPReplace:   "HTTPReplace",
	HTTPPatch:     "HTTPPatch",
}

// GetChaosTypeName 根据 ChaosType 获取名称
func GetChaosTypeName(c ChaosType) string {
	if name, ok := ChaosTypeMap[c]; ok {
		return name
	}
	return "Unknown"
}

type ChaosConfig struct {
	Type     ChaosType   `range:"1-9"`
	Spec     interface{} `optional:"true"`
	Duration int         `range:"1-60"`
}
type HTTPChaosTarget int

const (
	Request  HTTPChaosTarget = 1
	Response HTTPChaosTarget = 2
)

type Injection interface {
	Create(cli cli.Client, namespace string, appName string, config ChaosConfig) string
	GetGroudtruth() []Groudtruth
}

var httpChaosTargetMap = map[HTTPChaosTarget]chaosmeshv1alpha1.PodHttpChaosTarget{
	Request:  chaosmeshv1alpha1.PodHttpRequest,
	Response: chaosmeshv1alpha1.PodHttpResponse,
}

type HTTPReplaceBody int

const (
	Blank  HTTPReplaceBody = 1
	Random HTTPReplaceBody = 2
)

var httpReplaceBodyMap = map[HTTPReplaceBody]chaos.OptHTTPChaos{
	Blank:  chaos.WithReplaceBody([]byte("")),
	Random: chaos.WithRandomReplaceBody(),
}

type ContainerKillSpec struct{}

func (s *ContainerKillSpec) Create(cli cli.Client, namespace string, appName string, config ChaosConfig) string {
	duration := pointer.String(strconv.Itoa(config.Duration) + "m")
	action := chaosmeshv1alpha1.ContainerKillAction
	return controllers.CreatePodChaos(cli, namespace, appName, action, duration)
}

type PodFailureSpec struct{}

func (s *PodFailureSpec) Create(cli cli.Client, namespace string, appName string, config ChaosConfig) string {
	duration := pointer.String(strconv.Itoa(config.Duration) + "m")
	action := chaosmeshv1alpha1.PodFailureAction
	return controllers.CreatePodChaos(cli, namespace, appName, action, duration)
}

type PodKillSpec struct{}

func (s *PodKillSpec) Create(cli cli.Client, namespace string, appName string, config ChaosConfig) string {
	duration := pointer.String(strconv.Itoa(config.Duration) + "m")
	action := chaosmeshv1alpha1.PodKillAction
	return controllers.CreatePodChaos(cli, namespace, appName, action, duration)
}

type CPUStressChaosSpec struct {
	CPULoad   int `range:"1-100"`
	CPUWorker int `range:"1-3"`
}

func (s *CPUStressChaosSpec) Create(cli cli.Client, namespace string, appName string, config ChaosConfig) string {
	duration := pointer.String(strconv.Itoa(config.Duration) + "m")
	if cpuSpec, ok := config.Spec.(CPUStressChaosSpec); ok {
		stressors := controllers.MakeCPUStressors(
			cpuSpec.CPULoad,
			cpuSpec.CPUWorker,
		)
		return controllers.CreateStressChaos(cli, namespace, appName, stressors, "cpu-exhaustion", duration)
	} else {
		logrus.Error("Invalid cpu stress spec")
		return ""
	}
}

type MemoryStressChaosSpec struct {
	MemorySize int `range:"1-1024"`
	MemWorker  int `range:"1-4"`
}

func (s *MemoryStressChaosSpec) Create(cli cli.Client, namespace string, appName string, config ChaosConfig) string {
	duration := pointer.String(strconv.Itoa(config.Duration) + "m")
	if memorySpec, ok := config.Spec.(MemoryStressChaosSpec); ok {
		stressors := controllers.MakeMemoryStressors(
			strconv.Itoa(memorySpec.MemorySize)+"MiB",
			memorySpec.MemWorker,
		)
		return controllers.CreateStressChaos(cli, namespace, appName, stressors, "memory-exhaustion", duration)
	} else {
		logrus.Error("Invalid memory stress spec")
		return ""
	}
}

type HTTPChaosReplaceSpec struct {
	HTTPTarget  HTTPChaosTarget `range:"1-2"`
	ReplaceBody HTTPReplaceBody `range:"1-2"`
}

func (s *HTTPChaosReplaceSpec) Create(cli cli.Client, namespace string, appName string, config ChaosConfig) string {
	duration := pointer.String(strconv.Itoa(config.Duration) + "m")

	if replaceSpec, ok := config.Spec.(HTTPChaosReplaceSpec); ok {
		target := httpChaosTargetMap[replaceSpec.HTTPTarget]
		opts := []chaos.OptHTTPChaos{
			chaos.WithTarget(target),
			chaos.WithPort(8080),
			httpReplaceBodyMap[replaceSpec.ReplaceBody],
		}
		return controllers.CreateHTTPChaos(cli, namespace, appName, fmt.Sprintf("%s-replace", target), duration, opts...)
	} else {
		logrus.Error("Invalid http replace spec")
		return ""
	}
}

type HTTPChaosAbortSpec struct {
	HTTPTarget HTTPChaosTarget `range:"1-2"`
}

func (s *HTTPChaosAbortSpec) Create(cli cli.Client, namespace string, appName string, config ChaosConfig) string {
	duration := pointer.String(strconv.Itoa(config.Duration) + "m")

	abort := true
	if abortSpec, ok := config.Spec.(HTTPChaosAbortSpec); ok {
		target := httpChaosTargetMap[abortSpec.HTTPTarget]
		opts := []chaos.OptHTTPChaos{
			chaos.WithTarget(target),
			chaos.WithPort(8080),
			chaos.WithAbort(&abort),
		}
		return controllers.CreateHTTPChaos(cli, namespace, appName, fmt.Sprintf("%s-abort", target), duration, opts...)
	} else {
		logrus.Error("Invalid http abort spec")
		return ""
	}
}

var SpecMap = map[ChaosType]interface{}{
	CPUStress:    CPUStressChaosSpec{},
	MemoryStress: MemoryStressChaosSpec{},
	HTTPAbort:    HTTPChaosAbortSpec{},
	HTTPReplace:  HTTPChaosReplaceSpec{},
}

var ChaosHandlers = map[ChaosType]Injection{
	PodKill:       &PodKillSpec{},
	PodFailure:    &PodFailureSpec{},
	ContainerKill: &ContainerKillSpec{},
	MemoryStress:  &MemoryStressChaosSpec{},
	CPUStress:     &CPUStressChaosSpec{},
	HTTPAbort:     &HTTPChaosAbortSpec{},
	// HTTPDelay:  TODO
	HTTPReplace: &HTTPChaosReplaceSpec{},
	// HTTPPatch: TODO
	// TODO: Implement other chaos types
}

func Create(namespace string, appName string, config ChaosConfig) string {
	k8sClient := client.NewK8sClient()
	if handler, exists := ChaosHandlers[config.Type]; exists {
		return handler.Create(k8sClient, namespace, appName, config)
	}
	return ""
}
