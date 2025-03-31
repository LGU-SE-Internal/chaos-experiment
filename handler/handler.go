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

	// JVMChaos
	JVMLatency
	JVMReturn
	JVMException
	JVMGarbageCollector
	JVMRuleData
	JVMMySQL
	JVMCPUStress
	JVMMemoryStress

	// New MySQL chaos types
	JVMMySQLLatency
	JVMMySQLException
)

// 定义 ChaosType 对应的 map
var ChaosTypeMap = map[ChaosType]string{
	PodKill:             "PodKill",
	PodFailure:          "PodFailure",
	ContainerKill:       "ContainerKill",
	MemoryStress:        "MemoryStress",
	CPUStress:           "CPUStress",
	HTTPAbort:           "HTTPAbort",
	HTTPDelay:           "HTTPDelay",
	HTTPReplace:         "HTTPReplace",
	HTTPPatch:           "HTTPPatch",
	DNSError:            "DNSError",
	DNSRandom:           "DNSRandom",
	TimeSkew:            "TimeSkew",
	NetworkDelay:        "NetworkDelay",
	NetworkLoss:         "NetworkLoss",
	NetworkDuplicate:    "NetworkDuplicate",
	NetworkCorrupt:      "NetworkCorrupt",
	NetworkBandwidth:    "NetworkBandwidth",
	NetworkPartition:    "NetworkPartition",
	JVMLatency:          "JVMLatency",
	JVMReturn:           "JVMReturn",
	JVMException:        "JVMException",
	JVMGarbageCollector: "JVMGarbageCollector",
	JVMRuleData:         "JVMRuleData",
	JVMMySQL:            "JVMMySQL",
	JVMCPUStress:        "JVMCPUStress",
	JVMMemoryStress:     "JVMMemoryStress",
	JVMMySQLLatency:     "JVMMySQLLatency",
	JVMMySQLException:   "JVMMySQLException",
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

// DNSErrorSpec defines the DNS error chaos injection parameters
type DNSErrorSpec struct {
	Duration  int `range:"1-60" description:"Time Unit Minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"String"`
	AppName   int `range:"0-0" dynamic:"true" description:"Array"`
}

func (s *DNSErrorSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.ErrorAction

	// Use a simple pattern that matches all domains
	patterns := []string{"*"}

	return controllers.CreateDnsChaos(cli, TargetNamespace, labelArr[s.AppName], action, patterns, duration)
}

// DNSRandomSpec defines the DNS random chaos injection parameters
type DNSRandomSpec struct {
	Duration  int `range:"1-60" description:"Time Unit Minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"String"`
	AppName   int `range:"0-0" dynamic:"true" description:"Array"`
}

func (s *DNSRandomSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}

	duration := pointer.String(strconv.Itoa(s.Duration) + "m")
	action := chaosmeshv1alpha1.RandomAction

	// Use a simple pattern that matches all domains
	patterns := []string{"*"}

	return controllers.CreateDnsChaos(cli, TargetNamespace, labelArr[s.AppName], action, patterns, duration)
}

type JVMLatencySpec struct {
	Duration        int    `range:"1-60" description:"Time Unit Minute"`
	Namespace       int    `range:"0-0" dynamic:"true" description:"String"`
	AppName         int    `range:"0-0" dynamic:"true" description:"Array"`
	Class           string `range:"0-0" description:"Target Java Class"`
	Method          string `range:"0-0" description:"Target Method"`
	LatencyDuration int    `range:"1-5000" description:"Latency in ms"`
}

func (s *JVMLatencySpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	opts := []chaos.OptJVMChaos{
		chaos.WithJVMClass(s.Class),
		chaos.WithJVMMethod(s.Method),
		chaos.WithJVMLatencyDuration(s.LatencyDuration),
	}

	return controllers.CreateJVMChaos(cli, TargetNamespace, labelArr[s.AppName],
		chaosmeshv1alpha1.JVMLatencyAction, duration, opts...)
}

// JVM Return Value Type
type JVMReturnType int

const (
	StringReturn JVMReturnType = 1
	IntReturn    JVMReturnType = 2
)

