package main

import (
	chaos "github.com/LGU-SE-Internal/chaos-experiment/chaos"
	"github.com/LGU-SE-Internal/chaos-experiment/client"
	controllers "github.com/LGU-SE-Internal/chaos-experiment/controllers"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/pointer"
)

func main() {
	k8sClient := client.GetK8sClient()

	namespace := "ts"

	appList := []string{"ts-consign-service", "ts-route-service", "ts-train-service", "ts-travel-service", "ts-basic-service", "ts-food-service", "ts-security-service", "ts-seat-service", "ts-routeplan-service", "ts-travel2-service"}
	workflowSpec := controllers.NewWorkflowSpec(namespace)
	sleepTime := pointer.String("15m")
	injectTime := pointer.String("5m")
	// Add cpu
	stressors := controllers.MakeCPUStressors(100, 5)
	controllers.AddStressChaosWorkflowNodes(workflowSpec, namespace, appList, stressors, "cpu", injectTime, sleepTime)
	// Add memory
	stressors = controllers.MakeMemoryStressors("1GB", 1)
	controllers.AddStressChaosWorkflowNodes(workflowSpec, namespace, appList, stressors, "memory", injectTime, sleepTime)
	// Add Pod failure
	action := chaosmeshv1alpha1.PodFailureAction
	controllers.AddPodChaosWorkflowNodes(workflowSpec, namespace, appList, action, injectTime, sleepTime)
	// Add abort
	abort := true
	opts1 := []chaos.OptHTTPChaos{
		chaos.WithTarget(chaosmeshv1alpha1.PodHttpRequest),
		chaos.WithPort(8080),
		chaos.WithAbort(&abort),
	}
	controllers.AddHTTPChaosWorkflowNodes(workflowSpec, namespace, appList, "request-abort", injectTime, sleepTime, opts1...)
	// add replace
	opts2 := []chaos.OptHTTPChaos{
		chaos.WithTarget(chaosmeshv1alpha1.PodHttpResponse),
		chaos.WithPort(8080),
		chaos.WithReplaceBody([]byte(rand.String(6))),
	}
	controllers.AddHTTPChaosWorkflowNodes(workflowSpec, namespace, appList, "response-replace", injectTime, sleepTime, opts2...)
	// create workflow
	controllers.CreateWorkflow(k8sClient, workflowSpec, namespace)

}
