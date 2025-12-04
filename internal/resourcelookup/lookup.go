package resourcelookup

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/LGU-SE-Internal/chaos-experiment/client"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/databaseoperations"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/grpcoperations"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/javaclassmethods"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/networkdependencies"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/serviceendpoints"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"
	"github.com/LGU-SE-Internal/chaos-experiment/utils"
	"github.com/sirupsen/logrus"
)

// AppMethodPair represents a flattened app+method combination
type AppMethodPair struct {
	AppName    string `json:"app_name"`
	ClassName  string `json:"class_name"`
	MethodName string `json:"method_name"`
}

// AppEndpointPair represents a flattened app+endpoint combination
type AppEndpointPair struct {
	AppName       string `json:"app_name"`
	Route         string `json:"route"`
	Method        string `json:"method"`
	ServerAddress string `json:"server_address"`
	ServerPort    string `json:"server_port"`
	SpanName      string `json:"span_name"`
}

// AppNetworkPair represents a flattened source+target combination for network chaos
type AppNetworkPair struct {
	SourceService string   `json:"source_service"`
	TargetService string   `json:"target_service"`
	SpanNames     []string `json:"span_names"` // All span names between source and target services
}

// AppDNSPair represents a flattened app+domain combination for DNS chaos
type AppDNSPair struct {
	AppName   string   `json:"app_name"`
	Domain    string   `json:"domain"`
	SpanNames []string `json:"span_names"` // All span names for endpoints targeting this domain
}

// AppDatabasePair represents a flattened app+database+table+operation combination
type AppDatabasePair struct {
	AppName       string `json:"app_name"`
	DBName        string `json:"db_name"`
	TableName     string `json:"table_name"`
	OperationType string `json:"operation_type"`
}

// ContainerInfo represents container information with its pod and app
type ContainerInfo struct {
	PodName       string `json:"pod_name"`
	AppLabel      string `json:"app_label"`
	ContainerName string `json:"container_name"`
}

// Global cache for lookups - now system-aware
var (
	cachedAppLabels     map[string][]string
	cachedAppMethods    map[systemconfig.SystemType][]AppMethodPair
	cachedAppEndpoints  map[systemconfig.SystemType][]AppEndpointPair
	cachedNetworkPairs  map[systemconfig.SystemType][]AppNetworkPair
	cachedDNSEndpoints  map[systemconfig.SystemType][]AppDNSPair
	cachedContainerInfo map[string][]ContainerInfo
	cachedDBOperations  map[systemconfig.SystemType][]AppDatabasePair
)

// GetAllAppLabels returns all application labels sorted alphabetically
func GetAllAppLabels(namespace string, key string) ([]string, error) {
	prefix, err := utils.ExtractNsPrefix(namespace)
	if err != nil {
		return nil, err
	}

	if labels, exists := cachedAppLabels[prefix]; exists && len(labels) > 0 {
		return labels, nil
	}

	labels, err := client.GetLabels(context.Background(), namespace, key)
	logrus.Debugf("Fetched labels for namespace %s with key %s: %v", namespace, key, labels)
	if err != nil {
		return nil, err
	}

	// Sort alphabetically
	sort.Strings(labels)
	cachedAppLabels[prefix] = labels
	return labels, nil
}

// GetAllJVMMethods returns all app+method pairs sorted by app name
// This function uses the current system from systemconfig
func GetAllJVMMethods() ([]AppMethodPair, error) {
	currentSystem := systemconfig.GetCurrentSystem()
	
	if cachedAppMethods != nil {
		if result, exists := cachedAppMethods[currentSystem]; exists {
			return result, nil
		}
	}

	// Get all service names first
	services := javaclassmethods.ListAllServiceNames()
	result := make([]AppMethodPair, 0)

	// For each service, get its methods
	for _, serviceName := range services {
		methods := javaclassmethods.GetClassMethodsByService(serviceName)
		for _, method := range methods {
			result = append(result, AppMethodPair{
				AppName:    serviceName,
				ClassName:  method.ClassName,
				MethodName: method.MethodName,
			})
		}
	}

	// Sort by app name for consistency
	sort.Slice(result, func(i, j int) bool {
		if result[i].AppName != result[j].AppName {
			return result[i].AppName < result[j].AppName
		}
		if result[i].ClassName != result[j].ClassName {
			return result[i].ClassName < result[j].ClassName
		}
		return result[i].MethodName < result[j].MethodName
	})

	if cachedAppMethods == nil {
		cachedAppMethods = make(map[systemconfig.SystemType][]AppMethodPair)
	}
	cachedAppMethods[currentSystem] = result
	return result, nil
}

