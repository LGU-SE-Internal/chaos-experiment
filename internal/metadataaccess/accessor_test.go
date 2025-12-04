package metadataaccess

import (
	"testing"

	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"
)

// MockServiceEndpointAccessor implements ServiceEndpointAccessor for testing
type MockServiceEndpointAccessor struct {
	services  []string
	endpoints map[string][]ServiceEndpoint
}

func (m *MockServiceEndpointAccessor) GetEndpointsByService(serviceName string) []ServiceEndpoint {
	if m.endpoints == nil {
		return nil
	}
	return m.endpoints[serviceName]
}

func (m *MockServiceEndpointAccessor) GetAllServices() []string {
	return m.services
}

// MockDatabaseOperationAccessor implements DatabaseOperationAccessor for testing
type MockDatabaseOperationAccessor struct {
	services   []string
	operations map[string][]DatabaseOperation
}

func (m *MockDatabaseOperationAccessor) GetOperationsByService(serviceName string) []DatabaseOperation {
	if m.operations == nil {
		return nil
	}
	return m.operations[serviceName]
}

func (m *MockDatabaseOperationAccessor) GetAllDatabaseServices() []string {
	return m.services
}

// MockGRPCOperationAccessor implements GRPCOperationAccessor for testing
type MockGRPCOperationAccessor struct {
	services   []string
	operations map[string][]GRPCOperation
}

func (m *MockGRPCOperationAccessor) GetOperationsByService(serviceName string) []GRPCOperation {
	if m.operations == nil {
		return nil
	}
	return m.operations[serviceName]
}

func (m *MockGRPCOperationAccessor) GetAllGRPCServices() []string {
	return m.services
}

// MockJavaMethodAccessor implements JavaMethodAccessor for testing
type MockJavaMethodAccessor struct {
	services []string
	methods  map[string][]JavaClassMethod
}

func (m *MockJavaMethodAccessor) GetClassMethodsByService(serviceName string) []JavaClassMethod {
	if m.methods == nil {
		return nil
	}
	return m.methods[serviceName]
}

func (m *MockJavaMethodAccessor) GetAllServices() []string {
	return m.services
}

func TestMetadataAccessorServiceEndpoints(t *testing.T) {
	// Reset system and accessor
	_ = systemconfig.SetCurrentSystem(systemconfig.SystemTrainTicket)

	accessor := GetAccessor()
	accessor.Clear()

	// Create mock accessors
	tsAccessor := &MockServiceEndpointAccessor{
		services: []string{"ts-order-service", "ts-auth-service"},
		endpoints: map[string][]ServiceEndpoint{
			"ts-order-service": {
				{ServiceName: "ts-order-service", Route: "/api/v1/orders"},
			},
		},
	}

	otelAccessor := &MockServiceEndpointAccessor{
		services: []string{"frontend", "cart"},
		endpoints: map[string][]ServiceEndpoint{
			"frontend": {
				{ServiceName: "frontend", Route: "/api/products"},
			},
		},
	}

	// Register accessors
	accessor.RegisterTrainTicketServiceEndpoints(tsAccessor)
	accessor.RegisterOtelDemoServiceEndpoints(otelAccessor)

	// Test TrainTicket system
	_ = systemconfig.SetCurrentSystem(systemconfig.SystemTrainTicket)

	services := accessor.GetAllServices()
	if len(services) != 2 {
		t.Errorf("Expected 2 services for TrainTicket, got %d", len(services))
	}

	endpoints := accessor.GetEndpointsByService("ts-order-service")
	if len(endpoints) != 1 {
		t.Errorf("Expected 1 endpoint for ts-order-service, got %d", len(endpoints))
	}
	if endpoints[0].Route != "/api/v1/orders" {
		t.Errorf("Expected route /api/v1/orders, got %s", endpoints[0].Route)
	}

	// Test OtelDemo system
	_ = systemconfig.SetCurrentSystem(systemconfig.SystemOtelDemo)

	services = accessor.GetAllServices()
	if len(services) != 2 {
		t.Errorf("Expected 2 services for OtelDemo, got %d", len(services))
	}

	endpoints = accessor.GetEndpointsByService("frontend")
	if len(endpoints) != 1 {
		t.Errorf("Expected 1 endpoint for frontend, got %d", len(endpoints))
	}
	if endpoints[0].Route != "/api/products" {
		t.Errorf("Expected route /api/products, got %s", endpoints[0].Route)
	}
}

func TestMetadataAccessorDatabaseOperations(t *testing.T) {
	_ = systemconfig.SetCurrentSystem(systemconfig.SystemTrainTicket)

	accessor := GetAccessor()
	accessor.Clear()

	tsDbAccessor := &MockDatabaseOperationAccessor{
		services: []string{"ts-order-service"},
		operations: map[string][]DatabaseOperation{
			"ts-order-service": {
				{ServiceName: "ts-order-service", DBName: "ts", DBTable: "orders"},
			},
		},
	}

	otelDbAccessor := &MockDatabaseOperationAccessor{
		services: []string{"cart"},
		operations: map[string][]DatabaseOperation{
			"cart": {
				{ServiceName: "cart", DBName: "cart", DBSystem: "redis"},
			},
		},
	}

	accessor.RegisterTrainTicketDatabaseOperations(tsDbAccessor)
	accessor.RegisterOtelDemoDatabaseOperations(otelDbAccessor)

	// Test TrainTicket
	_ = systemconfig.SetCurrentSystem(systemconfig.SystemTrainTicket)

	dbServices := accessor.GetAllDatabaseServices()
	if len(dbServices) != 1 {
		t.Errorf("Expected 1 database service for TrainTicket, got %d", len(dbServices))
	}

	ops := accessor.GetDatabaseOperationsByService("ts-order-service")
	if len(ops) != 1 {
		t.Errorf("Expected 1 operation for ts-order-service, got %d", len(ops))
	}

	// Test OtelDemo
	_ = systemconfig.SetCurrentSystem(systemconfig.SystemOtelDemo)

	dbServices = accessor.GetAllDatabaseServices()
	if len(dbServices) != 1 {
		t.Errorf("Expected 1 database service for OtelDemo, got %d", len(dbServices))
	}

	ops = accessor.GetDatabaseOperationsByService("cart")
	if len(ops) != 1 {
		t.Errorf("Expected 1 operation for cart, got %d", len(ops))
	}
}

