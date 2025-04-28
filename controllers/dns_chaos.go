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

// CreateDnsChaos creates a DNS chaos experiment with the specified parameters
func CreateDnsChaos(cli client.Client, ctx context.Context, namespace string, appName string, action v1alpha1.DNSChaosAction, patterns []string, duration *string, annotations map[string]string, labels map[string]string) (string, error) {
	spec := chaos.GenerateDnsChaosSpec(namespace, appName, duration, action, patterns)
	name := strings.ToLower(fmt.Sprintf("%s-%s-dns-%s", namespace, appName, rand.String(6)))
	dnsChaos, err := chaos.NewDnsChaos(
		chaos.WithAnnotations(annotations),
		chaos.WithLabels(labels),
		chaos.WithName(name),
		chaos.WithNamespace(namespace),
		chaos.WithDnsChaosSpec(spec),
	)
	if err != nil {
		logrus.Errorf("Failed to create DNS chaos: %v", err)
		return "", err
	}
	create, err := dnsChaos.ValidateCreate()
	if err != nil {
		logrus.Errorf("Failed to validate create DNS chaos: %v", err)
		return "", err
	}
	logrus.Infof("Create warning: %v", create)
	err = cli.Create(ctx, dnsChaos)
	if err != nil {
		logrus.Errorf("Failed to create DNS chaos: %v", err)
		return "", err
	}
	return name, nil
}

// AddDnsChaosWorkflowNodes adds DNS chaos nodes to a workflow
func AddDnsChaosWorkflowNodes(workflowSpec *v1alpha1.WorkflowSpec, namespace string, appList []string, action v1alpha1.DNSChaosAction, patterns []string, injectTime *string, sleepTime *string) *v1alpha1.WorkflowSpec {
	for _, appName := range appList {
		spec := chaos.GenerateDnsChaosSpec(namespace, appName, nil, action, patterns)

		workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
			Name: strings.ToLower(fmt.Sprintf("%s-%s-%s-%s", namespace, appName, action, rand.String(6))),
			Type: v1alpha1.TypeDNSChaos,
			EmbedChaos: &v1alpha1.EmbedChaos{
				DNSChaos: spec,
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

// ScheduleDnsChaos schedules a sequence of DNS chaos experiments for multiple apps
func ScheduleDnsChaos(cli client.Client, namespace string, appList []string, action v1alpha1.DNSChaosAction, patterns []string) {
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
		spec := chaos.GenerateDnsChaosSpec(namespace, appName, nil, action, patterns)

		workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
			Name: strings.ToLower(fmt.Sprintf("%s-%s-%s", namespace, appName, action)),
			Type: v1alpha1.TypeDNSChaos,
			EmbedChaos: &v1alpha1.EmbedChaos{
				DNSChaos: spec,
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
