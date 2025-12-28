// Package grpcoperations provides a system-aware routing layer for gRPC operation data.
// This package delegates to the appropriate system-specific package based on the current system configuration.
// Note: gRPC operations are primarily used in OtelDemo; TrainTicket uses HTTP.
package grpcoperations

import (
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"

	hsgrpc "github.com/LGU-SE-Internal/chaos-experiment/internal/hs/grpcoperations"
	mediagrpc "github.com/LGU-SE-Internal/chaos-experiment/internal/media/grpcoperations"
	oteldemogrpc "github.com/LGU-SE-Internal/chaos-experiment/internal/oteldemo/grpcoperations"
	obgrpc "github.com/LGU-SE-Internal/chaos-experiment/internal/ob/grpcoperations"
	sngrpc "github.com/LGU-SE-Internal/chaos-experiment/internal/sn/grpcoperations"
	sockshopgrpc "github.com/LGU-SE-Internal/chaos-experiment/internal/sockshop/grpcoperations"
	teastoregrpc "github.com/LGU-SE-Internal/chaos-experiment/internal/teastore/grpcoperations"
)

// GRPCOperation represents a gRPC operation from ClickHouse analysis
type GRPCOperation struct {
	ServiceName    string
	RPCSystem      string
	RPCService     string
	RPCMethod      string
	GRPCStatusCode string
	ServerAddress  string
	ServerPort     string
	SpanKind       string
}

// GetOperationsByService returns all gRPC operations for a service based on current system
func GetOperationsByService(serviceName string) []GRPCOperation {
	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemOtelDemo:
		otelOps := oteldemogrpc.GetOperationsByService(serviceName)
		return convertOtelDemoOperations(otelOps)
	case systemconfig.SystemMediaMicroservices:
		mediaOps := mediagrpc.GetOperationsByService(serviceName)
		return convertMediaOperations(mediaOps)
	case systemconfig.SystemHotelReservation:
		hsOps := hsgrpc.GetOperationsByService(serviceName)
		return convertHSOperations(hsOps)
	case systemconfig.SystemSocialNetwork:
		snOps := sngrpc.GetOperationsByService(serviceName)
		return convertSNOperations(snOps)
	case systemconfig.SystemOnlineBoutique:
		obOps := obgrpc.GetOperationsByService(serviceName)
		return convertOBOperations(obOps)
	case systemconfig.SystemSockShop:
		sockshopOps := sockshopgrpc.GetOperationsByService(serviceName)
		return convertSockShopOperations(sockshopOps)
	case systemconfig.SystemTeaStore:
		teastoreOps := teastoregrpc.GetOperationsByService(serviceName)
		return convertTeaStoreOperations(teastoreOps)
	default:
		// TrainTicket doesn't have gRPC operations
		return []GRPCOperation{}
	}
}

// GetAllGRPCServices returns a list of all services that perform gRPC operations based on current system
func GetAllGRPCServices() []string {
	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemOtelDemo:
		return oteldemogrpc.GetAllGRPCServices()
	case systemconfig.SystemMediaMicroservices:
		return mediagrpc.GetAllGRPCServices()
	case systemconfig.SystemHotelReservation:
		return hsgrpc.GetAllGRPCServices()
	case systemconfig.SystemSocialNetwork:
		return sngrpc.GetAllGRPCServices()
	default:
		// TrainTicket doesn't have gRPC operations
		return []string{}
	}
}

// GetClientOperations returns all client-side gRPC operations based on current system
func GetClientOperations() []GRPCOperation {
	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemOtelDemo:
		otelOps := oteldemogrpc.GetClientOperations()
		return convertOtelDemoOperations(otelOps)
	case systemconfig.SystemMediaMicroservices:
		mediaOps := mediagrpc.GetClientOperations()
		return convertMediaOperations(mediaOps)
	case systemconfig.SystemHotelReservation:
		hsOps := hsgrpc.GetClientOperations()
		return convertHSOperations(hsOps)
	case systemconfig.SystemSocialNetwork:
		snOps := sngrpc.GetClientOperations()
		return convertSNOperations(snOps)
	default:
		// TrainTicket doesn't have gRPC operations
		return []GRPCOperation{}
	}
}

// GetServerOperations returns all server-side gRPC operations based on current system
func GetServerOperations() []GRPCOperation {
	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemOtelDemo:
		otelOps := oteldemogrpc.GetServerOperations()
		return convertOtelDemoOperations(otelOps)
	case systemconfig.SystemMediaMicroservices:
		mediaOps := mediagrpc.GetServerOperations()
		return convertMediaOperations(mediaOps)
	case systemconfig.SystemHotelReservation:
		hsOps := hsgrpc.GetServerOperations()
		return convertHSOperations(hsOps)
	case systemconfig.SystemSocialNetwork:
		snOps := sngrpc.GetServerOperations()
		return convertSNOperations(snOps)
	default:
		// TrainTicket doesn't have gRPC operations
		return []GRPCOperation{}
	}
}

// GetOperationsByRPCService returns all operations for a specific RPC service based on current system
func GetOperationsByRPCService(rpcService string) []GRPCOperation {
	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemOtelDemo:
		otelOps := oteldemogrpc.GetOperationsByRPCService(rpcService)
		return convertOtelDemoOperations(otelOps)
	case systemconfig.SystemMediaMicroservices:
		mediaOps := mediagrpc.GetOperationsByRPCService(rpcService)
		return convertMediaOperations(mediaOps)
	case systemconfig.SystemHotelReservation:
		hsOps := hsgrpc.GetOperationsByRPCService(rpcService)
		return convertHSOperations(hsOps)
	case systemconfig.SystemSocialNetwork:
		snOps := sngrpc.GetOperationsByRPCService(rpcService)
		return convertSNOperations(snOps)
	default:
		// TrainTicket doesn't have gRPC operations
		return []GRPCOperation{}
	}
}