type JVMReturnSpec struct {
	Duration       int           `range:"1-60" description:"Time Unit Minute"`
	Namespace      int           `range:"0-0" dynamic:"true" description:"String"`
	AppName        int           `range:"0-0" dynamic:"true" description:"Array"`
	Class          string        `range:"0-0" description:"Target Java Class"`
	Method         string        `range:"0-0" description:"Target Method"`
	ReturnType     JVMReturnType `range:"1-2" description:"Return Type (1=String, 2=Int)"`
	CustomValue    bool          `range:"0-1" description:"Use custom return value?"`
	CustomValueStr string        `range:"0-0" description:"Custom return value (if enabled)"`
}

func (s *JVMReturnSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	opts := []chaos.OptJVMChaos{
		chaos.WithJVMClass(s.Class),
		chaos.WithJVMMethod(s.Method),
	}

	if s.CustomValue && s.CustomValueStr != "" {
		// Use custom value
		if s.ReturnType == StringReturn && s.CustomValueStr[0] != '"' {
			// Add quotes for string type if not present
			opts = append(opts, chaos.WithJVMReturnValue(fmt.Sprintf("\"%s\"", s.CustomValueStr)))
		} else {
			opts = append(opts, chaos.WithJVMReturnValue(s.CustomValueStr))
		}
	} else {
		// Use random or default value
		if s.ReturnType == StringReturn {
			opts = append(opts, chaos.WithJVMRandomStringReturn(8))
		} else {
			opts = append(opts, chaos.WithJVMRandomIntReturn(1, 1000))
		}
	}

	return controllers.CreateJVMChaos(cli, TargetNamespace, labelArr[s.AppName],
		chaosmeshv1alpha1.JVMReturnAction, duration, opts...)
}

type JVMExceptionSpec struct {
	Duration     int    `range:"1-60" description:"Time Unit Minute"`
	Namespace    int    `range:"0-0" dynamic:"true" description:"String"`
	AppName      int    `range:"0-0" dynamic:"true" description:"Array"`
	Class        string `range:"0-0" description:"Target Java Class"`
	Method       string `range:"0-0" description:"Target Method"`
	CustomExp    bool   `range:"0-1" description:"Use custom exception?"`
	CustomExpStr string `range:"0-0" description:"Custom exception (if enabled)"`
}

func (s *JVMExceptionSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	opts := []chaos.OptJVMChaos{
		chaos.WithJVMClass(s.Class),
		chaos.WithJVMMethod(s.Method),
	}

	if s.CustomExp && s.CustomExpStr != "" {
		opts = append(opts, chaos.WithJVMException(s.CustomExpStr))
	} else {
		opts = append(opts, chaos.WithJVMDefaultException())
	}

	return controllers.CreateJVMChaos(cli, TargetNamespace, labelArr[s.AppName],
		chaosmeshv1alpha1.JVMExceptionAction, duration, opts...)
}

type JVMGCSpec struct {
	Duration  int `range:"1-60" description:"Time Unit Minute"`
	Namespace int `range:"0-0" dynamic:"true" description:"String"`
	AppName   int `range:"0-0" dynamic:"true" description:"Array"`
}

func (s *JVMGCSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	return controllers.CreateJVMChaos(cli, TargetNamespace, labelArr[s.AppName],
		chaosmeshv1alpha1.JVMGCAction, duration)
}

// JVM Stress Memory Type
type JVMMemoryType int

const (
	HeapMemory  JVMMemoryType = 1
	StackMemory JVMMemoryType = 2
)

// JVMStressSpec defines the JVM stress chaos injection parameters
type JVMStressSpec struct {
	Duration  int           `range:"1-60" description:"Time Unit Minute"`
	Namespace int           `range:"0-0" dynamic:"true" description:"String"`
	AppName   int           `range:"0-0" dynamic:"true" description:"Array"`
	Class     string        `range:"0-0" description:"Target Java Class"`
	Method    string        `range:"0-0" description:"Target Method"`
	CPUCount  int           `range:"1-8" description:"Number of CPU cores to stress"`
	MemType   JVMMemoryType `range:"1-2" description:"Memory Type (1=Heap, 2=Stack)"`
}

// JVMCPUStressSpec defines the JVM CPU stress chaos injection parameters
type JVMCPUStressSpec struct {
	Duration  int    `range:"1-60" description:"Time Unit Minute"`
	Namespace int    `range:"0-0" dynamic:"true" description:"String"`
	AppName   int    `range:"0-0" dynamic:"true" description:"Array"`
	Class     string `range:"0-0" description:"Target Java Class"`
	Method    string `range:"0-0" description:"Target Method"`
	CPUCount  int    `range:"1-8" description:"Number of CPU cores to stress"`
}

