package handler

import (
	"fmt"

	"github.com/CUHK-SE-Group/chaos-experiment/internal/networkdependencies"
)

// selectNetworkTargetForService selects a target service for a given source service
// based on the dependency index and returns the target name or an error
func selectNetworkTargetForService(sourceName string, targetIndex int) (targetName string, err error) {
	targetName, ok := networkdependencies.GetServicePairByServiceAndIndex(sourceName, targetIndex)
	if !ok {
		return "", fmt.Errorf("no network target found for service %s at index %d",
			sourceName, targetIndex)
	}
	return targetName, nil
}

// getServiceAndTargetForNetworkChaos is a helper function that retrieves the source and target
// services for a network chaos specification
func getServiceAndTargetForNetworkChaos(appNameIndex int, targetIndex int) (sourceName, targetName string, err error) {
	// Get the app labels
	labelArr, err := labelsGetter(TargetNamespace, TargetLabelKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to get service labels: %w", err)
	}

	if appNameIndex < 0 || appNameIndex >= len(labelArr) {
		return "", "", fmt.Errorf("app index %d out of range (max: %d)",
			appNameIndex, len(labelArr)-1)
	}

	sourceName = labelArr[appNameIndex]
	targetName, err = selectNetworkTargetForService(sourceName, targetIndex)
	if err != nil {
		return sourceName, "", err
	}

	return sourceName, targetName, nil
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
