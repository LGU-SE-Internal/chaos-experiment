package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/k0kubun/pp/v3"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	k8sClientInstance client.Client
	once              sync.Once
)

// GetK8sConfig returns Kubernetes configuration
func GetK8sConfig() *rest.Config {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config
}

// NewK8sClient initializes a new Kubernetes client using singleton pattern
func NewK8sClient() client.Client {
	once.Do(func() {
		cfg := GetK8sConfig()
		scheme := runtime.NewScheme()

		// Register Chaos Mesh CRD scheme
		err := chaosmeshv1alpha1.AddToScheme(scheme)
		if err != nil {
			logrus.Fatalf("Failed to add Chaos Mesh v1alpha1 scheme: %v", err)
		}

		// Register CoreV1 scheme
		err = corev1.AddToScheme(scheme)
		if err != nil {
			logrus.Fatalf("Failed to add CoreV1 scheme: %v", err)
		}

		// Create Kubernetes client
		k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
		if err != nil {
			logrus.Fatalf("Failed to create Kubernetes client: %v", err)
		}
		k8sClientInstance = k8sClient
	})
	return k8sClientInstance
}

func GetLabels(namespace string, key string) ([]string, error) {
	cli := NewK8sClient()
	labelValues := []string{}

	// List all pods in the specified namespace
	podList := &corev1.PodList{}
	err := cli.List(context.Background(), podList, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		fmt.Printf("Error listing pods in namespace %s: %v\n", namespace, err)
		return nil, err
	}
	pp.Print(podList.ListMeta)

	for _, pod := range podList.Items {
		for label, value := range pod.Labels {
			if key == label {
				labelValues = append(labelValues, value)
			}
		}
	}
	return labelValues, nil
}