func (s *JVMCPUStressSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	opts := []chaos.OptJVMChaos{
		chaos.WithJVMClass(s.Class),
		chaos.WithJVMMethod(s.Method),
		chaos.WithJVMStressCPUCount(s.CPUCount),
	}

	return controllers.CreateJVMChaos(cli, TargetNamespace, labelArr[s.AppName],
		chaosmeshv1alpha1.JVMStressAction, duration, opts...)
}

// JVMMemoryStressSpec defines the JVM memory stress chaos injection parameters
type JVMMemoryStressSpec struct {
	Duration  int           `range:"1-60" description:"Time Unit Minute"`
	Namespace int           `range:"0-0" dynamic:"true" description:"String"`
	AppName   int           `range:"0-0" dynamic:"true" description:"Array"`
	Class     string        `range:"0-0" description:"Target Java Class"`
	Method    string        `range:"0-0" description:"Target Method"`
	MemType   JVMMemoryType `range:"1-2" description:"Memory Type (1=Heap, 2=Stack)"`
}

func (s *JVMMemoryStressSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	// Convert memory type
	memType := "heap"
	if s.MemType == StackMemory {
		memType = "stack"
	}

	opts := []chaos.OptJVMChaos{
		chaos.WithJVMClass(s.Class),
		chaos.WithJVMMethod(s.Method),
		chaos.WithJVMStressMemType(memType),
	}

	return controllers.CreateJVMChaos(cli, TargetNamespace, labelArr[s.AppName],
		chaosmeshv1alpha1.JVMStressAction, duration, opts...)
}

// JVMRuleDataSpec defines the JVM custom rule injection parameters
type JVMRuleDataSpec struct {
	Duration  int    `range:"1-60" description:"Time Unit Minute"`
	Namespace int    `range:"0-0" dynamic:"true" description:"String"`
	AppName   int    `range:"0-0" dynamic:"true" description:"Array"`
	RuleName  string `range:"0-0" description:"Byteman Rule Name"`
	RuleData  string `range:"0-0" description:"Byteman Rule Data"`
}

func (s *JVMRuleDataSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	opts := []chaos.OptJVMChaos{
		chaos.WithJVMName(s.RuleName),
		chaos.WithJVMRuleData(s.RuleData),
	}

	return controllers.CreateJVMChaos(cli, TargetNamespace, labelArr[s.AppName],
		chaosmeshv1alpha1.JVMRuleDataAction, duration, opts...)
}

// SQL types for JVMMySQL
type MySQLType int

const (
	AllSQL     MySQLType = 0
	SelectSQL  MySQLType = 1
	InsertSQL  MySQLType = 2
	UpdateSQL  MySQLType = 3
	DeleteSQL  MySQLType = 4
	ReplaceSQL MySQLType = 5
)

// MySQL connector versions
type MySQLConnectorVersion int

const (
	MySQL5 MySQLConnectorVersion = 5
	MySQL8 MySQLConnectorVersion = 8
)

// Define available MySQL tables for selection
var AvailableMySQLTables = []string{
	"assurance",
	"auth_user",
	"config",
	"consign_price",
	"consign_record",
	"contacts",
	"delivery",
	"food_delivery_order",
	"food_delivery_order_food_list",
	"food_order",
	"inside_money",
	"inside_payment",
	"money",
	"notify_info",
	"office",
	"orders",
	"orders_other",
	"payment",
	"price_config",
	"route",
	"route_distances",
	"route_stations",
	"security_config",
	"station",
	"station_food_list",
	"station_food_store",
	"train_food",
	"train_food_list",
	"train_type",
	"trip",
	"trip2",
	"user",
	"user_roles",
	"voucher",
	"wait_list_order",
}

// JVMMySQLLatencySpec defines the JVM MySQL latency chaos injection parameters
type JVMMySQLLatencySpec struct {
	Duration   int `range:"1-60" description:"Time Unit Minute"`
	Namespace  int `range:"0-0" dynamic:"true" description:"String"`
	AppName    int `range:"0-0" dynamic:"true" description:"Array"`
	LatencyMs  int `range:"10-5000" description:"Latency in ms"`
	TableIndex int `range:"0-38" description:"Index of table to target (or -1 for all)"`
	SQLType    int `range:"0-5" description:"SQL Type (0=All, 1=Select, 2=Insert, 3=Update, 4=Delete, 5=Replace)"`
}

