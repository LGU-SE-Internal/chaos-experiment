package testdata

import (
	"github.com/CUHK-SE-Group/chaos-experiment/internal/javaclassmethods"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/networkdependencies"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/serviceendpoints"
)

// MockServiceLabels contains a list of mock service names
var MockServiceLabels = []string{
	"ts-auth-service",
	"ts-order-service",
	"ts-travel-service",
	"ts-empty-service",
	"ts-self-service",
	"ts-ui-dashboard",
}

// MockGetLabels provides a mock implementation of the GetLabels function
func MockGetLabels(namespace, labelKey string) ([]string, error) {
	return MockServiceLabels, nil
}

// MockServiceEndpoints contains mock service endpoints data
var MockServiceEndpoints = map[string][]serviceendpoints.ServiceEndpoint{
	"ts-auth-service": {
		{
			ServiceName:    "ts-auth-service",
			ServerAddress:  "ts-verification-code-service",
			ServerPort:     "8080",
			Route:          "/api/v1/verifycode",
			RequestMethod:  "POST",
			ResponseStatus: "200",
		},
		{
			ServiceName:    "ts-auth-service",
			ServerAddress:  "mysql",
			ServerPort:     "3306",
			Route:          "",
			RequestMethod:  "",
			ResponseStatus: "",
		},
	},
	"ts-order-service": {
		{
			ServiceName:    "ts-order-service",
			ServerAddress:  "mysql",
			ServerPort:     "3306",
			Route:          "",
			RequestMethod:  "",
			ResponseStatus: "",
		},
		{
			ServiceName:    "ts-order-service",
			ServerAddress:  "ts-payment-service",
			ServerPort:     "8080",
			Route:          "/api/v1/payment",
			RequestMethod:  "POST",
			ResponseStatus: "200",
		},
	},
	"ts-travel-service": {
		{
			ServiceName:    "ts-travel-service",
			ServerAddress:  "ts-route-service",
			ServerPort:     "8080",
			Route:          "/api/v1/routeservice",
			RequestMethod:  "GET",
			ResponseStatus: "200",
		},
	},
	"ts-ui-dashboard": {
		{
			ServiceName:    "ts-ui-dashboard",
			ServerAddress:  "ts-auth-service",
			ServerPort:     "8080",
			Route:          "/api/v1/users/login",
			RequestMethod:  "POST",
			ResponseStatus: "200",
		},
		{
			ServiceName:    "ts-ui-dashboard",
			ServerAddress:  "ts-travel-service",
			ServerPort:     "8080",
			Route:          "/api/v1/travel",
			RequestMethod:  "GET",
			ResponseStatus: "200",
		},
	},
	"ts-test-service": {
		{
			ServiceName:    "ts-test-service",
			ServerAddress:  "ts-dependency-1",
			ServerPort:     "8080",
			Route:          "/api/v1/dep1",
			RequestMethod:  "GET",
			ResponseStatus: "200",
		},
		{
			ServiceName:    "ts-test-service",
			ServerAddress:  "ts-dependency-2",
			ServerPort:     "8080",
			Route:          "/api/v1/dep2",
			RequestMethod:  "POST",
			ResponseStatus: "200",
		},
		{
			ServiceName:    "ts-test-service", // Self reference - should be filtered out for DNS
			ServerAddress:  "ts-test-service",
			ServerPort:     "8080",
			Route:          "/api/v1/self",
			RequestMethod:  "GET",
			ResponseStatus: "200",
		},
		{
			ServiceName:    "ts-test-service",
			ServerAddress:  "", // Empty server address - should be filtered out
			ServerPort:     "8080",
			Route:          "",
			RequestMethod:  "",
			ResponseStatus: "",
		},
	},
	"ts-empty-service": {
		{
			ServiceName:   "ts-empty-service",
			ServerAddress: "",
			ServerPort:    "8080",
		},
	},
	"ts-self-service": {
		{
			ServiceName:   "ts-self-service",
			ServerAddress: "ts-self-service", // Only self reference
			ServerPort:    "8080",
		},
	},
}

// MockGetEndpoints provides a mock implementation of the service endpoint getter
func MockGetEndpoints(serviceName string) []serviceendpoints.ServiceEndpoint {
	if endpoints, ok := MockServiceEndpoints[serviceName]; ok {
		return endpoints
	}
	return []serviceendpoints.ServiceEndpoint{}
}

// MockServiceDependencies contains mock network dependency data
var MockServiceDependencies = map[string][]networkdependencies.ServiceDependency{
	"ts-auth-service": {
		{
			SourceService:     "ts-auth-service",
			TargetService:     "ts-verification-code-service",
			ConnectionDetails: "HTTP REST API",
		},
		{
			SourceService:     "ts-auth-service",
			TargetService:     "ts-ui-dashboard",
			ConnectionDetails: "HTTP REST API",
		},
	},
	"ts-order-service": {
		{
			SourceService:     "ts-order-service",
			TargetService:     "ts-payment-service",
			ConnectionDetails: "HTTP REST API",
		},
	},
	"ts-travel-service": {
		{
			SourceService:     "ts-travel-service",
			TargetService:     "ts-route-service",
			ConnectionDetails: "HTTP REST API",
		},
	},
}

