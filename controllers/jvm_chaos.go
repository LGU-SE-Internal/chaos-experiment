package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/CUHK-SE-Group/chaos-experiment/chaos"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateJVMChaos(cli client.Client, namespace string, appName string, action v1alpha1.JVMChaosAction, duration *string, opts ...chaos.OptJVMChaos) string {
	spec := chaos.GenerateJVMChaosSpec(namespace, appName, duration, append([]chaos.OptJVMChaos{chaos.WithJVMAction(action)}, opts...)...)
	name := strings.ToLower(fmt.Sprintf("%s-%s-%s-%s", namespace, appName, action, rand.String(6)))
	jvmChaos, err := chaos.NewJvmChaos(chaos.WithName(name), chaos.WithNamespace(namespace), chaos.WithJVMChaosSpec(spec))
	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
		return ""
	}
	create, err := jvmChaos.ValidateCreate()
	if err != nil {
		logrus.Errorf("Failed to validate create chaos: %v", err)
		return ""
	}
	logrus.Infof("create warning: %v", create)
	err = cli.Create(context.Background(), jvmChaos)
	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
		return ""
	}
	return name
}

func AddJVMChaosWorkflowNodes(workflowSpec *v1alpha1.WorkflowSpec, namespace string, appList []string, action v1alpha1.JVMChaosAction, injectTime *string, sleepTime *string, opts ...chaos.OptJVMChaos) *v1alpha1.WorkflowSpec {
	for _, appName := range appList {
		spec := chaos.GenerateJVMChaosSpec(namespace, appName, nil, append([]chaos.OptJVMChaos{chaos.WithJVMAction(action)}, opts...)...)

		workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
			Name: strings.ToLower(fmt.Sprintf("%s-%s-%s-%s", namespace, appName, action, rand.String(6))),
			Type: v1alpha1.TypeJVMChaos,
			EmbedChaos: &v1alpha1.EmbedChaos{
				JVMChaos: spec,
			},
			Deadline: injectTime,
		})

		workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
			Name:     fmt.Sprintf("%s-%s", "sleep", rand.String(6)),
			Type:     v1alpha1.TypeSuspend,
			Deadline: sleepTime,
		})
	}

	return workflowSpec
}

func ScheduleJVMChaos(cli client.Client, namespace string, appList []string, action v1alpha1.JVMChaosAction, opts ...chaos.OptJVMChaos) {
	workflowName := strings.ToLower(fmt.Sprintf("%s-%s-%s", namespace, action, rand.String(6)))
	workflowSpec := v1alpha1.WorkflowSpec{
		Entry: workflowName,
		Templates: []v1alpha1.Template{
			{
				Name:     workflowName,
				Type:     v1alpha1.TypeSerial,
				Children: nil,
			},
		},
	}
	for idx, appName := range appList {
		spec := chaos.GenerateJVMChaosSpec(namespace, appName, nil, append([]chaos.OptJVMChaos{chaos.WithJVMAction(action)}, opts...)...)

		workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
			Name: strings.ToLower(fmt.Sprintf("%s-%s-%s", namespace, appName, action)),
			Type: v1alpha1.TypeJVMChaos,
			EmbedChaos: &v1alpha1.EmbedChaos{
				JVMChaos: spec,
			},
			Deadline: pointer.String("5m"),
		})
		if idx < len(appList)-1 {
			workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
				Name:     fmt.Sprintf("%s-%d", "sleep", idx),
				Type:     v1alpha1.TypeSuspend,
				Deadline: pointer.String("10m"),
			})
		}
	}

	for i, template := range workflowSpec.Templates {
		if i == 0 {
			continue
		}
		workflowSpec.Templates[0].Children = append(workflowSpec.Templates[0].Children, template.Name)
	}

	workflowChaos, err := chaos.NewWorkflowChaos(chaos.WithName(workflowName), chaos.WithNamespace(namespace), chaos.WithWorkflowSpec(&workflowSpec))
	if err != nil {
		logrus.Errorf("Failed to create chaos workflow: %v", err)
		return
	}

	create, err := workflowChaos.ValidateCreate()
	if err != nil {
		logrus.Errorf("Failed to validate create chaos: %v", err)
		return
	}
	logrus.Infof("create warning: %v", create)
	err = cli.Create(context.Background(), workflowChaos)
	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
	}
}
