// Package resourcetypes defines common types for resource data from ClickHouse analysis.
// These types represent the raw data stored in internal data packages.
package resourcetypes

// HTTPEndpoint represents an HTTP/REST endpoint from ClickHouse analysis
// This stores only HTTP-specific fields
type HTTPEndpoint struct {
	ServiceName    string // Service making the call
	RequestMethod  string // GET, POST, etc.
	ResponseStatus string // HTTP status code
	Route          string // URL route/path
	ServerAddress  string // Target service
	ServerPort     string // Target port
	SpanKind       string // Server or Client
	SpanName       string // Span name for groundtruth
}

// DatabaseOperation represents a database operation from ClickHouse analysis
// This stores only database-specific fields
type DatabaseOperation struct {
	ServiceName   string // Service making the database call
	DBName        string // Database name
	DBTable       string // Table name
	Operation     string // SELECT, INSERT, UPDATE, DELETE
	DBSystem      string // mysql, postgresql, redis, etc.
	ServerAddress string // Database server address
	ServerPort    string // Database server port
	SpanName      string // Span name for groundtruth
}

// RPCOperation represents a gRPC/RPC operation from ClickHouse analysis
// This stores only RPC-specific fields
type RPCOperation struct {
	ServiceName    string // Service making the RPC call
	RPCSystem      string // grpc, thrift, etc.
	RPCService     string // RPC service name
	RPCMethod      string // RPC method name
	StatusCode     string // RPC status code
	ServerAddress  string // Target service
	ServerPort     string // Target port
	SpanKind       string // Server or Client
	SpanName       string // Span name for groundtruth
}

// Legacy type aliases for backward compatibility with existing generated code
// These will be removed after regeneration
type ServiceEndpoint = HTTPEndpoint
type GRPCOperation = RPCOperation
