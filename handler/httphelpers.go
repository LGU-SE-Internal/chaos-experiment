package handler

import (
	"strconv"

	"github.com/CUHK-SE-Group/chaos-experiment/chaos"
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
