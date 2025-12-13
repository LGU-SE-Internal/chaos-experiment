// Package model defines the unified data model for all system metadata.
// This package provides a single structure that contains all endpoint types
// (HTTP, gRPC, Database) for a given system, eliminating the need for
// separate packages per endpoint type.
package model

import (
	"github.com/LGU-SE-Internal/chaos-experiment/internal/resourcetypes"
)

// SystemData represents all metadata for a single system.
// This unified structure contains all endpoint types and eliminates
// the need for separate serviceendpoints, databaseoperations, and
// grpcoperations packages per system.
type SystemData struct {
	// SystemName identifies the system (e.g., "ts", "otel-demo", "hs")
	SystemName string

	// ServiceEndpoints maps service names to their HTTP/REST endpoints
	ServiceEndpoints map[string][]resourcetypes.ServiceEndpoint

	// DatabaseOperations maps service names to their database operations
	DatabaseOperations map[string][]resourcetypes.DatabaseOperation

	// GRPCOperations maps service names to their gRPC operations
	GRPCOperations map[string][]resourcetypes.GRPCOperation

	// AllServices contains all unique service names (both callers and callees)
	AllServices []string
}

// GetEndpointsByService returns all HTTP endpoints for a service
func (sd *SystemData) GetEndpointsByService(serviceName string) []resourcetypes.ServiceEndpoint {
	if endpoints, exists := sd.ServiceEndpoints[serviceName]; exists {
		return endpoints
	}
	return []resourcetypes.ServiceEndpoint{}
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

// GetGRPCOperationsByService returns all gRPC operations for a service
func (sd *SystemData) GetGRPCOperationsByService(serviceName string) []resourcetypes.GRPCOperation {
	if operations, exists := sd.GRPCOperations[serviceName]; exists {
		return operations
	}
	return []resourcetypes.GRPCOperation{}
}

// GetAllGRPCServices returns a list of all services that perform gRPC operations
func (sd *SystemData) GetAllGRPCServices() []string {
	services := make([]string, 0, len(sd.GRPCOperations))
	for service := range sd.GRPCOperations {
		services = append(services, service)
	}
	return services
}

// GetClientGRPCOperations returns all client-side gRPC operations
func (sd *SystemData) GetClientGRPCOperations() []resourcetypes.GRPCOperation {
	var results []resourcetypes.GRPCOperation
	for _, operations := range sd.GRPCOperations {
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
