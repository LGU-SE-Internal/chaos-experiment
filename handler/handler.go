package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/LGU-SE-Internal/chaos-experiment/client"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/resourcelookup"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"
	"github.com/LGU-SE-Internal/chaos-experiment/utils"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
)

type SystemType systemconfig.SystemType

const (
	SystemTrainTicket        = SystemType(systemconfig.SystemTrainTicket)
	SystemOtelDemo           = SystemType(systemconfig.SystemOtelDemo)
	SystemMediaMicroservices = SystemType(systemconfig.SystemMediaMicroservices)
	SystemHotelReservation   = SystemType(systemconfig.SystemHotelReservation)
	SystemSocialNetwork      = SystemType(systemconfig.SystemSocialNetwork)
	SystemOnlineBoutique     = SystemType(systemconfig.SystemOnlineBoutique)
)

// String returns the string representation of SystemType
func (s SystemType) String() string {
	return systemconfig.SystemType(s).String()
}

// IsValid checks if the SystemType is valid
func (s SystemType) IsValid() bool {
	_, err := systemconfig.ParseSystemType(s.String())
	return err == nil
}

// GetAllSystemTypes returns all valid system types
func GetAllSystemTypes() []SystemType {
	systems := systemconfig.GetAllSystemTypes()
	result := make([]SystemType, len(systems))
	for i, sys := range systems {
		result[i] = SystemType(sys)
	}
	return result
}

type ChaosType int

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

// GetChaosTypeName returns the name of the given ChaosType
func GetChaosTypeName(c ChaosType) string {
	if name, ok := ChaosTypeMap[c]; ok {
		return name
	}
	return "Unknown"
}

type Conf struct {
	Annotations map[string]string
	Context     context.Context
	Labels      map[string]string
	Namespace   string
	System      systemconfig.SystemType
}
type Option func(*Conf)

