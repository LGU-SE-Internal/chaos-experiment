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

func CreateTimeChaos(cli client.Client, namespace string, appName string, timeOffset string, duration *string) string {
	spec := chaos.GenerateTimeChaosSpec(namespace, appName, duration, timeOffset)
	name := strings.ToLower(fmt.Sprintf("%s-%s-time-%s", namespace, appName, rand.String(6)))
	timeChaos, err := chaos.NewTimeChaos(chaos.WithName(name), chaos.WithNamespace(namespace), chaos.WithTimeChaosSpec(spec))
	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
		return ""
	}
	create, err := timeChaos.ValidateCreate()
	if err != nil {
		logrus.Errorf("Failed to validate create chaos: %v", err)
		return ""
	}
	logrus.Infof("create warning: %v", create)
	err = cli.Create(context.Background(), timeChaos)
	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
		return ""
	}
	return name
}

func AddTimeChaosWorkflowNodes(workflowSpec *v1alpha1.WorkflowSpec, namespace string, appList []string, timeOffset string, injectTime *string, sleepTime *string) *v1alpha1.WorkflowSpec {
	for _, appName := range appList {
		spec := chaos.GenerateTimeChaosSpec(namespace, appName, nil, timeOffset)

		workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
			Name: strings.ToLower(fmt.Sprintf("%s-%s-time-%s", namespace, appName, rand.String(6))),
			Type: v1alpha1.TypeTimeChaos,
			EmbedChaos: &v1alpha1.EmbedChaos{
				TimeChaos: spec,
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

func ScheduleTimeChaos(cli client.Client, namespace string, appList []string, timeOffset string) {
	workflowName := strings.ToLower(fmt.Sprintf("%s-time-%s", namespace, rand.String(6)))
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
		spec := chaos.GenerateTimeChaosSpec(namespace, appName, nil, timeOffset)

		workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
			Name: strings.ToLower(fmt.Sprintf("%s-%s-time", namespace, appName)),
			Type: v1alpha1.TypeTimeChaos,
			EmbedChaos: &v1alpha1.EmbedChaos{
				TimeChaos: spec,
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