// GetAllHTTPEndpoints returns all app+endpoint pairs sorted by app name
// This function uses the current system from systemconfig
func GetAllHTTPEndpoints() ([]AppEndpointPair, error) {
	currentSystem := systemconfig.GetCurrentSystem()
	
	if cachedAppEndpoints != nil {
		if result, exists := cachedAppEndpoints[currentSystem]; exists {
			return result, nil
		}
	}

	// Get all service names
	services := serviceendpoints.GetAllServices()
	result := make([]AppEndpointPair, 0)

	// For each service, get its endpoints
	for _, serviceName := range services {
		endpoints := serviceendpoints.GetEndpointsByService(serviceName)
		for _, endpoint := range endpoints {
			// Skip non-HTTP endpoints like rabbitmq
			if endpoint.ServerAddress == "ts-rabbitmq" {
				continue
			}

			// Only include endpoints with a valid route
			if endpoint.Route != "" {
				result = append(result, AppEndpointPair{
					AppName:       serviceName,
					Route:         endpoint.Route,
					Method:        endpoint.RequestMethod,
					ServerAddress: endpoint.ServerAddress,
					ServerPort:    endpoint.ServerPort,
					SpanName:      endpoint.SpanName,
				})
			}
		}
	}

	// Sort by app name for consistency
	sort.Slice(result, func(i, j int) bool {
		if result[i].AppName != result[j].AppName {
			return result[i].AppName < result[j].AppName
		}
		return result[i].Route < result[j].Route
	})

	if cachedAppEndpoints == nil {
		cachedAppEndpoints = make(map[systemconfig.SystemType][]AppEndpointPair)
	}
	cachedAppEndpoints[currentSystem] = result
	return result, nil
}

// GetAllNetworkPairs returns all network pairs sorted by source service
// This function uses the current system from systemconfig
func GetAllNetworkPairs() ([]AppNetworkPair, error) {
	currentSystem := systemconfig.GetCurrentSystem()
	
	if cachedNetworkPairs != nil {
		if result, exists := cachedNetworkPairs[currentSystem]; exists {
			return result, nil
		}
	}

	// Get all service-to-service pairs
	pairs := networkdependencies.GetAllServicePairs()
	result := make([]AppNetworkPair, 0, len(pairs))

	for _, pair := range pairs {
		// Get all span names between source and target services
		spanNames := getSpanNamesBetweenServices(pair.SourceService, pair.TargetService)
		result = append(result, AppNetworkPair{
			SourceService: pair.SourceService,
			TargetService: pair.TargetService,
			SpanNames:     spanNames,
		})
	}

	// Sort by source service for consistency
	sort.Slice(result, func(i, j int) bool {
		if result[i].SourceService != result[j].SourceService {
			return result[i].SourceService < result[j].SourceService
		}
		return result[i].TargetService < result[j].TargetService
	})

	if cachedNetworkPairs == nil {
		cachedNetworkPairs = make(map[systemconfig.SystemType][]AppNetworkPair)
	}
	cachedNetworkPairs[currentSystem] = result
	return result, nil
}

// getSpanNamesBetweenServices returns all unique span names for endpoints between two services
func getSpanNamesBetweenServices(sourceService, targetService string) []string {
	endpoints := serviceendpoints.GetEndpointsByService(sourceService)
	spanNameSet := make(map[string]bool)

	for _, endpoint := range endpoints {
		// Check if this endpoint targets the target service
		if endpoint.ServerAddress == targetService && endpoint.SpanName != "" {
			spanNameSet[endpoint.SpanName] = true
		}
	}

	// Convert set to sorted slice
	spanNames := make([]string, 0, len(spanNameSet))
	for spanName := range spanNameSet {
		spanNames = append(spanNames, spanName)
	}
	sort.Strings(spanNames)
	return spanNames
}

