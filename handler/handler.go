package handler

import (
	"fmt"
	"strconv"

	chaos "github.com/CUHK-SE-Group/chaos-experiment/chaos"
	"github.com/CUHK-SE-Group/chaos-experiment/client"
	controllers "github.com/CUHK-SE-Group/chaos-experiment/controllers"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"k8s.io/utils/pointer"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

type ChaosType int

const TargetNamespace = "ts" // todo: make it dynamic (e.g. from config)
const TargetLabelKey = "app"

const (
	// Default indicates an unknown Type.
	Default ChaosType = iota

	// PodChaos
	PodKill
	PodFailure
	ContainerKill

	// StressChaos
	MemoryStress
	CPUStress

	// HTTPChaos
	HTTPAbort
	HTTPDelay
	HTTPReplace
	HTTPPatch

	// DNSChaos
	DNSError
	DNSRandom

	// TimeChaos
	TimeSkew

	// NetworkChaos
	NetworkDelay
	NetworkLoss
	NetworkDuplicate
	NetworkCorrupt
	NetworkBandwidth
	NetworkPartition

	// ...
)

// 定义 ChaosType 对应的 map
var ChaosTypeMap = map[ChaosType]string{
	PodKill:          "PodKill",
	PodFailure:       "PodFailure",
	ContainerKill:    "ContainerKill",
	MemoryStress:     "MemoryStress",
	CPUStress:        "CPUStress",
	HTTPAbort:        "HTTPAbort",
	HTTPDelay:        "HTTPDelay",
	HTTPReplace:      "HTTPReplace",
	HTTPPatch:        "HTTPPatch",
	DNSError:         "DNSError",
	DNSRandom:        "DNSRandom",
	TimeSkew:         "TimeSkew",
	NetworkDelay:     "NetworkDelay",
	NetworkLoss:      "NetworkLoss",
	NetworkDuplicate: "NetworkDuplicate",
	NetworkCorrupt:   "NetworkCorrupt",
	NetworkBandwidth: "NetworkBandwidth",
	NetworkPartition: "NetworkPartition",
}

// GetChaosTypeName 根据 ChaosType 获取名称
func GetChaosTypeName(c ChaosType) string {
	if name, ok := ChaosTypeMap[c]; ok {
		return name
	}
	return "Unknown"
}

type HTTPChaosTarget int

const (
	Request  HTTPChaosTarget = 1
	Response HTTPChaosTarget = 2
)

type Injection interface {
	Create(cli cli.Client) string
}

var httpChaosTargetMap = map[HTTPChaosTarget]chaosmeshv1alpha1.PodHttpChaosTarget{
	Request:  chaosmeshv1alpha1.PodHttpRequest,
	Response: chaosmeshv1alpha1.PodHttpResponse,
}

type HTTPReplaceBody int

const (
	Blank  HTTPReplaceBody = 1
	Random HTTPReplaceBody = 2
)

var httpReplaceBodyMap = map[HTTPReplaceBody]chaos.OptHTTPChaos{
	Blank:  chaos.WithReplaceBody([]byte("")),
	Random: chaos.WithRandomReplaceBody(),
}

type ContainerKillSpec struct {
	Duration  int `range:"1-60" description:"Time Unit Minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"String"`
	AppName   int `range:"0-0" dynamic:"true" description:"Array"`
}

func (s *ContainerKillSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.ContainerKillAction
	return controllers.CreatePodChaos(cli, TargetNamespace, labelArr[s.AppName], action, duration)
}

type PodFailureSpec struct {
	Duration  int `range:"1-60" description:"Time Unit Minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"String"`
	AppName   int `range:"0-0" dynamic:"true" description:"Array"`
}

func (s *PodFailureSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.PodFailureAction
	return controllers.CreatePodChaos(cli, TargetNamespace, labelArr[s.AppName], action, duration)
}

type PodKillSpec struct {
	Duration  int `range:"1-60" description:"Time Unit Minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"String"`
	AppName   int `range:"0-0" dynamic:"true" description:"Array"`
}

func (s *PodKillSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.PodKillAction
	return controllers.CreatePodChaos(cli, TargetNamespace, labelArr[s.AppName], action, duration)
}

type CPUStressChaosSpec struct {
	CPULoad   int `range:"1-100" description:"CPU Load Percentage"`
	CPUWorker int `range:"1-3" description:"CPU Stress Threads"`
	Duration  int `range:"1-60" description:"Time Unit Minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"String"`
	AppName   int `range:"0-0" dynamic:"true" description:"Array"`
}

func (s *CPUStressChaosSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	stressors := controllers.MakeCPUStressors(
		s.CPULoad,
		s.CPUWorker,
	)
	return controllers.CreateStressChaos(cli, TargetNamespace, labelArr[s.AppName], stressors, "cpu-exhaustion", duration)
}

type MemoryStressChaosSpec struct {
	MemorySize int `range:"1-1024" description:"Memory Size Unit MB"`
	MemWorker  int `range:"1-4" description:"Memory Stress Threads"`
	Duration   int `range:"1-60" description:"Time Unit Minute"`
	Namespace  int `range:"0-0" dynamic:"true" description:"String"`
	AppName    int `range:"0-0" dynamic:"true" description:"Array"`
}

func (s *MemoryStressChaosSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	stressors := controllers.MakeMemoryStressors(
		strconv.Itoa(s.MemorySize)+"MiB",
		s.MemWorker,
	)
	return controllers.CreateStressChaos(cli, TargetNamespace, labelArr[s.AppName], stressors, "memory-exhaustion", pointer.String(strconv.Itoa(s.Duration)+"m"))
}

type HTTPChaosReplaceSpec struct {
	HTTPTarget  HTTPChaosTarget `range:"1-2" description:"HTTP Phase Request/Response"`
	ReplaceBody HTTPReplaceBody `range:"1-2" description:"Body Replacement Blank/Random"`
	Duration    int             `range:"1-60" description:"Time Unit Minute"`
	Namespace   int             `range:"0-0" dynamic:"true" description:"String"`
	AppName     int             `range:"0-0" dynamic:"true" description:"Array"`
}

func (s *HTTPChaosReplaceSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	target := httpChaosTargetMap[s.HTTPTarget]
	opts := []chaos.OptHTTPChaos{
		chaos.WithTarget(target),
		chaos.WithPort(8080),
		httpReplaceBodyMap[s.ReplaceBody],
	}
	return controllers.CreateHTTPChaos(cli, TargetNamespace, labelArr[s.AppName], fmt.Sprintf("%s-replace", target), duration, opts...)
}

type HTTPChaosAbortSpec struct {
	HTTPTarget HTTPChaosTarget `range:"1-2" description:"HTTP Phase Request/Response"`
	Duration   int             `range:"1-60" description:"Time Unit Minute"`
	Namespace  int             `range:"0-0" dynamic:"true" description:"String"`
	AppName    int             `range:"0-0" dynamic:"true" description:"Array"`
}

func (s *HTTPChaosAbortSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	abort := true
	target := httpChaosTargetMap[s.HTTPTarget]
	opts := []chaos.OptHTTPChaos{
		chaos.WithTarget(target),
		chaos.WithPort(8080),
		chaos.WithAbort(&abort),
	}
	return controllers.CreateHTTPChaos(cli, TargetNamespace, labelArr[s.AppName], fmt.Sprintf("%s-abort", target), duration, opts...)
}

type TimeSkewSpec struct {
	TimeOffset int `range:"-600-600" description:"Time offset in seconds"`
	Duration   int `range:"1-60" description:"Time Unit Minute"`
	Namespace  int `range:"0-0" dynamic:"true" description:"String"`
	AppName    int `range:"0-0" dynamic:"true" description:"Array"`
}

func (s *TimeSkewSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	// Format the TimeOffset with "s" unit
	timeOffset := fmt.Sprintf("%ds", s.TimeOffset)

	return controllers.CreateTimeChaos(cli, TargetNamespace, labelArr[s.AppName], timeOffset, duration)
}

// Map for Direction conversion from int to chaos-mesh Direction type
var directionMap = map[int]chaosmeshv1alpha1.Direction{
	1: chaosmeshv1alpha1.To,
	2: chaosmeshv1alpha1.From,
	3: chaosmeshv1alpha1.Both,
}

// Convert int direction code to chaos-mesh Direction
func getDirection(directionCode int) chaosmeshv1alpha1.Direction {
	if direction, ok := directionMap[directionCode]; ok {
		return direction
	}
	return chaosmeshv1alpha1.To // Default to "To" direction
}

// Common function to create network chaos with direction and optional target
func createNetworkChaosWithTargetDirection(cli cli.Client, action chaosmeshv1alpha1.NetworkChaosAction,
	labelArr []string, appNameIdx int, targetAppIdx int, directionCode int,
	duration *string, networkOpts ...chaos.OptNetworkChaos) string {

	direction := getDirection(directionCode)
	opts := []chaos.OptNetworkChaos{}

	// Add target and direction if specified and valid
	if targetAppIdx >= 0 && targetAppIdx < len(labelArr) {
		targetApp := labelArr[targetAppIdx]
		opts = append(opts, chaos.WithNetworkTargetAndDirection(TargetNamespace, targetApp, direction))
	} else {
		// Only set direction without target
		opts = append(opts, chaos.WithNetworkDirection(direction))
	}

	// Add specific network options
	opts = append(opts, networkOpts...)

	// Create network chaos
	return controllers.CreateNetworkChaos(cli, TargetNamespace, labelArr[appNameIdx],
		action, duration, opts...)
}

type NetworkDelaySpec struct {
	Latency     int `range:"1-2000" description:"Latency in milliseconds"`
	Correlation int `range:"0-100" description:"Correlation percentage"`
	Jitter      int `range:"0-1000" description:"Jitter in milliseconds"`
	Duration    int `range:"1-60" description:"Time Unit Minute"`
	Namespace   int `range:"0-0" dynamic:"true" description:"String"`
	AppName     int `range:"0-0" dynamic:"true" description:"Array"`
	Direction   int `range:"1-3" description:"Direction (1=to, 2=from, 3=both)"`
	TargetApp   int `range:"0-0" dynamic:"true" description:"Target application (if any)"`
}

func (s *NetworkDelaySpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}

	// Convert int values to appropriate string format
	latency := fmt.Sprintf("%dms", s.Latency)
	correlation := fmt.Sprintf("%d", s.Correlation)
	jitter := fmt.Sprintf("%dms", s.Jitter)
	duration := pointer.String(fmt.Sprintf("%dm", s.Duration))

	// Use the common function with specific delay options
	return createNetworkChaosWithTargetDirection(cli, chaosmeshv1alpha1.DelayAction,
		labelArr, s.AppName, s.TargetApp, s.Direction, duration,
		chaos.WithNetworkDelay(latency, correlation, jitter))
}

