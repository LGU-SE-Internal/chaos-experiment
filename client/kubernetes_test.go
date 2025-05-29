package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/k0kubun/pp/v3"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestGetLabel(t *testing.T) {
	labels, err := GetLabels(context.Background(), "ts0", "app")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(labels)
}
func TestCRDClient(t *testing.T) {
	k8sClient := NewK8sClient()
	ctx := context.Background()
	podChaosList := &v1alpha1.StressChaosList{}
	nameToQuery := "ts-ts-train-service-cpu-exhaustion-7mwd86"
	listOptions := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", nameToQuery),
		Namespace:     "ts",
	}
	err := k8sClient.List(ctx, podChaosList, listOptions)
	if err != nil {
		fmt.Printf("Failed to list PodChaos: %v\n", err)
		return
	}

	if len(podChaosList.Items) == 0 {
		fmt.Printf("No PodChaos found with name: %s\n", nameToQuery)
	} else {
		for _, podChaos := range podChaosList.Items {
			fmt.Printf("%+v", podChaos)
		}
	}
}
func TestCRDClient1(t *testing.T) {
	start, end, err := QueryCRDByName("ts", "ts-ts-train-service-cpu-exhaustion-7mwd86")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(start, end)
}

func TestGetContainersWithAppLabel(t *testing.T) {
	containerInfos, err := GetContainersWithAppLabel(context.Background(), "ts0")
	if err != nil {
		t.Error(err)
	}

	pp.Println(containerInfos)
}
