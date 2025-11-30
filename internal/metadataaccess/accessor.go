// Package metadataaccess provides a unified interface for accessing system-specific metadata.
// It uses the systemconfig package to determine the current system and routes calls to
// the appropriate underlying data package.
package metadataaccess

import (
	"sync"

	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"
)

// ServiceEndpoint represents a service endpoint from trace analysis
type ServiceEndpoint struct {
	ServiceName    string
	RequestMethod  string
	ResponseStatus string
	Route          string
	ServerAddress  string
	ServerPort     string
}

// DatabaseOperation represents a database operation from trace analysis
type DatabaseOperation struct {
	ServiceName   string
	DBName        string
	DBTable       string
	Operation     string
	DBSystem      string
	ServerAddress string
	ServerPort    string
}

// GRPCOperation represents a gRPC operation from trace analysis
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

// JavaClassMethod represents a Java class method for JVM chaos
type JavaClassMethod struct {
	ClassName  string
	MethodName string
}

// ServiceEndpointAccessor defines the interface for service endpoint access
type ServiceEndpointAccessor interface {
	GetEndpointsByService(serviceName string) []ServiceEndpoint
	GetAllServices() []string
}

// DatabaseOperationAccessor defines the interface for database operation access
type DatabaseOperationAccessor interface {
	GetOperationsByService(serviceName string) []DatabaseOperation
	GetAllDatabaseServices() []string
}

// GRPCOperationAccessor defines the interface for gRPC operation access
type GRPCOperationAccessor interface {
	GetOperationsByService(serviceName string) []GRPCOperation
	GetAllGRPCServices() []string
}

// JavaMethodAccessor defines the interface for Java method access
type JavaMethodAccessor interface {
	GetClassMethodsByService(serviceName string) []JavaClassMethod
	GetAllServices() []string
}

// MetadataAccessor provides unified access to system-specific metadata
type MetadataAccessor struct {
	mu sync.RWMutex

	// TrainTicket system accessors
	tsServiceEndpoints   ServiceEndpointAccessor
	tsDatabaseOperations DatabaseOperationAccessor
	tsJavaMethods        JavaMethodAccessor

	// OtelDemo system accessors
	otelServiceEndpoints   ServiceEndpointAccessor
	otelDatabaseOperations DatabaseOperationAccessor
	otelGRPCOperations     GRPCOperationAccessor
}

var (
	accessor     *MetadataAccessor
	accessorOnce sync.Once
)

// GetAccessor returns the singleton metadata accessor
func GetAccessor() *MetadataAccessor {
	accessorOnce.Do(func() {
		accessor = &MetadataAccessor{}
	})
	return accessor
}

// RegisterTrainTicketServiceEndpoints registers the TrainTicket service endpoint accessor
func (m *MetadataAccessor) RegisterTrainTicketServiceEndpoints(accessor ServiceEndpointAccessor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tsServiceEndpoints = accessor
}

// RegisterTrainTicketDatabaseOperations registers the TrainTicket database operation accessor
func (m *MetadataAccessor) RegisterTrainTicketDatabaseOperations(accessor DatabaseOperationAccessor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tsDatabaseOperations = accessor
}

// RegisterTrainTicketJavaMethods registers the TrainTicket Java method accessor
func (m *MetadataAccessor) RegisterTrainTicketJavaMethods(accessor JavaMethodAccessor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tsJavaMethods = accessor
}

// RegisterOtelDemoServiceEndpoints registers the OtelDemo service endpoint accessor
func (m *MetadataAccessor) RegisterOtelDemoServiceEndpoints(accessor ServiceEndpointAccessor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.otelServiceEndpoints = accessor
}

// RegisterOtelDemoDatabaseOperations registers the OtelDemo database operation accessor
func (m *MetadataAccessor) RegisterOtelDemoDatabaseOperations(accessor DatabaseOperationAccessor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.otelDatabaseOperations = accessor
}

// RegisterOtelDemoGRPCOperations registers the OtelDemo gRPC operation accessor
func (m *MetadataAccessor) RegisterOtelDemoGRPCOperations(accessor GRPCOperationAccessor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.otelGRPCOperations = accessor
}

// GetEndpointsByService returns endpoints for a service based on current system
func (m *MetadataAccessor) GetEndpointsByService(serviceName string) []ServiceEndpoint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemTrainTicket:
		if m.tsServiceEndpoints != nil {
			return m.tsServiceEndpoints.GetEndpointsByService(serviceName)
		}
	case systemconfig.SystemOtelDemo:
		if m.otelServiceEndpoints != nil {
			return m.otelServiceEndpoints.GetEndpointsByService(serviceName)
		}
	}
	return nil
}

