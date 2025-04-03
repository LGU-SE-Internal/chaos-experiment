package handler

import (
	"github.com/CUHK-SE-Group/chaos-experiment/internal/networkdependencies"
)

// selectNetworkTargetForService selects a target service for a given source service
// based on the dependency index and returns whether the selection was successful
func selectNetworkTargetForService(sourceName string, targetIndex int) (targetName string, ok bool) {
	return networkdependencies.GetServicePairByServiceAndIndex(sourceName, targetIndex)
}

// getServiceAndTargetForNetworkChaos is a helper function that retrieves the source and target
// services for a network chaos specification
func getServiceAndTargetForNetworkChaos(appNameIndex int, targetIndex int) (sourceName, targetName string, ok bool) {
	// Get the app labels
	labelArr, err := labelsGetter(TargetNamespace, TargetLabelKey)
	if err != nil || appNameIndex < 0 || appNameIndex >= len(labelArr) {
		return "", "", false
	}

	sourceName = labelArr[appNameIndex]
	targetName, ok = selectNetworkTargetForService(sourceName, targetIndex)
	return sourceName, targetName, ok
}

// GetNetworkDependencies returns all available network dependencies for a service
func GetNetworkDependencies(serviceName string) []string {
	return networkdependencies.GetDependenciesForService(serviceName)
}

// ListNetworkServiceNames returns a list of all available services with network dependencies
func ListNetworkServiceNames() []string {
	return networkdependencies.ListAllServiceNames()
}

// GetAllNetworkPairs returns all service-to-service communication pairs
func GetAllNetworkPairs() []networkdependencies.ServiceDependency {
	return networkdependencies.GetAllServicePairs()
}
