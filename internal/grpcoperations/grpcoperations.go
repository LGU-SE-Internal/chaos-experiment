// Package grpcoperations provides a system-aware routing layer for gRPC operation data.
// This package delegates to the appropriate system-specific package based on the current system configuration.
// Note: gRPC operations are primarily used in OtelDemo; TrainTicket uses HTTP.
package grpcoperations

import (
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"

	oteldemogrpc "github.com/LGU-SE-Internal/chaos-experiment/internal/oteldemo/grpcoperations"
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
