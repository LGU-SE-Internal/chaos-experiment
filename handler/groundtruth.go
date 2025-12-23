package handler

import (
	"context"
	"fmt"

	"github.com/LGU-SE-Internal/chaos-experiment/internal/resourcelookup"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"
)

// MetricType defines the type of metrics for groundtruth
type MetricType string

const (
	MetricCPU            MetricType = "cpu"
	MetricMemory         MetricType = "memory"
	MetricDisk           MetricType = "disk"
	MetricNetworkLatency MetricType = "network_latency"
	MetricHTTPLatency    MetricType = "http_latency"
	MetricSQLLatency     MetricType = "sql_latency"
)

// Groundtruth represents the expected impact of a chaos experiment
type Groundtruth struct {
	Service   []string `json:"service,omitempty"`
	Pod       []string `json:"pod,omitempty"`
	Container []string `json:"container,omitempty"`
	Metric    []string `json:"metric,omitempty"`
	Function  []string `json:"function,omitempty"`
	Span      []string `json:"span,omitempty"`
}

// GetGroundtruthFromAppIdx returns a Groundtruth object for a given app index
func GetGroundtruthFromAppIdx(ctx context.Context, system systemconfig.SystemType, namespace string, appIdx int) (Groundtruth, error) {
	systemCache := resourcelookup.GetSystemCache(system)

	appLabels, err := systemCache.GetAllAppLabels(ctx, namespace, defaultAppLabel)
	if err != nil || len(appLabels) == 0 {
		return Groundtruth{}, fmt.Errorf("failed to get app labels: %w", err)
	}

	if appIdx < 0 || appIdx >= len(appLabels) {
		return Groundtruth{}, fmt.Errorf("app index out of range: %d (max: %d)", appIdx, len(appLabels)-1)
	}

	appName := appLabels[appIdx]

	// Get containers and pods for the service
	containers, err := systemCache.GetContainersByService(ctx, namespace, appName)
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get containers: %w", err)
	}

	pods, err := systemCache.GetPodsByService(ctx, namespace, defaultAppLabel)
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get pods: %w", err)
	}

	// Create and populate the groundtruth
	gt := Groundtruth{
		Service:   []string{appName},
		Pod:       pods,
		Container: containers,
	}

	return gt, nil
}

// GetGroundtruthFromContainerIdx returns a Groundtruth object for a given container index
func GetGroundtruthFromContainerIdx(ctx context.Context, system systemconfig.SystemType, namespace string, containerIdx int) (Groundtruth, error) {
	systemCache := resourcelookup.GetSystemCache(system)

	containers, err := systemCache.GetAllContainers(ctx, namespace)
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get containers: %w", err)
	}

	if containerIdx < 0 || containerIdx >= len(containers) {
		return Groundtruth{}, fmt.Errorf("container index out of range: %d (max: %d)", containerIdx, len(containers)-1)
	}

	containerInfo := containers[containerIdx]

	// Create and populate the groundtruth
	gt := Groundtruth{
		Service:   []string{containerInfo.AppLabel},
		Pod:       []string{containerInfo.PodName},
		Container: []string{containerInfo.ContainerName},
	}

	return gt, nil
}

// GetGroundtruthFromDNSEndpointIdx returns a Groundtruth object for a given DNS endpoint index
func GetGroundtruthFromDNSEndpointIdx(ctx context.Context, system systemconfig.SystemType, namespace string, endpointIdx int) (Groundtruth, error) {
	systemCache := resourcelookup.GetSystemCache(system)

	endpoints, err := systemCache.GetAllDNSEndpoints()
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get DNS endpoints: %w", err)
	}

	if endpointIdx < 0 || endpointIdx >= len(endpoints) {
		return Groundtruth{}, fmt.Errorf("endpoint index out of range: %d (max: %d)", endpointIdx, len(endpoints)-1)
	}

	// Get the source and target services
	endpointPair := endpoints[endpointIdx]
	sourceService := endpointPair.AppName
	targetDomain := endpointPair.Domain

	// Get containers and pods for both services
	containers, pods, err := systemCache.GetContainersAndPodsByServices(ctx, namespace, []string{sourceService, targetDomain})
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get containers and pods: %w", err)
	}

	// For DNS chaos, use all span names between the source service and target domain
	// If no span names available, fall back to service names
	var spanNames []string
	if len(endpointPair.SpanNames) > 0 {
		spanNames = endpointPair.SpanNames
	} else {
		spanNames = []string{sourceService, targetDomain}
	}

	// Create and populate the groundtruth
	gt := Groundtruth{
		Service:   []string{sourceService, targetDomain},
		Pod:       pods,
		Container: containers,
		Span:      spanNames,
	}

	return gt, nil
}

