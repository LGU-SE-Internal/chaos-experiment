// Package serviceendpoints provides system-aware service endpoint data access.
// This file provides system-aware wrapper functions that delegate to the correct
// system-specific data based on the current system configuration.
package serviceendpoints

import (
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"

	oteldemoendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/oteldemo/serviceendpoints"
	tsendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/ts/serviceendpoints"
)

// GetEndpointsByServiceSystemAware returns all endpoints for a service based on current system
func GetEndpointsByServiceSystemAware(serviceName string) []ServiceEndpoint {
	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemTrainTicket:
		tsEndpoints := tsendpoints.GetEndpointsByService(serviceName)
		return convertTSEndpoints(tsEndpoints)
	case systemconfig.SystemOtelDemo:
		otelEndpoints := oteldemoendpoints.GetEndpointsByService(serviceName)
		return convertOtelDemoEndpoints(otelEndpoints)
	default:
		// Fall back to the hardcoded data in this package
		if endpoints, exists := ServiceEndpoints[serviceName]; exists {
			return endpoints
		}
		return []ServiceEndpoint{}
	}
}

// GetAllServicesSystemAware returns all service names based on current system
func GetAllServicesSystemAware() []string {
	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemTrainTicket:
		return tsendpoints.GetAllServices()
	case systemconfig.SystemOtelDemo:
		return oteldemoendpoints.GetAllServices()
	default:
		// Fall back to the hardcoded data in this package
		services := make([]string, 0, len(ServiceEndpoints))
		for service := range ServiceEndpoints {
			services = append(services, service)
		}
		return services
	}
}

// convertTSEndpoints converts TrainTicket endpoints to the common type
func convertTSEndpoints(tsEndpoints []tsendpoints.ServiceEndpoint) []ServiceEndpoint {
	result := make([]ServiceEndpoint, len(tsEndpoints))
	for i, ep := range tsEndpoints {
		result[i] = ServiceEndpoint{
			ServiceName:    ep.ServiceName,
			RequestMethod:  ep.RequestMethod,
			ResponseStatus: ep.ResponseStatus,
			Route:          ep.Route,
			ServerAddress:  ep.ServerAddress,
			ServerPort:     ep.ServerPort,
			SpanName:       ep.SpanName,
		}
	}
	return result
}

// convertOtelDemoEndpoints converts OtelDemo endpoints to the common type
func convertOtelDemoEndpoints(otelEndpoints []oteldemoendpoints.ServiceEndpoint) []ServiceEndpoint {
	result := make([]ServiceEndpoint, len(otelEndpoints))
	for i, ep := range otelEndpoints {
		result[i] = ServiceEndpoint{
			ServiceName:    ep.ServiceName,
			RequestMethod:  ep.RequestMethod,
			ResponseStatus: ep.ResponseStatus,
			Route:          ep.Route,
			ServerAddress:  ep.ServerAddress,
			ServerPort:     ep.ServerPort,
			SpanName:       ep.SpanName,
		}
	}
	return result
}
