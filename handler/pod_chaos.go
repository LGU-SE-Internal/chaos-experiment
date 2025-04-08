package handler

import (
	"fmt"
	"strconv"

	controllers "github.com/CUHK-SE-Group/chaos-experiment/controllers"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/resourcelookup"
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
	ns := TargetNamespace
	if conf.Namespace != "" {
		ns = conf.Namespace
	}

	appLabels, err := resourcelookup.GetAllAppLabels()
	if err != nil {
		return "", fmt.Errorf("failed to get app labels: %w", err)
	}

	if s.AppIdx < 0 || s.AppIdx >= len(appLabels) {
		return "", fmt.Errorf("app index out of range: %d (max: %d)", s.AppIdx, len(appLabels)-1)
	}

	appName := appLabels[s.AppIdx]
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.PodFailureAction

	return controllers.CreatePodChaos(cli, ns, appName, action, duration)
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
	ns := TargetNamespace
	if conf.Namespace != "" {
		ns = conf.Namespace
	}

	appLabels, err := resourcelookup.GetAllAppLabels()
	if err != nil {
		return "", fmt.Errorf("failed to get app labels: %w", err)
	}

	if s.AppIdx < 0 || s.AppIdx >= len(appLabels) {
		return "", fmt.Errorf("app index out of range: %d (max: %d)", s.AppIdx, len(appLabels)-1)
	}

	appName := appLabels[s.AppIdx]
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.PodKillAction

	return controllers.CreatePodChaos(cli, ns, appName, action, duration)
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
	ns := TargetNamespace
	if conf.Namespace != "" {
		ns = conf.Namespace
	}

	containers, err := resourcelookup.GetAllContainers()
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
	return controllers.CreatePodChaosWithContainer(cli, ns, appName, action, duration, []string{containerName})
}
