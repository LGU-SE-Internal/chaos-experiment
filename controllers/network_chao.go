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

// CreateNetworkChaos creates a NetworkChaos resource
func CreateNetworkChaos(cli client.Client, ctx context.Context, namespace string, appName string, action v1alpha1.NetworkChaosAction, duration *string, annotations map[string]string, labels map[string]string, opts ...chaos.OptNetworkChaos) (string, error) {
	spec := chaos.GenerateNetworkChaosSpec(namespace, appName, duration, action, opts...)
	name := strings.ToLower(fmt.Sprintf("%s-%s-%s-%s", namespace, appName, string(action), rand.String(6)))
	networkChaos, err := chaos.NewNetworkChaos(
		chaos.WithAnnotations(annotations),
		chaos.WithLabels(labels),
		chaos.WithName(name),
		chaos.WithNamespace(namespace),
		chaos.WithNetworkChaosSpec(spec),
	)
	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
		return "", err
	}
	create, err := networkChaos.ValidateCreate()
	if err != nil {
		logrus.Errorf("Failed to validate create chaos: %v", err)
		return "", err
	}
	logrus.Infof("create warning: %v", create)
	err = cli.Create(ctx, networkChaos)
	if err != nil {
		logrus.Errorf("Failed to create chaos: %v", err)
		return "", err
	}
	return name, nil
}

// Helper functions for common network chaos types with additional options support
func CreateNetworkDelayChaos(cli client.Client, ctx context.Context, namespace string, appName string, latency string, correlation string, jitter string, duration *string, annotations map[string]string, labels map[string]string, additionalOpts ...chaos.OptNetworkChaos) (string, error) {
	opts := []chaos.OptNetworkChaos{
		chaos.WithNetworkDelay(latency, correlation, jitter),
	}

	// Add any additional options provided
	opts = append(opts, additionalOpts...)

	return CreateNetworkChaos(
		cli,
		ctx,
		namespace,
		appName,
		v1alpha1.DelayAction,
		duration,
		annotations,
		labels,
		opts...,
	)
}

func CreateNetworkLossChaos(cli client.Client, ctx context.Context, namespace string, appName string, loss string, correlation string, duration *string, annotations map[string]string, labels map[string]string, additionalOpts ...chaos.OptNetworkChaos) (string, error) {
	opts := []chaos.OptNetworkChaos{
		chaos.WithNetworkLoss(loss, correlation),
	}

	// Add any additional options provided
	opts = append(opts, additionalOpts...)

	return CreateNetworkChaos(
		cli,
		ctx,
		namespace,
		appName,
		v1alpha1.LossAction,
		duration,
		annotations,
		labels,
		opts...,
	)
}

func CreateNetworkDuplicateChaos(cli client.Client, ctx context.Context, namespace string, appName string, duplicate string, correlation string, duration *string, annotations map[string]string, labels map[string]string, additionalOpts ...chaos.OptNetworkChaos) (string, error) {
	opts := []chaos.OptNetworkChaos{
		chaos.WithNetworkDuplicate(duplicate, correlation),
	}

	// Add any additional options provided
	opts = append(opts, additionalOpts...)

	return CreateNetworkChaos(
		cli,
		ctx,
		namespace,
		appName,
		v1alpha1.DuplicateAction,
		duration,
		annotations,
		labels,
		opts...,
	)
}

func CreateNetworkCorruptChaos(cli client.Client, ctx context.Context, namespace string, appName string, corrupt string, correlation string, duration *string, annotations map[string]string, labels map[string]string, additionalOpts ...chaos.OptNetworkChaos) (string, error) {
	opts := []chaos.OptNetworkChaos{
		chaos.WithNetworkCorrupt(corrupt, correlation),
	}

	// Add any additional options provided
	opts = append(opts, additionalOpts...)

	return CreateNetworkChaos(
		cli,
		ctx,
		namespace,
		appName,
		v1alpha1.CorruptAction,
		duration,
		annotations,
		labels,
		opts...,
	)
}

func CreateNetworkBandwidthChaos(cli client.Client, ctx context.Context, namespace string, appName string, rate string, limit uint32, buffer uint32, duration *string, annotations map[string]string, labels map[string]string, additionalOpts ...chaos.OptNetworkChaos) (string, error) {
	opts := []chaos.OptNetworkChaos{
		chaos.WithNetworkBandwidth(rate, limit, buffer),
	}

	// Add any additional options provided
	opts = append(opts, additionalOpts...)

	return CreateNetworkChaos(
		cli,
		ctx,
		namespace,
		appName,
		v1alpha1.BandwidthAction,
		duration,
		annotations,
		labels,
		opts...,
	)
}

// Updated signature to match other helper functions (without explicit target and direction)
func CreateNetworkPartitionChaos(cli client.Client, ctx context.Context, namespace string, appName string, duration *string, annotations map[string]string, labels map[string]string, additionalOpts ...chaos.OptNetworkChaos) (string, error) {
	return CreateNetworkChaos(
		cli,
		ctx,
		namespace,
		appName,
		v1alpha1.PartitionAction,
		duration,
		annotations,
		labels,
		additionalOpts...,
	)
}

// AddNetworkChaosWorkflowNodes adds network chaos nodes to a workflow
func AddNetworkChaosWorkflowNodes(workflowSpec *v1alpha1.WorkflowSpec, namespace string, appList []string, action v1alpha1.NetworkChaosAction, injectTime *string, sleepTime *string, opts ...chaos.OptNetworkChaos) *v1alpha1.WorkflowSpec {
	for _, appName := range appList {
		spec := chaos.GenerateNetworkChaosSpec(namespace, appName, nil, action, opts...)

		workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
			Name: strings.ToLower(fmt.Sprintf("%s-%s-%s-%s", namespace, appName, string(action), rand.String(6))),
			Type: v1alpha1.TypeNetworkChaos,
			EmbedChaos: &v1alpha1.EmbedChaos{
				NetworkChaos: spec,
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

// ScheduleNetworkChaos schedules a sequence of network chaos events
func ScheduleNetworkChaos(cli client.Client, namespace string, appList []string, action v1alpha1.NetworkChaosAction, opts ...chaos.OptNetworkChaos) {
	workflowName := strings.ToLower(fmt.Sprintf("%s-%s-%s", namespace, string(action), rand.String(6)))
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
		spec := chaos.GenerateNetworkChaosSpec(namespace, appName, nil, action, opts...)

		workflowSpec.Templates = append(workflowSpec.Templates, v1alpha1.Template{
			Name: strings.ToLower(fmt.Sprintf("%s-%s-%s", namespace, appName, string(action))),
			Type: v1alpha1.TypeNetworkChaos,
			EmbedChaos: &v1alpha1.EmbedChaos{
				NetworkChaos: spec,
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
