package resourcelookup

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/LGU-SE-Internal/chaos-experiment/client"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/databaseoperations"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/javaclassmethods"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/networkdependencies"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/serviceendpoints"
	"github.com/LGU-SE-Internal/chaos-experiment/utils"
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
}

// AppNetworkPair represents a flattened source+target combination for network chaos
type AppNetworkPair struct {
	SourceService string `json:"source_service"`
	TargetService string `json:"target_service"`
}

// AppDNSPair represents a flattened app+domain combination for DNS chaos
type AppDNSPair struct {
	AppName string `json:"app_name"`
	Domain  string `json:"domain"`
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

// Global cache for lookups
var (
	cachedAppLabels     map[string][]string
	cachedAppMethods    []AppMethodPair
	cachedAppEndpoints  []AppEndpointPair
	cachedNetworkPairs  []AppNetworkPair
	cachedDNSEndpoints  []AppDNSPair
	cachedContainerInfo map[string][]ContainerInfo
	cachedDBOperations  []AppDatabasePair
)

// GetAllAppLabels returns all application labels sorted alphabetically
func GetAllAppLabels(namespace string, key string) ([]string, error) {
	prefix, err := utils.ExtractNsPrefix(namespace)
	if err != nil {
		return nil, err
	}

	if labels, exists := cachedAppLabels[prefix]; exists {
		return labels, nil
	}

	labels, err := client.GetLabels(context.Background(), namespace, key)
	if err != nil {
		return nil, err
	}

	// Sort alphabetically
	sort.Strings(labels)
	cachedAppLabels[prefix] = labels
	return labels, nil
}

// GetAllJVMMethods returns all app+method pairs sorted by app name
func GetAllJVMMethods() ([]AppMethodPair, error) {
	if cachedAppMethods != nil {
		return cachedAppMethods, nil
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

	cachedAppMethods = result
	return result, nil
}

// GetAllHTTPEndpoints returns all app+endpoint pairs sorted by app name
func GetAllHTTPEndpoints() ([]AppEndpointPair, error) {
	if cachedAppEndpoints != nil {
		return cachedAppEndpoints, nil
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

	cachedAppEndpoints = result
	return result, nil
}

// GetAllNetworkPairs returns all network pairs sorted by source service
func GetAllNetworkPairs() ([]AppNetworkPair, error) {
	if cachedNetworkPairs != nil {
		return cachedNetworkPairs, nil
	}

	// Get all service-to-service pairs
	pairs := networkdependencies.GetAllServicePairs()
	result := make([]AppNetworkPair, 0, len(pairs))

	for _, pair := range pairs {
		result = append(result, AppNetworkPair{
			SourceService: pair.SourceService,
			TargetService: pair.TargetService,
		})
	}

	// Sort by source service for consistency
	sort.Slice(result, func(i, j int) bool {
		if result[i].SourceService != result[j].SourceService {
			return result[i].SourceService < result[j].SourceService
		}
		return result[i].TargetService < result[j].TargetService
	})

	cachedNetworkPairs = result
	return result, nil
}

// GetAllDNSEndpoints returns all app+domain pairs for DNS chaos sorted by app name
func GetAllDNSEndpoints() ([]AppDNSPair, error) {
	if cachedDNSEndpoints != nil {
		return cachedDNSEndpoints, nil
	}

	// Get all service names
	services := serviceendpoints.GetAllServices()
	result := make([]AppDNSPair, 0)

	// For each service, get its endpoints
	for _, serviceName := range services {
		endpoints := serviceendpoints.GetEndpointsByService(serviceName)
		uniqueDomains := make(map[string]bool)

		for _, endpoint := range endpoints {
			// Only include valid server addresses that are not the service itself
			if endpoint.ServerAddress != "" &&
				endpoint.ServerAddress != serviceName {
				uniqueDomains[endpoint.ServerAddress] = true
			}
		}

		// Convert unique domains to AppDNSPairs
		for domain := range uniqueDomains {
			result = append(result, AppDNSPair{
				AppName: serviceName,
				Domain:  domain,
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

	cachedDNSEndpoints = result
	return result, nil
}

// GetAllDatabaseOperations returns all app+database operations pairs sorted by app name
func GetAllDatabaseOperations() ([]AppDatabasePair, error) {
	if cachedDBOperations != nil {
		return cachedDBOperations, nil
	}

	// Get all service names that have database operations
	services := databaseoperations.GetAllDatabaseServices()
	result := make([]AppDatabasePair, 0)

	// For each service, get its database operations
	for _, serviceName := range services {
		operations := databaseoperations.GetOperationsByService(serviceName)
		for _, op := range operations {
			result = append(result, AppDatabasePair{
				AppName:       serviceName,
				DBName:        op.DBName,
				TableName:     op.DBTable,
				OperationType: op.Operation,
			})
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

	cachedDBOperations = result
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
		result = append(result, ContainerInfo{
			PodName:       c["podName"],
			AppLabel:      c["appLabel"],
			ContainerName: c["containerName"],
		})
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

	cachedAppMethods = nil
	cachedAppEndpoints = nil
	cachedNetworkPairs = nil
	cachedDNSEndpoints = nil
	cachedContainerInfo = nil
	cachedDBOperations = nil
}
