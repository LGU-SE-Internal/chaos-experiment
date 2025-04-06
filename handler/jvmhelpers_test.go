package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectJVMMethodForService(t *testing.T) {
	cleanup := setupJVMMocks()
	defer cleanup()

	tests := []struct {
		name          string
		serviceName   string
		methodIndex   int
		wantClassName string
		wantMethod    string
		wantErr       bool
	}{
		{
			name:          "Valid service and method index",
			serviceName:   "ts-auth-service",
			methodIndex:   0,
			wantClassName: "auth.AuthApplication",
			wantMethod:    "login",
			wantErr:       false,
		},
		{
			name:          "Valid service and second method index",
			serviceName:   "ts-auth-service",
			methodIndex:   1,
			wantClassName: "auth.AuthService",
			wantMethod:    "verifyCode",
			wantErr:       false,
		},
		{
			name:        "Valid service but out of bounds method index",
			serviceName: "ts-auth-service",
			methodIndex: 100,
			wantErr:     true, // Should fail with out of bounds index
		},
		{
			name:          "Another service with method",
			serviceName:   "ts-order-service",
			methodIndex:   0,
			wantClassName: "order.OrderService",
			wantMethod:    "createOrder",
			wantErr:       false,
		},
		{
			name:        "Non-existent service",
			serviceName: "non-existent-service",
			methodIndex: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			className, methodName, err := selectJVMMethodForService(tt.serviceName, tt.methodIndex)

			if (err != nil) != tt.wantErr {
				t.Errorf("selectJVMMethodForService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.Equal(t, tt.wantClassName, className, "Class name doesn't match expected")
				assert.Equal(t, tt.wantMethod, methodName, "Method name doesn't match expected")
			}
		})
	}
}

func TestGetAvailableJVMMethodsForApp(t *testing.T) {
	cleanup := setupJVMMocks()
	defer cleanup()

	tests := []struct {
		name      string
		appName   string
		wantCount int
		wantEmpty bool
		wantErr   bool
	}{
		{
			name:      "Service with multiple methods",
			appName:   "ts-auth-service",
			wantCount: 2,
			wantEmpty: false,
			wantErr:   false,
		},
		{
			name:      "Service with one method",
			appName:   "ts-order-service",
			wantCount: 1,
			wantEmpty: false,
			wantErr:   false,
		},
		{
			name:      "Non-existent service",
			appName:   "non-existent-service",
			wantCount: 0,
			wantEmpty: true,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			methods, err := getAvailableJVMMethodsForApp(tt.appName)

			if (err != nil) != tt.wantErr {
				t.Errorf("getAvailableJVMMethodsForApp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.Equal(t, tt.wantCount, len(methods), "Method count doesn't match expected")
				assert.Equal(t, tt.wantEmpty, len(methods) == 0, "Empty status doesn't match expected")

				// Verify the methods have the correct format if not empty
				if !tt.wantEmpty {
					for _, method := range methods {
						assert.NotEmpty(t, method, "Method descriptor should not be empty")
						// Check it has the expected format "ClassName.methodName"
						assert.Contains(t, method, ".", "Method descriptor should have Class.method format")
					}
				}
			}
		})
	}
}

func TestGetServiceAndMethodForChaosSpec(t *testing.T) {
	cleanup := setupJVMMocks()
	defer cleanup()

	tests := []struct {
		name         string
		appNameIndex int
		methodIndex  int
		wantAppName  string
		wantClass    string
		wantMethod   string
		wantErr      bool
	}{
		{
			name:         "Valid app and method indices",
			appNameIndex: 0, // ts-auth-service
			methodIndex:  0,
			wantAppName:  "ts-auth-service",
			wantClass:    "auth.AuthApplication",
			wantMethod:   "login",
			wantErr:      false,
		},
		{
			name:         "Valid app but out-of-bounds method index",
			appNameIndex: 0,
			methodIndex:  100,
			wantAppName:  "ts-auth-service",
			wantErr:      true,
		},
		{
			name:         "Invalid app index",
			appNameIndex: 20,
			methodIndex:  0,
			wantAppName:  "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appName, className, methodName, err := getServiceAndMethodForChaosSpec(tt.appNameIndex, tt.methodIndex)

			if (err != nil) != tt.wantErr {
				t.Errorf("getServiceAndMethodForChaosSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantAppName, appName, "App name doesn't match expected")

			if !tt.wantErr {
				assert.Equal(t, tt.wantClass, className, "Class name doesn't match expected")
				assert.Equal(t, tt.wantMethod, methodName, "Method name doesn't match expected")
			}
		})
	}
}

func TestGetJVMMethods(t *testing.T) {
	cleanup := setupJVMMocks()
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
			methods := GetJVMMethods(tt.serviceName)

			assert.Equal(t, tt.wantCount, len(methods), "Method count doesn't match expected")
			assert.Equal(t, tt.wantEmpty, len(methods) == 0, "Empty status doesn't match expected")

			// Verify method entries have expected properties
			if !tt.wantEmpty {
				for _, method := range methods {
					assert.NotNil(t, method, "Method entry should not be nil")
					assert.NotEmpty(t, method.ClassName, "Method entry should have a class name")
					assert.NotEmpty(t, method.MethodName, "Method entry should have a method name")
				}
			}
		})
	}
}