// GetAllDNSEndpoints returns all app+domain pairs for DNS chaos sorted by app name
// This function uses the current system from systemconfig
// Note: DNS chaos does NOT work for gRPC-only connections, so we filter those out
// We use grpcoperations data to identify gRPC-only service pairs
func GetAllDNSEndpoints() ([]AppDNSPair, error) {
	currentSystem := systemconfig.GetCurrentSystem()
	
	if cachedDNSEndpoints != nil {
		if result, exists := cachedDNSEndpoints[currentSystem]; exists {
			return result, nil
		}
	}

	// Build a set of gRPC-only service pairs (source -> target)
	// This uses the grpcoperations data to identify which service pairs only use gRPC
	grpcOnlyPairs := buildGRPCOnlyPairs()

	// Get all service names
	services := serviceendpoints.GetAllServices()
	result := make([]AppDNSPair, 0)

	// For each service, get its endpoints
	for _, serviceName := range services {
		endpoints := serviceendpoints.GetEndpointsByService(serviceName)
		// Map from domain to span names
		domainSpanNames := make(map[string]map[string]bool)

		for _, endpoint := range endpoints {
			// Only include valid server addresses that are not the service itself
			if endpoint.ServerAddress != "" &&
				endpoint.ServerAddress != serviceName {
				if domainSpanNames[endpoint.ServerAddress] == nil {
					domainSpanNames[endpoint.ServerAddress] = make(map[string]bool)
				}
				if endpoint.SpanName != "" {
					domainSpanNames[endpoint.ServerAddress][endpoint.SpanName] = true
				}
			}
		}

		// Convert to AppDNSPairs with span names, filtering out gRPC-only connections
		for domain, spanNameSet := range domainSpanNames {
			// Check if this service pair is gRPC-only
			pairKey := serviceName + "->" + domain
			if grpcOnlyPairs[pairKey] {
				// Skip gRPC-only connections - DNS chaos doesn't work for them
				continue
			}
			
			spanNames := make([]string, 0, len(spanNameSet))
			for spanName := range spanNameSet {
				spanNames = append(spanNames, spanName)
			}
			sort.Strings(spanNames)
			result = append(result, AppDNSPair{
				AppName:   serviceName,
				Domain:    domain,
				SpanNames: spanNames,
			})
		}
	}

	// Sort by app name for consistency
	sort.Slice(result, func(i, j int) bool {
		if result[i].AppName != result[j].AppName {
			return result[i].AppName < result[j].AppName
		}
		return result[i].Domain < result[j].Domain
	})

	if cachedDNSEndpoints == nil {
		cachedDNSEndpoints = make(map[systemconfig.SystemType][]AppDNSPair)
	}
	cachedDNSEndpoints[currentSystem] = result
	return result, nil
}

// buildGRPCOnlyPairs builds a set of service pairs that only communicate via gRPC
// Returns a map where key is "source->target" and value is true if gRPC-only
func buildGRPCOnlyPairs() map[string]bool {
	grpcOnlyPairs := make(map[string]bool)
	
	// Get all gRPC client operations (these represent outgoing gRPC calls)
	grpcOps := grpcoperations.GetClientOperations()
	
	// Track which service pairs have gRPC connections
	grpcPairs := make(map[string]bool)
	for _, op := range grpcOps {
		pairKey := op.ServiceName + "->" + op.ServerAddress
		grpcPairs[pairKey] = true
	}
	
	// Get all service endpoints to check which pairs also have HTTP
	services := serviceendpoints.GetAllServices()
	httpPairs := make(map[string]bool)
	
	for _, serviceName := range services {
		endpoints := serviceendpoints.GetEndpointsByService(serviceName)
		for _, endpoint := range endpoints {
			// HTTP endpoints have non-empty Route that doesn't look like gRPC
			// (simple heuristic: HTTP routes don't start with /package.Service/)
			if endpoint.ServerAddress != "" && endpoint.ServerAddress != serviceName {
				if endpoint.Route != "" && !grpcoperations.IsGRPCRoutePattern(endpoint.Route) {
					pairKey := serviceName + "->" + endpoint.ServerAddress
					httpPairs[pairKey] = true
				}
			}
		}
	}
	
	// A pair is gRPC-only if it has gRPC but no HTTP
	for pair := range grpcPairs {
		if !httpPairs[pair] {
			grpcOnlyPairs[pair] = true
		}
	}
	
	return grpcOnlyPairs
}

