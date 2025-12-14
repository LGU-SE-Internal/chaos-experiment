// Package adapter provides compatibility between old generated code and new registry pattern.
// This allows migration of resourcelookup and faultpoints without regenerating all data files.
package adapter

import (
	"github.com/LGU-SE-Internal/chaos-experiment/internal/model"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/registry"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/resourcetypes"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"

	// Import old generated packages
	hsdb "github.com/LGU-SE-Internal/chaos-experiment/internal/hs/databaseoperations"
	hsendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/hs/serviceendpoints"
	hsgrpc "github.com/LGU-SE-Internal/chaos-experiment/internal/hs/grpcoperations"
	mediadb "github.com/LGU-SE-Internal/chaos-experiment/internal/media/databaseoperations"
	mediaendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/media/serviceendpoints"
	mediagrpc "github.com/LGU-SE-Internal/chaos-experiment/internal/media/grpcoperations"
	obdb "github.com/LGU-SE-Internal/chaos-experiment/internal/ob/databaseoperations"
	obendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/ob/serviceendpoints"
	obgrpc "github.com/LGU-SE-Internal/chaos-experiment/internal/ob/grpcoperations"
	oteldemodb "github.com/LGU-SE-Internal/chaos-experiment/internal/oteldemo/databaseoperations"
	oteldemoendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/oteldemo/serviceendpoints"
	oteldemogrpc "github.com/LGU-SE-Internal/chaos-experiment/internal/oteldemo/grpcoperations"
	sndb "github.com/LGU-SE-Internal/chaos-experiment/internal/sn/databaseoperations"
	snendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/sn/serviceendpoints"
	sngrpc "github.com/LGU-SE-Internal/chaos-experiment/internal/sn/grpcoperations"
	tsdb "github.com/LGU-SE-Internal/chaos-experiment/internal/ts/databaseoperations"
	tsendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/ts/serviceendpoints"
)

func init() {
	// Auto-register all systems on package import
	registerTrainTicket()
	registerOtelDemo()
	registerMediaMicroservices()
	registerHotelReservation()
	registerSocialNetwork()
	registerOnlineBoutique()
}

func registerTrainTicket() {
	httpEps := convertTSEndpoints(tsendpoints.ServiceEndpoints)
	dbOps := convertTSDBOperations(tsdb.DatabaseOperations)
	
	registry.Register(systemconfig.SystemTrainTicket, &model.SystemData{
		SystemName:         "ts",
		HTTPEndpoints:      httpEps,
		DatabaseOperations: dbOps,
		RPCOperations:      make(map[string][]resourcetypes.RPCOperation), // TrainTicket has no RPC
		AllServices:        tsendpoints.GetAllServices(),
	})
}

func registerOtelDemo() {
	httpEps := convertOtelDemoEndpoints(oteldemoendpoints.ServiceEndpoints)
	dbOps := convertOtelDemoDBOperations(oteldemodb.DatabaseOperations)
	rpcOps := convertOtelDemoRPCOperations(oteldemogrpc.GRPCOperations)
	
	registry.Register(systemconfig.SystemOtelDemo, &model.SystemData{
		SystemName:         "otel-demo",
		HTTPEndpoints:      httpEps,
		DatabaseOperations: dbOps,
		RPCOperations:      rpcOps,
		AllServices:        oteldemoendpoints.GetAllServices(),
	})
}

func registerMediaMicroservices() {
	httpEps := convertMediaEndpoints(mediaendpoints.ServiceEndpoints)
	dbOps := convertMediaDBOperations(mediadb.DatabaseOperations)
	rpcOps := convertMediaRPCOperations(mediagrpc.GRPCOperations)
	
	registry.Register(systemconfig.SystemMediaMicroservices, &model.SystemData{
		SystemName:         "media",
		HTTPEndpoints:      httpEps,
		DatabaseOperations: dbOps,
		RPCOperations:      rpcOps,
		AllServices:        mediaendpoints.GetAllServices(),
	})
}