// getHTTPGroundtruth is a helper function that gets groundtruth information for HTTP chaos
func getHTTPGroundtruth(ctx context.Context, system systemconfig.SystemType, namespace string, endpointIdx int) (Groundtruth, error) {
	systemCache := resourcelookup.GetSystemCache(system)

	endpoints, err := systemCache.GetAllHTTPEndpoints()
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get HTTP endpoints: %w", err)
	}

	if endpointIdx < 0 || endpointIdx >= len(endpoints) {
		return Groundtruth{}, fmt.Errorf("endpoint index out of range: %d (max: %d)", endpointIdx, len(endpoints)-1)
	}

	// Get the source and target services
	endpointPair := endpoints[endpointIdx]
	sourceService := endpointPair.AppName
	targetService := endpointPair.ServerAddress

	// Get containers and pods for both services
	containers, pods, err := systemCache.GetContainersAndPodsByServices(ctx, namespace, []string{sourceService, targetService})
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get containers and pods: %w", err)
	}

	// Use the actual span name if available, otherwise use service names as fallback
	var spanNames []string
	if endpointPair.SpanName != "" {
		spanNames = []string{endpointPair.SpanName}
	} else {
		spanNames = []string{sourceService, targetService}
	}

	// Create and populate the groundtruth
	gt := Groundtruth{
		Service:   []string{sourceService, targetService},
		Pod:       pods,
		Container: containers,
		Span:      spanNames,
	}

	return gt, nil
}

// GetGroundtruthFromNetworkPairIdx returns a Groundtruth object for a given network pair index
func GetGroundtruthFromNetworkPairIdx(ctx context.Context, system systemconfig.SystemType, namespace string, networkPairIdx int) (Groundtruth, error) {
	systemCache := resourcelookup.GetSystemCache(system)

	networkPairs, err := systemCache.GetAllNetworkPairs()
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get network pairs: %w", err)
	}

	if networkPairIdx < 0 || networkPairIdx >= len(networkPairs) {
		return Groundtruth{}, fmt.Errorf("network pair index out of range: %d (max: %d)", networkPairIdx, len(networkPairs)-1)
	}

	// Get the source and target services
	pair := networkPairs[networkPairIdx]
	sourceService := pair.SourceService
	targetService := pair.TargetService

	// Get containers and pods for both services
	containers, pods, err := systemCache.GetContainersAndPodsByServices(ctx, namespace, []string{sourceService, targetService})
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get containers and pods: %w", err)
	}

	// For network faults, use all span names between the two services
	// If no span names available, fall back to service names
	var spanNames []string
	if len(pair.SpanNames) > 0 {
		spanNames = pair.SpanNames
	} else {
		spanNames = []string{sourceService, targetService}
	}

	// Create and populate the groundtruth
	gt := Groundtruth{
		Service:   []string{sourceService, targetService},
		Pod:       pods,
		Container: containers,
		Span:      spanNames,
	}

	return gt, nil
}

