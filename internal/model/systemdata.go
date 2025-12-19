// Package model defines the unified data model for all system metadata.
// This package provides a single structure that contains all endpoint types
// (HTTP, RPC, Database) for a given system.
package model

import (
	"github.com/LGU-SE-Internal/chaos-experiment/internal/resourcetypes"
)

// SystemData represents all metadata for a single system.
// This unified structure contains all endpoint types, stored separately
// by their type (HTTP, RPC, Database).
type SystemData struct {
	// SystemName identifies the system (e.g., "ts", "otel-demo", "hs")
	SystemName string

	// HTTPEndpoints maps service names to their HTTP/REST endpoints
	HTTPEndpoints map[string][]resourcetypes.HTTPEndpoint

	// DatabaseOperations maps service names to their database operations
	DatabaseOperations map[string][]resourcetypes.DatabaseOperation

	// RPCOperations maps service names to their RPC/gRPC operations
	RPCOperations map[string][]resourcetypes.RPCOperation

	// AllServices contains all unique service names (both callers and callees)
	AllServices []string
}

// GetHTTPEndpointsByService returns all HTTP endpoints for a service
func (sd *SystemData) GetHTTPEndpointsByService(serviceName string) []resourcetypes.HTTPEndpoint {
	if endpoints, exists := sd.HTTPEndpoints[serviceName]; exists {
		return endpoints
	}
	return []resourcetypes.HTTPEndpoint{}
}

// GetAllServices returns a list of all available service names
func (sd *SystemData) GetAllServices() []string {
	return sd.AllServices
}

// GetDatabaseOperationsByService returns all database operations for a service
func (sd *SystemData) GetDatabaseOperationsByService(serviceName string) []resourcetypes.DatabaseOperation {
	if operations, exists := sd.DatabaseOperations[serviceName]; exists {
		return operations
	}
	return []resourcetypes.DatabaseOperation{}
}

// GetAllDatabaseServices returns a list of all services that perform database operations
func (sd *SystemData) GetAllDatabaseServices() []string {
	services := make([]string, 0, len(sd.DatabaseOperations))
	for service := range sd.DatabaseOperations {
		services = append(services, service)
	}
	return services
}

// GetRPCOperationsByService returns all RPC operations for a service
func (sd *SystemData) GetRPCOperationsByService(serviceName string) []resourcetypes.RPCOperation {
	if operations, exists := sd.RPCOperations[serviceName]; exists {
		return operations
	}
	return []resourcetypes.RPCOperation{}
}

// GetAllRPCServices returns a list of all services that perform RPC operations
func (sd *SystemData) GetAllRPCServices() []string {
	services := make([]string, 0, len(sd.RPCOperations))
	for service := range sd.RPCOperations {
		services = append(services, service)
	}
	return services
}

// GetClientRPCOperations returns all client-side RPC operations
func (sd *SystemData) GetClientRPCOperations() []resourcetypes.RPCOperation {
	var results []resourcetypes.RPCOperation
	for _, operations := range sd.RPCOperations {
		for _, op := range operations {
			if op.SpanKind == "Client" {
				results = append(results, op)
			}
		}
	}
	return results
}

// GetDatabaseOperationsByDBSystem returns all operations for a specific database system
func (sd *SystemData) GetDatabaseOperationsByDBSystem(dbSystem string) []resourcetypes.DatabaseOperation {
	var results []resourcetypes.DatabaseOperation
	for _, operations := range sd.DatabaseOperations {
		for _, op := range operations {
			if op.DBSystem == dbSystem {
				results = append(results, op)
			}
		}
	}
	return results
}
