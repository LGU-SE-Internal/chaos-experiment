package handler

import (
	"github.com/CUHK-SE-Group/chaos-experiment/internal/serviceendpoints"
)


// selectDNSPatternsForService selects server addresses to use as patterns for DNS chaos
// based on the service name and endpoint index
func selectDNSPatternsForService(serviceName string, endpointIndex int) ([]string, bool) {
	endpoints := endpointsGetter(serviceName)

	// Filter out endpoints with empty server addresses
	validEndpoints := make([]serviceendpoints.ServiceEndpoint, 0)
	for _, ep := range endpoints {
		if ep.ServerAddress != "" && ep.ServerAddress != serviceName {
			validEndpoints = append(validEndpoints, ep)
		}
	}

	if len(validEndpoints) == 0 {
		return []string{"*"}, false // Fallback to all domains
	}

	// Use a specific endpoint if the index is valid
	if endpointIndex >= 0 && endpointIndex < len(validEndpoints) {
		return []string{validEndpoints[endpointIndex].ServerAddress}, true
	}

	// Otherwise collect all unique server addresses
	uniqueAddresses := make(map[string]bool)
	for _, ep := range validEndpoints {
		uniqueAddresses[ep.ServerAddress] = true
	}

	patterns := make([]string, 0, len(uniqueAddresses))
	for addr := range uniqueAddresses {
		patterns = append(patterns, addr)
	}

	return patterns, true
}

// getServiceAndPatternsForDNSChaos retrieves the service name and DNS patterns
// for a DNS chaos experiment
func getServiceAndPatternsForDNSChaos(appNameIndex int, endpointIndex int) (serviceName string, patterns []string, ok bool) {
	// Get the app labels
	labelArr, err := labelsGetter(TargetNamespace, TargetLabelKey)
	if err != nil || appNameIndex < 0 || appNameIndex >= len(labelArr) {
		return "", []string{"*"}, false
	}

	serviceName = labelArr[appNameIndex]
	patterns, ok = selectDNSPatternsForService(serviceName, endpointIndex)
	return serviceName, patterns, ok
}

// GetDNSEndpoints returns all unique server addresses that can be targeted by DNS chaos
func GetDNSEndpoints(serviceName string) []string {
	endpoints := endpointsGetter(serviceName)
	uniqueAddresses := make(map[string]bool)

	for _, ep := range endpoints {
		if ep.ServerAddress != "" && ep.ServerAddress != serviceName {
			uniqueAddresses[ep.ServerAddress] = true
		}
	}

	result := make([]string, 0, len(uniqueAddresses))
	for addr := range uniqueAddresses {
		result = append(result, addr)
	}

	return result
}

// CountDNSEndpoints returns the number of unique server addresses for a service
func CountDNSEndpoints(serviceName string) int {
	return len(GetDNSEndpoints(serviceName))
}
