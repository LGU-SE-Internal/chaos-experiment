package handler

import (
	"fmt"
	"strconv"

	"github.com/CUHK-SE-Group/chaos-experiment/chaos"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/serviceendpoints"
)

// HTTP Method enum
type HTTPMethod int

const (
	GET HTTPMethod = iota
	POST
	PUT
	DELETE
	HEAD
	OPTIONS
	PATCH
)

var httpMethodMap = map[HTTPMethod]string{
	GET:     "GET",
	POST:    "POST",
	PUT:     "PUT",
	DELETE:  "DELETE",
	HEAD:    "HEAD",
	OPTIONS: "OPTIONS",
	PATCH:   "PATCH",
}

// GetHTTPMethodName returns the string representation of an HTTP method
func GetHTTPMethodName(method HTTPMethod) string {
	if name, exists := httpMethodMap[method]; exists {
		return name
	}
	return "GET" // Default to GET
}

// HTTP Status Codes for replace
type HTTPStatusCode int

const (
	BadRequest HTTPStatusCode = iota
	Unauthorized
	Forbidden
	NotFound
	MethodNotAllowed
	RequestTimeout
	InternalServerError
	BadGateway
	ServiceUnavailable
	GatewayTimeout
)

var httpStatusCodeMap = map[HTTPStatusCode]int32{
	BadRequest:          400,
	Unauthorized:        401,
	Forbidden:           403,
	NotFound:            404,
	MethodNotAllowed:    405,
	RequestTimeout:      408,
	InternalServerError: 500,
	BadGateway:          502,
	ServiceUnavailable:  503,
	GatewayTimeout:      504,
}

// GetHTTPStatusCode returns the numeric HTTP status code
func GetHTTPStatusCode(statusCode HTTPStatusCode) int32 {
	if code, exists := httpStatusCodeMap[statusCode]; exists {
		return code
	}
	return 500 // Default to Internal Server Error
}

// HTTPEndpoint represents an HTTP endpoint for chaos testing
type HTTPEndpoint struct {
	ServiceName    string
	Route          string
	Method         string
	ResponseStatus string
	TargetService  string
	Port           string
}

// GetEndpointPort returns the port as an int32 with fallback to default 8080
func (e *HTTPEndpoint) GetEndpointPort() int32 {
	if e.Port == "" {
		return 8080 // Default port
	}

	if port, err := strconv.Atoi(e.Port); err == nil {
		return int32(port)
	}

	return 8080 // Default if conversion fails
}

// AddCommonHTTPOptions adds common HTTP options for prot, path and method if they exist
func AddCommonHTTPOptions(endpoint *HTTPEndpoint, opts []chaos.OptHTTPChaos) []chaos.OptHTTPChaos {

	opts = append(opts, chaos.WithPort(endpoint.GetEndpointPort())) // Always set the port
	// Add path if available
	if endpoint.Route != "" {
		path := endpoint.Route
		opts = append(opts, chaos.WithPath(&path))
	}

	// Add method if available
	if endpoint.Method != "" {
		method := endpoint.Method
		opts = append(opts, chaos.WithMethod(&method))
	}

	return opts
}

// selectHTTPEndpointForService selects an HTTP endpoint for a given service based on the endpoint index
// and returns the endpoint details and any error
func selectHTTPEndpointForService(serviceName string, endpointIndex int) (*HTTPEndpoint, error) {
	endpoints := endpointsGetter(serviceName)

	// Filter out non-HTTP endpoints (e.g., database connections)
	httpEndpoints := make([]serviceendpoints.ServiceEndpoint, 0)
	for _, ep := range endpoints {
		// Skip endpoints related to rabbitmq
		if ep.ServerAddress == "ts-rabbitmq" {
			continue
		}

		// Only include endpoints with a valid route
		if ep.Route != "" {
			httpEndpoints = append(httpEndpoints, ep)
		}
	}

	if len(httpEndpoints) == 0 {
		return nil, fmt.Errorf("no HTTP endpoints found for service %s", serviceName)
	}

	if endpointIndex < 0 || endpointIndex >= len(httpEndpoints) {
		return nil, fmt.Errorf("endpoint index %d out of range for service %s (max: %d)",
			endpointIndex, serviceName, len(httpEndpoints)-1)
	}

	ep := httpEndpoints[endpointIndex]
	return &HTTPEndpoint{
		ServiceName:    ep.ServiceName,
		Route:          ep.Route,
		Method:         ep.RequestMethod,
		ResponseStatus: ep.ResponseStatus,
		TargetService:  ep.ServerAddress,
		Port:           ep.ServerPort,
	}, nil
}

// getServiceAndEndpointForHTTPChaos is a helper function that retrieves the source and endpoint
// for an HTTP chaos specification
func getServiceAndEndpointForHTTPChaos(appNameIndex int, endpointIndex int) (serviceName string, endpoint *HTTPEndpoint, err error) {
	// Get the app labels
	labelArr, err := labelsGetter(TargetNamespace, TargetLabelKey)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get service labels: %w", err)
	}

	if appNameIndex < 0 || appNameIndex >= len(labelArr) {
		return "", nil, fmt.Errorf("app index %d out of range (max: %d)",
			appNameIndex, len(labelArr)-1)
	}

	serviceName = labelArr[appNameIndex]
	endpoint, err = selectHTTPEndpointForService(serviceName, endpointIndex)
	if err != nil {
		return serviceName, nil, err
	}

	return serviceName, endpoint, nil
}

// GetHTTPEndpoints returns all available HTTP endpoints for a service
func GetHTTPEndpoints(serviceName string) []HTTPEndpoint {
	endpoints := endpointsGetter(serviceName)
	result := make([]HTTPEndpoint, 0)

	for _, ep := range endpoints {
		// Skip endpoints related to rabbitmq
		if ep.ServerAddress == "ts-rabbitmq" {
			continue
		}

		// Only include endpoints with a valid route
		if ep.Route != "" {
			result = append(result, HTTPEndpoint{
				ServiceName:    ep.ServiceName,
				Route:          ep.Route,
				Method:         ep.RequestMethod,
				ResponseStatus: ep.ResponseStatus,
				TargetService:  ep.ServerAddress,
				Port:           ep.ServerPort,
			})
		}
	}

	return result
}

// ListHTTPServiceNames returns a list of all available services with HTTP endpoints
func ListHTTPServiceNames() []string {
	services := serviceendpoints.GetAllServices()
	result := make([]string, 0)

	for _, service := range services {
		endpoints := GetHTTPEndpoints(service)
		if len(endpoints) > 0 {
			result = append(result, service)
		}
	}

	return result
}

// CountHTTPEndpoints returns the number of HTTP endpoints for a service
func CountHTTPEndpoints(serviceName string) int {
	return len(GetHTTPEndpoints(serviceName))
}