func (s *JVMMySQLLatencySpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	// Convert SQL type to string
	sqlTypeStr := convertSQLTypeToString(s.SQLType)

	// Determine target table
	var tableStr string
	if s.TableIndex < 0 || s.TableIndex >= len(AvailableMySQLTables) {
		tableStr = "" // Empty means all tables
	} else {
		tableStr = AvailableMySQLTables[s.TableIndex]
	}

	opts := []chaos.OptJVMChaos{
		chaos.WithJVMMySQLConnector("5"), // Hardcoded to version 5
		chaos.WithJVMMySQLDatabase("ts"), // Hardcoded to ts database
		chaos.WithJVMMySQLTable(tableStr),
		chaos.WithJVMMySQLType(sqlTypeStr),
		chaos.WithJVMLatencyDuration(s.LatencyMs),
	}

	return controllers.CreateJVMChaos(cli, TargetNamespace, labelArr[s.AppName],
		chaosmeshv1alpha1.JVMMySQLAction, duration, opts...)
}

// JVMMySQLExceptionSpec defines the JVM MySQL exception chaos injection parameters
type JVMMySQLExceptionSpec struct {
	Duration     int    `range:"1-60" description:"Time Unit Minute"`
	Namespace    int    `range:"0-0" dynamic:"true" description:"String"`
	AppName      int    `range:"0-0" dynamic:"true" description:"Array"`
	ExceptionMsg string `range:"0-0" description:"Exception message"`
	TableIndex   int    `range:"0-38" description:"Index of table to target (or -1 for all)"`
	SQLType      int    `range:"0-5" description:"SQL Type (0=All, 1=Select, 2=Insert, 3=Update, 4=Delete, 5=Replace)"`
}

func (s *JVMMySQLExceptionSpec) Create(cli cli.Client) string {
	labelArr, err := client.GetLabels(TargetNamespace, TargetLabelKey)
	if err != nil {
		return ""
	}
	duration := pointer.String(strconv.Itoa(s.Duration) + "m")

	// Convert SQL type to string
	sqlTypeStr := convertSQLTypeToString(s.SQLType)

	// Determine target table
	var tableStr string
	if s.TableIndex < 0 || s.TableIndex >= len(AvailableMySQLTables) {
		tableStr = "" // Empty means all tables
	} else {
		tableStr = AvailableMySQLTables[s.TableIndex]
	}

	opts := []chaos.OptJVMChaos{
		chaos.WithJVMMySQLConnector("5"), // Hardcoded to version 5
		chaos.WithJVMMySQLDatabase("ts"), // Hardcoded to ts database
		chaos.WithJVMMySQLTable(tableStr),
		chaos.WithJVMMySQLType(sqlTypeStr),
		chaos.WithJVMException(s.ExceptionMsg),
	}

	return controllers.CreateJVMChaos(cli, TargetNamespace, labelArr[s.AppName],
		chaosmeshv1alpha1.JVMMySQLAction, duration, opts...)
}

// Helper function to convert SQL type from int to string
func convertSQLTypeToString(sqlType int) string {
	switch sqlType {
	case 1:
		return "select"
	case 2:
		return "insert"
	case 3:
		return "update"
	case 4:
		return "delete"
	case 5:
		return "replace"
	default:
		return "" // All SQL types
	}
}