type NetworkLossSpec struct {
	Loss        int `range:"1-100" description:"Packet loss percentage"`
	Correlation int `range:"0-100" description:"Correlation percentage"`
	Duration    int `range:"1-60" description:"Time Unit Minute"`
	Namespace   int `range:"0-0" dynamic:"true" description:"String"`
	AppName     int `range:"0-0" dynamic:"true" description:"Array"`
	Direction   int `range:"1-3" description:"Direction (1=to, 2=from, 3=both)"`
	TargetApp   int `range:"0-0" dynamic:"true" description:"Target application (if any)"`
}

func (s *NetworkLossSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}

	// Convert int values to appropriate string format
	loss := fmt.Sprintf("%d", s.Loss)
	correlation := fmt.Sprintf("%d", s.Correlation)
	duration := pointer.String(fmt.Sprintf("%dm", s.Duration))

	// Use the common function with specific loss options
	return createNetworkChaosWithTargetDirection(cli, chaosmeshv1alpha1.LossAction,
		labelArr, s.AppName, s.TargetApp, s.Direction, duration,
		chaos.WithNetworkLoss(loss, correlation))
}

type NetworkDuplicateSpec struct {
	Duplicate   int `range:"1-100" description:"Packet duplication percentage"`
	Correlation int `range:"0-100" description:"Correlation percentage"`
	Duration    int `range:"1-60" description:"Time Unit Minute"`
	Namespace   int `range:"0-0" dynamic:"true" description:"String"`
	AppName     int `range:"0-0" dynamic:"true" description:"Array"`
	Direction   int `range:"1-3" description:"Direction (1=to, 2=from, 3=both)"`
	TargetApp   int `range:"0-0" dynamic:"true" description:"Target application (if any)"`
}