func registerHotelReservation() {
	httpEps := convertHSEndpoints(hsendpoints.ServiceEndpoints)
	dbOps := convertHSDBOperations(hsdb.DatabaseOperations)
	rpcOps := convertHSRPCOperations(hsgrpc.GRPCOperations)
	
	registry.Register(systemconfig.SystemHotelReservation, &model.SystemData{
		SystemName:         "hs",
		HTTPEndpoints:      httpEps,
		DatabaseOperations: dbOps,
		RPCOperations:      rpcOps,
		AllServices:        hsendpoints.GetAllServices(),
	})
}

func registerSocialNetwork() {
	httpEps := convertSNEndpoints(snendpoints.ServiceEndpoints)
	dbOps := convertSNDBOperations(sndb.DatabaseOperations)
	rpcOps := convertSNRPCOperations(sngrpc.GRPCOperations)
	
	registry.Register(systemconfig.SystemSocialNetwork, &model.SystemData{
		SystemName:         "sn",
		HTTPEndpoints:      httpEps,
		DatabaseOperations: dbOps,
		RPCOperations:      rpcOps,
		AllServices:        snendpoints.GetAllServices(),
	})
}

func registerOnlineBoutique() {
	httpEps := convertOBEndpoints(obendpoints.ServiceEndpoints)
	dbOps := convertOBDBOperations(obdb.DatabaseOperations)
	rpcOps := convertOBRPCOperations(obgrpc.GRPCOperations)
	
	registry.Register(systemconfig.SystemOnlineBoutique, &model.SystemData{
		SystemName:         "ob",
		HTTPEndpoints:      httpEps,
		DatabaseOperations: dbOps,
		RPCOperations:      rpcOps,
		AllServices:        obendpoints.GetAllServices(),
	})
}

// Conversion functions from old generated types to new resourcetypes

func convertTSEndpoints(old map[string][]tsendpoints.ServiceEndpoint) map[string][]resourcetypes.HTTPEndpoint {
	result := make(map[string][]resourcetypes.HTTPEndpoint)
	for service, endpoints := range old {
		converted := make([]resourcetypes.HTTPEndpoint, len(endpoints))
		for i, ep := range endpoints {
			converted[i] = resourcetypes.HTTPEndpoint{
				ServiceName:    ep.ServiceName,
				RequestMethod:  ep.RequestMethod,
				ResponseStatus: ep.ResponseStatus,
				Route:          ep.Route,
				ServerAddress:  ep.ServerAddress,
				ServerPort:     ep.ServerPort,
				SpanName:       ep.SpanName,
				SpanKind:       "", // Old data doesn't have SpanKind
			}
		}
		result[service] = converted
	}
	return result
}

func convertTSDBOperations(old map[string][]tsdb.DatabaseOperation) map[string][]resourcetypes.DatabaseOperation {
	result := make(map[string][]resourcetypes.DatabaseOperation)
	for service, ops := range old {
		converted := make([]resourcetypes.DatabaseOperation, len(ops))
		for i, op := range ops {
			converted[i] = resourcetypes.DatabaseOperation{
				ServiceName:   op.ServiceName,
				DBName:        op.DBName,
				DBTable:       op.DBTable,
				Operation:     op.Operation,
				DBSystem:      op.DBSystem,
				ServerAddress: op.ServerAddress,
				ServerPort:    op.ServerPort,
				SpanName:      "", // Old data doesn't have SpanName
			}
		}
		result[service] = converted
	}
	return result
}

