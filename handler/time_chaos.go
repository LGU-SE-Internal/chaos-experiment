package handler

import (
	"context"
	"fmt"
	"strconv"

	controllers "github.com/LGU-SE-Internal/chaos-experiment/controllers"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/resourcelookup"
	"k8s.io/utils/pointer"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

type TimeSkewSpec struct {
	Duration     int `range:"1-60" description:"Time Unit Minute"`
	Namespace    int `range:"0-0" dynamic:"true"`
	ContainerIdx int `range:"0-0" dynamic:"true" description:"Container Index"`
	TimeOffset   int `range:"-600-600" description:"Time offset in seconds"`
}

func (s *TimeSkewSpec) Create(cli cli.Client, opts ...Option) (string, error) {
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

	ns := GetTargetNamespace(s.Namespace, DefaultStartIndex)
	if conf.Namespace != "" {
		ns = conf.Namespace
	}

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
	// Format the TimeOffset with "s" unit
	timeOffset := fmt.Sprintf("%ds", s.TimeOffset)

	return controllers.CreateTimeChaosWithContainer(cli, ctx, ns, appName, timeOffset, duration, annotations, labels, []string{containerName})
}
