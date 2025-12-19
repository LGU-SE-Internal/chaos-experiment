// Package endpoint provides conversion functions from internal data to handler endpoints.
package endpoint

import (
	"github.com/LGU-SE-Internal/chaos-experiment/internal/resourcetypes"
)

// FromHTTP converts an HTTPEndpoint to a unified Endpoint
func FromHTTP(http resourcetypes.HTTPEndpoint) Endpoint {
	return Endpoint{
		SourceService: http.ServiceName,
		TargetService: http.ServerAddress,
		TargetPort:    http.ServerPort,
		OperationType: "http",
		SpanName:      http.SpanName,
		Details: EndpointDetails{
			HTTPMethod: http.RequestMethod,
			HTTPRoute:  http.Route,
			HTTPStatus: http.ResponseStatus,
		},
	}
}

// FromRPC converts an RPCOperation to a unified Endpoint
func FromRPC(rpc resourcetypes.RPCOperation) Endpoint {
	return Endpoint{
		SourceService: rpc.ServiceName,
		TargetService: rpc.ServerAddress,
		TargetPort:    rpc.ServerPort,
		OperationType: "rpc",
		SpanName:      rpc.SpanName,
		Details: EndpointDetails{
			RPCSystem:  rpc.RPCSystem,
			RPCService: rpc.RPCService,
			RPCMethod:  rpc.RPCMethod,
			RPCStatus:  rpc.StatusCode,
		},
	}
}

// FromDatabase converts a DatabaseOperation to a unified Endpoint
func FromDatabase(db resourcetypes.DatabaseOperation) Endpoint {
	return Endpoint{
		SourceService: db.ServiceName,
		TargetService: db.ServerAddress,
		TargetPort:    db.ServerPort,
		OperationType: "db",
		SpanName:      db.SpanName,
		Details: EndpointDetails{
			DBName:      db.DBName,
			DBTable:     db.DBTable,
			DBOperation: db.Operation,
			DBSystem:    db.DBSystem,
		},
	}
}

// ToHTTPEndpointInfo converts an HTTPEndpoint to HTTPEndpointInfo for HTTP chaos
func ToHTTPEndpointInfo(http resourcetypes.HTTPEndpoint) HTTPEndpointInfo {
	return HTTPEndpointInfo{
		ServiceName:   http.ServiceName,
		Route:         http.Route,
		Method:        http.RequestMethod,
		ServerAddress: http.ServerAddress,
		ServerPort:    http.ServerPort,
		SpanName:      http.SpanName,
	}
}

// ToDatabaseInfo converts a DatabaseOperation to DatabaseInfo for MySQL chaos
func ToDatabaseInfo(db resourcetypes.DatabaseOperation) DatabaseInfo {
	return DatabaseInfo{
		ServiceName: db.ServiceName,
		DBName:      db.DBName,
		TableName:   db.DBTable,
		Operation:   db.Operation,
		SpanName:    db.SpanName,
	}
}
