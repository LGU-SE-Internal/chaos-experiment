package resourcelookup

import (
	"sort"

	"github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/javaclassmethods"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/networkdependencies"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/serviceendpoints"
)

// Constants
const (
	DefaultNamespace = "ts"
	DefaultLabelKey  = "app"
)

// AppMethodPair represents a flattened app+method combination
type AppMethodPair struct {
	AppName    string
	ClassName  string
	MethodName string
}

// AppEndpointPair represents a flattened app+endpoint combination
type AppEndpointPair struct {
	AppName       string
	Route         string
	Method        string
	ServerAddress string
	ServerPort    string
}

// AppNetworkPair represents a flattened source+target combination for network chaos
type AppNetworkPair struct {
	SourceService string
	TargetService string
}

// AppDNSPair represents a flattened app+domain combination for DNS chaos
type AppDNSPair struct {
	AppName string
	Domain  string
}

// ContainerInfo represents container information with its pod and app
type ContainerInfo struct {
	PodName       string
	AppLabel      string
	ContainerName string
}

// Global cache for lookups
var (
	cachedAppLabels     []string
	cachedAppMethods    []AppMethodPair
	cachedAppEndpoints  []AppEndpointPair
	cachedNetworkPairs  []AppNetworkPair
	cachedDNSEndpoints  []AppDNSPair
	cachedContainerInfo []ContainerInfo
)

// GetAllAppLabels returns all application labels sorted alphabetically
func GetAllAppLabels() ([]string, error) {
	if cachedAppLabels != nil {
		return cachedAppLabels, nil
	}

	labels, err := client.GetLabels(DefaultNamespace, DefaultLabelKey)
	if err != nil {
		return nil, err
	}

	// Sort alphabetically
	sort.Strings(labels)
	cachedAppLabels = labels
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

// GetAllContainers returns all containers with their info sorted by app label
func GetAllContainers() ([]ContainerInfo, error) {
	if cachedContainerInfo != nil {
		return cachedContainerInfo, nil
	}

	containers, err := client.GetContainersWithAppLabel(DefaultNamespace)
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

	cachedContainerInfo = result
	return result, nil
}

// InvalidateCache clears all cached data
func InvalidateCache() {
	cachedAppLabels = nil
	cachedAppMethods = nil
	cachedAppEndpoints = nil
	cachedNetworkPairs = nil
	cachedDNSEndpoints = nil
	cachedContainerInfo = nil
}