func (s *NetworkDuplicateSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}

	// Convert int values to appropriate string format
	duplicate := fmt.Sprintf("%d", s.Duplicate)
	correlation := fmt.Sprintf("%d", s.Correlation)
	duration := pointer.String(fmt.Sprintf("%dm", s.Duration))

	// Use the common function with specific duplicate options
	return createNetworkChaosWithTargetDirection(cli, chaosmeshv1alpha1.DuplicateAction,
		labelArr, s.AppName, s.TargetApp, s.Direction, duration,
		chaos.WithNetworkDuplicate(duplicate, correlation))
}

type NetworkCorruptSpec struct {
	Corrupt     int `range:"1-100" description:"Packet corruption percentage"`
	Correlation int `range:"0-100" description:"Correlation percentage"`
	Duration    int `range:"1-60" description:"Time Unit Minute"`
	Namespace   int `range:"0-0" dynamic:"true" description:"String"`
	AppName     int `range:"0-0" dynamic:"true" description:"Array"`
	Direction   int `range:"1-3" description:"Direction (1=to, 2=from, 3=both)"`
	TargetApp   int `range:"0-0" dynamic:"true" description:"Target application (if any)"`
}

func (s *NetworkCorruptSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}

	// Convert int values to appropriate string format
	corrupt := fmt.Sprintf("%d", s.Corrupt)
	correlation := fmt.Sprintf("%d", s.Correlation)
	duration := pointer.String(fmt.Sprintf("%dm", s.Duration))

	// Use the common function with specific corrupt options
	return createNetworkChaosWithTargetDirection(cli, chaosmeshv1alpha1.CorruptAction,
		labelArr, s.AppName, s.TargetApp, s.Direction, duration,
		chaos.WithNetworkCorrupt(corrupt, correlation))
}

