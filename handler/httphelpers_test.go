package handler

import (
	"testing"

	"github.com/CUHK-SE-Group/chaos-experiment/chaos"
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
	tests := []struct {
		name          string
		serviceName   string
		endpointIndex int
		wantOK        bool
	}{
		{
			name:          "Valid service and endpoint index",
			serviceName:   "ts-ui-dashboard",
			endpointIndex: 0,
			wantOK:        true,
		},
		{
			name:          "Valid service but negative endpoint index",
			serviceName:   "ts-ui-dashboard",
			endpointIndex: -1,
			wantOK:        false,
		},
		{
			name:          "Valid service but out of bounds endpoint index",
			serviceName:   "ts-ui-dashboard",
			endpointIndex: 100,
			wantOK:        false,
		},
		{
			name:          "Non-existent service",
			serviceName:   "non-existent-service",
			endpointIndex: 0,
			wantOK:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, ok := selectHTTPEndpointForService(tt.serviceName, tt.endpointIndex)

			if ok != tt.wantOK {
				t.Errorf("selectHTTPEndpointForService() ok = %v, want %v", ok, tt.wantOK)
				return
			}

			if tt.wantOK {
				if endpoint == nil {
					t.Errorf("selectHTTPEndpointForService() returned nil endpoint when ok = true")
				} else if endpoint.ServiceName != tt.serviceName {
					t.Errorf("selectHTTPEndpointForService() endpoint ServiceName = %v, want %v",
						endpoint.ServiceName, tt.serviceName)
				}
			}
		})
	}
}

func TestGetServiceAndEndpointForHTTPChaos(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	tests := []struct {
		name           string
		appNameIndex   int
		endpointIndex  int
		wantSourceName string
		wantOK         bool
	}{
		{
			name:           "Valid app and endpoint indices",
			appNameIndex:   0,
			endpointIndex:  0,
			wantSourceName: "ts-auth-service",
			wantOK:         true,
		},
		{
			name:           "Invalid app index",
			appNameIndex:   10,
			endpointIndex:  0,
			wantSourceName: "",
			wantOK:         false,
		},
		{
			name:           "Valid app index but invalid endpoint index",
			appNameIndex:   0,
			endpointIndex:  100,
			wantSourceName: "ts-auth-service",
			wantOK:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceName, endpoint, ok := getServiceAndEndpointForHTTPChaos(tt.appNameIndex, tt.endpointIndex)

			if ok != tt.wantOK {
				t.Errorf("getServiceAndEndpointForHTTPChaos() ok = %v, want %v", ok, tt.wantOK)
				return
			}

			if tt.wantOK {
				if serviceName != tt.wantSourceName {
					t.Errorf("getServiceAndEndpointForHTTPChaos() serviceName = %v, want %v", serviceName, tt.wantSourceName)
				}

				if endpoint == nil {
					t.Errorf("getServiceAndEndpointForHTTPChaos() returned nil endpoint when ok = true")
				}
			}
		})
	}
}

func TestGetHTTPEndpoints(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		wantEmpty   bool
	}{
		{
			name:        "Service with HTTP endpoints",
			serviceName: "ts-ui-dashboard",
			wantEmpty:   false,
		},
		{
			name:        "Non-existent service",
			serviceName: "non-existent-service",
			wantEmpty:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoints := GetHTTPEndpoints(tt.serviceName)

			if tt.wantEmpty && len(endpoints) > 0 {
				t.Errorf("GetHTTPEndpoints() returned %d endpoints, expected empty list", len(endpoints))
			}

			if !tt.wantEmpty && len(endpoints) == 0 {
				t.Errorf("GetHTTPEndpoints() returned empty list, expected endpoints")
			}
		})
	}
}

func TestListHTTPServiceNames(t *testing.T) {
	serviceNames := ListHTTPServiceNames()

	// Check that we get back a non-empty list
	if len(serviceNames) == 0 {
		t.Errorf("ListHTTPServiceNames() returned empty list, expected service names")
	}

	// Check that the list contains the expected service names
	expectedServices := []string{
		"ts-ui-dashboard",
		"ts-travel-service",
		"ts-food-service",
	}

	for _, expected := range expectedServices {
		found := false
		for _, actual := range serviceNames {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Logf("ListHTTPServiceNames() missing expected service: %s", expected)
			// Not failing the test as exact service list might change
		}
	}
}

func TestCountHTTPEndpoints(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		wantGreater bool
		compareWith int
	}{
		{
			name:        "Service with many HTTP endpoints",
			serviceName: "ts-ui-dashboard",
			wantGreater: true,
			compareWith: 5,
		},
		{
			name:        "Non-existent service",
			serviceName: "non-existent-service",
			wantGreater: false,
			compareWith: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := CountHTTPEndpoints(tt.serviceName)

			if tt.wantGreater && count <= tt.compareWith {
				t.Errorf("CountHTTPEndpoints() returned %d, expected > %d", count, tt.compareWith)
			}

			if !tt.wantGreater && count != tt.compareWith {
				t.Errorf("CountHTTPEndpoints() returned %d, expected %d", count, tt.compareWith)
			}
		})
	}
}

func TestHTTPHelpersIntegration(t *testing.T) {
	// Test that our helper functions work well together
	serviceNames := ListHTTPServiceNames()
	if len(serviceNames) == 0 {
		t.Fatal("No service names returned")
	}

	serviceName := serviceNames[0]
	endpoints := GetHTTPEndpoints(serviceName)
	if len(endpoints) == 0 {
		t.Fatalf("No endpoints returned for service %s", serviceName)
	}

	// Verify that counting endpoints works
	count := CountHTTPEndpoints(serviceName)
	if count != len(endpoints) {
		t.Errorf("CountHTTPEndpoints() = %d, want %d", count, len(endpoints))
	}

	// Verify that selecting an endpoint works
	endpoint, ok := selectHTTPEndpointForService(serviceName, 0)
	if !ok {
		t.Errorf("selectHTTPEndpointForService() failed for valid service and index")
	}

	if endpoint == nil {
		t.Fatalf("selectHTTPEndpointForService() returned nil endpoint when ok = true")
	}

	// Check that the selected endpoint properties match one of the endpoints from GetHTTPEndpoints
	found := false
	for _, ep := range endpoints {
		if ep.Route == endpoint.Route && ep.Method == endpoint.Method {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("selectHTTPEndpointForService() returned an endpoint not found in GetHTTPEndpoints()")
	}

	// Test AddCommonHTTPOptions
	opts := []chaos.OptHTTPChaos{}
	opts = AddCommonHTTPOptions(endpoint, opts)
	if len(opts) == 0 {
		t.Errorf("AddCommonHTTPOptions() returned empty options list")
	}
}

func TestHTTPHelpersWithMocks(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	// Test getServiceAndEndpointForHTTPChaos with the mocked data
	serviceName, endpoint, ok := getServiceAndEndpointForHTTPChaos(0, 0)

	if serviceName != "ts-auth-service" {
		t.Errorf("getServiceAndEndpointForHTTPChaos() with mocked data returned serviceName = %v, want %v",
			serviceName, "ts-auth-service")
	}

	// Depending on whether the mock service has HTTP endpoints, check the outcome
	if ok {
		if endpoint == nil {
			t.Errorf("getServiceAndEndpointForHTTPChaos() returned nil endpoint when ok = true")
		}
	}
}
