// Package serviceendpoints provides a system-aware routing layer for service endpoint data.
// This package delegates to the appropriate system-specific package based on the current system configuration.
package serviceendpoints

import (
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"

	hsendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/hs/serviceendpoints"
	mediaendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/media/serviceendpoints"
	obendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/ob/serviceendpoints"
	oteldemoendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/oteldemo/serviceendpoints"
	snendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/sn/serviceendpoints"
	sockshopendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/sockshop/serviceendpoints"
	teastoreendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/teastore/serviceendpoints"
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
	case systemconfig.SystemMediaMicroservices:
		mediaEps := mediaendpoints.GetEndpointsByService(serviceName)
		return convertMediaEndpoints(mediaEps)
	case systemconfig.SystemHotelReservation:
		hsEps := hsendpoints.GetEndpointsByService(serviceName)
		return convertHSEndpoints(hsEps)
	case systemconfig.SystemSocialNetwork:
		snEps := snendpoints.GetEndpointsByService(serviceName)
		return convertSNEndpoints(snEps)
	case systemconfig.SystemOnlineBoutique:
		obEps := obendpoints.GetEndpointsByService(serviceName)
		return convertOBEndpoints(obEps)
	case systemconfig.SystemSockShop:
		sockshopEps := sockshopendpoints.GetEndpointsByService(serviceName)
		return convertSockShopEndpoints(sockshopEps)
	case systemconfig.SystemTeaStore:
		teastoreEps := teastoreendpoints.GetEndpointsByService(serviceName)
		return convertTeaStoreEndpoints(teastoreEps)
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
	case systemconfig.SystemMediaMicroservices:
		return mediaendpoints.GetAllServices()
	case systemconfig.SystemHotelReservation:
		return hsendpoints.GetAllServices()
	case systemconfig.SystemSocialNetwork:
		return snendpoints.GetAllServices()
	case systemconfig.SystemOnlineBoutique:
		return obendpoints.GetAllServices()
	case systemconfig.SystemSockShop:
		return sockshopendpoints.GetAllServices()
	case systemconfig.SystemTeaStore:
		return teastoreendpoints.GetAllServices()
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

// convertMediaEndpoints converts media-specific endpoints to the common type
func convertMediaEndpoints(mediaEps []mediaendpoints.ServiceEndpoint) []ServiceEndpoint {
	result := make([]ServiceEndpoint, len(mediaEps))
	for i, ep := range mediaEps {
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

// convertHSEndpoints converts hs-specific endpoints to the common type
func convertHSEndpoints(hsEps []hsendpoints.ServiceEndpoint) []ServiceEndpoint {
	result := make([]ServiceEndpoint, len(hsEps))
	for i, ep := range hsEps {
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

// convertSNEndpoints converts sn-specific endpoints to the common type
func convertSNEndpoints(snEps []snendpoints.ServiceEndpoint) []ServiceEndpoint {
	result := make([]ServiceEndpoint, len(snEps))
	for i, ep := range snEps {
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

// convertOBEndpoints converts ob-specific endpoints to the common type
func convertOBEndpoints(obEps []obendpoints.ServiceEndpoint) []ServiceEndpoint {
	result := make([]ServiceEndpoint, len(obEps))
	for i, ep := range obEps {
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

// convertSockShopEndpoints converts sockshop-specific endpoints to the common type
func convertSockShopEndpoints(sockshopEps []sockshopendpoints.ServiceEndpoint) []ServiceEndpoint {
	result := make([]ServiceEndpoint, len(sockshopEps))
	for i, ep := range sockshopEps {
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

// convertTeaStoreEndpoints converts teastore-specific endpoints to the common type
func convertTeaStoreEndpoints(teastoreEps []teastoreendpoints.ServiceEndpoint) []ServiceEndpoint {
	result := make([]ServiceEndpoint, len(teastoreEps))
	for i, ep := range teastoreEps {
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
