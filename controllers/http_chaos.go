package controllers

import (
	"chaos-expriment/chaos"
	"context"
	"fmt"
	"strings"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/k0kubun/pp/v3"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ScheduleHTTPChaos(cli client.Client, namespace string, appList []string, stressType string, opts ...chaos.OptHTTPChaos) {
	workflowName := strings.ToLower(fmt.Sprintf("%s-%s-%s", namespace, stressType, rand.String(6)))
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

		spec := chaos.GenerateHttpChaosSpec(namespace, appName, opts...)

		workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
			Name: strings.ToLower(fmt.Sprintf("%s-%s-%s", namespace, appName, stressType)),
			Type: v1alpha1.TypeHTTPChaos,
			EmbedChaos: &v1alpha1.EmbedChaos{
				HTTPChaos: spec,
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

	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
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

func ScheduleSetsOfHTTPChaos(cli client.Client, namespace string) {
	ctx := context.Background()

	podList := &corev1.PodList{}

	listOptions := &client.ListOptions{
		Namespace: namespace,
	}

	if err := cli.List(ctx, podList, listOptions); err != nil {
		logrus.Errorf("Failed to list pods: %v", err)
		return
	}

	workflowSpec := v1alpha1.WorkflowSpec{
		Entry: "entry",
		Templates: []v1alpha1.Template{
			{
				Name:     "entry",
				Type:     v1alpha1.TypeSerial,
				Children: nil,
			},
		},
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
		specs := chaos.GenerateSetsOfHttpChaosSpec(namespace, pod.Name)

		for idx, spec := range specs {
			choice := ""
			if spec.PodHttpChaosActions.Abort != nil {
				choice = "abort"
			}
			if spec.PodHttpChaosActions.Delay != nil {
				choice = "delay-" + *spec.PodHttpChaosActions.Delay
			}
			if spec.PodHttpChaosActions.Replace != nil {
				choice = "replace"
			}
			if spec.PodHttpChaosActions.Patch != nil {
				choice = "patch"
			}

			workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
				Name: strings.ToLower(fmt.Sprintf("%s-%s-%s-%s", namespace, pod.Name, spec.Target, choice)),
				Type: v1alpha1.TypeHTTPChaos,
				EmbedChaos: &v1alpha1.EmbedChaos{
					HTTPChaos: &spec,
				},
				Deadline: pointer.String("5m"),
			})
			workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
				Name:     fmt.Sprintf("%s-%s-%s-%d", namespace, pod.Name, "sleep", idx),
				Type:     v1alpha1.TypeSuspend,
				Deadline: pointer.String("5m"),
			})
		}
	}

	for i, template := range workflowSpec.Templates {
		if i == 0 {
			continue
		}
		workflowSpec.Templates[0].Children = append(workflowSpec.Templates[0].Children, template.Name)
	}

	workflowChaos, err := chaos.NewWorkflowChaos(chaos.WithName("entry"), chaos.WithNamespace(namespace), chaos.WithWorkflowSpec(&workflowSpec))
	if err != nil {
		logrus.Errorf("Failed to create chaos workflow: %v", err)
	}

	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
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
