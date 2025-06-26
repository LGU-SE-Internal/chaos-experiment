package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/LGU-SE-Internal/chaos-experiment/client"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/resourcelookup"
	"github.com/LGU-SE-Internal/chaos-experiment/utils"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

type ChaosType int

type NamespaceInfo struct {
	Namespace string
	Count     int
}

var (
	NamespacePrefixs   []string
	NamespaceTargetMap map[string]int
	TargetLabelKey     string
)

func InitTargetConfig(namespaceTargetMap map[string]int, targetLabelKey string) error {
	NamespaceTargetMap = namespaceTargetMap
	TargetLabelKey = targetLabelKey

	allNamespaces, err := client.ListNamespaces()
	if err != nil {
		return err
	}

	allNamespaceMap := make(map[string]struct{}, len(allNamespaces))
	for _, ns := range allNamespaces {
		allNamespaceMap[ns] = struct{}{}
	}

	TargetLabelKey = targetLabelKey
	namespacePrefixs := make([]string, 0, len(namespaceTargetMap))
	for ns, count := range namespaceTargetMap {
		for i := DefaultStartIndex; i < count; i++ {
			namespace := fmt.Sprintf("%s%d", ns, i)
			_, exists := allNamespaceMap[namespace]
			if !exists {
				return fmt.Errorf("namespace %s does not exist in the cluster", namespace)
			}
		}

		namespacePrefixs = append(namespacePrefixs, ns)
	}

	sort.Strings(namespacePrefixs)
	NamespacePrefixs = namespacePrefixs

	resourcelookup.InitCaches()
	for _, ns := range namespacePrefixs {
		namespace := fmt.Sprintf("%s%d", ns, DefaultStartIndex)
		if err := resourcelookup.PreloadCaches(namespace, targetLabelKey); err != nil {
			return fmt.Errorf("failed to preload caches of namespace: %v", err)
		}
	}

	return nil
}

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

var ChaosNameMap = map[string]ChaosType{
	"PodKill":                  PodKill,
	"PodFailure":               PodFailure,
	"ContainerKill":            ContainerKill,
	"MemoryStress":             MemoryStress,
	"CPUStress":                CPUStress,
	"HTTPRequestAbort":         HTTPRequestAbort,
	"HTTPResponseAbort":        HTTPResponseAbort,
	"HTTPRequestDelay":         HTTPRequestDelay,
	"HTTPResponseDelay":        HTTPResponseDelay,
	"HTTPResponseReplaceBody":  HTTPResponseReplaceBody,
	"HTTPResponsePatchBody":    HTTPResponsePatchBody,
	"HTTPRequestReplacePath":   HTTPRequestReplacePath,
	"HTTPRequestReplaceMethod": HTTPRequestReplaceMethod,
	"HTTPResponseReplaceCode":  HTTPResponseReplaceCode,
	"DNSError":                 DNSError,
	"DNSRandom":                DNSRandom,
	"TimeSkew":                 TimeSkew,
	"NetworkDelay":             NetworkDelay,
	"NetworkLoss":              NetworkLoss,
	"NetworkDuplicate":         NetworkDuplicate,
	"NetworkCorrupt":           NetworkCorrupt,
	"NetworkBandwidth":         NetworkBandwidth,
	"NetworkPartition":         NetworkPartition,
	"JVMLatency":               JVMLatency,
	"JVMReturn":                JVMReturn,
	"JVMException":             JVMException,
	"JVMGarbageCollector":      JVMGarbageCollector,
	"JVMCPUStress":             JVMCPUStress,
	"JVMMemoryStress":          JVMMemoryStress,
	"JVMMySQLLatency":          JVMMySQLLatency,
	"JVMMySQLException":        JVMMySQLException,
}

// GetChaosTypeName 根据 ChaosType 获取名称
func GetChaosTypeName(c ChaosType) string {
	if name, ok := ChaosTypeMap[c]; ok {
		return name
	}
	return "Unknown"
}

type Conf struct {
	Annoations map[string]string
	Context    context.Context
	Labels     map[string]string
	Namespace  string
}
type Option func(*Conf)

func WithAnnotations(annotations map[string]string) Option {
	return func(c *Conf) {
		c.Annoations = annotations
	}
}

func WithContext(ctx context.Context) Option {
	return func(c *Conf) {
		c.Context = ctx
	}
}

func WithLabels(labels map[string]string) Option {
	return func(c *Conf) {
		c.Labels = labels
	}
}

func WithNs(ns string) Option {
	return func(c *Conf) {
		c.Namespace = ns
	}
}

