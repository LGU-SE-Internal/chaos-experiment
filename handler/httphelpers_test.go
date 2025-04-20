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
