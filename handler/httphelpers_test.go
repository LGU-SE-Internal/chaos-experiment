package handler

import (
	"testing"

	"github.com/LGU-SE-Internal/chaos-experiment/chaos"
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

func TestGetFilteredHTTPMethods(t *testing.T) {
	tests := []struct {
		name             string
		originalMethod   string
		wantLength       int
		shouldContain    []string
		shouldNotContain string
	}{
		{
			name:             "Filter out GET method",
			originalMethod:   "GET",
			wantLength:       6,
			shouldContain:    []string{"POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"},
			shouldNotContain: "GET",
		},
		{
			name:             "Filter out POST method",
			originalMethod:   "POST",
			wantLength:       6,
			shouldContain:    []string{"GET", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"},
			shouldNotContain: "POST",
		},
		{
			name:             "Filter out unknown method returns all",
			originalMethod:   "UNKNOWN",
			wantLength:       7,
			shouldContain:    []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"},
			shouldNotContain: "",
		},
		{
			name:             "Filter out PATCH method",
			originalMethod:   "PATCH",
			wantLength:       6,
			shouldContain:    []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},
			shouldNotContain: "PATCH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFilteredHTTPMethods(tt.originalMethod)

			// Check length
			if len(result) != tt.wantLength {
				t.Errorf("GetFilteredHTTPMethods() returned %d methods, want %d", len(result), tt.wantLength)
			}

			// Convert result to string slice for easier checking
			resultStrings := make([]string, len(result))
			for i, method := range result {
				resultStrings[i] = GetHTTPMethodName(method)
			}

			// Check that all expected methods are present
			for _, expectedMethod := range tt.shouldContain {
				found := false
				for _, resultMethod := range resultStrings {
					if resultMethod == expectedMethod {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetFilteredHTTPMethods() should contain %s but didn't", expectedMethod)
				}
			}

			// Check that the filtered method is not present
			if tt.shouldNotContain != "" {
				for _, resultMethod := range resultStrings {
					if resultMethod == tt.shouldNotContain {
						t.Errorf("GetFilteredHTTPMethods() should not contain %s but did", tt.shouldNotContain)
					}
				}
			}

			// Check ordering - methods should be in consistent order
			if len(result) > 1 {
				// Verify that GET comes before POST if both are present
				getIndex, postIndex := -1, -1
				for i, method := range resultStrings {
					if method == "GET" {
						getIndex = i
					}
					if method == "POST" {
						postIndex = i
					}
				}
				if getIndex != -1 && postIndex != -1 && getIndex > postIndex {
					t.Errorf("GetFilteredHTTPMethods() methods not in expected order: GET should come before POST")
				}
			}
		})
	}
}

func TestGetFilteredHTTPMethodByIndex(t *testing.T) {
	tests := []struct {
		name           string
		originalMethod string
		index          int
		wantMethod     string
	}{
		{
			name:           "Get first method when original is GET",
			originalMethod: "GET",
			index:          0,
			wantMethod:     "POST", // First in filtered list when GET is removed
		},
		{
			name:           "Get second method when original is GET",
			originalMethod: "GET",
			index:          1,
			wantMethod:     "PUT", // Second in filtered list when GET is removed
		},
		{
			name:           "Get first method when original is POST",
			originalMethod: "POST",
			index:          0,
			wantMethod:     "GET", // First in filtered list when POST is removed
		},
		{
			name:           "Index out of range returns first available",
			originalMethod: "GET",
			index:          10,
			wantMethod:     "POST", // Should return first available when index is out of range
		},
		{
			name:           "Negative index returns first available",
			originalMethod: "GET",
			index:          -1,
			wantMethod:     "POST", // Should return first available when index is negative
		},
		{
			name:           "Unknown original method with valid index",
			originalMethod: "UNKNOWN",
			index:          0,
			wantMethod:     "GET", // Should return GET as first method when original is unknown
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFilteredHTTPMethodByIndex(tt.originalMethod, tt.index)
			resultString := GetHTTPMethodName(result)

			if resultString != tt.wantMethod {
				t.Errorf("GetFilteredHTTPMethodByIndex() = %v, want %v", resultString, tt.wantMethod)
			}
		})
	}
}

func TestGetFilteredHTTPMethodByIndex_EdgeCases(t *testing.T) {
	// Test consistency - same original method and index should always return same result
	originalMethod := "GET"
	index := 2

	result1 := GetFilteredHTTPMethodByIndex(originalMethod, index)
	result2 := GetFilteredHTTPMethodByIndex(originalMethod, index)

	if result1 != result2 {
		t.Errorf("GetFilteredHTTPMethodByIndex() is not consistent: got %v and %v for same inputs", result1, result2)
	}

	// Test that filtering actually works - original method should never be returned
	for _, originalMethod := range []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"} {
		filtered := GetFilteredHTTPMethods(originalMethod)
		for i := 0; i < len(filtered); i++ {
			result := GetFilteredHTTPMethodByIndex(originalMethod, i)
			resultString := GetHTTPMethodName(result)
			if resultString == originalMethod {
				t.Errorf("GetFilteredHTTPMethodByIndex() returned original method %s, should be filtered out", originalMethod)
			}
		}
	}
}