func convertOtelDemoEndpoints(old map[string][]oteldemoendpoints.ServiceEndpoint) map[string][]resourcetypes.HTTPEndpoint {
	result := make(map[string][]resourcetypes.HTTPEndpoint)
	for service, endpoints := range old {
		converted := make([]resourcetypes.HTTPEndpoint, len(endpoints))
		for i, ep := range endpoints {
			converted[i] = resourcetypes.HTTPEndpoint{
				ServiceName:    ep.ServiceName,
				RequestMethod:  ep.RequestMethod,
				ResponseStatus: ep.ResponseStatus,
				Route:          ep.Route,
				ServerAddress:  ep.ServerAddress,
				ServerPort:     ep.ServerPort,
				SpanName:       ep.SpanName,
				SpanKind:       "",
			}
		}
		result[service] = converted
	}
	return result
}

func convertOtelDemoDBOperations(old map[string][]oteldemodb.DatabaseOperation) map[string][]resourcetypes.DatabaseOperation {
	result := make(map[string][]resourcetypes.DatabaseOperation)
	for service, ops := range old {
		converted := make([]resourcetypes.DatabaseOperation, len(ops))
		for i, op := range ops {
			converted[i] = resourcetypes.DatabaseOperation{
				ServiceName:   op.ServiceName,
				DBName:        op.DBName,
				DBTable:       op.DBTable,
				Operation:     op.Operation,
				DBSystem:      op.DBSystem,
				ServerAddress: op.ServerAddress,
				ServerPort:    op.ServerPort,
				SpanName:      "",
			}
		}
		result[service] = converted
	}
	return result
}

func convertOtelDemoRPCOperations(old map[string][]oteldemogrpc.GRPCOperation) map[string][]resourcetypes.RPCOperation {
	result := make(map[string][]resourcetypes.RPCOperation)
	for service, ops := range old {
		converted := make([]resourcetypes.RPCOperation, len(ops))
		for i, op := range ops {
			converted[i] = resourcetypes.RPCOperation{
				ServiceName:   op.ServiceName,
				RPCSystem:     op.RPCSystem,
				RPCService:    op.RPCService,
				RPCMethod:     op.RPCMethod,
				StatusCode:    op.GRPCStatusCode,
				ServerAddress: op.ServerAddress,
				ServerPort:    op.ServerPort,
				SpanKind:      op.SpanKind,
				SpanName:      "",
			}
		}
		result[service] = converted
	}
	return result
}

// Similar conversion functions for other systems
func convertMediaEndpoints(old map[string][]mediaendpoints.ServiceEndpoint) map[string][]resourcetypes.HTTPEndpoint {
	result := make(map[string][]resourcetypes.HTTPEndpoint)
	for service, endpoints := range old {
		converted := make([]resourcetypes.HTTPEndpoint, len(endpoints))
		for i, ep := range endpoints {
			converted[i] = resourcetypes.HTTPEndpoint{
				ServiceName:    ep.ServiceName,
				RequestMethod:  ep.RequestMethod,
				ResponseStatus: ep.ResponseStatus,
				Route:          ep.Route,
				ServerAddress:  ep.ServerAddress,
				ServerPort:     ep.ServerPort,
				SpanName:       ep.SpanName,
				SpanKind:       "",
			}
		}
		result[service] = converted
	}
	return result
}

func convertMediaDBOperations(old map[string][]mediadb.DatabaseOperation) map[string][]resourcetypes.DatabaseOperation {
	result := make(map[string][]resourcetypes.DatabaseOperation)
	for service, ops := range old {
		converted := make([]resourcetypes.DatabaseOperation, len(ops))
		for i, op := range ops {
			converted[i] = resourcetypes.DatabaseOperation{
				ServiceName:   op.ServiceName,
				DBName:        op.DBName,
				DBTable:       op.DBTable,
				Operation:     op.Operation,
				DBSystem:      op.DBSystem,
				ServerAddress: op.ServerAddress,
				ServerPort:    op.ServerPort,
				SpanName:      "",
			}
		}
		result[service] = converted
	}
	return result
}