type NetworkBandwidthSpec struct {
	Rate      int `range:"1-1000000" description:"Bandwidth rate in kbps"`
	Limit     int `range:"1-10000" description:"Number of bytes that can be queued"`
	Buffer    int `range:"1-10000" description:"Maximum amount of bytes available instantaneously"`
	Duration  int `range:"1-60" description:"Time Unit Minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"String"`
	AppName   int `range:"0-0" dynamic:"true" description:"Array"`
	Direction int `range:"1-3" description:"Direction (1=to, 2=from, 3=both)"`
	TargetApp int `range:"0-0" dynamic:"true" description:"Target application (if any)"`
}

func (s *NetworkBandwidthSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""

	}

	// Convert rate from kbps to string with unit
	rate := fmt.Sprintf("%dkbps", s.Rate)
	limit := uint32(s.Limit)
	buffer := uint32(s.Buffer)
	duration := pointer.String(fmt.Sprintf("%dm", s.Duration))

	// Use the common function with specific bandwidth options
	return createNetworkChaosWithTargetDirection(cli, chaosmeshv1alpha1.BandwidthAction,
		labelArr, s.AppName, s.TargetApp, s.Direction, duration,
		chaos.WithNetworkBandwidth(rate, limit, buffer))
}

type NetworkPartitionSpec struct {
	Duration  int `range:"1-60" description:"Time Unit Minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"String"`
	AppName   int `range:"0-0" dynamic:"true" description:"Array"`
	TargetApp int `range:"0-0" dynamic:"true" description:"Target application"`
	Direction int `range:"1-3" description:"Direction (1=to, 2=from, 3=both)"`
}

func (s *NetworkPartitionSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}

	if s.TargetApp < 0 || s.TargetApp >= len(labelArr) {
		return ""
	}

	duration := pointer.String(fmt.Sprintf("%dm", s.Duration))

	// Use the common function for partition - this handles target and direction consistently
	return createNetworkChaosWithTargetDirection(cli, chaosmeshv1alpha1.PartitionAction,
		labelArr, s.AppName, s.TargetApp, s.Direction, duration)
}