// GetAllDatabaseOperations returns all app+database operations pairs sorted by app name
// This function uses the current system from systemconfig
// Note: DB chaos only supports MySQL, so we filter to only return MySQL operations
func GetAllDatabaseOperations() ([]AppDatabasePair, error) {
	currentSystem := systemconfig.GetCurrentSystem()
	
	if cachedDBOperations != nil {
		if result, exists := cachedDBOperations[currentSystem]; exists {
			return result, nil
		}
	}

	// Get all service names that have database operations
	services := databaseoperations.GetAllDatabaseServices()
	result := make([]AppDatabasePair, 0)

	// For each service, get its database operations
	for _, serviceName := range services {
		operations := databaseoperations.GetOperationsByService(serviceName)
		for _, op := range operations {
			// Only include MySQL operations (DB chaos only supports MySQL)
			if op.DBSystem == "mysql" {
				result = append(result, AppDatabasePair{
					AppName:       serviceName,
					DBName:        op.DBName,
					TableName:     op.DBTable,
					OperationType: op.Operation,
				})
			}
		}
	}

	// Sort by app name for consistency
	sort.Slice(result, func(i, j int) bool {
		if result[i].AppName != result[j].AppName {
			return result[i].AppName < result[j].AppName
		}
		if result[i].DBName != result[j].DBName {
			return result[i].DBName < result[j].DBName
		}
		if result[i].TableName != result[j].TableName {
			return result[i].TableName < result[j].TableName
		}
		return result[i].OperationType < result[j].OperationType
	})

	if cachedDBOperations == nil {
		cachedDBOperations = make(map[systemconfig.SystemType][]AppDatabasePair)
	}
	cachedDBOperations[currentSystem] = result
	return result, nil
}

// GetAllContainers returns all containers with their info sorted by app label
func GetAllContainers(namespace string) ([]ContainerInfo, error) {
	prefix, err := utils.ExtractNsPrefix(namespace)
	if err != nil {
		return nil, err
	}

	if result, exists := cachedContainerInfo[prefix]; exists {
		return result, nil
	}

	containers, err := client.GetContainersWithAppLabel(context.Background(), namespace)
	if err != nil {
		return nil, err
	}

	result := make([]ContainerInfo, 0, len(containers))
	for _, c := range containers {
		if c["appLabel"] != "" {
			result = append(result, ContainerInfo{
				PodName:       c["podName"],
				AppLabel:      c["appLabel"],
				ContainerName: c["containerName"],
			})
		}
	}

	// Sort by app label for consistency
	sort.Slice(result, func(i, j int) bool {
		if result[i].AppLabel != result[j].AppLabel {
			return result[i].AppLabel < result[j].AppLabel
		}
		return result[i].ContainerName < result[j].ContainerName
	})

	cachedContainerInfo[prefix] = result
	return result, nil
}

// GetContainersByService returns all container names for a specific service
func GetContainersByService(namespace string, serviceName string) ([]string, error) {
	allContainers, err := GetAllContainers(namespace)
	if err != nil {
		return nil, err
	}

	containerNames := []string{}
	for _, container := range allContainers {
		if container.AppLabel == serviceName {
			containerNames = append(containerNames, container.ContainerName)
		}
	}

	// Sort for consistency
	sort.Strings(containerNames)
	return containerNames, nil
}

// GetPodsByService returns all pod names for a specific service
func GetPodsByService(namespace string, serviceName string) ([]string, error) {
	allContainers, err := GetAllContainers(namespace)
	if err != nil {
		return nil, err
	}

	// Use a map to ensure uniqueness
	podMap := make(map[string]bool)
	for _, container := range allContainers {
		if container.AppLabel == serviceName {
			podMap[container.PodName] = true
		}
	}

	// Convert map to slice
	pods := make([]string, 0, len(podMap))
	for pod := range podMap {
		pods = append(pods, pod)
	}

	// Sort for consistency
	sort.Strings(pods)
	return pods, nil
}

