package resourcelookup

import (
	"context"
	"sort"
	"sync"

	"github.com/LGU-SE-Internal/chaos-experiment/client"
	_ "github.com/LGU-SE-Internal/chaos-experiment/internal/adapter" // Auto-register all systems
	"github.com/LGU-SE-Internal/chaos-experiment/internal/registry"
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
	cacheMutex          sync.RWMutex
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

	cacheMutex.RLock()
	if labels, exists := cachedAppLabels[prefix]; exists && len(labels) > 0 {
		cacheMutex.RUnlock()
		return labels, nil
	}
	cacheMutex.RUnlock()

	labels, err := client.GetLabels(context.Background(), namespace, key)
	logrus.Debugf("Fetched labels for namespace %s with key %s: %v", namespace, key, labels)
	if err != nil {
		return nil, err
	}

	// Sort alphabetically
	sort.Strings(labels)
	
	cacheMutex.Lock()
	if cachedAppLabels == nil {
		cachedAppLabels = make(map[string][]string)
	}
	cachedAppLabels[prefix] = labels
	cacheMutex.Unlock()
	
	return labels, nil
}

// GetAllHTTPEndpoints returns all app+endpoint pairs sorted by app name
// This function uses the current system from systemconfig and the NEW registry pattern
func GetAllHTTPEndpoints() ([]AppEndpointPair, error) {
	currentSystem := systemconfig.GetCurrentSystem()
	
	cacheMutex.RLock()
	if cachedAppEndpoints != nil {
		if result, exists := cachedAppEndpoints[currentSystem]; exists {
			cacheMutex.RUnlock()
			return result, nil
		}
	}
	cacheMutex.RUnlock()

	// NEW: Use registry pattern instead of direct imports
	sysData := registry.MustGetCurrent()
	result := make([]AppEndpointPair, 0)

	// For each service, get its endpoints
	for _, serviceName := range sysData.GetAllServices() {
		endpoints := sysData.GetHTTPEndpointsByService(serviceName)
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

	cacheMutex.Lock()
	if cachedAppEndpoints == nil {
		cachedAppEndpoints = make(map[systemconfig.SystemType][]AppEndpointPair)
	}
	cachedAppEndpoints[currentSystem] = result
	cacheMutex.Unlock()
	
	return result, nil
}

// GetAllNetworkPairs returns all network pairs sorted by source service
// This function uses the current system from systemconfig and the NEW registry pattern
func GetAllNetworkPairs() ([]AppNetworkPair, error) {
	currentSystem := systemconfig.GetCurrentSystem()
	
	cacheMutex.RLock()
	if cachedNetworkPairs != nil {
		if result, exists := cachedNetworkPairs[currentSystem]; exists {
			cacheMutex.RUnlock()
			return result, nil
		}
	}
	cacheMutex.RUnlock()

	// NEW: Use registry pattern to build network pairs from all operation types
	sysData := registry.MustGetCurrent()
	pairMap := make(map[string]*AppNetworkPair)

	// Add HTTP endpoints
	for _, service := range sysData.GetAllServices() {
		for _, ep := range sysData.GetHTTPEndpointsByService(service) {
			if ep.ServerAddress != "" && ep.ServerAddress != service {
				key := ep.ServiceName + "->" + ep.ServerAddress
				if pairMap[key] == nil {
					pairMap[key] = &AppNetworkPair{
						SourceService: ep.ServiceName,
						TargetService: ep.ServerAddress,
						SpanNames:     []string{},
					}
				}
				if ep.SpanName != "" {
					pairMap[key].SpanNames = append(pairMap[key].SpanNames, ep.SpanName)
				}
			}
		}
	}

	// Add RPC operations
	for _, service := range sysData.GetAllRPCServices() {
		for _, op := range sysData.GetRPCOperationsByService(service) {
			if op.ServerAddress != "" && op.ServerAddress != service {
				key := op.ServiceName + "->" + op.ServerAddress
				if pairMap[key] == nil {
					pairMap[key] = &AppNetworkPair{
						SourceService: op.ServiceName,
						TargetService: op.ServerAddress,
						SpanNames:     []string{},
					}
				}
				if op.SpanName != "" {
					pairMap[key].SpanNames = append(pairMap[key].SpanNames, op.SpanName)
				}
			}
		}
	}

	// Add Database operations
	for _, service := range sysData.GetAllDatabaseServices() {
		for _, op := range sysData.GetDatabaseOperationsByService(service) {
			if op.ServerAddress != "" && op.ServerAddress != service {
				key := op.ServiceName + "->" + op.ServerAddress
				if pairMap[key] == nil {
					pairMap[key] = &AppNetworkPair{
						SourceService: op.ServiceName,
						TargetService: op.ServerAddress,
						SpanNames:     []string{},
					}
				}
				if op.SpanName != "" {
					pairMap[key].SpanNames = append(pairMap[key].SpanNames, op.SpanName)
				}
			}
		}
	}

	// Convert map to slice
	result := make([]AppNetworkPair, 0, len(pairMap))
	for _, pair := range pairMap {
		// Deduplicate and sort span names
		pair.SpanNames = uniqueSorted(pair.SpanNames)
		result = append(result, *pair)
	}

	// Sort by source service for consistency
	sort.Slice(result, func(i, j int) bool {
		if result[i].SourceService != result[j].SourceService {
			return result[i].SourceService < result[j].SourceService
		}
		return result[i].TargetService < result[j].TargetService
	})

	cacheMutex.Lock()
	if cachedNetworkPairs == nil {
		cachedNetworkPairs = make(map[systemconfig.SystemType][]AppNetworkPair)
	}
	cachedNetworkPairs[currentSystem] = result
	cacheMutex.Unlock()
	
	return result, nil
}

// GetAllDNSEndpoints returns all DNS endpoints (HTTP + DB, excludes RPC)
// This function uses the current system from systemconfig and the NEW registry pattern
func GetAllDNSEndpoints() ([]AppDNSPair, error) {
	currentSystem := systemconfig.GetCurrentSystem()
	
	cacheMutex.RLock()
	if cachedDNSEndpoints != nil {
		if result, exists := cachedDNSEndpoints[currentSystem]; exists {
			cacheMutex.RUnlock()
			return result, nil
		}
	}
	cacheMutex.RUnlock()

	// NEW: Use registry pattern for DNS endpoints (HTTP + DB, no RPC)
	sysData := registry.MustGetCurrent()
	domainMap := make(map[string]*AppDNSPair)

	// Add HTTP endpoints
	for _, service := range sysData.GetAllServices() {
		for _, ep := range sysData.GetHTTPEndpointsByService(service) {
			if ep.ServerAddress != "" && ep.ServerAddress != service {
				key := ep.ServiceName + "->" + ep.ServerAddress
				if domainMap[key] == nil {
					domainMap[key] = &AppDNSPair{
						AppName:   ep.ServiceName,
						Domain:    ep.ServerAddress,
						SpanNames: []string{},
					}
				}
				if ep.SpanName != "" {
					domainMap[key].SpanNames = append(domainMap[key].SpanNames, ep.SpanName)
				}
			}
		}
	}

	// Add Database operations
	for _, service := range sysData.GetAllDatabaseServices() {
		for _, op := range sysData.GetDatabaseOperationsByService(service) {
			if op.ServerAddress != "" && op.ServerAddress != service {
				key := op.ServiceName + "->" + op.ServerAddress
				if domainMap[key] == nil {
					domainMap[key] = &AppDNSPair{
						AppName:   op.ServiceName,
						Domain:    op.ServerAddress,
						SpanNames: []string{},
					}
				}
				if op.SpanName != "" {
					domainMap[key].SpanNames = append(domainMap[key].SpanNames, op.SpanName)
				}
			}
		}
	}

	// Convert map to slice
	result := make([]AppDNSPair, 0, len(domainMap))
	for _, pair := range domainMap {
		pair.SpanNames = uniqueSorted(pair.SpanNames)
		result = append(result, *pair)
	}

	// Sort by app name for consistency
	sort.Slice(result, func(i, j int) bool {
		if result[i].AppName != result[j].AppName {
			return result[i].AppName < result[j].AppName
		}
		return result[i].Domain < result[j].Domain
	})

	cacheMutex.Lock()
	if cachedDNSEndpoints == nil {
		cachedDNSEndpoints = make(map[systemconfig.SystemType][]AppDNSPair)
	}
	cachedDNSEndpoints[currentSystem] = result
	cacheMutex.Unlock()
	
	return result, nil
}

// GetAllDatabaseOperations returns all database operations
// This function uses the current system from systemconfig and the NEW registry pattern
func GetAllDatabaseOperations() ([]AppDatabasePair, error) {
	currentSystem := systemconfig.GetCurrentSystem()
	
	cacheMutex.RLock()
	if cachedDBOperations != nil {
		if result, exists := cachedDBOperations[currentSystem]; exists {
			cacheMutex.RUnlock()
			return result, nil
		}
	}
	cacheMutex.RUnlock()

	// NEW: Use registry pattern for database operations
	sysData := registry.MustGetCurrent()
	result := make([]AppDatabasePair, 0)

	for _, service := range sysData.GetAllDatabaseServices() {
		for _, op := range sysData.GetDatabaseOperationsByService(service) {
			result = append(result, AppDatabasePair{
				AppName:       op.ServiceName,
				DBName:        op.DBName,
				TableName:     op.DBTable,
				OperationType: op.Operation,
			})
		}
	}

	// Sort by app name, then database name
	sort.Slice(result, func(i, j int) bool {
		if result[i].AppName != result[j].AppName {
			return result[i].AppName < result[j].AppName
		}
		if result[i].DBName != result[j].DBName {
			return result[i].DBName < result[j].DBName
		}
		return result[i].TableName < result[j].TableName
	})

	cacheMutex.Lock()
	if cachedDBOperations == nil {
		cachedDBOperations = make(map[systemconfig.SystemType][]AppDatabasePair)
	}
	cachedDBOperations[currentSystem] = result
	cacheMutex.Unlock()
	
	return result, nil
}

// GetAllContainers returns all containers with their info sorted by app label
func GetAllContainers(namespace string) ([]ContainerInfo, error) {
	prefix, err := utils.ExtractNsPrefix(namespace)
	if err != nil {
		return nil, err
	}

	cacheKey := prefix + ":all"

	cacheMutex.RLock()
	if result, exists := cachedContainerInfo[cacheKey]; exists && len(result) > 0 {
		cacheMutex.RUnlock()
		return result, nil
	}
	cacheMutex.RUnlock()

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

	cacheMutex.Lock()
	if cachedContainerInfo == nil {
		cachedContainerInfo = make(map[string][]ContainerInfo)
	}
	cachedContainerInfo[cacheKey] = result
	cacheMutex.Unlock()

	return result, nil
}

// GetContainersByAppLabel returns all containers for a given app label
func GetContainersByAppLabel(appLabel, namespace string) ([]ContainerInfo, error) {
	prefix, err := utils.ExtractNsPrefix(namespace)
	if err != nil {
		return nil, err
	}

	cacheKey := prefix + ":" + appLabel

	cacheMutex.RLock()
	if containers, exists := cachedContainerInfo[cacheKey]; exists && len(containers) > 0 {
		cacheMutex.RUnlock()
		return containers, nil
	}
	cacheMutex.RUnlock()

	// Get all containers with app labels
	allContainers, err := client.GetContainersWithAppLabel(context.Background(), namespace)
	if err != nil {
		return nil, err
	}

	// Filter by the specific app label
	containers := make([]ContainerInfo, 0)
	for _, c := range allContainers {
		if c["appLabel"] == appLabel {
			containers = append(containers, ContainerInfo{
				PodName:       c["podName"],
				AppLabel:      c["appLabel"],
				ContainerName: c["containerName"],
			})
		}
	}

	cacheMutex.Lock()
	if cachedContainerInfo == nil {
		cachedContainerInfo = make(map[string][]ContainerInfo)
	}
	cachedContainerInfo[cacheKey] = containers
	cacheMutex.Unlock()

	return containers, nil
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

// InitCaches initializes resource caches
func InitCaches() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	
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
	var wg sync.WaitGroup
	errChan := make(chan error, 6)

	wg.Add(6)

	// Preload app labels
	go func() {
		defer wg.Done()
		_, err := GetAllAppLabels(namespace, labelKey)
		if err != nil {
			errChan <- err
		}
	}()

	// Preload HTTP endpoints
	go func() {
		defer wg.Done()
		_, err := GetAllHTTPEndpoints()
		if err != nil {
			errChan <- err
		}
	}()

	// Preload network pairs
	go func() {
		defer wg.Done()
		_, err := GetAllNetworkPairs()
		if err != nil {
			errChan <- err
		}
	}()

	// Preload DNS endpoints
	go func() {
		defer wg.Done()
		_, err := GetAllDNSEndpoints()
		if err != nil {
			errChan <- err
		}
	}()

	// Preload database operations
	go func() {
		defer wg.Done()
		_, err := GetAllDatabaseOperations()
		if err != nil {
			errChan <- err
		}
	}()

	// Preload container info
	go func() {
		defer wg.Done()
		_, err := GetAllContainers(namespace)
		if err != nil {
			errChan <- err
		}
	}()

	// Wait for all initialization to complete
	wg.Wait()
	close(errChan)

	// Return the first error if any
	for err := range errChan {
		return err
	}

	return nil
}

// ClearCache clears all cached data (alias for InvalidateCache)
func ClearCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	
	cachedAppLabels = nil
	cachedAppMethods = nil
	cachedAppEndpoints = nil
	cachedNetworkPairs = nil
	cachedDNSEndpoints = nil
	cachedContainerInfo = nil
	cachedDBOperations = nil
}

// InvalidateCache clears all cached data
func InvalidateCache() {
	ClearCache()
}

// Helper functions

func uniqueSorted(items []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(items))
	for _, item := range items {
		if !seen[item] && item != "" {
			seen[item] = true
			result = append(result, item)
		}
	}
	sort.Strings(result)
	return result
}

// GetAllJVMMethods is not migrated yet - would need JVM data in registry
// Keeping stub for compatibility
func GetAllJVMMethods() ([]AppMethodPair, error) {
	// TODO: Migrate JVM methods to registry pattern
	return []AppMethodPair{}, nil
}
