package handler

import (
	"fmt"
	"strconv"

	controllers "github.com/CUHK-SE-Group/chaos-experiment/controllers"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/resourcelookup"
	"k8s.io/utils/pointer"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

type TimeSkewSpec struct {
	Duration   int `range:"15-15" description:"Time Unit Minute"`
	Namespace  int `range:"0-0" dynamic:"true" description:"String"`
	AppIdx     int `range:"0-0" dynamic:"true" description:"App Index"`
	TimeOffset int `range:"-600-600" description:"Time offset in seconds"`
}

func (s *TimeSkewSpec) Create(cli cli.Client, opts ...Option) (string, error) {
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
	// Format the TimeOffset with "s" unit
	timeOffset := fmt.Sprintf("%ds", s.TimeOffset)

	return controllers.CreateTimeChaos(cli, ns, appName, timeOffset, duration)
}
