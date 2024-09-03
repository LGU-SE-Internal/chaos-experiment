package controllers

import (
	"chaos-expriment/chaos"
	"context"
	"encoding/json"
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ScheduleChaos(cli client.Client, namespace string) {
	ctx := context.Background()

	podList := &corev1.PodList{}

	listOptions := &client.ListOptions{
		Namespace: namespace,
	}

	if err := cli.List(ctx, podList, listOptions); err != nil {
		logrus.Errorf("Failed to list pods: %v", err)
		return
	}

	for _, pod := range podList.Items {
		if pod.Status.Phase != corev1.PodRunning {
			logrus.Infof("Pod %s in namespace %s is not running (status: %s). Deleting pod.", pod.Name, pod.Namespace, pod.Status.Phase)

			if err := cli.Delete(ctx, &pod, &client.DeleteOptions{}); err != nil {
				logrus.Errorf("Failed to delete pod %s/%s: %v", pod.Namespace, pod.Name, err)
			} else {
				logrus.Infof("Successfully deleted pod %s/%s", pod.Namespace, pod.Name)
			}
		}
	}

	spec := v1alpha1.HTTPChaosSpec{
		PodSelector: v1alpha1.PodSelector{
			Selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					Namespaces: []string{namespace},
					//FieldSelectors:      nil,
					//LabelSelectors: nil,
					//ExpressionSelectors: nil,
					//AnnotationSelectors: nil,
				},
				//Nodes:             nil,
				Pods: nil,
				//NodeSelectors:     nil,
				//PodPhaseSelectors: nil,
			},
			Mode:  v1alpha1.OneMode,
			Value: "",
		},
		Target: v1alpha1.PodHttpRequest,
		PodHttpChaosActions: v1alpha1.PodHttpChaosActions{
			//Abort:   nil,
			Delay: pointer.String("5s"),
			//Replace: nil,
			//Patch:   nil,
		},
		Port: 8080,
		//Path:            nil,
		//Method:          nil,
		//Code:            nil,
		//RequestHeaders:  nil,
		//ResponseHeaders: nil,
		//TLS:             nil,
		Duration: pointer.String("1m"),
		//RemoteCluster:       "",
	}
	httpChaos, err := chaos.NewHttpChaos(chaos.WithNamespace(namespace), chaos.WithName("test111"), chaos.WithPodHttpChaosSpec(&spec))
	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
	}
	jsonDataIndented, err := json.MarshalIndent(httpChaos, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to indented JSON:", err)
		return
	}

	// 打印带缩进的 JSON 字符串
	fmt.Println("Indented JSON format:")
	fmt.Println(string(jsonDataIndented))
}
