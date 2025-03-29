package handler

import (
	"fmt"
	"strconv"

	chaos "github.com/CUHK-SE-Group/chaos-experiment/chaos"
	"github.com/CUHK-SE-Group/chaos-experiment/client"
	controllers "github.com/CUHK-SE-Group/chaos-experiment/controllers"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"k8s.io/utils/pointer"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

type ChaosType int

const TargetNamespace = "ts" // todo: make it dynamic (e.g. from config)
const TargetLabelKey = "app"

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

type HTTPChaosTarget int

const (
	Request  HTTPChaosTarget = 1
	Response HTTPChaosTarget = 2
)

type Injection interface {
	Create(cli cli.Client) string
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

type ContainerKillSpec struct {
	Duration  int `range:"1-60" description:"time unit minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"string"`
	AppName   int `range:"0-0" dynamic:"true" description:"array"`
}

func (s *ContainerKillSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.ContainerKillAction
	return controllers.CreatePodChaos(cli, TargetNamespace, labelArr[s.AppName], action, duration)
}

type PodFailureSpec struct {
	Duration  int `range:"1-60" description:"time unit minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"string"`
	AppName   int `range:"0-0" dynamic:"true" description:"array"`
}

func (s *PodFailureSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.PodFailureAction
	return controllers.CreatePodChaos(cli, TargetNamespace, labelArr[s.AppName], action, duration)
}

type PodKillSpec struct {
	Duration  int `range:"1-60" description:"time unit minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"string"`
	AppName   int `range:"0-0" dynamic:"true" description:"array"`
}

func (s *PodKillSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.PodKillAction
	return controllers.CreatePodChaos(cli, TargetNamespace, labelArr[s.AppName], action, duration)
}

type CPUStressChaosSpec struct {
	CPULoad   int `range:"1-100" description:"CPU Load Percentage"`
	CPUWorker int `range:"1-3" description:"CPU Stress Threads"`
	Duration  int `range:"1-60" description:"time unit minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"string"`
	AppName   int `range:"0-0" dynamic:"true" description:"array"`
}

func (s *CPUStressChaosSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	stressors := controllers.MakeCPUStressors(
		s.CPULoad,
		s.CPUWorker,
	)
	return controllers.CreateStressChaos(cli, TargetNamespace, labelArr[s.AppName], stressors, "cpu-exhaustion", duration)
}

type MemoryStressChaosSpec struct {
	MemorySize int `range:"1-1024" description:"Memory Size Unit MB"`
	MemWorker  int `range:"1-4" description:"Memory Stress Threads"`
	Duration   int `range:"1-60" description:"Time Unit Minute"`
	Namespace  int `range:"0-0" dynamic:"true" description:"String"`
	AppName    int `range:"0-0" dynamic:"true" description:"Array"`
}

func (s *MemoryStressChaosSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	stressors := controllers.MakeMemoryStressors(
		strconv.Itoa(s.MemorySize)+"MiB",
		s.MemWorker,
	)
	return controllers.CreateStressChaos(cli, TargetNamespace, labelArr[s.AppName], stressors, "memory-exhaustion", pointer.String(strconv.Itoa(s.Duration)+"m"))
}

type HTTPChaosReplaceSpec struct {
	HTTPTarget  HTTPChaosTarget `range:"1-2" description:"HTTP Phase Request/Response"`
	ReplaceBody HTTPReplaceBody `range:"1-2" description:"Body Replacement Blank/Random"`
	Duration    int             `range:"1-60" description:"Time Unit Minute"`
	Namespace   int             `range:"0-0" dynamic:"true" description:"String"`
	AppName     int             `range:"0-0" dynamic:"true" description:"Array"`
}

func (s *HTTPChaosReplaceSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	target := httpChaosTargetMap[s.HTTPTarget]
	opts := []chaos.OptHTTPChaos{
		chaos.WithTarget(target),
		chaos.WithPort(8080),
		httpReplaceBodyMap[s.ReplaceBody],
	}
	return controllers.CreateHTTPChaos(cli, TargetNamespace, labelArr[s.AppName], fmt.Sprintf("%s-replace", target), duration, opts...)
}

type HTTPChaosAbortSpec struct {
	HTTPTarget HTTPChaosTarget `range:"1-2" description:"HTTP Phase Request/Response"`
	Duration   int             `range:"1-60" description:"time unit minute"`
	Namespace  int             `range:"0-0" dynamic:"true" description:"string"`
	AppName    int             `range:"0-0" dynamic:"true" description:"array"`
}

func (s *HTTPChaosAbortSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	abort := true
	target := httpChaosTargetMap[s.HTTPTarget]
	opts := []chaos.OptHTTPChaos{
		chaos.WithTarget(target),
		chaos.WithPort(8080),
		chaos.WithAbort(&abort),
	}
	return controllers.CreateHTTPChaos(cli, TargetNamespace, labelArr[s.AppName], fmt.Sprintf("%s-abort", target), duration, opts...)
}

var SpecMap = map[ChaosType]any{
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
	HTTPReplace:   &HTTPChaosReplaceSpec{},
}

type InjectionConf struct {
	PodKill       *PodKillSpec           `range:"0-2"`
	PodFailure    *PodFailureSpec        `range:"0-2"`
	ContainerKill *ContainerKillSpec     `range:"0-2"`
	MemoryStress  *MemoryStressChaosSpec `range:"0-4"`
	CPUStress     *CPUStressChaosSpec    `range:"0-4"`
	HTTPAbort     *HTTPChaosAbortSpec    `range:"0-3"`
	HTTPReplace   *HTTPChaosReplaceSpec  `range:"0-4"`
}
