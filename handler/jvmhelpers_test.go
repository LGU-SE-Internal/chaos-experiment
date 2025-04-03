package handler

import (
	"reflect"
	"testing"
)

// Mock for GetLabels
func mockGetLabels(namespace, labelKey string) ([]string, error) {
	return []string{"ts-auth-service", "ts-order-service", "ts-travel-service"}, nil
}

// Setup and teardown for tests, overriding labelsGetter
func setupMocks() func() {
	originalGetLabels := labelsGetter
	labelsGetter = mockGetLabels

	return func() {
		labelsGetter = originalGetLabels
	}
}

func TestSelectJVMMethodForService(t *testing.T) {
	tests := []struct {
		name          string
		appName       string
		methodIndex   int
		wantClassName string
		wantOK        bool
	}{
		{
			name:          "Valid method index",
			appName:       "ts-auth-service",
			methodIndex:   0,
			wantClassName: "auth.AuthApplication",
			wantOK:        true,
		},
		{
			name:        "Out of bounds method index should return random method",
			appName:     "ts-auth-service",
			methodIndex: 1000,
			wantOK:      true, // Should still return a method
		},
		{
			name:          "Non-existent service",
			appName:       "non-existent-service",
			methodIndex:   0,
			wantClassName: "",
			wantOK:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			className, methodName, ok := selectJVMMethodForService(tt.appName, tt.methodIndex)

			if ok != tt.wantOK {
				t.Errorf("selectJVMMethodForService() ok = %v, want %v", ok, tt.wantOK)
				return
			}

			if tt.wantOK && tt.wantClassName != "" && className != tt.wantClassName {
				t.Errorf("selectJVMMethodForService() className = %v, want %v", className, tt.wantClassName)
			}

			if tt.wantOK && methodName == "" {
				t.Errorf("selectJVMMethodForService() methodName is empty, expected a value")
			}
		})
	}
}

func TestGetAvailableJVMMethodsForApp(t *testing.T) {
	tests := []struct {
		name      string
		appName   string
		wantEmpty bool
	}{
		{
			name:      "Existing service",
			appName:   "ts-auth-service",
			wantEmpty: false,
		},
		{
			name:      "Non-existent service",
			appName:   "non-existent-service",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			methods := getAvailableJVMMethodsForApp(tt.appName)

			if tt.wantEmpty && len(methods) > 0 {
				t.Errorf("getAvailableJVMMethodsForApp() returned %d methods, expected empty list", len(methods))
			}

			if !tt.wantEmpty && len(methods) == 0 {
				t.Errorf("getAvailableJVMMethodsForApp() returned empty list, expected methods")
			}
		})
	}
}

func TestGetServiceAndMethodForChaosSpec(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	tests := []struct {
		name         string
		appNameIndex int
		methodIndex  int
		wantAppName  string
		wantOK       bool
	}{
		{
			name:         "Valid app and method indices",
			appNameIndex: 0,
			methodIndex:  0,
			wantAppName:  "ts-auth-service",
			wantOK:       true,
		},
		{
			name:         "Invalid app index",
			appNameIndex: 10,
			methodIndex:  0,
			wantAppName:  "",
			wantOK:       false,
		},
		{
			name:         "Valid app but invalid method index should return a random method",
			appNameIndex: 0,
			methodIndex:  1000,
			wantAppName:  "ts-auth-service",
			wantOK:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appName, className, methodName, ok := getServiceAndMethodForChaosSpec(tt.appNameIndex, tt.methodIndex)

			if ok != tt.wantOK {
				t.Errorf("getServiceAndMethodForChaosSpec() ok = %v, want %v", ok, tt.wantOK)
				return
			}

			if tt.wantOK {
				if appName != tt.wantAppName {
					t.Errorf("getServiceAndMethodForChaosSpec() appName = %v, want %v", appName, tt.wantAppName)
				}

				if className == "" {
					t.Errorf("getServiceAndMethodForChaosSpec() className is empty")
				}

				if methodName == "" {
					t.Errorf("getServiceAndMethodForChaosSpec() methodName is empty")
				}
			}
		})
	}
}

func TestGetJVMMethods(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		wantEmpty   bool
	}{
		{
			name:        "Existing service",
			serviceName: "ts-auth-service",
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
			methods := GetJVMMethods(tt.serviceName)

			if tt.wantEmpty && len(methods) > 0 {
				t.Errorf("GetJVMMethods() returned %d methods, expected empty list", len(methods))
			}

			if !tt.wantEmpty && len(methods) == 0 {
				t.Errorf("GetJVMMethods() returned empty list, expected methods")
			}
		})
	}
}

func TestGetJVMMethodsForApp(t *testing.T) {
	tests := []struct {
		name      string
		appName   string
		wantEmpty bool
	}{
		{
			name:      "Existing app",
			appName:   "ts-auth-service",
			wantEmpty: false,
		},
		{
			name:      "Non-existent app",
			appName:   "non-existent-app",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			methods := GetJVMMethodsForApp(tt.appName)

			if tt.wantEmpty && len(methods) > 0 {
				t.Errorf("GetJVMMethodsForApp() returned %d methods, expected empty list", len(methods))
			}

			if !tt.wantEmpty && len(methods) == 0 {
				t.Errorf("GetJVMMethodsForApp() returned empty list, expected methods")
			}
		})
	}
}

func TestListJVMServiceNames(t *testing.T) {
	serviceNames := ListJVMServiceNames()

	// Check that we get back a non-empty list
	if len(serviceNames) == 0 {
		t.Errorf("ListJVMServiceNames() returned empty list, expected service names")
	}

	// Check that the list contains the expected service names
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
			t.Errorf("ListJVMServiceNames() missing expected service: %s", expected)
		}
	}
}

func TestJVMHelpersIntegration(t *testing.T) {
	// Test that our helper functions work well together
	serviceNames := ListJVMServiceNames()
	if len(serviceNames) == 0 {
		t.Fatal("No service names returned")
	}

	serviceName := serviceNames[0]
	methods := GetJVMMethods(serviceName)
	if len(methods) == 0 {
		t.Fatalf("No methods returned for service %s", serviceName)
	}

	// Verify that the methods returned by different functions are consistent
	methodsFromApp := GetJVMMethodsForApp(serviceName)
	if !reflect.DeepEqual(methods, methodsFromApp) {
		t.Errorf("Methods from GetJVMMethods and GetJVMMethodsForApp are not consistent")
	}

	// Check that listing available methods works for this service
	methodNames := getAvailableJVMMethodsForApp(serviceName)
	if len(methodNames) != len(methods) {
		t.Errorf("Method count mismatch: got %d names but %d methods", len(methodNames), len(methods))
	}
}
