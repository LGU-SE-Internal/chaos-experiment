// Package resourcetypes defines common types for resource data from ClickHouse analysis.
// These types are used across all system-specific resource packages to avoid duplication.
package resourcetypes

// EndpointType represents the type of endpoint for filtering by fault type
type EndpointType string

const (
	// EndpointTypeHTTP represents HTTP/REST endpoints
	EndpointTypeHTTP EndpointType = "http"
	// EndpointTypeDatabase represents database operations
	EndpointTypeDatabase EndpointType = "database"
	// EndpointTypeGRPC represents gRPC operations
	EndpointTypeGRPC EndpointType = "grpc"
)

// ServiceEndpoint represents a service endpoint from ClickHouse analysis
type ServiceEndpoint struct {
	ServiceName    string
	RequestMethod  string
	ResponseStatus string
	Route          string
	ServerAddress  string
	ServerPort     string
	SpanName       string
	EndpointType   EndpointType // Type marker for filtering
}

// DatabaseOperation represents a database operation from ClickHouse analysis
type DatabaseOperation struct {
	ServiceName   string
	DBName        string
	DBTable       string
	Operation     string
	DBSystem      string
	ServerAddress string
	ServerPort    string
}

// GRPCOperation represents a gRPC operation from ClickHouse analysis
type GRPCOperation struct {
	ServiceName    string
	RPCSystem      string
	RPCService     string
	RPCMethod      string
	GRPCStatusCode string
	ServerAddress  string
	ServerPort     string
	SpanKind       string
}
