package handler

import (
	"fmt"

	"github.com/LGU-SE-Internal/chaos-experiment/internal/resourcelookup"
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
func GetGroundtruthFromAppIdx(namespace string, appIdx int) (Groundtruth, error) {
	appLabels, err := resourcelookup.GetAllAppLabels(namespace, TargetLabelKey)
	if err != nil || len(appLabels) == 0 {
		return Groundtruth{}, fmt.Errorf("failed to get app labels: %w", err)
	}

	if appIdx < 0 || appIdx >= len(appLabels) {
		return Groundtruth{}, fmt.Errorf("app index out of range: %d (max: %d)", appIdx, len(appLabels)-1)
	}

	appName := appLabels[appIdx]

	// Get containers and pods for the service
	containers, err := resourcelookup.GetContainersByService(namespace, appName)
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get containers: %w", err)
	}

	pods, err := resourcelookup.GetPodsByService(namespace, appName)
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
func GetGroundtruthFromContainerIdx(namespace string, containerIdx int) (Groundtruth, error) {
	containers, err := resourcelookup.GetAllContainers(namespace)
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
func GetGroundtruthFromDNSEndpointIdx(namespace string, endpointIdx int) (Groundtruth, error) {
	endpoints, err := resourcelookup.GetAllDNSEndpoints()
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
	containers, pods, err := resourcelookup.GetContainersAndPodsByServices(namespace, []string{sourceService, targetDomain})
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get containers and pods: %w", err)
	}

	// Create and populate the groundtruth
	gt := Groundtruth{
		Service:   []string{sourceService, targetDomain},
		Pod:       pods,
		Container: containers,
		Span:      []string{sourceService, targetDomain},
	}

	return gt, nil
}

// getHTTPGroundtruth is a helper function that gets groundtruth information for HTTP chaos
func getHTTPGroundtruth(namespace string, endpointIdx int) (Groundtruth, error) {
	endpoints, err := resourcelookup.GetAllHTTPEndpoints()
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
	containers, pods, err := resourcelookup.GetContainersAndPodsByServices(namespace, []string{sourceService, targetService})
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get containers and pods: %w", err)
	}

	// Create and populate the groundtruth
	gt := Groundtruth{
		Service:   []string{sourceService, targetService},
		Pod:       pods,
		Container: containers,
		Span:      []string{sourceService, targetService},
	}

	return gt, nil
}

// GetGroundtruthFromNetworkPairIdx returns a Groundtruth object for a given network pair index
func GetGroundtruthFromNetworkPairIdx(namespace string, networkPairIdx int) (Groundtruth, error) {
	networkPairs, err := resourcelookup.GetAllNetworkPairs()
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
	containers, pods, err := resourcelookup.GetContainersAndPodsByServices(namespace, []string{sourceService, targetService})
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get containers and pods: %w", err)
	}

	// Create and populate the groundtruth
	gt := Groundtruth{
		Service:   []string{sourceService, targetService},
		Pod:       pods,
		Container: containers,
		Span:      []string{sourceService, targetService},
	}

	return gt, nil
}

// GetGroundtruthFromMethodIdx returns a Groundtruth object for a given JVM method index
func GetGroundtruthFromMethodIdx(namespace string, methodIdx int) (Groundtruth, error) {
	methods, err := resourcelookup.GetAllJVMMethods()
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
	containers, err := resourcelookup.GetContainersByService(namespace, appName)
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get containers: %w", err)
	}

	pods, err := resourcelookup.GetPodsByService(namespace, appName)
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
func GetGroundtruthFromDatabaseIdx(namespace string, dbOpIdx int) (Groundtruth, error) {
	dbOps, err := resourcelookup.GetAllDatabaseOperations()
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get database operations: %w", err)
	}

	if dbOpIdx < 0 || dbOpIdx >= len(dbOps) {
		return Groundtruth{}, fmt.Errorf("database operation index out of range: %d (max: %d)", dbOpIdx, len(dbOps)-1)
	}

	dbOp := dbOps[dbOpIdx]
	appName := dbOp.AppName

	// Get containers and pods for the service
	containers, err := resourcelookup.GetContainersByService(namespace, appName)
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get containers: %w", err)
	}

	pods, err := resourcelookup.GetPodsByService(namespace, appName)
	if err != nil {
		return Groundtruth{}, fmt.Errorf("failed to get pods: %w", err)
	}

	// Try to get MySQL service information
	mysqlPods, err := resourcelookup.GetPodsByService(namespace, "mysql")
	if err != nil {
		// If error, just continue without MySQL pods
		mysqlPods = []string{}
	}

	mysqlContainers, err := resourcelookup.GetContainersByService(namespace, "mysql")
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

