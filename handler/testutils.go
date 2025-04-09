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