var SpecMap = map[ChaosType]any{
	CPUStress:           CPUStressChaosSpec{},
	MemoryStress:        MemoryStressChaosSpec{},
	HTTPAbort:           HTTPChaosAbortSpec{},
	HTTPReplace:         HTTPChaosReplaceSpec{},
	DNSError:            DNSErrorSpec{},
	DNSRandom:           DNSRandomSpec{},
	TimeSkew:            TimeSkewSpec{},
	NetworkDelay:        NetworkDelaySpec{},
	NetworkLoss:         NetworkLossSpec{},
	NetworkDuplicate:    NetworkDuplicateSpec{},
	NetworkCorrupt:      NetworkCorruptSpec{},
	NetworkBandwidth:    NetworkBandwidthSpec{},
	NetworkPartition:    NetworkPartitionSpec{},
	JVMLatency:          JVMLatencySpec{},
	JVMReturn:           JVMReturnSpec{},
	JVMException:        JVMExceptionSpec{},
	JVMGarbageCollector: JVMGCSpec{},
	JVMRuleData:         JVMRuleDataSpec{},
	JVMCPUStress:        JVMCPUStressSpec{},
	JVMMemoryStress:     JVMMemoryStressSpec{},
	JVMMySQLLatency:     JVMMySQLLatencySpec{},
	JVMMySQLException:   JVMMySQLExceptionSpec{},
}

var ChaosHandlers = map[ChaosType]Injection{
	PodKill:             &PodKillSpec{},
	PodFailure:          &PodFailureSpec{},
	ContainerKill:       &ContainerKillSpec{},
	MemoryStress:        &MemoryStressChaosSpec{},
	CPUStress:           &CPUStressChaosSpec{},
	HTTPAbort:           &HTTPChaosAbortSpec{},
	HTTPReplace:         &HTTPChaosReplaceSpec{},
	DNSError:            &DNSErrorSpec{},
	DNSRandom:           &DNSRandomSpec{},
	TimeSkew:            &TimeSkewSpec{},
	NetworkDelay:        &NetworkDelaySpec{},
	NetworkLoss:         &NetworkLossSpec{},
	NetworkDuplicate:    &NetworkDuplicateSpec{},
	NetworkCorrupt:      &NetworkCorruptSpec{},
	NetworkBandwidth:    &NetworkBandwidthSpec{},
	NetworkPartition:    &NetworkPartitionSpec{},
	JVMLatency:          &JVMLatencySpec{},
	JVMReturn:           &JVMReturnSpec{},
	JVMException:        &JVMExceptionSpec{},
	JVMGarbageCollector: &JVMGCSpec{},
	JVMRuleData:         &JVMRuleDataSpec{},
	JVMCPUStress:        &JVMCPUStressSpec{},
	JVMMemoryStress:     &JVMMemoryStressSpec{},
	JVMMySQLLatency:     &JVMMySQLLatencySpec{},
	JVMMySQLException:   &JVMMySQLExceptionSpec{},
}

type InjectionConf struct {
	PodKill             *PodKillSpec           `range:"0-2"`
	PodFailure          *PodFailureSpec        `range:"0-2"`
	ContainerKill       *ContainerKillSpec     `range:"0-2"`
	MemoryStress        *MemoryStressChaosSpec `range:"0-4"`
	CPUStress           *CPUStressChaosSpec    `range:"0-4"`
	HTTPAbort           *HTTPChaosAbortSpec    `range:"0-3"`
	HTTPReplace         *HTTPChaosReplaceSpec  `range:"0-4"`
	DNSError            *DNSErrorSpec          `range:"0-2"`
	DNSRandom           *DNSRandomSpec         `range:"0-2"`
	TimeSkew            *TimeSkewSpec          `range:"0-3"`
	NetworkDelay        *NetworkDelaySpec      `range:"0-7"`
	NetworkLoss         *NetworkLossSpec       `range:"0-6"`
	NetworkDuplicate    *NetworkDuplicateSpec  `range:"0-6"`
	NetworkCorrupt      *NetworkCorruptSpec    `range:"0-6"`
	NetworkBandwidth    *NetworkBandwidthSpec  `range:"0-7"`
	NetworkPartition    *NetworkPartitionSpec  `range:"0-4"`
	JVMLatency          *JVMLatencySpec        `range:"0-5"`
	JVMReturn           *JVMReturnSpec         `range:"0-7"`
	JVMException        *JVMExceptionSpec      `range:"0-5"`
	JVMGarbageCollector *JVMGCSpec             `range:"0-2"`
	JVMRuleData         *JVMRuleDataSpec       `range:"0-4"`
	JVMCPUStress        *JVMCPUStressSpec      `range:"0-5"`
	JVMMemoryStress     *JVMMemoryStressSpec   `range:"0-4"`
	JVMMySQLLatency     *JVMMySQLLatencySpec   `range:"0-5"`
	JVMMySQLException   *JVMMySQLExceptionSpec `range:"0-5"`
}