func WithAnnotations(annotations map[string]string) Option {
	return func(c *Conf) {
		c.Annotations = annotations
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

func WithSystem(system systemconfig.SystemType) Option {
	return func(c *Conf) {
		c.System = system
	}
}

type Injection interface {
	Create(cli cli.Client, opt ...Option) (string, error)
}
type GroundtruthProvider interface {
	GetGroundtruth(ctx context.Context) (Groundtruth, error)
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

func (ic *InjectionConf) GetDisplayConfig(ctx context.Context) (map[string]any, error) {
	instance, err := ic.getActiveInjection()
	if err != nil {
		return nil, err
	}

	parsed, err := parseInjection(ctx, instance)
	if err != nil {
		return nil, err
	}

	return parsed, nil
}

func (ic *InjectionConf) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	instance, err := ic.getActiveInjection()
	if err != nil {
		return Groundtruth{}, err
	}

	// Check if the injection supports GetGroundtruth
	if provider, ok := instance.(GroundtruthProvider); ok {
		return provider.GetGroundtruth(ctx)
	}

	return Groundtruth{}, fmt.Errorf("injection does not support groundtruth calculation")
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

func BatchCreate(ctx context.Context, confs []InjectionConf, system SystemType, namespace string, annotations, labels map[string]string) ([]string, error) {
	if len(confs) == 0 {
		return nil, fmt.Errorf("no injection configurations provided")
	}

	// Validate system type
	if !system.IsValid() {
		return nil, fmt.Errorf("invalid system type: %s", system)
	}

	systemconfig.SetCurrentSystem(systemconfig.SystemType(system))

	type result struct {
		name  string
		err   error
		index int
	}

	resultChan := make(chan result, len(confs))
	for idx, conf := range confs {
		go func(ic InjectionConf, i int) {
			injection, err := ic.getActiveInjection()
			if err != nil {
				resultChan <- result{err: fmt.Errorf("failed to get active injection for conf at index %d: %w", i, err), index: i}
			}

			name, err := injection.Create(
				client.GetK8sClient(),
				WithAnnotations(annotations),
				WithContext(ctx),
				WithLabels(labels),
				WithNs(namespace),
				WithSystem(systemconfig.GetCurrentSystem()),
			)
			if err != nil {
				resultChan <- result{err: fmt.Errorf("failed to inject chaos for %v: %w", injection, err), index: i}
			} else {
				resultChan <- result{name: name, index: i}
			}
		}(conf, idx)
	}

	results := make([]result, len(confs))
	for i := 0; i < len(confs); i++ {
		res := <-resultChan
		results[res.index] = res
	}

	names := make([]string, 0, len(confs))
	var errs []error

	for _, res := range results {
		if res.err != nil {
			errs = append(errs, res.err)
		} else {
			names = append(names, res.name)
		}
	}

	if len(errs) > 0 {
		return names, fmt.Errorf("some injections failed: %v", errs)
	}

	return names, nil
}

func parseInjection(ctx context.Context, instance Injection) (map[string]any, error) {
	instanceValue := reflect.ValueOf(instance).Elem()
	instanceType := instanceValue.Type()

	result := make(map[string]any, instanceValue.NumField())

	var system systemconfig.SystemType
	var endpointMethod string

	systems := systemconfig.GetAllSystemTypes()
	for i := range instanceValue.NumField() {
		if instanceType.Field(i).Name == keySystem {
			index, err := getIntValue(instanceValue.Field(i))
			if err != nil {
				return nil, err
			}

			if index >= 0 && int(index) < len(systems) {
				system = systems[index]
				break
			}
		}
	}

	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace for system %s: %v", system, err)
	}

	systemCache := resourcelookup.GetSystemCache(system)

	for i := range instanceValue.NumField() {
		key := utils.ToSnakeCase(instanceType.Field(i).Name)

		index, err := getIntValue(instanceValue.Field(i))
		if err != nil {
			return nil, err
		}

		var value any
		switch i {
		case 1:
			result[key] = system.String()
		case 2:
			switch instanceType.Field(i).Name {
			case keyApp:
				labels, err := systemCache.GetAllAppLabels(ctx, namespace, defaultAppLabel)
				if err != nil || len(labels) == 0 {
					return nil, err
				}

				value = map[string]any{"app_name": labels[index]}
			case keyMethod:
				methods, err := systemCache.GetAllJVMMethods()
				if err != nil {
					return nil, err
				}

				value = methods[index]
			case keyEndpoint:
				endpoints, err := systemCache.GetAllHTTPEndpoints()
				if err != nil {
					return nil, err
				}

				value = endpoints[index]

				endpointMethod = endpoints[index].Method

			case keyNetworkPair:
				networkpairs, err := systemCache.GetAllNetworkPairs()
				if err != nil {
					return nil, err
				}

				value = networkpairs[index]
			case keyContainer:
				containers, err := systemCache.GetAllContainers(ctx, namespace)
				if err != nil {
					return nil, err
				}

				value = containers[index]
			case keyDNSEndpoint:
				endpoints, err := systemCache.GetAllDNSEndpoints()
				if err != nil {
					return nil, err
				}

				value = endpoints[index]
			case keyDatabase:
				operations, err := systemCache.GetAllDatabaseOperations()
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
			} else if key == "replace_method" && endpointMethod != "" {
				// Get the actual HTTP method name for replace_method
				filteredMethod := GetFilteredHTTPMethodByIndex(endpointMethod, int(value))
				result[key] = GetHTTPMethodName(filteredMethod)
			}
		}
	}

	return result, nil
}

func getIntValue(field reflect.Value) (int64, error) {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int(), nil
	default:
		return 0, fmt.Errorf("unsupported field type: %v", field.Kind())
	}
}

type Pair struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type SystemResource struct {
	AppLabels        []string `json:"app_labels"`
	JVMAppNames      []string `json:"jvm_app_names"`
	HTTPAppNames     []string `json:"http_app_names"`
	NetworkPairs     []Pair   `json:"network_pairs"`
	DNSAppNames      []string `json:"dns_app_names"`
	DatabaseAppNames []string `json:"database_app_names"`
	ContainerNames   []string `json:"container_names"`
}

func (sr *SystemResource) ToMap() map[string][]string {
	result := make(map[string][]string)

	result["app_labels"] = sr.AppLabels
	result["jvm_app_names"] = sr.JVMAppNames
	result["http_app_names"] = sr.HTTPAppNames
	result["dns_app_names"] = sr.DNSAppNames
	result["database_app_names"] = sr.DatabaseAppNames
	result["container_names"] = sr.ContainerNames

	if len(sr.NetworkPairs) > 0 {
		var networkPairStrings []string
		for _, pair := range sr.NetworkPairs {
			networkPairStrings = append(networkPairStrings, fmt.Sprintf("%s->%s", pair.Source, pair.Target))
		}

		result["network_pairs"] = networkPairStrings
	}

	return result
}

func (sr *SystemResource) ToDeduplicatedMap() map[string][]string {
	result := make(map[string][]string)
	for key, value := range sr.ToMap() {
		result[key] = utils.RemoveDuplicates(value)
	}

	return result
}

// GetSystemResourceMap retrieves system resources for all systems
func GetSystemResourceMap(ctx context.Context) (map[SystemType]SystemResource, error) {
	resourceMap := make(map[SystemType]SystemResource, len(systemconfig.GetAllSystemTypes()))
	for _, system := range systemconfig.GetAllSystemTypes() {
		systemCache := resourcelookup.GetSystemCache(system)
		namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to get namespace prefix for system %s: %v", system, err)
		}

		appLabels, err := systemCache.GetAllAppLabels(ctx, namespace, defaultAppLabel)
		if err != nil {
			return nil, fmt.Errorf("failed to get app labels namespace %s for system %s: %v", namespace, system, err)
		}

		methods, err := systemCache.GetAllJVMMethods()
		if err != nil {
			return nil, fmt.Errorf("failed to get JVM methods: %v", err)
		}

		jvmAppNames := make([]string, 0, len(methods))
		for _, method := range methods {
			jvmAppNames = append(jvmAppNames, method.AppName)
		}

		endpoints, err := systemCache.GetAllHTTPEndpoints()
		if err != nil {
			return nil, fmt.Errorf("failed to get HTTP endpoints: %v", err)
		}

		httpAppNames := make([]string, 0, len(endpoints))
		for _, endpoint := range endpoints {
			httpAppNames = append(httpAppNames, endpoint.AppName)
		}

		pairs, err := systemCache.GetAllNetworkPairs()
		if err != nil {
			return nil, fmt.Errorf("failed to get network pairs: %v", err)
		}

		networkPairs := make([]Pair, 0, len(pairs))
		for _, pair := range pairs {
			networkPairs = append(networkPairs, Pair{
				Source: pair.SourceService,
				Target: pair.TargetService,
			})
		}

		dnsEndpoints, err := systemCache.GetAllDNSEndpoints()
		if err != nil {
			return nil, fmt.Errorf("failed to get DNS endpoints: %v", err)
		}

		dnsAppNames := make([]string, 0, len(dnsEndpoints))
		for _, endpoint := range dnsEndpoints {
			dnsAppNames = append(dnsAppNames, endpoint.AppName)
		}

		operations, err := systemCache.GetAllDatabaseOperations()
		if err != nil {
			return nil, fmt.Errorf("failed to get database operations: %v", err)
		}

		databaseAppNames := make([]string, 0, len(operations))
		for _, operation := range operations {
			databaseAppNames = append(databaseAppNames, operation.AppName)
		}

		containers, err := systemCache.GetAllContainers(ctx, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get containers of namespace %s for system %s: %v", namespace, system, err)
		}

		containerNames := make([]string, 0, len(containers))
		for _, container := range containers {
			containerNames = append(containerNames, container.AppLabel)
		}

		resourceMap[SystemType(system)] = SystemResource{
			AppLabels:        appLabels,
			JVMAppNames:      jvmAppNames,
			HTTPAppNames:     httpAppNames,
			NetworkPairs:     networkPairs,
			DNSAppNames:      dnsAppNames,
			DatabaseAppNames: databaseAppNames,
			ContainerNames:   containerNames,
		}
	}

	return resourceMap, nil
}

