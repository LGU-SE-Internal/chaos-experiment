package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/CUHK-SE-Group/chaos-experiment/chaos"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/k0kubun/pp/v3"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateIOChaos creates an IO chaos experiment
func CreateIOChaos(cli client.Client, namespace string, appName string, volumePath string, chaosType string, duration *string, opts ...chaos.OptIOChaos) string {
	spec := chaos.GenerateIOChaosSpec(namespace, appName, duration, volumePath, opts...)
	name := strings.ToLower(fmt.Sprintf("%s-%s-%s-%s", namespace, appName, chaosType, rand.String(6)))
	ioChaos, err := chaos.NewIOChaos(chaos.WithName(name), chaos.WithNamespace(namespace), chaos.WithIOChaosSpec(spec))
	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
		return ""
	}
	create, err := ioChaos.ValidateCreate()
	if err != nil {
		logrus.Errorf("Failed to validate create chaos: %v", err)
		return ""
	}
	logrus.Infof("create warning: %v", create)
	err = cli.Create(context.Background(), ioChaos)
	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
		return ""
	}
	return name
}

// CreateIODelayExperiment creates an IO delay experiment
func CreateIODelayExperiment(cli client.Client, namespace string, appName string, volumePath string, path string, delay string, duration *string) string {
	opts := []chaos.OptIOChaos{
		chaos.WithIODelayAction(delay),
		chaos.WithIOPath(path),
	}
	return CreateIOChaos(cli, namespace, appName, volumePath, "io-delay", duration, opts...)
}

// CreateIOErrorExperiment creates an IO error experiment
func CreateIOErrorExperiment(cli client.Client, namespace string, appName string, volumePath string, path string, errno uint32, duration *string) string {
	opts := []chaos.OptIOChaos{
		chaos.WithIOErrorAction(errno),
		chaos.WithIOPath(path),
	}
	return CreateIOChaos(cli, namespace, appName, volumePath, "io-error", duration, opts...)
}

// CreateIOMistakeExperiment creates an IO mistake experiment
func CreateIOMistakeExperiment(cli client.Client, namespace string, appName string, volumePath string, path string, filling v1alpha1.FillingType, maxOccurrences int64, maxLength int64, duration *string) string {
	opts := []chaos.OptIOChaos{
		chaos.WithIOMistakeAction(filling, maxOccurrences, maxLength),
		chaos.WithIOPath(path),
	}
	return CreateIOChaos(cli, namespace, appName, volumePath, "io-mistake", duration, opts...)
}

// AddIOChaosWorkflowNodes adds IO chaos nodes to a workflow
func AddIOChaosWorkflowNodes(workflowSpec *v1alpha1.WorkflowSpec, namespace string, appList []string, volumePath string, chaosType string, injectTime *string, sleepTime *string, opts ...chaos.OptIOChaos) *v1alpha1.WorkflowSpec {
	for _, appName := range appList {
		spec := chaos.GenerateIOChaosSpec(namespace, appName, nil, volumePath, opts...)

		workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
			Name: strings.ToLower(fmt.Sprintf("%s-%s-%s-%s", namespace, appName, chaosType, rand.String(6))),
			Type: v1alpha1.TypeIOChaos,
			EmbedChaos: &v1alpha1.EmbedChaos{
				IOChaos: spec,
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

// ScheduleIOChaos schedules IO chaos experiments for a list of applications
func ScheduleIOChaos(cli client.Client, namespace string, appList []string, volumePath string, chaosType string, opts ...chaos.OptIOChaos) {
	workflowName := strings.ToLower(fmt.Sprintf("%s-%s-%s", namespace, chaosType, rand.String(6)))
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
		spec := chaos.GenerateIOChaosSpec(namespace, appName, nil, volumePath, opts...)

		workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
			Name: strings.ToLower(fmt.Sprintf("%s-%s-%s", namespace, appName, chaosType)),
			Type: v1alpha1.TypeIOChaos,
			EmbedChaos: &v1alpha1.EmbedChaos{
				IOChaos: spec,
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

	pp.Print("%+v", workflowChaos)
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

// ScheduleIODelayExperiments schedules IO delay experiments
func ScheduleIODelayExperiments(cli client.Client, namespace string, appList []string, volumePath string, path string, delay string) {
	opts := []chaos.OptIOChaos{
		chaos.WithIODelayAction(delay),
		chaos.WithIOPath(path),
	}
	ScheduleIOChaos(cli, namespace, appList, volumePath, "io-delay", opts...)
}

// ScheduleIOErrorExperiments schedules IO error experiments
func ScheduleIOErrorExperiments(cli client.Client, namespace string, appList []string, volumePath string, path string, errno uint32) {
	opts := []chaos.OptIOChaos{
		chaos.WithIOErrorAction(errno),
		chaos.WithIOPath(path),
	}
	ScheduleIOChaos(cli, namespace, appList, volumePath, "io-error", opts...)
}

// ScheduleIOMistakeExperiments schedules IO mistake experiments
func ScheduleIOMistakeExperiments(cli client.Client, namespace string, appList []string, volumePath string, path string, filling v1alpha1.FillingType, maxOccurrences int64, maxLength int64) {
	opts := []chaos.OptIOChaos{
		chaos.WithIOMistakeAction(filling, maxOccurrences, maxLength),
		chaos.WithIOPath(path),
	}
	ScheduleIOChaos(cli, namespace, appList, volumePath, "io-mistake", opts...)
}
