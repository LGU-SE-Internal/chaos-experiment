package systemconfig

import (
	"fmt"
	"sync"
)

// MetadataType represents the type of metadata
type MetadataType string

const (
	// MetadataServiceEndpoints represents service endpoint metadata
	MetadataServiceEndpoints MetadataType = "service_endpoints"
	// MetadataDatabaseOperations represents database operation metadata
	MetadataDatabaseOperations MetadataType = "database_operations"
	// MetadataJavaClassMethods represents Java class method metadata
	MetadataJavaClassMethods MetadataType = "java_class_methods"
	// MetadataNetworkDependencies represents network dependency metadata
	MetadataNetworkDependencies MetadataType = "network_dependencies"
	// MetadataGRPCOperations represents gRPC operation metadata (primarily for OtelDemo)
	MetadataGRPCOperations MetadataType = "grpc_operations"
)

// MetadataProvider is an interface for providing system-specific metadata
type MetadataProvider interface {
	// GetServiceNames returns a list of all service names
	GetServiceNames() []string
}

// ServiceEndpointProvider provides service endpoint data
type ServiceEndpointProvider interface {
	MetadataProvider
	// GetEndpointsByService returns endpoints for a specific service
	GetEndpointsByService(serviceName string) []ServiceEndpointData
}

// ServiceEndpointData represents a service endpoint
type ServiceEndpointData struct {
	ServiceName    string
	RequestMethod  string
	ResponseStatus string
	Route          string
	ServerAddress  string
	ServerPort     string
}

// DatabaseOperationProvider provides database operation data
type DatabaseOperationProvider interface {
	MetadataProvider
	// GetOperationsByService returns database operations for a specific service
	GetOperationsByService(serviceName string) []DatabaseOperationData
}

// DatabaseOperationData represents a database operation
type DatabaseOperationData struct {
	ServiceName   string
	DBName        string
	DBTable       string
	Operation     string
	DBSystem      string
	ServerAddress string
	ServerPort    string
}

// GRPCOperationProvider provides gRPC operation data
type GRPCOperationProvider interface {
	MetadataProvider
	// GetOperationsByService returns gRPC operations for a specific service
	GetOperationsByService(serviceName string) []GRPCOperationData
}

// GRPCOperationData represents a gRPC operation
type GRPCOperationData struct {
	ServiceName    string
	RPCSystem      string
	RPCService     string
	RPCMethod      string
	GRPCStatusCode string
	ServerAddress  string
	ServerPort     string
	SpanKind       string
}

// MetadataRegistry holds registered metadata providers for each system
type MetadataRegistry struct {
	mu                 sync.RWMutex
	serviceEndpoints   map[SystemType]ServiceEndpointProvider
	databaseOperations map[SystemType]DatabaseOperationProvider
	grpcOperations     map[SystemType]GRPCOperationProvider
}

var (
	// globalRegistry is the singleton metadata registry
	globalRegistry *MetadataRegistry
	registryOnce   sync.Once
)

// GetRegistry returns the global metadata registry
func GetRegistry() *MetadataRegistry {
	registryOnce.Do(func() {
		globalRegistry = &MetadataRegistry{
			serviceEndpoints:   make(map[SystemType]ServiceEndpointProvider),
			databaseOperations: make(map[SystemType]DatabaseOperationProvider),
			grpcOperations:     make(map[SystemType]GRPCOperationProvider),
		}
	})
	return globalRegistry
}

// RegisterServiceEndpointProvider registers a service endpoint provider for a system
func (r *MetadataRegistry) RegisterServiceEndpointProvider(system SystemType, provider ServiceEndpointProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.serviceEndpoints[system] = provider
}

// RegisterDatabaseOperationProvider registers a database operation provider for a system
func (r *MetadataRegistry) RegisterDatabaseOperationProvider(system SystemType, provider DatabaseOperationProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.databaseOperations[system] = provider
}

// RegisterGRPCOperationProvider registers a gRPC operation provider for a system
func (r *MetadataRegistry) RegisterGRPCOperationProvider(system SystemType, provider GRPCOperationProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.grpcOperations[system] = provider
}

// GetServiceEndpointProvider returns the service endpoint provider for the current system
func (r *MetadataRegistry) GetServiceEndpointProvider() (ServiceEndpointProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	system := GetCurrentSystem()
	provider, exists := r.serviceEndpoints[system]
	if !exists {
		return nil, fmt.Errorf("no service endpoint provider registered for system: %s", system)
	}
	return provider, nil
}

// GetDatabaseOperationProvider returns the database operation provider for the current system
func (r *MetadataRegistry) GetDatabaseOperationProvider() (DatabaseOperationProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	system := GetCurrentSystem()
	provider, exists := r.databaseOperations[system]
	if !exists {
		return nil, fmt.Errorf("no database operation provider registered for system: %s", system)
	}
	return provider, nil
}

// GetGRPCOperationProvider returns the gRPC operation provider for the current system
func (r *MetadataRegistry) GetGRPCOperationProvider() (GRPCOperationProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	system := GetCurrentSystem()
	provider, exists := r.grpcOperations[system]
	if !exists {
		return nil, fmt.Errorf("no gRPC operation provider registered for system: %s", system)
	}
	return provider, nil
}

// HasServiceEndpointProvider checks if a service endpoint provider is registered for the current system
func (r *MetadataRegistry) HasServiceEndpointProvider() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.serviceEndpoints[GetCurrentSystem()]
	return exists
}

// HasDatabaseOperationProvider checks if a database operation provider is registered for the current system
func (r *MetadataRegistry) HasDatabaseOperationProvider() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.databaseOperations[GetCurrentSystem()]
	return exists
}

// HasGRPCOperationProvider checks if a gRPC operation provider is registered for the current system
func (r *MetadataRegistry) HasGRPCOperationProvider() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.grpcOperations[GetCurrentSystem()]
	return exists
}

// Clear removes all registered providers (useful for testing)
func (r *MetadataRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.serviceEndpoints = make(map[SystemType]ServiceEndpointProvider)
	r.databaseOperations = make(map[SystemType]DatabaseOperationProvider)
	r.grpcOperations = make(map[SystemType]GRPCOperationProvider)
}
