package handler

import (
	"fmt"
	"reflect"

	"github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/chaos-experiment/utils"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

type ChaosType int

const TargetNamespace = "ts" // todo: make it dynamic (e.g. from config)
const TargetLabelKey = "app"

const (
	// PodChaos
	PodKill ChaosType = iota
	PodFailure
	ContainerKill

	// StressChaos
	MemoryStress
	CPUStress

	// HTTPChaos
	HTTPRequestAbort
	HTTPResponseAbort
	HTTPRequestDelay
	HTTPResponseDelay
	HTTPResponseReplaceBody
	HTTPResponsePatchBody
	HTTPRequestReplacePath
	HTTPRequestReplaceMethod
	HTTPResponseReplaceCode

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
	JVMCPUStress
	JVMMemoryStress
	JVMMySQLLatency
	JVMMySQLException
)

// Define ChaosType to name mapping
var ChaosTypeMap = map[ChaosType]string{
	PodKill:                  "PodKill",
	PodFailure:               "PodFailure",
	ContainerKill:            "ContainerKill",
	MemoryStress:             "MemoryStress",
	CPUStress:                "CPUStress",
	HTTPRequestAbort:         "HTTPRequestAbort",
	HTTPResponseAbort:        "HTTPResponseAbort",
	HTTPRequestDelay:         "HTTPRequestDelay",
	HTTPResponseDelay:        "HTTPResponseDelay",
	HTTPResponseReplaceBody:  "HTTPResponseReplaceBody",
	HTTPResponsePatchBody:    "HTTPResponsePatchBody",
	HTTPRequestReplacePath:   "HTTPRequestReplacePath",
	HTTPRequestReplaceMethod: "HTTPRequestReplaceMethod",
	HTTPResponseReplaceCode:  "HTTPResponseReplaceCode",
	DNSError:                 "DNSError",
	DNSRandom:                "DNSRandom",
	TimeSkew:                 "TimeSkew",
	NetworkDelay:             "NetworkDelay",
	NetworkLoss:              "NetworkLoss",
	NetworkDuplicate:         "NetworkDuplicate",
	NetworkCorrupt:           "NetworkCorrupt",
	NetworkBandwidth:         "NetworkBandwidth",
	NetworkPartition:         "NetworkPartition",
	JVMLatency:               "JVMLatency",
	JVMReturn:                "JVMReturn",
	JVMException:             "JVMException",
	JVMGarbageCollector:      "JVMGarbageCollector",
	JVMCPUStress:             "JVMCPUStress",
	JVMMemoryStress:          "JVMMemoryStress",
	JVMMySQLLatency:          "JVMMySQLLatency",
	JVMMySQLException:        "JVMMySQLException",
}

// GetChaosTypeName 根据 ChaosType 获取名称
func GetChaosTypeName(c ChaosType) string {
	if name, ok := ChaosTypeMap[c]; ok {
		return name
	}
	return "Unknown"
}

type Conf struct {
	Namespace string
}
type Option func(*Conf)

func WithNs(ns string) Option {
	return func(c *Conf) {
		c.Namespace = ns
	}
}

type Injection interface {
	Create(cli cli.Client, opt ...Option) (string, error)
}

var SpecMap = map[ChaosType]any{

	CPUStress:                CPUStressChaosSpec{},
	MemoryStress:             MemoryStressChaosSpec{},
	HTTPRequestAbort:         HTTPRequestAbortSpec{},
	HTTPResponseAbort:        HTTPResponseAbortSpec{},
	HTTPRequestDelay:         HTTPRequestDelaySpec{},
	HTTPResponseDelay:        HTTPResponseDelaySpec{},
	HTTPResponseReplaceBody:  HTTPResponseReplaceBodySpec{},
	HTTPResponsePatchBody:    HTTPResponsePatchBodySpec{},
	HTTPRequestReplacePath:   HTTPRequestReplacePathSpec{},
	HTTPRequestReplaceMethod: HTTPRequestReplaceMethodSpec{},
	HTTPResponseReplaceCode:  HTTPResponseReplaceCodeSpec{},
	DNSError:                 DNSErrorSpec{},
	DNSRandom:                DNSRandomSpec{},
	TimeSkew:                 TimeSkewSpec{},
	NetworkDelay:             NetworkDelaySpec{},
	NetworkLoss:              NetworkLossSpec{},
	NetworkDuplicate:         NetworkDuplicateSpec{},
	NetworkCorrupt:           NetworkCorruptSpec{},
	NetworkBandwidth:         NetworkBandwidthSpec{},
	NetworkPartition:         NetworkPartitionSpec{},
	JVMLatency:               JVMLatencySpec{},
	JVMReturn:                JVMReturnSpec{},
	JVMException:             JVMExceptionSpec{},
	JVMGarbageCollector:      JVMGCSpec{},
	JVMCPUStress:             JVMCPUStressSpec{},
	JVMMemoryStress:          JVMMemoryStressSpec{},
	JVMMySQLLatency:          JVMMySQLLatencySpec{},
	JVMMySQLException:        JVMMySQLExceptionSpec{},
}

var ChaosHandlers = map[ChaosType]Injection{
	PodKill:                  &PodKillSpec{},
	PodFailure:               &PodFailureSpec{},
	ContainerKill:            &ContainerKillSpec{},
	MemoryStress:             &MemoryStressChaosSpec{},
	CPUStress:                &CPUStressChaosSpec{},
	HTTPRequestAbort:         &HTTPRequestAbortSpec{},
	HTTPResponseAbort:        &HTTPResponseAbortSpec{},
	HTTPRequestDelay:         &HTTPRequestDelaySpec{},
	HTTPResponseDelay:        &HTTPResponseDelaySpec{},
	HTTPResponseReplaceBody:  &HTTPResponseReplaceBodySpec{},
	HTTPResponsePatchBody:    &HTTPResponsePatchBodySpec{},
	HTTPRequestReplacePath:   &HTTPRequestReplacePathSpec{},
	HTTPRequestReplaceMethod: &HTTPRequestReplaceMethodSpec{},
	HTTPResponseReplaceCode:  &HTTPResponseReplaceCodeSpec{},
	DNSError:                 &DNSErrorSpec{},
	DNSRandom:                &DNSRandomSpec{},
	TimeSkew:                 &TimeSkewSpec{},
	NetworkDelay:             &NetworkDelaySpec{},
	NetworkLoss:              &NetworkLossSpec{},
	NetworkDuplicate:         &NetworkDuplicateSpec{},
	NetworkCorrupt:           &NetworkCorruptSpec{},
	NetworkBandwidth:         &NetworkBandwidthSpec{},
	NetworkPartition:         &NetworkPartitionSpec{},
	JVMLatency:               &JVMLatencySpec{},
	JVMReturn:                &JVMReturnSpec{},
	JVMException:             &JVMExceptionSpec{},
	JVMGarbageCollector:      &JVMGCSpec{},
	JVMCPUStress:             &JVMCPUStressSpec{},
	JVMMemoryStress:          &JVMMemoryStressSpec{},
	JVMMySQLLatency:          &JVMMySQLLatencySpec{},
	JVMMySQLException:        &JVMMySQLExceptionSpec{},
}

type InjectionConf struct {
	PodKill                  *PodKillSpec                  `range:"0-2"`
	PodFailure               *PodFailureSpec               `range:"0-2"`
	ContainerKill            *ContainerKillSpec            `range:"0-2"`
	MemoryStress             *MemoryStressChaosSpec        `range:"0-4"`
	CPUStress                *CPUStressChaosSpec           `range:"0-4"`
	HTTPRequestAbort         *HTTPRequestAbortSpec         `range:"0-2"`
	HTTPResponseAbort        *HTTPResponseAbortSpec        `range:"0-2"`
	HTTPRequestDelay         *HTTPRequestDelaySpec         `range:"0-3"`
	HTTPResponseDelay        *HTTPResponseDelaySpec        `range:"0-3"`
	HTTPResponseReplaceBody  *HTTPResponseReplaceBodySpec  `range:"0-3"`
	HTTPResponsePatchBody    *HTTPResponsePatchBodySpec    `range:"0-2"`
	HTTPRequestReplacePath   *HTTPRequestReplacePathSpec   `range:"0-2"`
	HTTPRequestReplaceMethod *HTTPRequestReplaceMethodSpec `range:"0-3"`
	HTTPResponseReplaceCode  *HTTPResponseReplaceCodeSpec  `range:"0-3"`
	DNSError                 *DNSErrorSpec                 `range:"0-2"`
	DNSRandom                *DNSRandomSpec                `range:"0-2"`
	TimeSkew                 *TimeSkewSpec                 `range:"0-3"`
	NetworkDelay             *NetworkDelaySpec             `range:"0-6"`
	NetworkLoss              *NetworkLossSpec              `range:"0-5"`
	NetworkDuplicate         *NetworkDuplicateSpec         `range:"0-5"`
	NetworkCorrupt           *NetworkCorruptSpec           `range:"0-5"`
	NetworkBandwidth         *NetworkBandwidthSpec         `range:"0-6"`
	NetworkPartition         *NetworkPartitionSpec         `range:"0-3"`
	JVMLatency               *JVMLatencySpec               `range:"0-3"`
	JVMReturn                *JVMReturnSpec                `range:"0-4"`
	JVMException             *JVMExceptionSpec             `range:"0-3"`
	JVMGarbageCollector      *JVMGCSpec                    `range:"0-2"`
	JVMCPUStress             *JVMCPUStressSpec             `range:"0-3"`
	JVMMemoryStress          *JVMMemoryStressSpec          `range:"0-3"`
	JVMMySQLLatency          *JVMMySQLLatencySpec          `range:"0-3"`
	JVMMySQLException        *JVMMySQLExceptionSpec        `range:"0-2"`
}

func (ic *InjectionConf) Create(opts ...Option) (map[string]any, string, error) {
	cli := client.NewK8sClient()
	instance, config, err := ic.getActiveInjection()
	if err != nil {
		return nil, "", err
	}

	name, err := instance.Create(cli, opts...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to inject chaos for %T: %w", instance, err)
	}

	return config, name, nil
}

func (ic *InjectionConf) getActiveInjection() (Injection, map[string]any, error) {
	val := reflect.ValueOf(ic).Elem()

	for i := range val.NumField() {
		field := val.Field(i)
		if field.IsNil() {
			continue
		}

		instance := field.Interface().(Injection)
		instanceValue := reflect.ValueOf(instance).Elem()
		instanceType := instanceValue.Type()

		result := make(map[string]any, instanceValue.NumField())
		for i := range instanceValue.NumField() {
			key := utils.ToSnakeCase(instanceType.Field(i).Name)
			switch i {
			case 1:
				result[key] = TargetNamespace
			case 2:
				labels, err := client.GetLabels(TargetNamespace, TargetLabelKey)
				if err != nil {
					return nil, nil, err
				}

				index, err := getIntValue(instanceValue.Field(i))
				if err != nil {
					return nil, nil, err
				}

				result[key] = labels[index]
			default:
				value, err := getIntValue(instanceValue.Field(i))
				if err != nil {
					return nil, nil, err
				}

				result[key] = value
			}
		}

		return instance, result, nil
	}

	return nil, nil, fmt.Errorf("failed to get the non-empty injection")
}

func getIntValue(field reflect.Value) (int64, error) {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int(), nil
	default:
		return 0, fmt.Errorf("unsupported field type: %v", field.Kind())
	}
}
