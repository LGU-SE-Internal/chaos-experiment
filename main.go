package main

import (
	chaos "chaos-expriment/chaos"
	controllers "chaos-expriment/controllers"
	"os"
	"path/filepath"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Load configuration
func getK8sConfig() *rest.Config {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config
}

func main() {
	cfg := getK8sConfig()
	scheme := runtime.NewScheme()
	err := chaosmeshv1alpha1.AddToScheme(scheme)
	if err != nil {
		logrus.Fatalf("add chaosmeshv1alpha1 scheme: %v", err)
	}
	err = corev1.AddToScheme(scheme)
	if err != nil {
		logrus.Fatalf("add corev1 scheme: %v", err)
	}

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		logrus.Fatalf("create k8sClient: %v", err)
	}

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