type ChaosResourceMapping struct {
	IndexFieldName string `json:"index_field_name"`
	ResourceType   string `json:"resource_type"`
}

var keyResourceMap = map[string]string{
	keyApp:         "app_labels",
	keyMethod:      "jvm_app_names",
	keyEndpoint:    "http_app_names",
	keyNetworkPair: "network_pairs",
	keyDNSEndpoint: "dns_app_names",
	keyDatabase:    "database_app_names",
	keyContainer:   "container_names",
}

func GetChaosTypeResourceMappings() (map[string]ChaosResourceMapping, error) {
	injectionConfType := reflect.TypeFor[InjectionConf]()

	result := make(map[string]ChaosResourceMapping, injectionConfType.NumField())
	for i := 0; i < injectionConfType.NumField(); i++ {
		field := injectionConfType.Field(i)
		fieldName := field.Name
		fieldType := field.Type

		if fieldType.Kind() == reflect.Ptr {
			elemType := fieldType.Elem()
			if elemType.NumField() >= 3 {
				resourceIndexName := elemType.Field(2).Name
				resourceKey, exists := keyResourceMap[resourceIndexName]
				if !exists {
					return nil, fmt.Errorf("unknown resource index name: %s", resourceIndexName)
				}

				result[fieldName] = ChaosResourceMapping{
					IndexFieldName: resourceIndexName,
					ResourceType:   resourceKey,
				}
			} else {
				return nil, fmt.Errorf("field %s does not have enough fields", fieldName)
			}
		}
	}

	return result, nil
}