func TestMetadataAccessorGRPCOperations(t *testing.T) {
	accessor := GetAccessor()
	accessor.Clear()

	grpcAccessor := &MockGRPCOperationAccessor{
		services: []string{"checkout", "frontend"},
		operations: map[string][]GRPCOperation{
			"checkout": {
				{ServiceName: "checkout", RPCService: "oteldemo.CartService"},
			},
		},
	}

	accessor.RegisterOtelDemoGRPCOperations(grpcAccessor)

	// Test TrainTicket - should not have gRPC
	_ = systemconfig.SetCurrentSystem(systemconfig.SystemTrainTicket)

	if accessor.HasGRPCOperations() {
		t.Error("TrainTicket should not have gRPC operations")
	}

	grpcServices := accessor.GetAllGRPCServices()
	if grpcServices != nil {
		t.Error("Expected nil gRPC services for TrainTicket")
	}

	// Test OtelDemo
	_ = systemconfig.SetCurrentSystem(systemconfig.SystemOtelDemo)

	if !accessor.HasGRPCOperations() {
		t.Error("OtelDemo should have gRPC operations")
	}

	grpcServices = accessor.GetAllGRPCServices()
	if len(grpcServices) != 2 {
		t.Errorf("Expected 2 gRPC services for OtelDemo, got %d", len(grpcServices))
	}

	ops := accessor.GetGRPCOperationsByService("checkout")
	if len(ops) != 1 {
		t.Errorf("Expected 1 gRPC operation for checkout, got %d", len(ops))
	}
}

func TestMetadataAccessorJavaMethods(t *testing.T) {
	accessor := GetAccessor()
	accessor.Clear()

	javaAccessor := &MockJavaMethodAccessor{
		services: []string{"ts-order-service"},
		methods: map[string][]JavaClassMethod{
			"ts-order-service": {
				{ClassName: "order.service.OrderServiceImpl", MethodName: "createOrder"},
			},
		},
	}

	accessor.RegisterTrainTicketJavaMethods(javaAccessor)

	// Test TrainTicket
	_ = systemconfig.SetCurrentSystem(systemconfig.SystemTrainTicket)

	if !accessor.HasJavaMethods() {
		t.Error("TrainTicket should have Java methods")
	}

	javaServices := accessor.GetAllJavaServices()
	if len(javaServices) != 1 {
		t.Errorf("Expected 1 Java service, got %d", len(javaServices))
	}

	methods := accessor.GetJavaMethodsByService("ts-order-service")
	if len(methods) != 1 {
		t.Errorf("Expected 1 Java method, got %d", len(methods))
	}
	if methods[0].MethodName != "createOrder" {
		t.Errorf("Expected method createOrder, got %s", methods[0].MethodName)
	}

	// Test OtelDemo - should not have Java methods
	_ = systemconfig.SetCurrentSystem(systemconfig.SystemOtelDemo)

	if accessor.HasJavaMethods() {
		t.Error("OtelDemo should not have Java methods")
	}

	methods = accessor.GetJavaMethodsByService("ts-order-service")
	if methods != nil {
		t.Error("Expected nil Java methods for OtelDemo")
	}
}

func TestMetadataAccessorHasProviders(t *testing.T) {
	accessor := GetAccessor()
	accessor.Clear()

	_ = systemconfig.SetCurrentSystem(systemconfig.SystemTrainTicket)

	// Initially no providers
	if accessor.HasServiceEndpoints() {
		t.Error("Should not have service endpoints before registration")
	}

	if accessor.HasDatabaseOperations() {
		t.Error("Should not have database operations before registration")
	}

	// Register a provider
	accessor.RegisterTrainTicketServiceEndpoints(&MockServiceEndpointAccessor{})

	if !accessor.HasServiceEndpoints() {
		t.Error("Should have service endpoints after registration")
	}

	// Switch to OtelDemo - should not have provider
	_ = systemconfig.SetCurrentSystem(systemconfig.SystemOtelDemo)

	if accessor.HasServiceEndpoints() {
		t.Error("OtelDemo should not have service endpoints (not registered)")
	}
}

func TestMetadataAccessorClear(t *testing.T) {
	accessor := GetAccessor()

	accessor.RegisterTrainTicketServiceEndpoints(&MockServiceEndpointAccessor{})
	accessor.RegisterOtelDemoServiceEndpoints(&MockServiceEndpointAccessor{})

	_ = systemconfig.SetCurrentSystem(systemconfig.SystemTrainTicket)
	if !accessor.HasServiceEndpoints() {
		t.Error("Should have service endpoints before clear")
	}

	accessor.Clear()

	if accessor.HasServiceEndpoints() {
		t.Error("Should not have service endpoints after clear")
	}

	_ = systemconfig.SetCurrentSystem(systemconfig.SystemOtelDemo)
	if accessor.HasServiceEndpoints() {
		t.Error("Should not have service endpoints after clear")
	}
}
