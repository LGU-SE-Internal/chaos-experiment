package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/LGU-SE-Internal/chaos-experiment/chaos"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/k0kubun/pp/v3"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreatePodChaos(cli client.Client, ctx context.Context, namespace string, appName string, action v1alpha1.PodChaosAction, duration *string, annotations map[string]string, labels map[string]string) (string, error) {
	spec := chaos.GeneratePodChaosSpec(namespace, appName, duration, action)
	name := strings.ToLower(fmt.Sprintf("%s-%s-%s-%s", namespace, appName, action, rand.String(6)))
	podChaos, err := chaos.NewPodChaos(
		chaos.WithAnnotations(annotations),
		chaos.WithLabels(labels),
		chaos.WithName(name),
		chaos.WithNamespace(namespace),
		chaos.WithPodChaosSpec(spec),
	)
	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
		return "", err
	}
	create, err := podChaos.ValidateCreate()
	if err != nil {
		logrus.Errorf("Failed to validate create chaos: %v", err)
		return "", err
	}
	logrus.Infof("create warning: %v", create)
	err = cli.Create(ctx, podChaos)
	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
		return "", err
	}
	return name, nil
}

// CreatePodChaosWithContainer creates a pod chaos experiment with specified container names
func CreatePodChaosWithContainer(cli client.Client, ctx context.Context, namespace string, appName string, action v1alpha1.PodChaosAction, duration *string, annotations map[string]string, labels map[string]string, containerNames []string) (string, error) {
	spec := chaos.GeneratePodChaosSpecWithContainers(namespace, appName, duration, action, containerNames)
	name := strings.ToLower(fmt.Sprintf("%s-%s-%s-%s", namespace, appName, action, rand.String(6)))
	podChaos, err := chaos.NewPodChaos(
		chaos.WithAnnotations(annotations),
		chaos.WithLabels(labels),
		chaos.WithName(name),
		chaos.WithNamespace(namespace),
		chaos.WithPodChaosSpec(spec),
	)
	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
		return "", err
	}
	create, err := podChaos.ValidateCreate()
	if err != nil {
		logrus.Errorf("Failed to validate create chaos: %v", err)
		return "", err
	}
	logrus.Infof("create warning: %v", create)
	err = cli.Create(ctx, podChaos)
	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
		return "", err
	}
	return name, nil
}

func AddPodChaosWorkflowNodes(workflowSpec *v1alpha1.WorkflowSpec, namespace string, appList []string, action v1alpha1.PodChaosAction, injectTime *string, sleepTime *string) *v1alpha1.WorkflowSpec {
	for _, appName := range appList {

		spec := chaos.GeneratePodChaosSpec(namespace, appName, nil, action)

		workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
			Name: strings.ToLower(fmt.Sprintf("%s-%s-%s-%s", namespace, appName, action, rand.String(6))),
			Type: v1alpha1.TypePodChaos,
			EmbedChaos: &v1alpha1.EmbedChaos{
				PodChaos: spec,
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

func SchedulePodChaos(cli client.Client, namespace string, appList []string, action v1alpha1.PodChaosAction) {
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

		spec := chaos.GeneratePodChaosSpec(namespace, appName, nil, action)

		workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
			Name: strings.ToLower(fmt.Sprintf("%s-%s-%s", namespace, appName, action)),
			Type: v1alpha1.TypePodChaos,
			EmbedChaos: &v1alpha1.EmbedChaos{
				PodChaos: spec,
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

	pp.Print("%+v", workflowChaos)
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