// GetGroundtruthFromMethodIdx returns a Groundtruth object for a given JVM method index
func GetGroundtruthFromMethodIdx(ctx context.Context, system systemconfig.SystemType, namespace string, methodIdx int) (Groundtruth, error) {
	systemCache := resourcelookup.GetSystemCache(system)

	methods, err := systemCache.GetAllJVMMethods()
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get JVM methods: %w", err)
	}

	if methodIdx < 0 || methodIdx >= len(methods) {
		return Groundtruth{}, fmt.Errorf("method index out of range: %d (max: %d)", methodIdx, len(methods)-1)
	}

	methodPair := methods[methodIdx]
	appName := methodPair.AppName

	// Format function identifier as className.methodName
	className := methodPair.ClassName

	functionName := fmt.Sprintf("%s.%s", className, methodPair.MethodName)

	// Get containers and pods for the service
	containers, err := systemCache.GetContainersByService(ctx, namespace, appName)
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get containers: %w", err)
	}

	pods, err := systemCache.GetPodsByService(ctx, namespace, appName)
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get pods: %w", err)
	}

	// Create and populate the groundtruth
	gt := Groundtruth{
		Service:   []string{appName},
		Pod:       pods,
		Container: containers,
		Function:  []string{functionName},
	}

	return gt, nil
}

// GetGroundtruthFromDatabaseIdx returns a Groundtruth object for a given database operation index
func GetGroundtruthFromDatabaseIdx(ctx context.Context, system systemconfig.SystemType, namespace string, dbOpIdx int) (Groundtruth, error) {
	systemCache := resourcelookup.GetSystemCache(system)

	dbOps, err := systemCache.GetAllDatabaseOperations()
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get database operations: %w", err)
	}

	if dbOpIdx < 0 || dbOpIdx >= len(dbOps) {
		return Groundtruth{}, fmt.Errorf("database operation index out of range: %d (max: %d)", dbOpIdx, len(dbOps)-1)
	}

	dbOp := dbOps[dbOpIdx]
	appName := dbOp.AppName

	// Get containers and pods for the service
	containers, err := systemCache.GetContainersByService(ctx, namespace, appName)
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get containers: %w", err)
	}

	pods, err := systemCache.GetPodsByService(ctx, namespace, appName)
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get pods: %w", err)
	}

	// Try to get MySQL service information
	mysqlPods, err := systemCache.GetPodsByService(ctx, namespace, "mysql")
	if err != nil {
		// If error, just continue without MySQL pods
		mysqlPods = []string{}
	}

	mysqlContainers, err := systemCache.GetContainersByService(ctx, namespace, "mysql")
	if err != nil {
		// If error, just continue without MySQL containers
		mysqlContainers = []string{}
	}

	// Combine service and MySQL pods/containers
	allPods := append(pods, mysqlPods...)
	allContainers := append(containers, mysqlContainers...)

	// Create and populate the groundtruth - removed Function field as requested
	gt := Groundtruth{
		Service:   []string{appName, "mysql"},
		Pod:       allPods,
		Container: allContainers,
		Span:      []string{appName, "mysql"}, // Include span information for tracking
	}

	return gt, nil
}

func (s *PodFailureSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return GetGroundtruthFromAppIdx(ctx, system, namespace, s.AppIdx)
}

func (s *PodKillSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return GetGroundtruthFromAppIdx(ctx, system, namespace, s.AppIdx)
}

func (s *ContainerKillSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return GetGroundtruthFromContainerIdx(ctx, system, namespace, s.ContainerIdx)
}

func (s *MemoryStressChaosSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	gt, err := GetGroundtruthFromContainerIdx(ctx, system, namespace, s.ContainerIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricMemory))
	return gt, nil
}

func (s *CPUStressChaosSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	gt, err := GetGroundtruthFromContainerIdx(ctx, system, namespace, s.ContainerIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricCPU))
	return gt, nil
}

func (s *TimeSkewSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return GetGroundtruthFromContainerIdx(ctx, system, namespace, s.ContainerIdx)
}

func (s *DNSErrorSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return GetGroundtruthFromDNSEndpointIdx(ctx, system, namespace, s.DNSEndpointIdx)
}

func (s *DNSRandomSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return GetGroundtruthFromDNSEndpointIdx(ctx, system, namespace, s.DNSEndpointIdx)
}

