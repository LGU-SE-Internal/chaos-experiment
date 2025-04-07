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
	Duration  int `range:"15-15" description:"Time Unit Minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"String"`
	AppIdx    int `range:"0-0" dynamic:"true" description:"App Index"`
	CPULoad   int `range:"1-100" description:"CPU Load Percentage"`
	CPUWorker int `range:"1-3" description:"CPU Stress Threads"`
}

func (s *CPUStressChaosSpec) Create(cli cli.Client, opts ...Option) (string, error) {
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

	stressors := controllers.MakeCPUStressors(
		s.CPULoad,
		s.CPUWorker,
	)
	return controllers.CreateStressChaos(cli, ns, appName, stressors, "cpu-exhaustion", duration)
}

// Update MemoryStressChaosSpec to use flattened app index
type MemoryStressChaosSpec struct {
	Duration   int `range:"15-15" description:"Time Unit Minute"`
	Namespace  int `range:"0-0" dynamic:"true" description:"String"`
	AppIdx     int `range:"0-0" dynamic:"true" description:"App Index"`
	MemorySize int `range:"1-1024" description:"Memory Size Unit MB"`
	MemWorker  int `range:"1-4" description:"Memory Stress Threads"`
}

func (s *MemoryStressChaosSpec) Create(cli cli.Client, opts ...Option) (string, error) {
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

	stressors := controllers.MakeMemoryStressors(
		strconv.Itoa(s.MemorySize)+"MiB",
		s.MemWorker,
	)
	return controllers.CreateStressChaos(cli, ns, appName, stressors, "memory-exhaustion", duration)
}
