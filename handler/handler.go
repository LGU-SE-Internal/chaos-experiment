package handler

import (
	"fmt"
	"strconv"

	chaos "github.com/CUHK-SE-Group/chaos-experiment/chaos"
	"github.com/CUHK-SE-Group/chaos-experiment/client"
	controllers "github.com/CUHK-SE-Group/chaos-experiment/controllers"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"k8s.io/utils/pointer"
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

type ChaosConfig struct {
	Type ChaosType `range:"1-9"`
	Spec ChaosSpec
}
type ChaosSpec struct {
	// common
	InjectTime int `range:"1-60"`
	SleepTime  int `range:"1-60"`
	
	// StressChaos
	MemorySize int `range:"1-262144" optional:"true"` 
	MemWorker  int `range:"1-8192" optional:"true"`
	CPULoad    int `range:"1-100" optional:"true"`
	CPUWorker  int `range:"1-8192" optional:"true"`

	// HTTPChaos
	HTTPTarget  HTTPChaosTarget `range:"1-2" optional:"true"`
	ReplaceBody HTTPReplaceBody `range:"1-2" optional:"true"`

	// TODO: Implement other chaos types
}
type HTTPChaosTarget int

const (
	Request  HTTPChaosTarget = 1
	Response HTTPChaosTarget = 2
)

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

func CreateChaosHandlers(workflowSpec *chaosmeshv1alpha1.WorkflowSpec, namespace string, appList []string, config ChaosConfig) map[ChaosType]func() {
	injectTime := pointer.String(strconv.Itoa(config.Spec.InjectTime) + "m")
	sleepTime := pointer.String(strconv.Itoa(config.Spec.SleepTime) + "m")
	return map[ChaosType]func(){
		PodKill: func() {
			action := chaosmeshv1alpha1.PodKillAction
			controllers.AddPodChaosWorkflowNodes(workflowSpec, namespace, appList, action, injectTime, sleepTime)
		},
		PodFailure: func() {
			action := chaosmeshv1alpha1.PodFailureAction
			controllers.AddPodChaosWorkflowNodes(workflowSpec, namespace, appList, action, injectTime, sleepTime)
		},
		ContainerKill: func() {
			action := chaosmeshv1alpha1.ContainerKillAction
			controllers.AddPodChaosWorkflowNodes(workflowSpec, namespace, appList, action, injectTime, sleepTime)
		},
		MemoryStress: func() {
			stressors := controllers.MakeMemoryStressors(strconv.Itoa(config.Spec.MemorySize)+"MiB", config.Spec.MemWorker)
			controllers.AddStressChaosWorkflowNodes(workflowSpec, namespace, appList, stressors, "memory-exhaustion", injectTime, sleepTime)
		},
		CPUStress: func() {
			stressors := controllers.MakeCPUStressors(config.Spec.CPULoad, config.Spec.CPUWorker)
			controllers.AddStressChaosWorkflowNodes(workflowSpec, namespace, appList, stressors, "cpu-exhaustion", injectTime, sleepTime)
		},

		HTTPAbort: func() {
			abort := true
			target := httpChaosTargetMap[config.Spec.HTTPTarget]
			opts := []chaos.OptHTTPChaos{
				chaos.WithTarget(target),
				chaos.WithPort(8080),
				chaos.WithAbort(&abort),
			}
			controllers.AddHTTPChaosWorkflowNodes(workflowSpec, namespace, appList, fmt.Sprintf("%s-abort", target), injectTime, sleepTime, opts...)
		},
		HTTPDelay: func() {
			// TODO
		},
		HTTPReplace: func() {
			target := httpChaosTargetMap[config.Spec.HTTPTarget]
			opts := []chaos.OptHTTPChaos{
				chaos.WithTarget(target),
				chaos.WithPort(8080),
				httpReplaceBodyMap[config.Spec.ReplaceBody],
			}
			controllers.AddHTTPChaosWorkflowNodes(workflowSpec, namespace, appList, fmt.Sprintf("%s-replace", target), injectTime, sleepTime, opts...)

		},
		HTTPPatch: func() {
			// TODO
		},
		// TODO: Implement other chaos types
	}
}

func Create(config ChaosConfig) {
	k8sClient := client.NewK8sClient()
	namespace := "ts"

	appList := []string{"ts-train-service"}
	workflowSpec := controllers.NewWorkflowSpec(namespace)

	handlers := CreateChaosHandlers(workflowSpec, namespace, appList, config)

	if handler, exists := handlers[config.Type]; exists {
		handler()
	}
	controllers.CreateWorkflow(k8sClient, workflowSpec, namespace)
}
