package handler

import (
	"github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/javaclassmethods"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/serviceendpoints"
)

type (
	// Label getter function
	LabelsGetterFunc func(string, string) ([]string, error)

	// Endpoints getter function
	EndpointsGetterFunc func(string) []serviceendpoints.ServiceEndpoint

	// JVM related function types
	JavaMethodGetterFunc   func(string, int) *javaclassmethods.ClassMethodEntry
	JavaMethodsGetterFunc  func(string) []javaclassmethods.ClassMethodEntry
	JavaServicesGetterFunc func() []string
	JavaMethodsListerFunc  func(string) []string
)

// Package variables for dependency injection and mocking
var (
	// Common label getter
	labelsGetter = client.GetLabels

	// Endpoints getter
	endpointsGetter EndpointsGetterFunc = serviceendpoints.GetEndpointsByService

	// JVM function variables
	javaMethodGetter   JavaMethodGetterFunc   = javaclassmethods.GetMethodByIndexOrRandom
	javaMethodsGetter  JavaMethodsGetterFunc  = javaclassmethods.GetClassMethodsByService
	javaServicesGetter JavaServicesGetterFunc = javaclassmethods.ListAllServiceNames
	javaMethodsLister  JavaMethodsListerFunc  = javaclassmethods.ListAvailableMethods
)