// MockGetServicePairByServiceAndIndex provides mock implementation for network dependencies
func MockGetServicePairByServiceAndIndex(sourceName string, targetIndex int) (string, bool) {
	if pairs, ok := MockServiceDependencies[sourceName]; ok {
		if targetIndex >= 0 && targetIndex < len(pairs) {
			return pairs[targetIndex].TargetService, true
		}
	}
	return "", false
}

// MockGetDependenciesForService provides a mock implementation to get target dependencies
func MockGetDependenciesForService(serviceName string) []string {
	if pairs, ok := MockServiceDependencies[serviceName]; ok {
		result := make([]string, len(pairs))
		for i, pair := range pairs {
			result[i] = pair.TargetService
		}
		return result
	}
	return []string{}
}

// MockListAllServiceNames provides mock implementation to list all service names with dependencies
func MockListAllServiceNames() []string {
	result := []string{}
	for service := range MockServiceDependencies {
		result = append(result, service)
	}
	return result
}

// MockGetAllServicePairs returns all mock service dependency pairs
func MockGetAllServicePairs() []networkdependencies.ServiceDependency {
	result := []networkdependencies.ServiceDependency{}
	for _, pairs := range MockServiceDependencies {
		result = append(result, pairs...)
	}
	return result
}

// MockJavaClassMethods contains mock Java class and method data
var MockJavaClassMethods = map[string][]struct {
	ClassName  string
	MethodName string
}{
	"ts-auth-service": {
		{
			ClassName:  "auth.AuthApplication",
			MethodName: "login",
		},
		{
			ClassName:  "auth.AuthService",
			MethodName: "verifyCode",
		},
	},
	"ts-order-service": {
		{
			ClassName:  "order.OrderService",
			MethodName: "createOrder",
		},
	},
	"ts-travel-service": {
		{
			ClassName:  "travel.TravelService",
			MethodName: "queryTravel",
		},
	},
}

// MockGetJavaClassMethod returns a mock Java class method entry based on service name and method index
func MockGetJavaClassMethod(serviceName string, methodIndex int) *javaclassmethods.ClassMethodEntry {
	methods, exists := MockJavaClassMethods[serviceName]
	if !exists || methodIndex < 0 || methodIndex >= len(methods) {
		return nil
	}

	mockMethod := methods[methodIndex]
	return &javaclassmethods.ClassMethodEntry{
		ClassName:  mockMethod.ClassName,
		MethodName: mockMethod.MethodName,
	}
}

// MockGetJavaClassMethodsByService returns all mock Java class methods for a service
func MockGetJavaClassMethodsByService(serviceName string) []*javaclassmethods.ClassMethodEntry {
	methods, exists := MockJavaClassMethods[serviceName]
	if !exists {
		return []*javaclassmethods.ClassMethodEntry{}
	}

	result := make([]*javaclassmethods.ClassMethodEntry, len(methods))
	for i, mockMethod := range methods {
		result[i] = &javaclassmethods.ClassMethodEntry{
			ClassName:  mockMethod.ClassName,
			MethodName: mockMethod.MethodName,
		}
	}
	return result
}

// MockListJavaClassMethodServices returns the list of services that have Java class methods
func MockListJavaClassMethodServices() []string {
	result := []string{}
	for service := range MockJavaClassMethods {
		result = append(result, service)
	}
	return result
}

// MockGetClassMethodsByService returns all mock Java class methods for a service
// This returns non-pointer slice to match the interface in jvmhelpers.go
func MockGetClassMethodsByService(serviceName string) []javaclassmethods.ClassMethodEntry {
	methods, exists := MockJavaClassMethods[serviceName]
	if !exists {
		return []javaclassmethods.ClassMethodEntry{}
	}

	result := make([]javaclassmethods.ClassMethodEntry, len(methods))
	for i, mockMethod := range methods {
		result[i] = javaclassmethods.ClassMethodEntry{
			ClassName:  mockMethod.ClassName,
			MethodName: mockMethod.MethodName,
		}
	}
	return result
}

// MockListAvailableMethods returns a list of method names in the format ClassName.methodName
func MockListAvailableMethods(serviceName string) []string {
	methods, exists := MockJavaClassMethods[serviceName]
	if !exists {
		return []string{}
	}

	result := make([]string, len(methods))
	for i, mockMethod := range methods {
		// Get simple class name without package
		className := mockMethod.ClassName
		result[i] = className + "." + mockMethod.MethodName
	}
	return result
}