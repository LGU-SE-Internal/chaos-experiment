package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"

	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
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

	for _, pod := range podList.Items {
		if value, exists := pod.Labels[key]; exists {
			labelValues = append(labelValues, value)
		}
	}

	slices.Sort(labelValues)
	labelValues = slices.Compact(labelValues)
	return labelValues, nil
}

// QueryCRDByName 查询指定命名空间和名称的 CRD，并检查其状态
func QueryCRDByName(namespace, nameToQuery string) (time.Time, time.Time, error) {
	k8sClient := NewK8sClient()
	ctx := context.Background()

	// 定义支持的 CRD 类型和对应的 GVR 映射
	// TODO: 添加需要的类型
	crdMapping := map[schema.GroupVersionResource]client.Object{
		{Group: "chaos-mesh.org", Version: "v1alpha1", Resource: "podchaos"}:     &chaosmeshv1alpha1.PodChaos{},
		{Group: "chaos-mesh.org", Version: "v1alpha1", Resource: "networkchaos"}: &chaosmeshv1alpha1.NetworkChaos{},
		{Group: "chaos-mesh.org", Version: "v1alpha1", Resource: "stresschaos"}:  &chaosmeshv1alpha1.StressChaos{},
	}

	for gvr, obj := range crdMapping {
		objCopy := obj.DeepCopyObject().(client.Object)
		err := k8sClient.Get(ctx, client.ObjectKey{Name: nameToQuery, Namespace: namespace}, objCopy)
		if err == nil {
			logrus.Infof("Found resource in GroupVersionResource: %s\n", gvr)

			switch resource := objCopy.(type) {
			case *chaosmeshv1alpha1.PodChaos:
				return checkStatus(resource.Status.ChaosStatus)

			case *chaosmeshv1alpha1.NetworkChaos:
				return checkStatus(resource.Status.ChaosStatus)

			case *chaosmeshv1alpha1.StressChaos:
				return checkStatus(resource.Status.ChaosStatus)
			}
			return time.Time{}, time.Time{}, fmt.Errorf("CRD type not found")
		}
	}

	return time.Time{}, time.Time{}, fmt.Errorf("No resource found for name '%s' in namespace '%s'\n", nameToQuery, namespace)
}

// checkStatus 检查 Chaos 状态是否注入成功和恢复成功
func checkStatus(status chaosmeshv1alpha1.ChaosStatus) (time.Time, time.Time, error) {
	var (
		apply time.Time
		reco  time.Time
	)

	for _, record := range status.Experiment.Records {
		for _, event := range record.Events {
			if event.Operation == chaosmeshv1alpha1.Apply && event.Type == chaosmeshv1alpha1.TypeSucceeded {
				apply = event.Timestamp.Time
			}
			if event.Operation == chaosmeshv1alpha1.Recover && event.Type == chaosmeshv1alpha1.TypeSucceeded {
				reco = event.Timestamp.Time
			}
		}
	}

	// 判断是否找到注入和恢复事件
	if apply.IsZero() && reco.IsZero() {
		return apply, reco, fmt.Errorf("no successful Apply or Recover events found")
	}
	if apply.IsZero() {
		return apply, reco, fmt.Errorf("injection not successful: Apply event missing")
	}
	if reco.IsZero() {
		return apply, reco, fmt.Errorf("injection successful but recovery not successful")
	}

	// 检查注入和恢复的逻辑关系
	if apply.After(reco) {
		return apply, reco, fmt.Errorf("recovery occurred before injection, which is invalid")
	}

	return apply, reco, nil
}
