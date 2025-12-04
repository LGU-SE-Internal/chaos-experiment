package networkdependencies

import (
	"sync"

	"github.com/LGU-SE-Internal/chaos-experiment/internal/serviceendpoints"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"
)

// ServiceDependency represents a dependency between services
type ServiceDependency struct {
	SourceService     string
	TargetService     string
	ConnectionDetails string
}

// Function variables that can be replaced during testing
var (
	// GetServicePairByServiceAndIndexFunc is the implementation for GetServicePairByServiceAndIndex
	GetServicePairByServiceAndIndexFunc = getServicePairByServiceAndIndexImpl

	// GetDependenciesForServiceFunc is the implementation for GetDependenciesForService
	GetDependenciesForServiceFunc = getDependenciesForServiceImpl

	// ListAllServiceNamesFunc is the implementation for ListAllServiceNames
	ListAllServiceNamesFunc = listAllServiceNamesImpl

	// GetAllServicePairsFunc is the implementation for GetAllServicePairs
	GetAllServicePairsFunc = getAllServicePairsImpl
)

// dependencyGraphs maps system types to their dependency graphs
var dependencyGraphs = make(map[systemconfig.SystemType]map[string][]string)
var graphMu sync.RWMutex

// getDependencyGraph returns the dependency graph for the current system, building it if necessary
func getDependencyGraph() map[string][]string {
	system := systemconfig.GetCurrentSystem()

	graphMu.RLock()
	if graph, exists := dependencyGraphs[system]; exists {
		graphMu.RUnlock()
		return graph
	}
	graphMu.RUnlock()

	// Build the graph for this system
	graphMu.Lock()
	defer graphMu.Unlock()

	// Double-check after acquiring write lock
	if graph, exists := dependencyGraphs[system]; exists {
		return graph
	}

	graph := buildDependencyGraphForSystem()
	dependencyGraphs[system] = graph
	return graph
}

// buildDependencyGraph builds a map of service dependencies based on service endpoints
func buildDependencyGraph() {
	dependencyGraph = make(map[string][]string)

	// Get all service names
	allServices := serviceendpoints.GetAllServices()

	// For each service, find all targets it communicates with
	for _, service := range allServices {
		endpoints := serviceendpoints.GetEndpointsByService(service)

		// Track services this service depends on
		for _, endpoint := range endpoints {
			// Add this dependency to the graph if not already present
			addDependency(service, endpoint.ServerAddress)

			// Also add the reverse dependency for bidirectional relationships
			addDependency(endpoint.ServerAddress, service)
		}
	}
}

// addDependency adds a target service to the dependency list of a source service
func addDependency(sourceService, targetService string) {
	// Skip if source and target are the same
	if sourceService == targetService {
		return
	}

	// Initialize the slice if it doesn't exist
	if _, exists := dependencyGraph[sourceService]; !exists {
		dependencyGraph[sourceService] = []string{}
	}

	// Check if target already exists in the dependency list
	for _, existing := range dependencyGraph[sourceService] {
		if existing == targetService {
			return
		}
	}

	// Add the target to the dependency list
	dependencyGraph[sourceService] = append(dependencyGraph[sourceService], targetService)
}

// GetDependenciesForService returns all services that a given service communicates with
func GetDependenciesForService(serviceName string) []string {
	return GetDependenciesForServiceFunc(serviceName)
}

// getDependenciesForServiceImpl is the actual implementation of GetDependenciesForService
func getDependenciesForServiceImpl(serviceName string) []string {
	if dependencies, exists := dependencyGraph[serviceName]; exists {
		return dependencies
	}
	return []string{}
}

// GetAllServicePairs returns a list of all available service communication pairs
func GetAllServicePairs() []ServiceDependency {
	return GetAllServicePairsFunc()
}

// getAllServicePairsImpl is the actual implementation of GetAllServicePairs
func getAllServicePairsImpl() []ServiceDependency {
	var pairs []ServiceDependency

	for source, targets := range dependencyGraph {
		for _, target := range targets {
			pairs = append(pairs, ServiceDependency{
				SourceService:     source,
				TargetService:     target,
				ConnectionDetails: "HTTP/gRPC Communication",
			})
		}
	}

	return pairs
}

// GetServicePair returns a specific service pair by index
func GetServicePair(index int) (source, target string, ok bool) {
	pairs := GetAllServicePairs()
	if index < 0 || index >= len(pairs) {
		return "", "", false
	}

	return pairs[index].SourceService, pairs[index].TargetService, true
}

// GetServicePairByServiceAndIndex returns a target service for a given source service by index
func GetServicePairByServiceAndIndex(serviceName string, index int) (string, bool) {
	return GetServicePairByServiceAndIndexFunc(serviceName, index)
}

// getServicePairByServiceAndIndexImpl is the actual implementation of GetServicePairByServiceAndIndex
func getServicePairByServiceAndIndexImpl(serviceName string, index int) (string, bool) {
	dependencies := GetDependenciesForService(serviceName)

	if index < 0 || index >= len(dependencies) {
		return "", false
	}

	return dependencies[index], true
}

// CountDependencies returns the number of dependencies for a service
func CountDependencies(serviceName string) int {
	return len(GetDependenciesForService(serviceName))
}

// ListAllServiceNames returns a list of all available service names with dependencies
func ListAllServiceNames() []string {
	return ListAllServiceNamesFunc()
}

// listAllServiceNamesImpl is the actual implementation of ListAllServiceNames
func listAllServiceNamesImpl() []string {
	serviceNames := []string{}

	for service := range dependencyGraph {
		if len(dependencyGraph[service]) > 0 {
			serviceNames = append(serviceNames, service)
		}
	}

	return serviceNames
}