func TestGetJVMMethodsForApp(t *testing.T) {
	cleanup := setupJVMMocks()
	defer cleanup()

	tests := []struct {
		name      string
		appName   string
		wantCount int
		wantEmpty bool
	}{
		{
			name:      "Existing app with methods",
			appName:   "ts-auth-service",
			wantCount: 2,
			wantEmpty: false,
		},
		{
			name:      "Non-existent app",
			appName:   "non-existent-app",
			wantCount: 0,
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			methods := GetJVMMethodsForApp(tt.appName)

			assert.Equal(t, tt.wantCount, len(methods), "Method count doesn't match expected")
			assert.Equal(t, tt.wantEmpty, len(methods) == 0, "Empty status doesn't match expected")

			// Verify method entries have expected properties
			if !tt.wantEmpty {
				for _, method := range methods {
					assert.NotNil(t, method, "Method entry should not be nil")
					assert.NotEmpty(t, method.ClassName, "Method entry should have a class name")
					assert.NotEmpty(t, method.MethodName, "Method entry should have a method name")
				}
			}
		})
	}
}

func TestListJVMServiceNames(t *testing.T) {
	cleanup := setupJVMMocks()
	defer cleanup()

	serviceNames := ListJVMServiceNames()

	// Check that we get back a non-empty list
	assert.NotEmpty(t, serviceNames, "ListJVMServiceNames() returned empty list")

	// Check that the list contains all expected service names
	expectedServices := []string{
		"ts-auth-service",
		"ts-order-service",
		"ts-travel-service",
	}

	// Verify all expected services are in the list
	for _, expected := range expectedServices {
		assert.Contains(t, serviceNames, expected, "ListJVMServiceNames() missing expected service: %s", expected)
	}

	// Verify no unexpected services are in the list
	assert.Equal(t, len(expectedServices), len(serviceNames), "Unexpected number of service names")
}

func TestJVMHelpersIntegration(t *testing.T) {
	cleanup := setupJVMMocks()
	defer cleanup()

	// Test that our helper functions work well together
	serviceNames := ListJVMServiceNames()
	assert.NotEmpty(t, serviceNames, "No service names returned")

	serviceName := "ts-auth-service" // Use a service we know has methods
	methods := GetJVMMethods(serviceName)
	assert.NotEmpty(t, methods, "No methods returned for service")

	// Verify that the methods returned by different functions are consistent
	methodsFromApp := GetJVMMethodsForApp(serviceName)
	assert.Equal(t, methods, methodsFromApp, "Methods from GetJVMMethods and GetJVMMethodsForApp are not consistent")

	// Check that selectJVMMethodForService works with our mock data
	className, methodName, err := selectJVMMethodForService(serviceName, 0)
	assert.NoError(t, err, "selectJVMMethodForService() failed for valid service and index")
	assert.Equal(t, "auth.AuthApplication", className, "Unexpected class name")
	assert.Equal(t, "login", methodName, "Unexpected method name")

	// Test getServiceAndMethodForChaosSpec with the mocked data
	appName, className, methodName, err := getServiceAndMethodForChaosSpec(0, 0)
	assert.NoError(t, err, "getServiceAndMethodForChaosSpec() failed with valid indices")
	assert.Equal(t, "ts-auth-service", appName, "Unexpected app name")
	assert.Equal(t, "auth.AuthApplication", className, "Unexpected class name")
	assert.Equal(t, "login", methodName, "Unexpected method name")
}