// GetContainersAndPodsByServices returns containers and pods for multiple services
// This is useful for chaos that affects multiple services
func GetContainersAndPodsByServices(namespace string, serviceNames []string) ([]string, []string, error) {
	allContainers, err := GetAllContainers(namespace)
	if err != nil {
		return nil, nil, err
	}

	// Use maps to ensure uniqueness
	containerMap := make(map[string]bool)
	podMap := make(map[string]bool)

	// Create a map of service names for faster lookup
	serviceMap := make(map[string]bool)
	for _, service := range serviceNames {
		serviceMap[service] = true
	}

	// Filter containers for the specified services
	for _, container := range allContainers {
		if serviceMap[container.AppLabel] {
			containerMap[container.ContainerName] = true
			podMap[container.PodName] = true
		}
	}

	// Convert maps to slices
	containers := make([]string, 0, len(containerMap))
	for container := range containerMap {
		containers = append(containers, container)
	}

	pods := make([]string, 0, len(podMap))
	for pod := range podMap {
		pods = append(pods, pod)
	}

	// Sort for consistency
	sort.Strings(containers)
	sort.Strings(pods)

	return containers, pods, nil
}

func InitCaches() {
	cachedAppLabels = make(map[string][]string)
	cachedContainerInfo = make(map[string][]ContainerInfo)
	cachedAppMethods = make(map[systemconfig.SystemType][]AppMethodPair)
	cachedAppEndpoints = make(map[systemconfig.SystemType][]AppEndpointPair)
	cachedNetworkPairs = make(map[systemconfig.SystemType][]AppNetworkPair)
	cachedDNSEndpoints = make(map[systemconfig.SystemType][]AppDNSPair)
	cachedDBOperations = make(map[systemconfig.SystemType][]AppDatabasePair)
}

// PreloadCaches preloads resource caches to reduce first-access latency
func PreloadCaches(namespace string, labelKey string) error {
	// Create error channel to collect all errors
	errChan := make(chan error, 7)

	var wg sync.WaitGroup
	wg.Add(7)

	// Preload app labels
	go func() {
		defer wg.Done()
		_, err := GetAllAppLabels(namespace, labelKey)
		if err != nil {
			errChan <- fmt.Errorf("failed to preload app labels cache: %v", err)
		}
	}()

	// Preload JVM methods
	go func() {
		defer wg.Done()
		_, err := GetAllJVMMethods()
		if err != nil {
			errChan <- fmt.Errorf("failed to preload JVM methods cache: %v", err)
		}
	}()

	// Preload HTTP endpoints
	go func() {
		defer wg.Done()
		_, err := GetAllHTTPEndpoints()
		if err != nil {
			errChan <- fmt.Errorf("failed to preload HTTP endpoints cache: %v", err)
		}
	}()

	// Preload network pairs
	go func() {
		defer wg.Done()
		_, err := GetAllNetworkPairs()
		if err != nil {
			errChan <- fmt.Errorf("failed to preload network pairs cache: %v", err)
		}
	}()

	// Preload DNS endpoints
	go func() {
		defer wg.Done()
		_, err := GetAllDNSEndpoints()
		if err != nil {
			errChan <- fmt.Errorf("failed to preload DNS endpoints cache: %v", err)
		}
	}()

	// Preload database operations
	go func() {
		defer wg.Done()
		_, err := GetAllDatabaseOperations()
		if err != nil {
			errChan <- fmt.Errorf("failed to preload database operations cache: %v", err)
		}
	}()

	// Preload container info
	go func() {
		defer wg.Done()
		_, err := GetAllContainers(namespace)
		if err != nil {
			errChan <- fmt.Errorf("failed to preload container info cache: %v", err)
		}
	}()

	// Wait for all initialization to complete
	wg.Wait()
	close(errChan)

	// Collect all errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	// If there are errors, return the first one
	if len(errs) > 0 {
		return fmt.Errorf("cache preloading encountered errors: %v", errs[0])
	}

	return nil
}

// InvalidateCache clears all cached data
func InvalidateCache() {
	cachedAppLabels = make(map[string][]string)
	cachedContainerInfo = make(map[string][]ContainerInfo)
	cachedAppMethods = make(map[systemconfig.SystemType][]AppMethodPair)
	cachedAppEndpoints = make(map[systemconfig.SystemType][]AppEndpointPair)
	cachedNetworkPairs = make(map[systemconfig.SystemType][]AppNetworkPair)
	cachedDNSEndpoints = make(map[systemconfig.SystemType][]AppDNSPair)
	cachedDBOperations = make(map[systemconfig.SystemType][]AppDatabasePair)
}
