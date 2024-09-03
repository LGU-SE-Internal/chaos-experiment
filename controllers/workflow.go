package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Workflow(cli client.Client, namespace string) {
	spec := v1alpha1.WorkflowList{}
	err := cli.List(context.Background(), &spec)
	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
	}
	logrus.Infof("Chaos will become ready %+v", spec)

	jsonDataIndented, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to indented JSON:", err)
		return
	}
	fmt.Println(string(jsonDataIndented))

}