func convertHSEndpoints(old map[string][]hsendpoints.ServiceEndpoint) map[string][]resourcetypes.HTTPEndpoint {
	result := make(map[string][]resourcetypes.HTTPEndpoint)
	for service, endpoints := range old {
		converted := make([]resourcetypes.HTTPEndpoint, len(endpoints))
		for i, ep := range endpoints {
			converted[i] = resourcetypes.HTTPEndpoint{
				ServiceName:    ep.ServiceName,
				RequestMethod:  ep.RequestMethod,
				ResponseStatus: ep.ResponseStatus,
				Route:          ep.Route,
				ServerAddress:  ep.ServerAddress,
				ServerPort:     ep.ServerPort,
				SpanName:       ep.SpanName,
				SpanKind:       "",
			}
		}
		result[service] = converted
	}
	return result
}

func convertHSDBOperations(old map[string][]hsdb.DatabaseOperation) map[string][]resourcetypes.DatabaseOperation {
	result := make(map[string][]resourcetypes.DatabaseOperation)
	for service, ops := range old {
		converted := make([]resourcetypes.DatabaseOperation, len(ops))
		for i, op := range ops {
			converted[i] = resourcetypes.DatabaseOperation{
				ServiceName:   op.ServiceName,
				DBName:        op.DBName,
				DBTable:       op.DBTable,
				Operation:     op.Operation,
				DBSystem:      op.DBSystem,
				ServerAddress: op.ServerAddress,
				ServerPort:    op.ServerPort,
				SpanName:      "",
			}
		}
		result[service] = converted
	}
	return result
}

func convertSNEndpoints(old map[string][]snendpoints.ServiceEndpoint) map[string][]resourcetypes.HTTPEndpoint {
	result := make(map[string][]resourcetypes.HTTPEndpoint)
	for service, endpoints := range old {
		converted := make([]resourcetypes.HTTPEndpoint, len(endpoints))
		for i, ep := range endpoints {
			converted[i] = resourcetypes.HTTPEndpoint{
				ServiceName:    ep.ServiceName,
				RequestMethod:  ep.RequestMethod,
				ResponseStatus: ep.ResponseStatus,
				Route:          ep.Route,
				ServerAddress:  ep.ServerAddress,
				ServerPort:     ep.ServerPort,
				SpanName:       ep.SpanName,
				SpanKind:       "",
			}
		}
		result[service] = converted
	}
	return result
}

func convertSNDBOperations(old map[string][]sndb.DatabaseOperation) map[string][]resourcetypes.DatabaseOperation {
	result := make(map[string][]resourcetypes.DatabaseOperation)
	for service, ops := range old {
		converted := make([]resourcetypes.DatabaseOperation, len(ops))
		for i, op := range ops {
			converted[i] = resourcetypes.DatabaseOperation{
				ServiceName:   op.ServiceName,
				DBName:        op.DBName,
				DBTable:       op.DBTable,
				Operation:     op.Operation,
				DBSystem:      op.DBSystem,
				ServerAddress: op.ServerAddress,
				ServerPort:    op.ServerPort,
				SpanName:      "",
			}
		}
		result[service] = converted
	}
	return result
}

func convertOBEndpoints(old map[string][]obendpoints.ServiceEndpoint) map[string][]resourcetypes.HTTPEndpoint {
	result := make(map[string][]resourcetypes.HTTPEndpoint)
	for service, endpoints := range old {
		converted := make([]resourcetypes.HTTPEndpoint, len(endpoints))
		for i, ep := range endpoints {
			converted[i] = resourcetypes.HTTPEndpoint{
				ServiceName:    ep.ServiceName,
				RequestMethod:  ep.RequestMethod,
				ResponseStatus: ep.ResponseStatus,
				Route:          ep.Route,
				ServerAddress:  ep.ServerAddress,
				ServerPort:     ep.ServerPort,
				SpanName:       ep.SpanName,
				SpanKind:       "",
			}
		}
		result[service] = converted
	}
	return result
}

