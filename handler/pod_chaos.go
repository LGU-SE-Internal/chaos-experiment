package handler

import (
	"context"
	"fmt"
	"strconv"

	controllers "github.com/LGU-SE-Internal/chaos-experiment/controllers"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/resourcelookup"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"k8s.io/utils/pointer"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

type PodFailureSpec struct {
	Duration  int `range:"1-60" description:"Time Unit Minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"String"`
	AppIdx    int `range:"0-0" dynamic:"true" description:"App Index"`
}

func (s *PodFailureSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annoations != nil {
		annotations = conf.Annoations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace

	appLabels, err := resourcelookup.GetAllAppLabels(ns, TargetLabelKey)
	if err != nil {
		return "", fmt.Errorf("failed to get app labels: %w", err)
	}

	if s.AppIdx < 0 || s.AppIdx >= len(appLabels) {
		return "", fmt.Errorf("app index out of range: %d (max: %d)", s.AppIdx, len(appLabels)-1)
	}

	appName := appLabels[s.AppIdx]
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.PodFailureAction

	return controllers.CreatePodChaos(cli, ctx, ns, appName, action, duration, annotations, labels)
}

// Update PodKillSpec to use flattened app index
type PodKillSpec struct {
	Duration  int `range:"1-60" description:"Time Unit Minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"String"`
	AppIdx    int `range:"0-0" dynamic:"true" description:"App Index"`
}

func (s *PodKillSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annoations != nil {
		annotations = conf.Annoations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace

	appLabels, err := resourcelookup.GetAllAppLabels(ns, TargetLabelKey)
	if err != nil {
		return "", fmt.Errorf("failed to get app labels: %w", err)
	}

	if s.AppIdx < 0 || s.AppIdx >= len(appLabels) {
		return "", fmt.Errorf("app index out of range: %d (max: %d)", s.AppIdx, len(appLabels)-1)
	}

	appName := appLabels[s.AppIdx]
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.PodKillAction

	return controllers.CreatePodChaos(cli, ctx, ns, appName, action, duration, annotations, labels)
}

type ContainerKillSpec struct {
	Duration     int `range:"1-60" description:"Time Unit Minute"`
	Namespace    int `range:"0-0" dynamic:"true" description:"String"`
	ContainerIdx int `range:"0-0" dynamic:"true" description:"Container Index"`
}

func (s *ContainerKillSpec) Create(cli cli.Client, opts ...Option) (string, error) {
	conf := Conf{}
	for _, opt := range opts {
		opt(&conf)
	}

	annotations := make(map[string]string)
	if conf.Annoations != nil {
		annotations = conf.Annoations
	}

	ctx := context.Background()
	if conf.Context != nil {
		ctx = conf.Context
	}

	labels := make(map[string]string)
	if conf.Labels != nil {
		labels = conf.Labels
	}

	ns := conf.Namespace

	containers, err := resourcelookup.GetAllContainers(ns)
	if err != nil {
		return "", fmt.Errorf("failed to get containers: %w", err)
	}

	if s.ContainerIdx < 0 || s.ContainerIdx >= len(containers) {
		return "", fmt.Errorf("container index out of range: %d (max: %d)", s.ContainerIdx, len(containers)-1)
	}

	containerInfo := containers[s.ContainerIdx]
	appName := containerInfo.AppLabel
	containerName := containerInfo.ContainerName

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.ContainerKillAction

	// Use the updated CreatePodChaosWithContainer function
	return controllers.CreatePodChaosWithContainer(cli, ctx, ns, appName, action, duration, annotations, labels, []string{containerName})
}