func (s *PodFailureSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return GetGroundtruthFromAppIdx(namespace, s.AppIdx)
}

func (s *PodKillSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return GetGroundtruthFromAppIdx(namespace, s.AppIdx)
}

func (s *ContainerKillSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return GetGroundtruthFromContainerIdx(namespace, s.ContainerIdx)
}

func (s *MemoryStressChaosSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	gt, err := GetGroundtruthFromContainerIdx(namespace, s.ContainerIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricMemory))
	return gt, nil
}

func (s *CPUStressChaosSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	gt, err := GetGroundtruthFromContainerIdx(namespace, s.ContainerIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricCPU))
	return gt, nil
}

func (s *TimeSkewSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return GetGroundtruthFromContainerIdx(namespace, s.ContainerIdx)
}

func (s *DNSErrorSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return GetGroundtruthFromDNSEndpointIdx(namespace, s.DNSEndpointIdx)
}

func (s *DNSRandomSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return GetGroundtruthFromDNSEndpointIdx(namespace, s.DNSEndpointIdx)
}

func (s *HTTPRequestAbortSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return getHTTPGroundtruth(namespace, s.EndpointIdx)
}

func (s *HTTPResponseAbortSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return getHTTPGroundtruth(namespace, s.EndpointIdx)
}

func (s *HTTPRequestDelaySpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	gt, err := getHTTPGroundtruth(namespace, s.EndpointIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricHTTPLatency))
	return gt, nil
}

func (s *HTTPResponseDelaySpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	gt, err := getHTTPGroundtruth(namespace, s.EndpointIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricHTTPLatency))
	return gt, nil
}

func (s *HTTPResponseReplaceBodySpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return getHTTPGroundtruth(namespace, s.EndpointIdx)
}

func (s *HTTPResponsePatchBodySpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return getHTTPGroundtruth(namespace, s.EndpointIdx)
}

func (s *HTTPRequestReplacePathSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return getHTTPGroundtruth(namespace, s.EndpointIdx)
}

func (s *HTTPRequestReplaceMethodSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return getHTTPGroundtruth(namespace, s.EndpointIdx)
}

func (s *HTTPResponseReplaceCodeSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return getHTTPGroundtruth(namespace, s.EndpointIdx)
}

func (s *NetworkDelaySpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	gt, err := GetGroundtruthFromNetworkPairIdx(namespace, s.NetworkPairIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricNetworkLatency))
	return gt, nil
}

func (s *NetworkLossSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return GetGroundtruthFromNetworkPairIdx(namespace, s.NetworkPairIdx)
}

func (s *NetworkDuplicateSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return GetGroundtruthFromNetworkPairIdx(namespace, s.NetworkPairIdx)
}

func (s *NetworkCorruptSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return GetGroundtruthFromNetworkPairIdx(namespace, s.NetworkPairIdx)
}

func (s *NetworkBandwidthSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return GetGroundtruthFromNetworkPairIdx(namespace, s.NetworkPairIdx)
}

func (s *NetworkPartitionSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return GetGroundtruthFromNetworkPairIdx(namespace, s.NetworkPairIdx)
}

// JVM chaos GetGroundtruth implementations
func (s *JVMLatencySpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	gt, err := GetGroundtruthFromMethodIdx(namespace, s.MethodIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricNetworkLatency))
	return gt, nil
}

func (s *JVMReturnSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return GetGroundtruthFromMethodIdx(namespace, s.MethodIdx)
}

func (s *JVMExceptionSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return GetGroundtruthFromMethodIdx(namespace, s.MethodIdx)
}

func (s *JVMGCSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return GetGroundtruthFromAppIdx(namespace, s.AppIdx)
}

func (s *JVMCPUStressSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	gt, err := GetGroundtruthFromMethodIdx(namespace, s.MethodIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricCPU))
	return gt, nil
}

func (s *JVMMemoryStressSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	gt, err := GetGroundtruthFromMethodIdx(namespace, s.MethodIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricMemory))
	return gt, nil
}

func (s *JVMMySQLLatencySpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	gt, err := GetGroundtruthFromDatabaseIdx(namespace, s.DatabaseIdx)
	if err != nil {
		return Groundtruth{}, err
	}

	gt.Metric = append(gt.Metric, string(MetricSQLLatency))
	return gt, nil
}

func (s *JVMMySQLExceptionSpec) GetGroundtruth() (Groundtruth, error) {
	namespace := fmt.Sprintf("%s%d", NamespacePrefixs[s.Namespace], DefaultStartIndex)
	return GetGroundtruthFromDatabaseIdx(namespace, s.DatabaseIdx)
}