func convertOBDBOperations(old map[string][]obdb.DatabaseOperation) map[string][]resourcetypes.DatabaseOperation {
	result := make(map[string][]resourcetypes.DatabaseOperation)
	for service, ops := range old {
		converted := make([]resourcetypes.DatabaseOperation, len(ops))
		for i, op := range ops {
			converted[i] = resourcetypes.DatabaseOperation{
				ServiceName:   op.ServiceName,
				DBName:        op.DBName,
				DBTable:       op.DBTable,
				Operation:     op.Operation,
				DBSystem:      op.DBSystem,
				ServerAddress: op.ServerAddress,
				ServerPort:    op.ServerPort,
				SpanName:      "",
			}
		}
		result[service] = converted
	}
	return result
}

func convertMediaRPCOperations(old map[string][]mediagrpc.GRPCOperation) map[string][]resourcetypes.RPCOperation {
	result := make(map[string][]resourcetypes.RPCOperation)
	for service, ops := range old {
		converted := make([]resourcetypes.RPCOperation, len(ops))
		for i, op := range ops {
			converted[i] = resourcetypes.RPCOperation{
				ServiceName:   op.ServiceName,
				RPCSystem:     op.RPCSystem,
				RPCService:    op.RPCService,
				RPCMethod:     op.RPCMethod,
				StatusCode:    op.GRPCStatusCode,
				ServerAddress: op.ServerAddress,
				ServerPort:    op.ServerPort,
				SpanKind:      op.SpanKind,
				SpanName:      "",
			}
		}
		result[service] = converted
	}
	return result
}

func convertHSRPCOperations(old map[string][]hsgrpc.GRPCOperation) map[string][]resourcetypes.RPCOperation {
	result := make(map[string][]resourcetypes.RPCOperation)
	for service, ops := range old {
		converted := make([]resourcetypes.RPCOperation, len(ops))
		for i, op := range ops {
			converted[i] = resourcetypes.RPCOperation{
				ServiceName:   op.ServiceName,
				RPCSystem:     op.RPCSystem,
				RPCService:    op.RPCService,
				RPCMethod:     op.RPCMethod,
				StatusCode:    op.GRPCStatusCode,
				ServerAddress: op.ServerAddress,
				ServerPort:    op.ServerPort,
				SpanKind:      op.SpanKind,
				SpanName:      "",
			}
		}
		result[service] = converted
	}
	return result
}

func convertSNRPCOperations(old map[string][]sngrpc.GRPCOperation) map[string][]resourcetypes.RPCOperation {
	result := make(map[string][]resourcetypes.RPCOperation)
	for service, ops := range old {
		converted := make([]resourcetypes.RPCOperation, len(ops))
		for i, op := range ops {
			converted[i] = resourcetypes.RPCOperation{
				ServiceName:   op.ServiceName,
				RPCSystem:     op.RPCSystem,
				RPCService:    op.RPCService,
				RPCMethod:     op.RPCMethod,
				StatusCode:    op.GRPCStatusCode,
				ServerAddress: op.ServerAddress,
				ServerPort:    op.ServerPort,
				SpanKind:      op.SpanKind,
				SpanName:      "",
			}
		}
		result[service] = converted
	}
	return result
}

func convertOBRPCOperations(old map[string][]obgrpc.GRPCOperation) map[string][]resourcetypes.RPCOperation {
	result := make(map[string][]resourcetypes.RPCOperation)
	for service, ops := range old {
		converted := make([]resourcetypes.RPCOperation, len(ops))
		for i, op := range ops {
			converted[i] = resourcetypes.RPCOperation{
				ServiceName:   op.ServiceName,
				RPCSystem:     op.RPCSystem,
				RPCService:    op.RPCService,
				RPCMethod:     op.RPCMethod,
				StatusCode:    op.GRPCStatusCode,
				ServerAddress: op.ServerAddress,
				ServerPort:    op.ServerPort,
				SpanKind:      op.SpanKind,
				SpanName:      "",
			}
		}
		result[service] = converted
	}
	return result
}