// DNSChaosSpec defines the DNS chaos injection parameters
type DNSChaosSpec struct {
	Duration  int `range:"1-60" description:"Time Unit Minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"String"`
	AppName   int `range:"0-0" dynamic:"true" description:"Array"`
	Action    int `range:"1-2" description:"Action (1=error, 2=random)"`
}

// Convert int action code to chaos-mesh DNSChaosAction
func getDNSAction(actionCode int) chaosmeshv1alpha1.DNSChaosAction {
	switch actionCode {
	case 1:
		return chaosmeshv1alpha1.ErrorAction
	case 2:
		return chaosmeshv1alpha1.RandomAction
	default:
		return chaosmeshv1alpha1.ErrorAction // Default to error action
	}
}

func (s *DNSChaosSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := getDNSAction(s.Action)

	// Use a simple pattern that matches all domains
	patterns := []string{"*"}

	return controllers.CreateDnsChaos(cli, TargetNamespace, labelArr[s.AppName], action, patterns, duration)
}

var SpecMap = map[ChaosType]any{
	CPUStress:        CPUStressChaosSpec{},
	MemoryStress:     MemoryStressChaosSpec{},
	HTTPAbort:        HTTPChaosAbortSpec{},
	HTTPReplace:      HTTPChaosReplaceSpec{},
	DNSError:         DNSChaosSpec{},
	DNSRandom:        DNSChaosSpec{},
	TimeSkew:         TimeSkewSpec{},
	NetworkDelay:     NetworkDelaySpec{},
	NetworkLoss:      NetworkLossSpec{},
	NetworkDuplicate: NetworkDuplicateSpec{},
	NetworkCorrupt:   NetworkCorruptSpec{},
	NetworkBandwidth: NetworkBandwidthSpec{},
	NetworkPartition: NetworkPartitionSpec{},
}

var ChaosHandlers = map[ChaosType]Injection{
	PodKill:          &PodKillSpec{},
	PodFailure:       &PodFailureSpec{},
	ContainerKill:    &ContainerKillSpec{},
	MemoryStress:     &MemoryStressChaosSpec{},
	CPUStress:        &CPUStressChaosSpec{},
	HTTPAbort:        &HTTPChaosAbortSpec{},
	HTTPReplace:      &HTTPChaosReplaceSpec{},
	DNSError:         &DNSChaosSpec{Action: 1}, // Default to error action
	DNSRandom:        &DNSChaosSpec{Action: 2}, // Default to random action
	TimeSkew:         &TimeSkewSpec{},
	NetworkDelay:     &NetworkDelaySpec{},
	NetworkLoss:      &NetworkLossSpec{},
	NetworkDuplicate: &NetworkDuplicateSpec{},
	NetworkCorrupt:   &NetworkCorruptSpec{},
	NetworkBandwidth: &NetworkBandwidthSpec{},
	NetworkPartition: &NetworkPartitionSpec{},
}

type InjectionConf struct {
	PodKill          *PodKillSpec           `range:"0-2"`
	PodFailure       *PodFailureSpec        `range:"0-2"`
	ContainerKill    *ContainerKillSpec     `range:"0-2"`
	MemoryStress     *MemoryStressChaosSpec `range:"0-4"`
	CPUStress        *CPUStressChaosSpec    `range:"0-4"`
	HTTPAbort        *HTTPChaosAbortSpec    `range:"0-3"`
	HTTPReplace      *HTTPChaosReplaceSpec  `range:"0-4"`
	DNSError         *DNSChaosSpec          `range:"0-3"`
	DNSRandom        *DNSChaosSpec          `range:"0-3"`
	TimeSkew         *TimeSkewSpec          `range:"0-3"`
	NetworkDelay     *NetworkDelaySpec      `range:"0-7"`
	NetworkLoss      *NetworkLossSpec       `range:"0-6"`
	NetworkDuplicate *NetworkDuplicateSpec  `range:"0-6"`
	NetworkCorrupt   *NetworkCorruptSpec    `range:"0-6"`
	NetworkBandwidth *NetworkBandwidthSpec  `range:"0-7"`
	NetworkPartition *NetworkPartitionSpec  `range:"0-4"`
}
