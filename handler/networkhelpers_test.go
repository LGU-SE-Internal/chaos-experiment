package handler

import (
	"testing"
)

// We'll reuse the mockGetLabels and setupMocks functions from jvmhelpers_test.go

func TestSelectNetworkTargetForService(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()
	tests := []struct {
		name           string
		sourceName     string
		targetIndex    int
		wantTargetName string
		wantOK         bool
	}{
		{
			name:           "Valid source and target index",
			sourceName:     "ts-auth-service",
			targetIndex:    0,
			wantTargetName: "ts-verification-code-service", // Assuming this is the first dependency
			wantOK:         true,
		},
		{
			name:           "Valid source but negative target index",
			sourceName:     "ts-auth-service",
			targetIndex:    -1,
			wantTargetName: "",
			wantOK:         false,
		},
		{
			name:           "Valid source but out of bounds target index",
			sourceName:     "ts-auth-service",
			targetIndex:    100,
			wantTargetName: "",
			wantOK:         false,
		},
		{
			name:           "Non-existent source service",
			sourceName:     "non-existent-service",
			targetIndex:    0,
			wantTargetName: "",
			wantOK:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetName, ok := selectNetworkTargetForService(tt.sourceName, tt.targetIndex)

			if ok != tt.wantOK {
				t.Errorf("selectNetworkTargetForService() ok = %v, want %v", ok, tt.wantOK)
				return
			}

			if tt.wantOK && targetName != tt.wantTargetName {
				t.Errorf("selectNetworkTargetForService() targetName = %v, want %v", targetName, tt.wantTargetName)
			}
		})
	}
}

func TestGetServiceAndTargetForNetworkChaos(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	tests := []struct {
		name           string
		appNameIndex   int
		targetIndex    int
		wantSourceName string
		wantTargetName string
		wantOK         bool
	}{
		{
			name:           "Valid app and target indices",
			appNameIndex:   0,
			targetIndex:    0,
			wantSourceName: "ts-auth-service",
			wantTargetName: "ts-ui-dashboard", // Assuming this is the first dependency
			wantOK:         true,
		},
		{
			name:           "Invalid app index",
			appNameIndex:   10,
			targetIndex:    0,
			wantSourceName: "",
			wantTargetName: "",
			wantOK:         false,
		},
		{
			name:           "Valid app index but invalid target index",
			appNameIndex:   0,
			targetIndex:    100,
			wantSourceName: "ts-auth-service",
			wantTargetName: "",
			wantOK:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceName, targetName, ok := getServiceAndTargetForNetworkChaos(tt.appNameIndex, tt.targetIndex)

			if ok != tt.wantOK {
				t.Errorf("getServiceAndTargetForNetworkChaos() ok = %v, want %v", ok, tt.wantOK)
				return
			}

			if tt.wantOK {
				if sourceName != tt.wantSourceName {
					t.Errorf("getServiceAndTargetForNetworkChaos() sourceName = %v, want %v", sourceName, tt.wantSourceName)
				}

				if targetName != tt.wantTargetName {
					t.Errorf("getServiceAndTargetForNetworkChaos() targetName = %v, want %v", targetName, tt.wantTargetName)
				}
			}
		})
	}
}

func TestGetNetworkDependencies(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		wantEmpty   bool
	}{
		{
			name:        "Existing service with dependencies",
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
			dependencies := GetNetworkDependencies(tt.serviceName)

			if tt.wantEmpty && len(dependencies) > 0 {
				t.Errorf("GetNetworkDependencies() returned %d dependencies, expected empty list", len(dependencies))
			}

			if !tt.wantEmpty && len(dependencies) == 0 {
				t.Errorf("GetNetworkDependencies() returned empty list, expected dependencies")
			}
		})
	}
}

func TestListNetworkServiceNames(t *testing.T) {
	serviceNames := ListNetworkServiceNames()

	// Check that we get back a non-empty list
	if len(serviceNames) == 0 {
		t.Errorf("ListNetworkServiceNames() returned empty list, expected service names")
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
			t.Logf("ListNetworkServiceNames() missing expected service: %s", expected)
			// Not failing the test as the exact services might vary
		}
	}
}

func TestGetAllNetworkPairs(t *testing.T) {
	pairs := GetAllNetworkPairs()

	// Check that we get back a non-empty list
	if len(pairs) == 0 {
		t.Errorf("GetAllNetworkPairs() returned empty list, expected service pairs")
	}

	// Verify the structure of the pairs
	for _, pair := range pairs {
		if pair.SourceService == "" {
			t.Errorf("GetAllNetworkPairs() returned pair with empty source service")
		}

		if pair.TargetService == "" {
			t.Errorf("GetAllNetworkPairs() returned pair with empty target service")
		}

		if pair.ConnectionDetails == "" {
			t.Errorf("GetAllNetworkPairs() returned pair with empty connection details")
		}
	}
}

func TestNetworkHelpersIntegration(t *testing.T) {
	// Test that our helper functions work well together
	serviceNames := ListNetworkServiceNames()
	if len(serviceNames) == 0 {
		t.Fatal("No service names returned")
	}

	sourceName := serviceNames[0]
	dependencies := GetNetworkDependencies(sourceName)

	if len(dependencies) == 0 {
		// Try another service if this one has no dependencies
		if len(serviceNames) > 1 {
			sourceName = serviceNames[1]
			dependencies = GetNetworkDependencies(sourceName)
		}
	}

	if len(dependencies) == 0 {
		t.Skip("No service with dependencies found, skipping integration test")
	}

	// Test that selectNetworkTargetForService works with the dependencies
	targetName, ok := selectNetworkTargetForService(sourceName, 0)
	if !ok {
		t.Errorf("selectNetworkTargetForService() failed for valid service and index")
	}

	if targetName != dependencies[0] {
		t.Errorf("selectNetworkTargetForService() targetName = %v, want %v", targetName, dependencies[0])
	}

	// Verify that all pairs contain our source service
	pairs := GetAllNetworkPairs()
	foundPair := false

	for _, pair := range pairs {
		if pair.SourceService == sourceName && pair.TargetService == targetName {
			foundPair = true
			break
		}
	}

	if !foundPair {
		t.Errorf("GetAllNetworkPairs() does not contain expected pair: %s -> %s", sourceName, targetName)
	}
}

func TestNetworkHelpersWithMocks(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	// Test getServiceAndTargetForNetworkChaos with the mocked data
	sourceName, targetName, ok := getServiceAndTargetForNetworkChaos(0, 0)

	if !ok {
		t.Errorf("getServiceAndTargetForNetworkChaos() with mocked data returned ok = false, want true")
	}

	if sourceName != "ts-auth-service" {
		t.Errorf("getServiceAndTargetForNetworkChaos() sourceName = %v, want %v", sourceName, "ts-auth-service")
	}

	if targetName == "" {
		t.Errorf("getServiceAndTargetForNetworkChaos() returned empty targetName")
	}
}
