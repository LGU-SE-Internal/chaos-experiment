package javaclassmethods_test

import (
	"testing"

	"github.com/LGU-SE-Internal/chaos-experiment/internal/javaclassmethods"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/testdata"
)

// SetupJavaClassMethodsMock sets up mock functions for Java class methods
func SetupJavaClassMethodsMock() func() {
	// Store original functions
	originalGetClassMethodsByService := javaclassmethods.GetClassMethodsByServiceFunc
	originalGetAllServices := javaclassmethods.GetAllServicesFunc

	// Replace with mock implementations
	javaclassmethods.GetClassMethodsByServiceFunc = testdata.MockGetClassMethodsByService
	javaclassmethods.GetAllServicesFunc = testdata.MockListJavaClassMethodServices

	// Return cleanup function
	return func() {
		javaclassmethods.GetClassMethodsByServiceFunc = originalGetClassMethodsByService
		javaclassmethods.GetAllServicesFunc = originalGetAllServices
	}
}

func TestGetMethodByIndex(t *testing.T) {
	// Setup mocks
	cleanup := SetupJavaClassMethodsMock()
	defer cleanup()

	tests := []struct {
		name        string
		serviceName string
		index       int
		wantClass   string
		wantMethod  string
		wantNil     bool
	}{
		{
			name:        "Valid service and index",
			serviceName: "ts-auth-service",
			index:       0,
			wantClass:   "auth.AuthApplication",
			wantMethod:  "login",
			wantNil:     false,
		},
		{
			name:        "Valid service and index 1",
			serviceName: "ts-auth-service",
			index:       1,
			wantClass:   "auth.AuthService",
			wantMethod:  "verifyCode",
			wantNil:     false,
		},
		{
			name:        "Valid service but out of bounds index",
			serviceName: "ts-auth-service",
			index:       10,
			wantClass:   "auth.AuthApplication", // Should return first method
			wantMethod:  "login",
			wantNil:     false,
		},
		{
			name:        "Non-existent service",
			serviceName: "non-existent-service",
			index:       0,
			wantNil:     true,
		},
		{
			name:        "Negative index",
			serviceName: "ts-auth-service",
			index:       -1,
			wantClass:   "auth.AuthApplication", // Should return first method
			wantMethod:  "login",
			wantNil:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := javaclassmethods.GetMethodByIndex(tt.serviceName, tt.index)

			if tt.wantNil {
				if method != nil {
					t.Errorf("GetMethodByIndex() = %v, want nil", method)
				}
				return
			}

			if method == nil {
				t.Errorf("GetMethodByIndex() = nil, want non-nil")
				return
			}

			if method.ClassName != tt.wantClass {
				t.Errorf("GetMethodByIndex() class = %v, want %v", method.ClassName, tt.wantClass)
			}

			if method.MethodName != tt.wantMethod {
				t.Errorf("GetMethodByIndex() method = %v, want %v", method.MethodName, tt.wantMethod)
			}
		})
	}
}

func TestGetRandomMethod(t *testing.T) {
	// Setup mocks
	cleanup := SetupJavaClassMethodsMock()
	defer cleanup()

	tests := []struct {
		name        string
		serviceName string
		wantNil     bool
	}{
		{
			name:        "Service with methods",
			serviceName: "ts-auth-service",
			wantNil:     false,
		},
		{
			name:        "Non-existent service",
			serviceName: "non-existent-service",
			wantNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := javaclassmethods.GetRandomMethod(tt.serviceName)

			if tt.wantNil {
				if method != nil {
					t.Errorf("GetRandomMethod() = %v, want nil", method)
				}
				return
			}

			if method == nil {
				t.Errorf("GetRandomMethod() = nil, want non-nil")
			}
		})
	}
}

func TestGetMethodByIndexOrRandom(t *testing.T) {
	// Setup mocks
	cleanup := SetupJavaClassMethodsMock()
	defer cleanup()

	tests := []struct {
		name        string
		serviceName string
		index       int
		wantNil     bool
	}{
		{
			name:        "Valid service and index",
			serviceName: "ts-auth-service",
			index:       0,
			wantNil:     false,
		},
		{
			name:        "Valid service and out of bounds index",
			serviceName: "ts-auth-service",
			index:       10,
			wantNil:     false, // Should return a random method
		},
		{
			name:        "Non-existent service",
			serviceName: "non-existent-service",
			index:       0,
			wantNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := javaclassmethods.GetMethodByIndexOrRandom(tt.serviceName, tt.index)

			if tt.wantNil {
				if method != nil {
					t.Errorf("GetMethodByIndexOrRandom() = %v, want nil", method)
				}
				return
			}

			if method == nil {
				t.Errorf("GetMethodByIndexOrRandom() = nil, want non-nil")
			}
		})
	}
}

