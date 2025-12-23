package resourcelookup

import (
	"context"
	"sort"
	"sync"

	"github.com/LGU-SE-Internal/chaos-experiment/client"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/databaseoperations"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/grpcoperations"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/javaclassmethods"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/networkdependencies"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/serviceendpoints"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"
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

type systemCache struct {
	system        systemconfig.SystemType
	appLabels     map[string][]string
	appMethods    []AppMethodPair
	appEndpoints  []AppEndpointPair
	networkPairs  []AppNetworkPair
	dnsEndpoints  []AppDNSPair
	containerInfo map[string][]ContainerInfo
	dbOperations  []AppDatabasePair
}

func GetSystemCache(system systemconfig.SystemType) *systemCache {
	return getCacheManager().getSystemCache(system)
}

// newSystemCache creates a new systemCache instance
func newSystemCache(system systemconfig.SystemType) *systemCache {
	return &systemCache{
		system:        system,
		appLabels:     make(map[string][]string),
		appMethods:    []AppMethodPair{},
		appEndpoints:  []AppEndpointPair{},
		networkPairs:  []AppNetworkPair{},
		dnsEndpoints:  []AppDNSPair{},
		dbOperations:  []AppDatabasePair{},
		containerInfo: make(map[string][]ContainerInfo),
	}
}

// cacheManager manages caches for different namespaces (singleton)
type cacheManager struct {
	caches map[systemconfig.SystemType]*systemCache
	mu     sync.RWMutex
}

var (
	managerInstance *cacheManager
	managerOnce     sync.Once
)

// getCacheManager returns the singleton CacheManager instance
func getCacheManager() *cacheManager {
	managerOnce.Do(func() {
		allSystemTypes := systemconfig.GetAllSystemTypes()
		managerInstance = &cacheManager{
			caches: make(map[systemconfig.SystemType]*systemCache, len(allSystemTypes)),
		}
	})
	return managerInstance
}

func (cm *cacheManager) getSystemCache(system systemconfig.SystemType) *systemCache {
	cm.mu.RLock()
	cache, exists := cm.caches[system]
	cm.mu.RUnlock()

	if !exists {
		cm.mu.Lock()
		defer cm.mu.Unlock()
		// Double-check existence
		cache, exists = cm.caches[system]
		if !exists {
			cache = newSystemCache(system)
			cm.caches[system] = cache
		}
	}

	return cache
}

// GetAllAppLabels returns all application labels sorted alphabetically
func (s *systemCache) GetAllAppLabels(ctx context.Context, namespace string, key string) ([]string, error) {
	if len(s.appLabels) > 0 {
		if labels, exists := s.appLabels[key]; exists {
			return labels, nil
		}
	}

	labels, err := client.GetLabels(ctx, namespace, key)
	logrus.Debugf("Fetched labels for namespace %s with key %s: %v", namespace, key, labels)
	if err != nil {
		return nil, err
	}

	// Sort alphabetically
	sort.Strings(labels)
	return labels, nil
}

// GetAllJVMMethods returns all app+method pairs sorted by app name
// This function uses the current system from systemconfig
func (s *systemCache) GetAllJVMMethods() ([]AppMethodPair, error) {
	if len(s.appMethods) > 0 {
		return s.appMethods, nil
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

	return result, nil
}

// GetAllHTTPEndpoints returns all app+endpoint pairs sorted by app name
// This function uses the current system from systemconfig
func (s *systemCache) GetAllHTTPEndpoints() ([]AppEndpointPair, error) {
	if len(s.appEndpoints) > 0 {
		return s.appEndpoints, nil
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

	return result, nil
}

// GetAllNetworkPairs returns all network pairs sorted by source service
// This function uses the current system from systemconfig
func (s *systemCache) GetAllNetworkPairs() ([]AppNetworkPair, error) {
	if len(s.networkPairs) > 0 {
		return s.networkPairs, nil
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
func (s *systemCache) GetAllDNSEndpoints() ([]AppDNSPair, error) {
	if len(s.dnsEndpoints) > 0 {
		return s.dnsEndpoints, nil
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
func (s *systemCache) GetAllDatabaseOperations() ([]AppDatabasePair, error) {
	if len(s.dbOperations) > 0 {
		return s.dbOperations, nil
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

	return result, nil
}

// GetAllContainers returns all containers with their info sorted by app label
func (s *systemCache) GetAllContainers(ctx context.Context, namespace string) ([]ContainerInfo, error) {
	if len(s.containerInfo) > 0 {
		if containers, exists := s.containerInfo[namespace]; exists {
			return containers, nil
		}
	}

	containers, err := client.GetContainersWithAppLabel(ctx, namespace)
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

	s.containerInfo[namespace] = result
	return result, nil
}

// GetContainersByService returns all container names for a specific service
func (s *systemCache) GetContainersByService(ctx context.Context, namespace string, serviceName string) ([]string, error) {
	allContainers, err := s.GetAllContainers(ctx, namespace)
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
func (s *systemCache) GetPodsByService(ctx context.Context, namespace string, serviceName string) ([]string, error) {
	allContainers, err := s.GetAllContainers(ctx, namespace)
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
func (s *systemCache) GetContainersAndPodsByServices(ctx context.Context, namespace string, serviceNames []string) ([]string, []string, error) {
	allContainers, err := s.GetAllContainers(ctx, namespace)
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

// InvalidateCache clears all cached data
func (s *systemCache) InvalidateCache() {
	s.appLabels = make(map[string][]string)
	s.appMethods = []AppMethodPair{}
	s.appEndpoints = []AppEndpointPair{}
	s.networkPairs = []AppNetworkPair{}
	s.dnsEndpoints = []AppDNSPair{}
	s.containerInfo = make(map[string][]ContainerInfo)
	s.dbOperations = []AppDatabasePair{}
}