func (s *HTTPRequestAbortSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return getHTTPGroundtruth(ctx, system, namespace, s.EndpointIdx)
}

func (s *HTTPResponseAbortSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return getHTTPGroundtruth(ctx, system, namespace, s.EndpointIdx)
}

func (s *HTTPRequestDelaySpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	gt, err := getHTTPGroundtruth(ctx, system, namespace, s.EndpointIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricHTTPLatency))
	return gt, nil
}

func (s *HTTPResponseDelaySpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	gt, err := getHTTPGroundtruth(ctx, system, namespace, s.EndpointIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricHTTPLatency))
	return gt, nil
}

func (s *HTTPResponseReplaceBodySpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return getHTTPGroundtruth(ctx, system, namespace, s.EndpointIdx)
}

func (s *HTTPResponsePatchBodySpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return getHTTPGroundtruth(ctx, system, namespace, s.EndpointIdx)
}

func (s *HTTPRequestReplacePathSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return getHTTPGroundtruth(ctx, system, namespace, s.EndpointIdx)
}

func (s *HTTPRequestReplaceMethodSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return getHTTPGroundtruth(ctx, system, namespace, s.EndpointIdx)
}

func (s *HTTPResponseReplaceCodeSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return getHTTPGroundtruth(ctx, system, namespace, s.EndpointIdx)
}

func (s *NetworkDelaySpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	gt, err := GetGroundtruthFromNetworkPairIdx(ctx, system, namespace, s.NetworkPairIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricNetworkLatency))
	return gt, nil
}

func (s *NetworkLossSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return GetGroundtruthFromNetworkPairIdx(ctx, system, namespace, s.NetworkPairIdx)
}

func (s *NetworkDuplicateSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return GetGroundtruthFromNetworkPairIdx(ctx, system, namespace, s.NetworkPairIdx)
}

func (s *NetworkCorruptSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return GetGroundtruthFromNetworkPairIdx(ctx, system, namespace, s.NetworkPairIdx)
}

func (s *NetworkBandwidthSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return GetGroundtruthFromNetworkPairIdx(ctx, system, namespace, s.NetworkPairIdx)
}

func (s *NetworkPartitionSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return GetGroundtruthFromNetworkPairIdx(ctx, system, namespace, s.NetworkPairIdx)
}

// JVM chaos GetGroundtruth implementations
func (s *JVMLatencySpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	gt, err := GetGroundtruthFromMethodIdx(ctx, system, namespace, s.MethodIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricNetworkLatency))
	return gt, nil
}

func (s *JVMReturnSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return GetGroundtruthFromMethodIdx(ctx, system, namespace, s.MethodIdx)
}

func (s *JVMExceptionSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return GetGroundtruthFromMethodIdx(ctx, system, namespace, s.MethodIdx)
}

func (s *JVMGCSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return GetGroundtruthFromAppIdx(ctx, system, namespace, s.AppIdx)
}

func (s *JVMCPUStressSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	gt, err := GetGroundtruthFromMethodIdx(ctx, system, namespace, s.MethodIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricCPU))
	return gt, nil
}

func (s *JVMMemoryStressSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	gt, err := GetGroundtruthFromMethodIdx(ctx, system, namespace, s.MethodIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricMemory))
	return gt, nil
}

func (s *JVMMySQLLatencySpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	gt, err := GetGroundtruthFromDatabaseIdx(ctx, system, namespace, s.DatabaseIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricSQLLatency))
	return gt, nil
}

func (s *JVMMySQLExceptionSpec) GetGroundtruth(ctx context.Context) (Groundtruth, error) {
	system := systemconfig.GetAllSystemTypes()[s.System]
	namespace, err := systemconfig.GetNamespaceByIndex(system, defaultStartIndex)
	if err != nil {
		return Groundtruth{}, err
	}
	return GetGroundtruthFromDatabaseIdx(ctx, system, namespace, s.DatabaseIdx)
}