// convertOtelDemoOperations converts otel-demo-specific operations to the common type
func convertOtelDemoOperations(otelOps []oteldemogrpc.GRPCOperation) []GRPCOperation {
	result := make([]GRPCOperation, len(otelOps))
	for i, op := range otelOps {
		result[i] = GRPCOperation{
			ServiceName:    op.ServiceName,
			RPCSystem:      op.RPCSystem,
			RPCService:     op.RPCService,
			RPCMethod:      op.RPCMethod,
			GRPCStatusCode: op.GRPCStatusCode,
			ServerAddress:  op.ServerAddress,
			ServerPort:     op.ServerPort,
			SpanKind:       op.SpanKind,
		}
	}
	return result
}

// convertMediaOperations converts media-specific operations to the common type
func convertMediaOperations(mediaOps []mediagrpc.GRPCOperation) []GRPCOperation {
	result := make([]GRPCOperation, len(mediaOps))
	for i, op := range mediaOps {
		result[i] = GRPCOperation{
			ServiceName:    op.ServiceName,
			RPCSystem:      op.RPCSystem,
			RPCService:     op.RPCService,
			RPCMethod:      op.RPCMethod,
			GRPCStatusCode: op.GRPCStatusCode,
			ServerAddress:  op.ServerAddress,
			ServerPort:     op.ServerPort,
			SpanKind:       op.SpanKind,
		}
	}
	return result
}

// convertHSOperations converts hs-specific operations to the common type
func convertHSOperations(hsOps []hsgrpc.GRPCOperation) []GRPCOperation {
	result := make([]GRPCOperation, len(hsOps))
	for i, op := range hsOps {
		result[i] = GRPCOperation{
			ServiceName:    op.ServiceName,
			RPCSystem:      op.RPCSystem,
			RPCService:     op.RPCService,
			RPCMethod:      op.RPCMethod,
			GRPCStatusCode: op.GRPCStatusCode,
			ServerAddress:  op.ServerAddress,
			ServerPort:     op.ServerPort,
			SpanKind:       op.SpanKind,
		}
	}
	return result
}

// convertSNOperations converts sn-specific operations to the common type
func convertSNOperations(snOps []sngrpc.GRPCOperation) []GRPCOperation {
	result := make([]GRPCOperation, len(snOps))
	for i, op := range snOps {
		result[i] = GRPCOperation{
			ServiceName:    op.ServiceName,
			RPCSystem:      op.RPCSystem,
			RPCService:     op.RPCService,
			RPCMethod:      op.RPCMethod,
			GRPCStatusCode: op.GRPCStatusCode,
			ServerAddress:  op.ServerAddress,
			ServerPort:     op.ServerPort,
			SpanKind:       op.SpanKind,
		}
	}
	return result
}
// convertOBOperations converts ob-specific operations to the common type
func convertOBOperations(obOps []obgrpc.GRPCOperation) []GRPCOperation {
result := make([]GRPCOperation, len(obOps))
for i, op := range obOps {
result[i] = GRPCOperation{
ServiceName:    op.ServiceName,
RPCSystem:      op.RPCSystem,
RPCService:     op.RPCService,
RPCMethod:      op.RPCMethod,
GRPCStatusCode: op.GRPCStatusCode,
ServerAddress:  op.ServerAddress,
ServerPort:     op.ServerPort,
SpanKind:       op.SpanKind,
}
}
return result
}

// convertSockShopOperations converts sockshop-specific operations to the common type
func convertSockShopOperations(sockshopOps []sockshopgrpc.GRPCOperation) []GRPCOperation {
	result := make([]GRPCOperation, len(sockshopOps))
	for i, op := range sockshopOps {
		result[i] = GRPCOperation{
			ServiceName:    op.ServiceName,
			RPCSystem:      op.RPCSystem,
			RPCService:     op.RPCService,
			RPCMethod:      op.RPCMethod,
			GRPCStatusCode: op.GRPCStatusCode,
			ServerAddress:  op.ServerAddress,
			ServerPort:     op.ServerPort,
			SpanKind:       op.SpanKind,
		}
	}
	return result
}

// convertTeaStoreOperations converts teastore-specific operations to the common type
func convertTeaStoreOperations(teastoreOps []teastoregrpc.GRPCOperation) []GRPCOperation {
	result := make([]GRPCOperation, len(teastoreOps))
	for i, op := range teastoreOps {
		result[i] = GRPCOperation{
			ServiceName:    op.ServiceName,
			RPCSystem:      op.RPCSystem,
			RPCService:     op.RPCService,
			RPCMethod:      op.RPCMethod,
			GRPCStatusCode: op.GRPCStatusCode,
			ServerAddress:  op.ServerAddress,
			ServerPort:     op.ServerPort,
			SpanKind:       op.SpanKind,
		}
	}
	return result
}


// IsGRPCRoutePattern checks if a route looks like a gRPC route pattern
// gRPC routes typically follow the format: /package.Service/Method
// Examples: /oteldemo.CartService/AddItem, /flagd.evaluation.v1.Service/EventStream
func IsGRPCRoutePattern(route string) bool {
	if route == "" || len(route) < 3 {
		return false
	}
	// gRPC routes start with / and contain package.Service/Method pattern
	if route[0] != '/' {
		return false
	}
	// Look for patterns like /oteldemo.CartService/AddItem
	// These have a dot in the first segment (before second slash)
	hasDot := false
	for i := 1; i < len(route); i++ {
		if route[i] == '/' {
			break
		}
		if route[i] == '.' {
			hasDot = true
		}
	}
	return hasDot
}