func TestCountMethods(t *testing.T) {
	// Setup mocks
	cleanup := SetupJavaClassMethodsMock()
	defer cleanup()

	tests := []struct {
		name        string
		serviceName string
		want        int
	}{
		{
			name:        "Service with multiple methods",
			serviceName: "ts-auth-service",
			want:        2,
		},
		{
			name:        "Service with one method",
			serviceName: "ts-order-service",
			want:        1,
		},
		{
			name:        "Non-existent service",
			serviceName: "non-existent-service",
			want:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := javaclassmethods.CountMethods(tt.serviceName)

			if count != tt.want {
				t.Errorf("CountMethods() = %v, want %v", count, tt.want)
			}
		})
	}
}

func TestGetMethodDisplayName(t *testing.T) {
	tests := []struct {
		name  string
		entry javaclassmethods.ClassMethodEntry
		want  string
	}{
		{
			name: "Simple class name",
			entry: javaclassmethods.ClassMethodEntry{
				ClassName:  "SimpleClass",
				MethodName: "testMethod",
			},
			want: "SimpleClass.testMethod",
		},
		{
			name: "Class with package",
			entry: javaclassmethods.ClassMethodEntry{
				ClassName:  "com.example.TestClass",
				MethodName: "doSomething",
			},
			want: "TestClass.doSomething",
		},
		{
			name: "Multiple packages",
			entry: javaclassmethods.ClassMethodEntry{
				ClassName:  "org.apache.commons.lang.StringUtils",
				MethodName: "isEmpty",
			},
			want: "StringUtils.isEmpty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			displayName := javaclassmethods.GetMethodDisplayName(tt.entry)

			if displayName != tt.want {
				t.Errorf("GetMethodDisplayName() = %v, want %v", displayName, tt.want)
			}
		})
	}
}

func TestListAllServiceNames(t *testing.T) {
	// Setup mocks
	cleanup := SetupJavaClassMethodsMock()
	defer cleanup()

	serviceNames := javaclassmethods.ListAllServiceNames()

	if len(serviceNames) == 0 {
		t.Errorf("ListAllServiceNames() returned empty list, expected service names")
	}

	// Check that the expected service names are in the list
	expectedServices := []string{
		"ts-auth-service",
		"ts-order-service",
		"ts-travel-service",
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
			t.Errorf("ListAllServiceNames() missing expected service: %s", expected)
		}
	}
}

func TestListAvailableMethods(t *testing.T) {
	// Setup mocks
	cleanup := SetupJavaClassMethodsMock()
	defer cleanup()

	tests := []struct {
		name        string
		serviceName string
		wantCount   int
		wantEmpty   bool
	}{
		{
			name:        "Service with multiple methods",
			serviceName: "ts-auth-service",
			wantCount:   2,
			wantEmpty:   false,
		},
		{
			name:        "Service with one method",
			serviceName: "ts-order-service",
			wantCount:   1,
			wantEmpty:   false,
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
			methods := javaclassmethods.ListAvailableMethods(tt.serviceName)

			if len(methods) != tt.wantCount {
				t.Errorf("ListAvailableMethods() returned %d methods, want %d", len(methods), tt.wantCount)
			}

			if (len(methods) == 0) != tt.wantEmpty {
				t.Errorf("ListAvailableMethods() empty status = %v, want %v", len(methods) == 0, tt.wantEmpty)
			}

			// Check format of method names
			if !tt.wantEmpty {
				for _, method := range methods {
					if len(method) == 0 {
						t.Errorf("ListAvailableMethods() returned empty method name")
					}

					// Should have a dot separating class and method
					if !ContainsChar(method, '.') {
						t.Errorf("ListAvailableMethods() returned method without class separator: %s", method)
					}
				}
			}
		})
	}
}

// Helper function to check if a string contains a specific character
func ContainsChar(s string, c rune) bool {
	for _, r := range s {
		if r == c {
			return true
		}
	}
	return false
}
