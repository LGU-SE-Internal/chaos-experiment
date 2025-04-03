package testdata

import (
	"github.com/CUHK-SE-Group/chaos-experiment/internal/javaclassmethods"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/networkdependencies"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/serviceendpoints"
)

// Type definitions for function types that will be mocked
type (
	EndpointsGetterFunc      func(string) []serviceendpoints.ServiceEndpoint
	LabelsGetterFunc         func(string, string) ([]string, error)
	NetworkDependencyFunc    func(string, int) (string, bool)
	JavaMethodGetterFunc     func(string, int) *javaclassmethods.ClassMethodEntry
)

// SetupLabelsMock replaces the labels getter function with mock implementation
// and returns a cleanup function
func SetupLabelsMock(originalGetter LabelsGetterFunc) func() {
	return func() {
		// This would be set by the test: labelsGetter = originalGetter
	}
}

// SetupEndpointsMock replaces the endpoints getter with a mock implementation
// and returns a cleanup function
func SetupEndpointsMock(originalGetter EndpointsGetterFunc) func() {
	return func() {
		// This would be set by the test: endpointsGetter = originalGetter
	}
}

// SetupNetworkDependenciesMock replaces the network dependency functions with mock implementations
// and returns a cleanup function
func SetupNetworkDependenciesMock() func() {
	// Store original functions
	originalGetServicePair := networkdependencies.GetServicePairByServiceAndIndex
	originalGetDependencies := networkdependencies.GetDependenciesForService
	originalListServiceNames := networkdependencies.ListAllServiceNames
	originalGetAllPairs := networkdependencies.GetAllServicePairs

	// Replace with mock implementations
	networkdependencies.GetServicePairByServiceAndIndex = MockGetServicePairByServiceAndIndex
	networkdependencies.GetDependenciesForService = MockGetDependenciesForService
	networkdependencies.ListAllServiceNames = MockListAllServiceNames
	networkdependencies.GetAllServicePairs = MockGetAllServicePairs

	// Return cleanup function
	return func() {
		networkdependencies.GetServicePairByServiceAndIndex = originalGetServicePair
		networkdependencies.GetDependenciesForService = originalGetDependencies
		networkdependencies.ListAllServiceNames = originalListServiceNames
		networkdependencies.GetAllServicePairs = originalGetAllPairs
	}
}