// GetAllServices returns all services based on current system
func (m *MetadataAccessor) GetAllServices() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemTrainTicket:
		if m.tsServiceEndpoints != nil {
			return m.tsServiceEndpoints.GetAllServices()
		}
	case systemconfig.SystemOtelDemo:
		if m.otelServiceEndpoints != nil {
			return m.otelServiceEndpoints.GetAllServices()
		}
	}
	return nil
}

// GetDatabaseOperationsByService returns database operations for a service based on current system
func (m *MetadataAccessor) GetDatabaseOperationsByService(serviceName string) []DatabaseOperation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemTrainTicket:
		if m.tsDatabaseOperations != nil {
			return m.tsDatabaseOperations.GetOperationsByService(serviceName)
		}
	case systemconfig.SystemOtelDemo:
		if m.otelDatabaseOperations != nil {
			return m.otelDatabaseOperations.GetOperationsByService(serviceName)
		}
	}
	return nil
}

// GetAllDatabaseServices returns all database services based on current system
func (m *MetadataAccessor) GetAllDatabaseServices() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemTrainTicket:
		if m.tsDatabaseOperations != nil {
			return m.tsDatabaseOperations.GetAllDatabaseServices()
		}
	case systemconfig.SystemOtelDemo:
		if m.otelDatabaseOperations != nil {
			return m.otelDatabaseOperations.GetAllDatabaseServices()
		}
	}
	return nil
}

// GetGRPCOperationsByService returns gRPC operations for a service (OtelDemo only)
func (m *MetadataAccessor) GetGRPCOperationsByService(serviceName string) []GRPCOperation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// gRPC operations are primarily for OtelDemo
	if systemconfig.IsOtelDemo() && m.otelGRPCOperations != nil {
		return m.otelGRPCOperations.GetOperationsByService(serviceName)
	}
	return nil
}

// GetAllGRPCServices returns all gRPC services (OtelDemo only)
func (m *MetadataAccessor) GetAllGRPCServices() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// gRPC operations are primarily for OtelDemo
	if systemconfig.IsOtelDemo() && m.otelGRPCOperations != nil {
		return m.otelGRPCOperations.GetAllGRPCServices()
	}
	return nil
}

// GetJavaMethodsByService returns Java methods for a service (TrainTicket only)
func (m *MetadataAccessor) GetJavaMethodsByService(serviceName string) []JavaClassMethod {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Java methods are for TrainTicket (Java-based microservices)
	if systemconfig.IsTrainTicket() && m.tsJavaMethods != nil {
		return m.tsJavaMethods.GetClassMethodsByService(serviceName)
	}
	return nil
}

// GetAllJavaServices returns all Java services (TrainTicket only)
func (m *MetadataAccessor) GetAllJavaServices() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Java methods are for TrainTicket
	if systemconfig.IsTrainTicket() && m.tsJavaMethods != nil {
		return m.tsJavaMethods.GetAllServices()
	}
	return nil
}

// HasServiceEndpoints returns true if service endpoints are available for current system
func (m *MetadataAccessor) HasServiceEndpoints() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemTrainTicket:
		return m.tsServiceEndpoints != nil
	case systemconfig.SystemOtelDemo:
		return m.otelServiceEndpoints != nil
	}
	return false
}

// HasDatabaseOperations returns true if database operations are available for current system
func (m *MetadataAccessor) HasDatabaseOperations() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemTrainTicket:
		return m.tsDatabaseOperations != nil
	case systemconfig.SystemOtelDemo:
		return m.otelDatabaseOperations != nil
	}
	return false
}

// HasGRPCOperations returns true if gRPC operations are available for current system
func (m *MetadataAccessor) HasGRPCOperations() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return systemconfig.IsOtelDemo() && m.otelGRPCOperations != nil
}

// HasJavaMethods returns true if Java methods are available for current system
func (m *MetadataAccessor) HasJavaMethods() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return systemconfig.IsTrainTicket() && m.tsJavaMethods != nil
}

// Clear removes all registered accessors (useful for testing)
func (m *MetadataAccessor) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tsServiceEndpoints = nil
	m.tsDatabaseOperations = nil
	m.tsJavaMethods = nil
	m.otelServiceEndpoints = nil
	m.otelDatabaseOperations = nil
	m.otelGRPCOperations = nil
}
