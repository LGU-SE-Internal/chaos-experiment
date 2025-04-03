package networkdependencies

import (
	"github.com/CUHK-SE-Group/chaos-experiment/internal/serviceendpoints"
)

// ServiceDependency represents a dependency between services
type ServiceDependency struct {
	SourceService     string
	TargetService     string
	ConnectionDetails string
}

// Function variables that can be replaced during testing
var (
	// GetServicePairByServiceAndIndex retrieves the target service for a source service based on index
	GetServicePairByServiceAndIndex = getServicePairByServiceAndIndex

	// GetDependenciesForService returns all target services for a given source service
	GetDependenciesForService = getDependenciesForService

	// ListAllServiceNames returns a list of all service names with dependencies
	ListAllServiceNames = listAllServiceNames

	// GetAllServicePairs returns all service dependency pairs
	GetAllServicePairs = getAllServicePairs
)

// dependencyGraph is a map of service to its dependent services
var dependencyGraph map[string][]string

// Initialize the dependency graph from service endpoints
func init() {
	buildDependencyGraph()
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
func getDependenciesForService(serviceName string) []string {
	if dependencies, exists := dependencyGraph[serviceName]; exists {
		return dependencies
	}
	return []string{}
}

// GetAllServicePairs returns a list of all available service communication pairs
func getAllServicePairs() []ServiceDependency {
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
func getServicePair(index int) (source, target string, ok bool) {
	pairs := GetAllServicePairs()
	if index < 0 || index >= len(pairs) {
		return "", "", false
	}

	return pairs[index].SourceService, pairs[index].TargetService, true
}

// GetServicePairByServiceAndIndex returns a target service for a given source service by index
func getServicePairByServiceAndIndex(serviceName string, index int) (target string, ok bool) {
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
func listAllServiceNames() []string {
	serviceNames := []string{}

	for service := range dependencyGraph {
		if len(dependencyGraph[service]) > 0 {
			serviceNames = append(serviceNames, service)
		}
	}

	return serviceNames
}