type Injection interface {
	Create(cli cli.Client, opt ...Option) (string, error)
}
type GroundtruthProvider interface {
	GetGroundtruth() (Groundtruth, error)
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

func (ic *InjectionConf) Create(ctx context.Context, namespace string, annotations map[string]string, labels map[string]string) (string, error) {
	activeField, err := ic.getActiveField()
	if err != nil {
		return "", err
	}

	instance := activeField.Interface().(Injection)
	name, err := instance.Create(
		client.NewK8sClient(),
		WithAnnotations(annotations),
		WithContext(ctx),
		WithLabels(labels),
		WithNs(namespace),
	)
	if err != nil {
		return "", fmt.Errorf("failed to inject chaos for %T: %w", instance, err)
	}

	return name, nil
}

func (ic *InjectionConf) getActiveField() (reflect.Value, error) {
	val := reflect.ValueOf(ic).Elem()

	var activeField reflect.Value
	for i := range val.NumField() {
		field := val.Field(i)
		if !field.IsNil() {
			activeField = field
			break
		}
	}

	if !activeField.IsValid() {
		return reflect.Value{}, fmt.Errorf("failed to get the non-empty injection")
	}

	return activeField, nil
}

func (ic *InjectionConf) getActiveInjection() (Injection, error) {
	activeField, err := ic.getActiveField()
	if err != nil {
		return nil, err
	}

	return activeField.Interface().(Injection), nil
}

func (ic *InjectionConf) GetDisplayConfig() (map[string]any, error) {
	instance, err := ic.getActiveInjection()
	if err != nil {
		return nil, err
	}

	instanceValue := reflect.ValueOf(instance).Elem()
	instanceType := instanceValue.Type()

	result := make(map[string]any, instanceValue.NumField())

	var prefix string
	for i := range instanceValue.NumField() {
		if instanceType.Field(i).Name == KeyNamespace {
			index, err := getIntValue(instanceValue.Field(i))
			if err != nil {
				return nil, err
			}

			if index >= 0 && int(index) < len(NamespacePrefixs) {
				prefix = NamespacePrefixs[index]
				break
			}
		}
	}

	for i := range instanceValue.NumField() {
		key := utils.ToSnakeCase(instanceType.Field(i).Name)

		index, err := getIntValue(instanceValue.Field(i))
		if err != nil {
			return nil, err
		}

		var value any
		switch i {
		case 1:
			result[key] = prefix
		case 2:
			switch instanceType.Field(i).Name {
			case KeyApp:
				namespace := fmt.Sprintf("%s%d", prefix, DefaultStartIndex)
				labels, err := resourcelookup.GetAllAppLabels(namespace, TargetLabelKey)
				if err != nil || len(labels) == 0 {
					return nil, err
				}

				value = map[string]any{"app_name": labels[index]}
			case KeyMethod:
				methods, err := resourcelookup.GetAllJVMMethods()
				if err != nil {
					return nil, err
				}

				value = methods[index]
			case KeyEndpoint:
				endpoints, err := resourcelookup.GetAllHTTPEndpoints()
				if err != nil {
					return nil, err
				}

				value = endpoints[index]
			case KeyNetworkPair:
				networkpairs, err := resourcelookup.GetAllNetworkPairs()
				if err != nil {
					return nil, err
				}

				value = networkpairs[index]
			case KeyContainer:
				namespace := fmt.Sprintf("%s%d", prefix, DefaultStartIndex)
				containers, err := resourcelookup.GetAllContainers(namespace)
				if err != nil {
					return nil, err
				}

				value = containers[index]
			case KeyDNSEndpoint:
				endpoints, err := resourcelookup.GetAllDNSEndpoints()
				if err != nil {
					return nil, err
				}

				value = endpoints[index]
			case KeyDatabase:
				operations, err := resourcelookup.GetAllDatabaseOperations()
				if err != nil {
					return nil, err
				}

				value = operations[index]
			}

			jsonData, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal injection point: %v", err)
			}

			var injectionPoint map[string]any
			if err := json.Unmarshal(jsonData, &injectionPoint); err != nil {
				return nil, fmt.Errorf("failed to unmarshal injection point: %v", err)
			}

			result["injection_point"] = injectionPoint
		default:
			value, err := getIntValue(instanceValue.Field(i))
			if err != nil {
				return nil, err
			}

			result[key] = value
			if key == "direction" {
				result[key] = directionMap[int(value)]
			}
		}
	}

	return result, nil
}

func (ic *InjectionConf) GetGroundtruth() (Groundtruth, error) {
	instance, err := ic.getActiveInjection()
	if err != nil {
		return Groundtruth{}, err
	}

	// Check if the injection supports GetGroundtruth
	if provider, ok := instance.(GroundtruthProvider); ok {
		return provider.GetGroundtruth()
	}

	return Groundtruth{}, fmt.Errorf("injection does not support groundtruth calculation")
}

func getIntValue(field reflect.Value) (int64, error) {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int(), nil
	default:
		return 0, fmt.Errorf("unsupported field type: %v", field.Kind())
	}
}
