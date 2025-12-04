// Package serviceendpoints provides a system-aware routing layer for service endpoint data.
// This package delegates to the appropriate system-specific package based on the current system configuration.
package serviceendpoints

import (
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"

	oteldemoendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/oteldemo/serviceendpoints"
	tsendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/ts/serviceendpoints"
)

// ServiceEndpoint represents a service endpoint from ClickHouse analysis
type ServiceEndpoint struct {
	ServiceName    string
	RequestMethod  string
	ResponseStatus string
	Route          string
	ServerAddress  string
	ServerPort     string
	SpanName       string
}

// GetEndpointsByService returns all endpoints for a service based on current system
func GetEndpointsByService(serviceName string) []ServiceEndpoint {
	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemTrainTicket:
		tsEps := tsendpoints.GetEndpointsByService(serviceName)
		return convertTSEndpoints(tsEps)
	case systemconfig.SystemOtelDemo:
		otelEps := oteldemoendpoints.GetEndpointsByService(serviceName)
		return convertOtelDemoEndpoints(otelEps)
	default:
		// Default to TrainTicket
		tsEps := tsendpoints.GetEndpointsByService(serviceName)
		return convertTSEndpoints(tsEps)
	}
}

// GetAllServices returns a list of all available service names based on current system
func GetAllServices() []string {
	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemTrainTicket:
		return tsendpoints.GetAllServices()
	case systemconfig.SystemOtelDemo:
		return oteldemoendpoints.GetAllServices()
	default:
		// Default to TrainTicket
		return tsendpoints.GetAllServices()
	}
}

// convertTSEndpoints converts ts-specific endpoints to the common type
func convertTSEndpoints(tsEps []tsendpoints.ServiceEndpoint) []ServiceEndpoint {
	result := make([]ServiceEndpoint, len(tsEps))
	for i, ep := range tsEps {
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

// convertOtelDemoEndpoints converts otel-demo-specific endpoints to the common type
func convertOtelDemoEndpoints(otelEps []oteldemoendpoints.ServiceEndpoint) []ServiceEndpoint {
	result := make([]ServiceEndpoint, len(otelEps))
	for i, ep := range otelEps {
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
