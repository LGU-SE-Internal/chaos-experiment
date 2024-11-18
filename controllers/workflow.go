package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/CUHK-SE-Group/chaos-experiment/chaos"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/rand"
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

func NewWorkflowSpec(namespace string) *v1alpha1.WorkflowSpec {
	workflowName := strings.ToLower(fmt.Sprintf("%s-%s", namespace, rand.String(6)))
	return &v1alpha1.WorkflowSpec{
		Entry: workflowName,
		Templates: []v1alpha1.Template{
			{
				Name:     workflowName,
				Type:     v1alpha1.TypeSerial,
				Children: nil,
			},
		},
	}
}

func CreateWorkflow(cli client.Client, workflowSpec *v1alpha1.WorkflowSpec, namespace string) {
	for i, template := range workflowSpec.Templates {
		if i == 0 {
			continue
		}
		workflowSpec.Templates[0].Children = append(workflowSpec.Templates[0].Children, template.Name)
	}

	workflowChaos, err := chaos.NewWorkflowChaos(chaos.WithName(workflowSpec.Entry), chaos.WithNamespace(namespace), chaos.WithWorkflowSpec(workflowSpec))
	if err != nil {
		logrus.Errorf("Failed to create chaos workflow: %v", err)
	}

	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
	}

	create, err := workflowChaos.ValidateCreate()
	if err != nil {
		logrus.Errorf("Failed to validate create chaos: %v", err)
	}
	logrus.Infof("create warning: %v", create)
	err = cli.Create(context.Background(), workflowChaos)
	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
	}
}
