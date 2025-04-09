package handler

import (
	"github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/serviceendpoints"
)

type (
	// Label getter function
	LabelsGetterFunc func(string, string) ([]string, error)

	// Endpoints getter function
	EndpointsGetterFunc func(string) []serviceendpoints.ServiceEndpoint
)

// Package variables for dependency injection and mocking
var (
	// Common label getter
	labelsGetter = client.GetLabels

	// Endpoints getter
	endpointsGetter EndpointsGetterFunc = serviceendpoints.GetEndpointsByService
)
