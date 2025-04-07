package testdata

import (
	"github.com/CUHK-SE-Group/chaos-experiment/internal/javaclassmethods"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/networkdependencies"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/serviceendpoints"
)

// Type definitions for function types that will be mocked
type (
	EndpointsGetterFunc  func(string) []serviceendpoints.ServiceEndpoint
	LabelsGetterFunc     func(string, string) ([]string, error)
	JavaMethodGetterFunc func(string, int) *javaclassmethods.ClassMethodEntry
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
	originalGetServicePair := networkdependencies.GetServicePairByServiceAndIndexFunc
	originalGetDependencies := networkdependencies.GetDependenciesForServiceFunc
	originalListServiceNames := networkdependencies.ListAllServiceNamesFunc
	originalGetAllPairs := networkdependencies.GetAllServicePairsFunc

	// Replace with mock implementations
	networkdependencies.GetServicePairByServiceAndIndexFunc = MockGetServicePairByServiceAndIndex
	networkdependencies.GetDependenciesForServiceFunc = MockGetDependenciesForService
	networkdependencies.ListAllServiceNamesFunc = MockListAllServiceNames
	networkdependencies.GetAllServicePairsFunc = MockGetAllServicePairs

	// Return cleanup function
	return func() {
		networkdependencies.GetServicePairByServiceAndIndexFunc = originalGetServicePair
		networkdependencies.GetDependenciesForServiceFunc = originalGetDependencies
		networkdependencies.ListAllServiceNamesFunc = originalListServiceNames
		networkdependencies.GetAllServicePairsFunc = originalGetAllPairs
	}
}
