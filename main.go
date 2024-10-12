package main

import (
	controllers "chaos-expriment/controllers"
	"os"
	"path/filepath"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

	namespace := "onlineboutique"
	appList := []string{"checkoutservice", "recommendationservice", "emailservice", "paymentservice", "productcatalogservice"}
	stressors := controllers.MakeCPUStressors(100, 2)
	controllers.ScheduleStressChaos(k8sClient, namespace, appList, stressors, "cpu")
}
