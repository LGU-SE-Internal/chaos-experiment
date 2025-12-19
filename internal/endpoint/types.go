// Package endpoint defines the unified endpoint types that resourcelookup provides to handlers.
// These are conversion targets from internal data (HTTP, RPC, DB) based on fault type needs.
package endpoint

// Endpoint represents a unified endpoint for fault injection.
// This is the common interface that all handlers work with.
type Endpoint struct {
	// Source service making the call
	SourceService string
	
	// Target service receiving the call
	TargetService string
	
	// Target port
	TargetPort string
	
	// Type of operation: "http", "rpc", "db"
	OperationType string
	
	// SpanName for groundtruth generation
	SpanName string
	
	// Type-specific details (optional, used by specific handlers)
	Details EndpointDetails
}

// EndpointDetails contains type-specific information
type EndpointDetails struct {
	// HTTP-specific
	HTTPMethod string
	HTTPRoute  string
	HTTPStatus string
	
	// RPC-specific
	RPCSystem  string
	RPCService string
	RPCMethod  string
	RPCStatus  string
	
	// Database-specific
	DBName      string
	DBTable     string
	DBOperation string // SELECT, INSERT, UPDATE, DELETE
	DBSystem    string // mysql, postgresql, redis, etc.
}

// CallPair represents a service-to-service call relationship.
// Used for network chaos which needs all types of calls.
type CallPair struct {
	SourceService string
	TargetService string
	SpanNames     []string // All span names for this pair
	OperationTypes []string // Types of operations: ["http", "rpc", "db"]
}

// HTTPEndpointInfo represents HTTP endpoint information for HTTP chaos
type HTTPEndpointInfo struct {
	ServiceName   string
	Route         string
	Method        string
	ServerAddress string
	ServerPort    string
	SpanName      string
}

// DNSEndpointInfo represents DNS endpoint information for DNS chaos
// DNS chaos works for HTTP and DB but NOT for RPC
type DNSEndpointInfo struct {
	ServiceName string
	Domain      string // Target service/domain
	SpanNames   []string
	HasHTTP     bool // Has HTTP calls to this domain
	HasDB       bool // Has DB calls to this domain
}

// DatabaseInfo represents database operation information for MySQL chaos
type DatabaseInfo struct {
	ServiceName string
	DBName      string
	TableName   string
	Operation   string // SELECT, INSERT, UPDATE, DELETE
	SpanName    string
}
