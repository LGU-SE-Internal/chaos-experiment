package handler

import (
	"fmt"
	"strconv"

	controllers "github.com/CUHK-SE-Group/chaos-experiment/controllers"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/resourcelookup"
	"k8s.io/utils/pointer"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

type CPUStressChaosSpec struct {
	Duration     int `range:"1-60" description:"Time Unit Minute"`
	Namespace    int `range:"0-0" dynamic:"true" description:"String"`
	ContainerIdx int `range:"0-0" dynamic:"true" description:"Container Index"`
	CPULoad      int `range:"1-100" description:"CPU Load Percentage"`
	CPUWorker    int `range:"1-3" description:"CPU Stress Threads"`
}

func (s *CPUStressChaosSpec) Create(cli cli.Client, labels map[string]string, opts ...Option) (string, error) {
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

	stressors := controllers.MakeCPUStressors(
		s.CPULoad,
		s.CPUWorker,
	)
	return controllers.CreateStressChaosWithContainer(cli, ns, appName, stressors, "cpu-exhaustion", duration, labels, []string{containerName})
}

type MemoryStressChaosSpec struct {
	Duration     int `range:"1-60" description:"Time Unit Minute"`
	Namespace    int `range:"0-0" dynamic:"true" description:"String"`
	ContainerIdx int `range:"0-0" dynamic:"true" description:"Container Index"`
	MemorySize   int `range:"1-1024" description:"Memory Size Unit MB"`
	MemWorker    int `range:"1-4" description:"Memory Stress Threads"`
}

func (s *MemoryStressChaosSpec) Create(cli cli.Client, labels map[string]string, opts ...Option) (string, error) {
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

	stressors := controllers.MakeMemoryStressors(
		strconv.Itoa(s.MemorySize)+"MiB",
		s.MemWorker,
	)
	return controllers.CreateStressChaosWithContainer(cli, ns, appName, stressors, "memory-exhaustion", duration, labels, []string{containerName})
}
