package handler

import (
	"testing"

	"github.com/CUHK-SE-Group/chaos-experiment/chaos"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/pointer"
)

func TestGetHTTPMethodName(t *testing.T) {
	tests := []struct {
		name       string
		method     HTTPMethod
		wantResult string
	}{
		{
			name:       "GET method",
			method:     GET,
			wantResult: "GET",
		},
		{
			name:       "POST method",
			method:     POST,
			wantResult: "POST",
		},
		{
			name:       "PUT method",
			method:     PUT,
			wantResult: "PUT",
		},
		{
			name:       "DELETE method",
			method:     DELETE,
			wantResult: "DELETE",
		},
		{
			name:       "Invalid method falls back to GET",
			method:     HTTPMethod(999),
			wantResult: "GET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetHTTPMethodName(tt.method)
			if result != tt.wantResult {
				t.Errorf("GetHTTPMethodName() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func TestGetHTTPStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode HTTPStatusCode
		wantResult int32
	}{
		{
			name:       "Bad Request",
			statusCode: BadRequest,
			wantResult: 400,
		},
		{
			name:       "Unauthorized",
			statusCode: Unauthorized,
			wantResult: 401,
		},
		{
			name:       "Internal Server Error",
			statusCode: InternalServerError,
			wantResult: 500,
		},
		{
			name:       "Invalid status code falls back to 500",
			statusCode: HTTPStatusCode(999),
			wantResult: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetHTTPStatusCode(tt.statusCode)
			if result != tt.wantResult {
				t.Errorf("GetHTTPStatusCode() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func TestHTTPEndpointGetEndpointPort(t *testing.T) {
	tests := []struct {
		name       string
		endpoint   HTTPEndpoint
		wantResult int32
	}{
		{
			name: "Valid port",
			endpoint: HTTPEndpoint{
				Port: "8080",
			},
			wantResult: 8080,
		},
		{
			name: "Empty port defaults to 8080",
			endpoint: HTTPEndpoint{
				Port: "",
			},
			wantResult: 8080,
		},
		{
			name: "Non-numeric port defaults to 8080",
			endpoint: HTTPEndpoint{
				Port: "invalid",
			},
			wantResult: 8080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.endpoint.GetEndpointPort()
			if result != tt.wantResult {
				t.Errorf("GetEndpointPort() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func TestAddCommonHTTPOptions(t *testing.T) {
	tests := []struct {
		name       string
		endpoint   *HTTPEndpoint
		inputOpts  []chaos.OptHTTPChaos
		wantLength int
	}{
		{
			name: "Endpoint with all fields",
			endpoint: &HTTPEndpoint{
				Route:  "/api/test",
				Method: "GET",
				Port:   "8080",
			},
			inputOpts:  []chaos.OptHTTPChaos{},
			wantLength: 3, // Port + Path + Method
		},
		{
			name: "Endpoint with no route",
			endpoint: &HTTPEndpoint{
				Method: "POST",
				Port:   "9090",
			},
			inputOpts:  []chaos.OptHTTPChaos{},
			wantLength: 2, // Port + Method
		},
		{
			name: "Endpoint with no method",
			endpoint: &HTTPEndpoint{
				Route: "/api/test",
				Port:  "8080",
			},
			inputOpts:  []chaos.OptHTTPChaos{},
			wantLength: 2, // Port + Path
		},
		{
			name: "Endpoint with existing options",
			endpoint: &HTTPEndpoint{
				Route:  "/api/test",
				Method: "GET",
				Port:   "8080",
			},
			inputOpts:  []chaos.OptHTTPChaos{chaos.WithDelay(pointer.String("100ms"))},
			wantLength: 4, // Existing + Port + Path + Method
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddCommonHTTPOptions(tt.endpoint, tt.inputOpts)
			if len(result) != tt.wantLength {
				t.Errorf("AddCommonHTTPOptions() returned %d options, want %d", len(result), tt.wantLength)
			}
		})
	}
}

func TestSelectHTTPEndpointForService(t *testing.T) {
	cleanup := setupHTTPMocks()
	defer cleanup()

	tests := []struct {
		name          string
		serviceName   string
		endpointIndex int
		wantErr       bool
		wantRoute     string
	}{
		{
			name:          "Valid service with API endpoint",
			serviceName:   "ts-auth-service",
			endpointIndex: 0,
			wantErr:       false,
			wantRoute:     "/api/v1/verifycode",
		},
		{
			name:          "Valid service with database endpoint should be filtered out",
			serviceName:   "ts-auth-service",
			endpointIndex: 1, // After filtering, this would be out of bounds
			wantErr:       true,
		},
		{
			name:          "Valid service but negative endpoint index",
			serviceName:   "ts-ui-dashboard",
			endpointIndex: -1,
			wantErr:       true,
		},
		{
			name:          "Valid service but out of bounds endpoint index",
			serviceName:   "ts-ui-dashboard",
			endpointIndex: 100,
			wantErr:       true,
		},
		{
			name:          "Non-existent service",
			serviceName:   "non-existent-service",
			endpointIndex: 0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, err := selectHTTPEndpointForService(tt.serviceName, tt.endpointIndex)

			if (err != nil) != tt.wantErr {
				t.Errorf("selectHTTPEndpointForService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.NotNil(t, endpoint, "Endpoint should not be nil when successful")
				assert.Equal(t, tt.serviceName, endpoint.ServiceName, "Endpoint ServiceName doesn't match")

				if tt.wantRoute != "" {
					assert.Equal(t, tt.wantRoute, endpoint.Route, "Endpoint Route doesn't match expected")
				}
			}
		})
	}
}

func TestGetServiceAndEndpointForHTTPChaos(t *testing.T) {
	cleanup := setupHTTPMocks()
	defer cleanup()

	tests := []struct {
		name           string
		appNameIndex   int
		endpointIndex  int
		wantSourceName string
		wantErr        bool
		wantRoute      string
	}{
		{
			name:           "Valid app and endpoint indices",
			appNameIndex:   0, // ts-auth-service
			endpointIndex:  0,
			wantSourceName: "ts-auth-service",
			wantErr:        false,
			wantRoute:      "/api/v1/verifycode",
		},
		{
			name:           "UI Dashboard service with login endpoint",
			appNameIndex:   5, // ts-ui-dashboard
			endpointIndex:  0,
			wantSourceName: "ts-ui-dashboard",
			wantErr:        false,
			wantRoute:      "/api/v1/users/login",
		},
		{
			name:           "Invalid app index",
			appNameIndex:   20,
			endpointIndex:  0,
			wantSourceName: "",
			wantErr:        true,
		},
		{
			name:           "Valid app index but invalid endpoint index",
			appNameIndex:   0,
			endpointIndex:  100,
			wantSourceName: "ts-auth-service",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceName, endpoint, err := getServiceAndEndpointForHTTPChaos(tt.appNameIndex, tt.endpointIndex)

			if (err != nil) != tt.wantErr {
				t.Errorf("getServiceAndEndpointForHTTPChaos() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantSourceName, serviceName, "Service name doesn't match expected")

			if !tt.wantErr {
				assert.NotNil(t, endpoint, "Endpoint should not be nil when successful")

				if tt.wantRoute != "" {
					assert.Equal(t, tt.wantRoute, endpoint.Route, "Endpoint Route doesn't match expected")
				}
			}
		})
	}
}

func TestGetHTTPEndpoints(t *testing.T) {
	cleanup := setupHTTPMocks()
	defer cleanup()

	tests := []struct {
		name        string
		serviceName string
		wantCount   int
		wantEmpty   bool
	}{
		{
			name:        "Service with valid HTTP endpoints",
			serviceName: "ts-ui-dashboard",
			wantCount:   2, // Two HTTP endpoints in mock data
			wantEmpty:   false,
		},
		{
			name:        "Service with mixed endpoints (HTTP and DB)",
			serviceName: "ts-auth-service",
			wantCount:   1, // Only one HTTP endpoint, filtering out DB
			wantEmpty:   false,
		},
		{
			name:        "Service with no endpoints",
			serviceName: "ts-empty-service",
			wantCount:   0,
			wantEmpty:   true,
		},
		{
			name:        "Non-existent service",
			serviceName: "non-existent-service",
			wantCount:   0,
			wantEmpty:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoints := GetHTTPEndpoints(tt.serviceName)

			assert.Equal(t, tt.wantCount, len(endpoints), "Unexpected number of endpoints")
			assert.Equal(t, tt.wantEmpty, len(endpoints) == 0, "Unexpected empty status")

			// Verify some properties of non-empty endpoint lists
			if !tt.wantEmpty {
				for _, ep := range endpoints {
					assert.Equal(t, tt.serviceName, ep.ServiceName, "Endpoint ServiceName doesn't match service")
					assert.NotEmpty(t, ep.Route, "HTTP endpoint should have a route")
				}
			}
		})
	}
}

func TestListHTTPServiceNames(t *testing.T) {
	cleanup := setupHTTPMocks()
	defer cleanup()

	serviceNames := ListHTTPServiceNames()

	// Check that we get back a non-empty list
	assert.NotEmpty(t, serviceNames, "ListHTTPServiceNames() returned empty list")

	// Check that the list contains expected service names
	expectedServices := []string{
		"ts-auth-service",
		"ts-order-service",
		"ts-travel-service",
		"ts-ui-dashboard",
	}

	// Verify at least some of our expected services are in the list
	foundCount := 0
	for _, expected := range expectedServices {
		for _, actual := range serviceNames {
			if actual == expected {
				foundCount++
				break
			}
		}
	}

	// At least some of our expected services should be present
	assert.Greater(t, foundCount, 0, "None of the expected services found in result")
}

func TestCountHTTPEndpoints(t *testing.T) {
	cleanup := setupHTTPMocks()
	defer cleanup()

	tests := []struct {
		name        string
		serviceName string
		wantCount   int
	}{
		{
			name:        "Service with multiple HTTP endpoints",
			serviceName: "ts-ui-dashboard",
			wantCount:   2,
		},
		{
			name:        "Service with one HTTP endpoint",
			serviceName: "ts-auth-service",
			wantCount:   1,
		},
		{
			name:        "Service with no HTTP endpoints",
			serviceName: "ts-empty-service",
			wantCount:   0,
		},
		{
			name:        "Non-existent service",
			serviceName: "non-existent-service",
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := CountHTTPEndpoints(tt.serviceName)
			assert.Equal(t, tt.wantCount, count, "Count doesn't match expected")
		})
	}
}

func TestHTTPHelpersIntegration(t *testing.T) {
	cleanup := setupHTTPMocks()
	defer cleanup()

	// Test that our helper functions work well together
	serviceNames := ListHTTPServiceNames()
	assert.NotEmpty(t, serviceNames, "No service names returned")

	serviceName := "ts-ui-dashboard" // Use a service we know has HTTP endpoints
	endpoints := GetHTTPEndpoints(serviceName)
	assert.NotEmpty(t, endpoints, "No endpoints returned for service")

	// Verify that counting endpoints works
	count := CountHTTPEndpoints(serviceName)
	assert.Equal(t, len(endpoints), count, "CountHTTPEndpoints() returned unexpected count")

	// Verify that selecting an endpoint works
	endpoint, err := selectHTTPEndpointForService(serviceName, 0)
	assert.NoError(t, err, "selectHTTPEndpointForService() failed for valid service and index")
	assert.NotNil(t, endpoint, "selectHTTPEndpointForService() returned nil endpoint when successful")

	// Check that the selected endpoint properties match one from GetHTTPEndpoints
	found := false
	for _, ep := range endpoints {
		if ep.Route == endpoint.Route && ep.Method == endpoint.Method {
			found = true
			break
		}
	}
	assert.True(t, found, "selectHTTPEndpointForService() returned an endpoint not in GetHTTPEndpoints()")

	// Test AddCommonHTTPOptions
	opts := []chaos.OptHTTPChaos{}
	opts = AddCommonHTTPOptions(endpoint, opts)
	assert.NotEmpty(t, opts, "AddCommonHTTPOptions() returned empty options list")
}

func TestHTTPHelpersWithMocks(t *testing.T) {
	cleanup := setupHTTPMocks()
	defer cleanup()

	// Test getServiceAndEndpointForHTTPChaos with the mocked data
	serviceName, endpoint, err := getServiceAndEndpointForHTTPChaos(0, 0)

	assert.NoError(t, err, "Expected no error with valid mock data")
	assert.Equal(t, "ts-auth-service", serviceName, "Unexpected service name")
	assert.NotNil(t, endpoint, "Expected non-nil endpoint with valid mock data")

	if endpoint != nil {
		assert.Equal(t, "/api/v1/verifycode", endpoint.Route, "Unexpected route in endpoint")
		assert.Equal(t, "POST", endpoint.Method, "Unexpected method in endpoint")
		assert.Equal(t, "ts-verification-code-service", endpoint.TargetService, "Unexpected target service in endpoint")
	}
}
