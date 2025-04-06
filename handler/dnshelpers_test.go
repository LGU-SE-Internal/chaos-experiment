package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectDNSPatternsForService(t *testing.T) {
	cleanup := setupDNSMocks()
	defer cleanup()

	tests := []struct {
		name          string
		serviceName   string
		endpointIndex int
		wantPatterns  []string
		wantErr       bool
	}{
		{
			name:          "Service with endpoint with server address",
			serviceName:   "ts-auth-service",
			endpointIndex: 0,
			wantPatterns:  []string{"ts-verification-code-service"},
			wantErr:       false,
		},
		{
			name:          "Service with database endpoint",
			serviceName:   "ts-auth-service",
			endpointIndex: 1,
			wantPatterns:  []string{"mysql"},
			wantErr:       false,
		},
		{
			name:          "Service with multiple endpoints, all patterns",
			serviceName:   "ts-auth-service",
			endpointIndex: -1,
			wantPatterns:  []string{"ts-verification-code-service", "mysql"},
			wantErr:       false,
		},
		{
			name:          "Service with no valid patterns",
			serviceName:   "ts-empty-service",
			endpointIndex: 0,
			wantPatterns:  []string{"*"},
			wantErr:       true,
		},
		{
			name:          "Service with only self reference",
			serviceName:   "ts-self-service",
			endpointIndex: 0,
			wantPatterns:  []string{"*"},
			wantErr:       true,
		},
		{
			name:          "Service with index out of range",
			serviceName:   "ts-auth-service",
			endpointIndex: 10, // Out of range
			wantPatterns:  []string{"ts-verification-code-service", "mysql"},
			wantErr:       false, // Still returns all patterns
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns, err := selectDNSPatternsForService(tt.serviceName, tt.endpointIndex)

			if (err != nil) != tt.wantErr {
				t.Errorf("selectDNSPatternsForService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For multiple patterns we need to check they contain the same elements
			// regardless of order
			if !tt.wantErr {
				if len(patterns) > 1 && len(tt.wantPatterns) > 1 {
					assert.ElementsMatch(t, tt.wantPatterns, patterns, "Patterns don't match expected")
				} else {
					assert.Equal(t, tt.wantPatterns, patterns, "Patterns don't match expected")
				}
			}
		})
	}
}

func TestGetServiceAndPatternsForDNSChaos(t *testing.T) {
	cleanupEndpoints := setupDNSMocks()
	defer cleanupEndpoints()

	cleanupLabels := setupDNSMocks()
	defer cleanupLabels()

	tests := []struct {
		name            string
		appNameIndex    int
		endpointIndex   int
		wantServiceName string
		wantPatterns    []string
		wantErr         bool
	}{
		{
			name:            "Valid app name and endpoint index",
			appNameIndex:    0, // ts-auth-service is at index 0 in mockLabels
			endpointIndex:   0,
			wantServiceName: "ts-auth-service",
			wantPatterns:    []string{"ts-verification-code-service"},
			wantErr:         false,
		},
		{
			name:            "Valid app name and all endpoints",
			appNameIndex:    0, // ts-auth-service
			endpointIndex:   -1,
			wantServiceName: "ts-auth-service",
			wantPatterns:    []string{"ts-verification-code-service", "mysql"},
			wantErr:         false,
		},
		{
			name:            "Service with database endpoint",
			appNameIndex:    1, // ts-order-service is at index 1 and has mysql endpoint
			endpointIndex:   0,
			wantServiceName: "ts-order-service",
			wantPatterns:    []string{"mysql"},
			wantErr:         false,
		},
		{
			name:            "Invalid app name index",
			appNameIndex:    10, // Out of range
			endpointIndex:   0,
			wantServiceName: "",
			wantPatterns:    []string{"*"},
			wantErr:         true,
		},
		{
			name:            "Service with no valid endpoints",
			appNameIndex:    3, // ts-empty-service is at index 3
			endpointIndex:   0,
			wantServiceName: "ts-empty-service",
			wantPatterns:    []string{"*"},
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceName, patterns, err := getServiceAndPatternsForDNSChaos(tt.appNameIndex, tt.endpointIndex)

			if (err != nil) != tt.wantErr {
				t.Errorf("getServiceAndPatternsForDNSChaos() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantServiceName, serviceName, "Service name doesn't match expected")

			// For multiple patterns we need to check they contain the same elements
			// regardless of order
			if !tt.wantErr && len(patterns) > 0 {
				if len(patterns) > 1 && len(tt.wantPatterns) > 1 {
					assert.ElementsMatch(t, tt.wantPatterns, patterns, "Patterns don't match expected")
				} else {
					assert.Equal(t, tt.wantPatterns, patterns, "Patterns don't match expected")
				}
			}
		})
	}
}

func TestGetDNSEndpoints(t *testing.T) {
	cleanup := setupDNSMocks()
	defer cleanup()

	tests := []struct {
		name          string
		serviceName   string
		wantEndpoints []string
	}{
		{
			name:          "Service with multiple endpoints",
			serviceName:   "ts-auth-service",
			wantEndpoints: []string{"ts-verification-code-service", "mysql"},
		},
		{
			name:          "Service with one endpoint",
			serviceName:   "ts-travel-service",
			wantEndpoints: []string{"ts-route-service"},
		},
		{
			name:          "Service with no valid endpoints",
			serviceName:   "ts-empty-service",
			wantEndpoints: []string{},
		},
		{
			name:          "Service with only self reference",
			serviceName:   "ts-self-service",
			wantEndpoints: []string{},
		},
		{
			name:          "Non-existent service",
			serviceName:   "non-existent",
			wantEndpoints: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoints := GetDNSEndpoints(tt.serviceName)

			// Check we have the same elements regardless of order
			assert.ElementsMatch(t, tt.wantEndpoints, endpoints, "Endpoints don't match expected")
		})
	}
}

func TestCountDNSEndpoints(t *testing.T) {
	cleanup := setupDNSMocks()
	defer cleanup()

	tests := []struct {
		name        string
		serviceName string
		wantCount   int
	}{
		{
			name:        "Service with multiple endpoints",
			serviceName: "ts-auth-service",
			wantCount:   2,
		},
		{
			name:        "Service with one endpoint",
			serviceName: "ts-travel-service",
			wantCount:   1,
		},
		{
			name:        "Service with no valid endpoints",
			serviceName: "ts-empty-service",
			wantCount:   0,
		},
		{
			name:        "Non-existent service",
			serviceName: "non-existent",
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := CountDNSEndpoints(tt.serviceName)
			assert.Equal(t, tt.wantCount, count, "Count doesn't match expected")
		})
	}
}
