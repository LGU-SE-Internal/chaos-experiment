package systemconfig

import (
	"testing"
)

// MockServiceEndpointProvider is a mock implementation for testing
type MockServiceEndpointProvider struct {
	services  []string
	endpoints map[string][]ServiceEndpointData
}

func (m *MockServiceEndpointProvider) GetServiceNames() []string {
	return m.services
}

func (m *MockServiceEndpointProvider) GetEndpointsByService(serviceName string) []ServiceEndpointData {
	if m.endpoints == nil {
		return nil
	}
	return m.endpoints[serviceName]
}

// MockDatabaseOperationProvider is a mock implementation for testing
type MockDatabaseOperationProvider struct {
	services   []string
	operations map[string][]DatabaseOperationData
}

func (m *MockDatabaseOperationProvider) GetServiceNames() []string {
	return m.services
}

func (m *MockDatabaseOperationProvider) GetOperationsByService(serviceName string) []DatabaseOperationData {
	if m.operations == nil {
		return nil
	}
	return m.operations[serviceName]
}

// MockGRPCOperationProvider is a mock implementation for testing
type MockGRPCOperationProvider struct {
	services   []string
	operations map[string][]GRPCOperationData
}

func (m *MockGRPCOperationProvider) GetServiceNames() []string {
	return m.services
}

func (m *MockGRPCOperationProvider) GetOperationsByService(serviceName string) []GRPCOperationData {
	if m.operations == nil {
		return nil
	}
	return m.operations[serviceName]
}

func TestMetadataRegistry(t *testing.T) {
	// Reset the registry and system for testing
	registry := GetRegistry()
	registry.Clear()
	_ = SetCurrentSystem(SystemTrainTicket)

	// Create mock providers
	tsEndpointProvider := &MockServiceEndpointProvider{
		services: []string{"ts-service-1", "ts-service-2"},
		endpoints: map[string][]ServiceEndpointData{
			"ts-service-1": {
				{ServiceName: "ts-service-1", Route: "/api/v1/test", RequestMethod: "GET"},
			},
		},
	}

	otelEndpointProvider := &MockServiceEndpointProvider{
		services: []string{"otel-service-1", "otel-service-2"},
		endpoints: map[string][]ServiceEndpointData{
			"otel-service-1": {
				{ServiceName: "otel-service-1", Route: "/api/products", RequestMethod: "GET"},
			},
		},
	}

	// Register providers
	registry.RegisterServiceEndpointProvider(SystemTrainTicket, tsEndpointProvider)
	registry.RegisterServiceEndpointProvider(SystemOtelDemo, otelEndpointProvider)

	// Test getting provider for TrainTicket system
	_ = SetCurrentSystem(SystemTrainTicket)
	provider, err := registry.GetServiceEndpointProvider()
	if err != nil {
		t.Fatalf("GetServiceEndpointProvider() error = %v", err)
	}

	services := provider.GetServiceNames()
	if len(services) != 2 {
		t.Errorf("Expected 2 services for TrainTicket, got %d", len(services))
	}
	if services[0] != "ts-service-1" {
		t.Errorf("Expected first service to be ts-service-1, got %s", services[0])
	}

	// Test getting provider for OtelDemo system
	_ = SetCurrentSystem(SystemOtelDemo)
	provider, err = registry.GetServiceEndpointProvider()
	if err != nil {
		t.Fatalf("GetServiceEndpointProvider() error = %v", err)
	}

	services = provider.GetServiceNames()
	if len(services) != 2 {
		t.Errorf("Expected 2 services for OtelDemo, got %d", len(services))
	}
	if services[0] != "otel-service-1" {
		t.Errorf("Expected first service to be otel-service-1, got %s", services[0])
	}
}

func TestRegistryHasProviders(t *testing.T) {
	registry := GetRegistry()
	registry.Clear()
	_ = SetCurrentSystem(SystemTrainTicket)

	// Initially no providers
	if registry.HasServiceEndpointProvider() {
		t.Error("HasServiceEndpointProvider() should return false when no provider is registered")
	}

	// Register a provider
	registry.RegisterServiceEndpointProvider(SystemTrainTicket, &MockServiceEndpointProvider{})

	// Now it should exist
	if !registry.HasServiceEndpointProvider() {
		t.Error("HasServiceEndpointProvider() should return true after registration")
	}

	// Switch to OtelDemo - should not have provider
	_ = SetCurrentSystem(SystemOtelDemo)
	if registry.HasServiceEndpointProvider() {
		t.Error("HasServiceEndpointProvider() should return false for OtelDemo system")
	}
}

func TestRegistryGetProviderNotRegistered(t *testing.T) {
	registry := GetRegistry()
	registry.Clear()
	_ = SetCurrentSystem(SystemTrainTicket)

	_, err := registry.GetServiceEndpointProvider()
	if err == nil {
		t.Error("GetServiceEndpointProvider() should return error when no provider is registered")
	}

	_, err = registry.GetDatabaseOperationProvider()
	if err == nil {
		t.Error("GetDatabaseOperationProvider() should return error when no provider is registered")
	}

	_, err = registry.GetGRPCOperationProvider()
	if err == nil {
		t.Error("GetGRPCOperationProvider() should return error when no provider is registered")
	}
}

func TestDatabaseOperationProviderRegistry(t *testing.T) {
	registry := GetRegistry()
	registry.Clear()
	_ = SetCurrentSystem(SystemTrainTicket)

	mockProvider := &MockDatabaseOperationProvider{
		services: []string{"ts-order-service"},
		operations: map[string][]DatabaseOperationData{
			"ts-order-service": {
				{ServiceName: "ts-order-service", DBName: "ts", DBTable: "orders", Operation: "SELECT"},
			},
		},
	}

	registry.RegisterDatabaseOperationProvider(SystemTrainTicket, mockProvider)

	if !registry.HasDatabaseOperationProvider() {
		t.Error("HasDatabaseOperationProvider() should return true after registration")
	}

	provider, err := registry.GetDatabaseOperationProvider()
	if err != nil {
		t.Fatalf("GetDatabaseOperationProvider() error = %v", err)
	}

	ops := provider.GetOperationsByService("ts-order-service")
	if len(ops) != 1 {
		t.Errorf("Expected 1 operation, got %d", len(ops))
	}
}

func TestGRPCOperationProviderRegistry(t *testing.T) {
	registry := GetRegistry()
	registry.Clear()
	_ = SetCurrentSystem(SystemOtelDemo)

	mockProvider := &MockGRPCOperationProvider{
		services: []string{"checkout"},
		operations: map[string][]GRPCOperationData{
			"checkout": {
				{ServiceName: "checkout", RPCService: "oteldemo.CartService", RPCMethod: "GetCart"},
			},
		},
	}

	registry.RegisterGRPCOperationProvider(SystemOtelDemo, mockProvider)

	if !registry.HasGRPCOperationProvider() {
		t.Error("HasGRPCOperationProvider() should return true after registration")
	}

	provider, err := registry.GetGRPCOperationProvider()
	if err != nil {
		t.Fatalf("GetGRPCOperationProvider() error = %v", err)
	}

	ops := provider.GetOperationsByService("checkout")
	if len(ops) != 1 {
		t.Errorf("Expected 1 operation, got %d", len(ops))
	}
	if ops[0].RPCService != "oteldemo.CartService" {
		t.Errorf("Expected RPCService to be oteldemo.CartService, got %s", ops[0].RPCService)
	}
}
