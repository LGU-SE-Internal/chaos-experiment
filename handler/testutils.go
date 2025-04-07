package handler

import (
	"github.com/CUHK-SE-Group/chaos-experiment/internal/testdata"
)

func setupCommonMocks() func() {
	// Save original
	originalLabelsGetter := labelsGetter

	// Replace with mock
	labelsGetter = testdata.MockGetLabels

	// Return cleanup function
	return func() {
		labelsGetter = originalLabelsGetter
	}
}

// Save original functions for restoration
var originalHTTPEndpointsGetter EndpointsGetterFunc

// setupHTTPMocks sets up mocks for HTTP tests
func setupHTTPMocks() func() {
	// Setup common mocks
	commonCleanup := setupCommonMocks()

	// Save original
	originalEndpointsGetter := endpointsGetter

	// Replace with mock
	endpointsGetter = testdata.MockGetEndpoints

	// Return cleanup
	return func() {
		commonCleanup()
		endpointsGetter = originalEndpointsGetter
	}
}

// Save original functions for restoration
var originalEndpointsGetter EndpointsGetterFunc
var originalLabelsGetter func(string, string) ([]string, error)

// setupDNSMocks sets up mocks for DNS tests
func setupDNSMocks() func() {
	// Setup common mocks
	commonCleanup := setupCommonMocks()

	// Save original
	originalEndpointsGetter := endpointsGetter

	// Replace with mock
	endpointsGetter = testdata.MockGetEndpoints

	// Return cleanup
	return func() {
		commonCleanup()
		endpointsGetter = originalEndpointsGetter
	}
}

// Save original functions for restoration
var originalJVMMethodGetter JavaMethodGetterFunc
var originalJVMMethodsGetter JavaMethodsGetterFunc
var originalJVMServicesGetter JavaServicesGetterFunc
var originalJVMMethodsLister JavaMethodsListerFunc
var originalLabelsGetterForJVM func(string, string) ([]string, error)

// setupJVMMocks sets up mocks for JVM tests
func setupJVMMocks() func() {
	// Setup common mocks
	commonCleanup := setupCommonMocks()

	// Save originals
	originalJVMMethodGetter := javaMethodGetter
	originalJVMMethodsGetter := javaMethodsGetter
	originalJVMServicesGetter := javaServicesGetter
	originalJVMMethodsLister := javaMethodsLister

	// Replace with mock implementations
	javaMethodGetter = testdata.MockGetJavaClassMethod
	javaMethodsGetter = testdata.MockGetClassMethodsByService
	javaServicesGetter = testdata.MockListJavaClassMethodServices
	javaMethodsLister = testdata.MockListAvailableMethods

	// Return cleanup
	return func() {
		commonCleanup()
		javaMethodGetter = originalJVMMethodGetter
		javaMethodsGetter = originalJVMMethodsGetter
		javaServicesGetter = originalJVMServicesGetter
		javaMethodsLister = originalJVMMethodsLister
	}
}

// setupNetworkMocks sets up mocks for network tests
func setupNetworkMocks() func() {
	// Network mocks rely on testdata.SetupNetworkDependenciesMock()
	networkCleanup := testdata.SetupNetworkDependenciesMock()

	// Setup common mocks
	commonCleanup := setupCommonMocks()

	// Return cleanup
	return func() {
		commonCleanup()
		networkCleanup()
	}
}
