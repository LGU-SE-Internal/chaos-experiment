// Package databaseoperations provides a system-aware routing layer for database operation data.
// This package delegates to the appropriate system-specific package based on the current system configuration.
package databaseoperations

import (
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"

	hsdb "github.com/LGU-SE-Internal/chaos-experiment/internal/hs/databaseoperations"
	mediadb "github.com/LGU-SE-Internal/chaos-experiment/internal/media/databaseoperations"
	oteldemodb "github.com/LGU-SE-Internal/chaos-experiment/internal/oteldemo/databaseoperations"
	obdb "github.com/LGU-SE-Internal/chaos-experiment/internal/ob/databaseoperations"
	sndb "github.com/LGU-SE-Internal/chaos-experiment/internal/sn/databaseoperations"
	tsdb "github.com/LGU-SE-Internal/chaos-experiment/internal/ts/databaseoperations"
)

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

// GetOperationsByService returns all database operations for a service based on current system
func GetOperationsByService(serviceName string) []DatabaseOperation {
	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemTrainTicket:
		tsOps := tsdb.GetOperationsByService(serviceName)
		return convertTSOperations(tsOps)
	case systemconfig.SystemOtelDemo:
		otelOps := oteldemodb.GetOperationsByService(serviceName)
		return convertOtelDemoOperations(otelOps)
	case systemconfig.SystemMediaMicroservices:
		mediaOps := mediadb.GetOperationsByService(serviceName)
		return convertMediaOperations(mediaOps)
	case systemconfig.SystemHotelReservation:
		hsOps := hsdb.GetOperationsByService(serviceName)
		return convertHSOperations(hsOps)
	case systemconfig.SystemSocialNetwork:
		snOps := sndb.GetOperationsByService(serviceName)
		return convertSNOperations(snOps)
	case systemconfig.SystemOnlineBoutique:
		obOps := obdb.GetOperationsByService(serviceName)
		return convertOBOperations(obOps)
	default:
		// Default to TrainTicket
		tsOps := tsdb.GetOperationsByService(serviceName)
		return convertTSOperations(tsOps)
	}
}

// GetAllDatabaseServices returns a list of all services that perform database operations based on current system
func GetAllDatabaseServices() []string {
	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemTrainTicket:
		return tsdb.GetAllDatabaseServices()
	case systemconfig.SystemOtelDemo:
		return oteldemodb.GetAllDatabaseServices()
	case systemconfig.SystemMediaMicroservices:
		return mediadb.GetAllDatabaseServices()
	case systemconfig.SystemHotelReservation:
		return hsdb.GetAllDatabaseServices()
	case systemconfig.SystemSocialNetwork:
		return sndb.GetAllDatabaseServices()
	default:
		// Default to TrainTicket
		return tsdb.GetAllDatabaseServices()
	}
}

// GetOperationsByDatabase returns all operations for a specific database based on current system
func GetOperationsByDatabase(dbName string) []DatabaseOperation {
	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemTrainTicket:
		tsOps := tsdb.GetOperationsByDatabase(dbName)
		return convertTSOperations(tsOps)
	case systemconfig.SystemOtelDemo:
		otelOps := oteldemodb.GetOperationsByDatabase(dbName)
		return convertOtelDemoOperations(otelOps)
	case systemconfig.SystemMediaMicroservices:
		mediaOps := mediadb.GetOperationsByDatabase(dbName)
		return convertMediaOperations(mediaOps)
	case systemconfig.SystemHotelReservation:
		hsOps := hsdb.GetOperationsByDatabase(dbName)
		return convertHSOperations(hsOps)
	case systemconfig.SystemSocialNetwork:
		snOps := sndb.GetOperationsByDatabase(dbName)
		return convertSNOperations(snOps)
	default:
		// Default to TrainTicket
		tsOps := tsdb.GetOperationsByDatabase(dbName)
		return convertTSOperations(tsOps)
	}
}

// GetOperationsByTable returns all operations for a specific table based on current system
func GetOperationsByTable(dbTable string) []DatabaseOperation {
	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemTrainTicket:
		tsOps := tsdb.GetOperationsByTable(dbTable)
		return convertTSOperations(tsOps)
	case systemconfig.SystemOtelDemo:
		otelOps := oteldemodb.GetOperationsByTable(dbTable)
		return convertOtelDemoOperations(otelOps)
	case systemconfig.SystemMediaMicroservices:
		mediaOps := mediadb.GetOperationsByTable(dbTable)
		return convertMediaOperations(mediaOps)
	case systemconfig.SystemHotelReservation:
		hsOps := hsdb.GetOperationsByTable(dbTable)
		return convertHSOperations(hsOps)
	case systemconfig.SystemSocialNetwork:
		snOps := sndb.GetOperationsByTable(dbTable)
		return convertSNOperations(snOps)
	default:
		// Default to TrainTicket
		tsOps := tsdb.GetOperationsByTable(dbTable)
		return convertTSOperations(tsOps)
	}
}

// convertTSOperations converts ts-specific operations to the common type
func convertTSOperations(tsOps []tsdb.DatabaseOperation) []DatabaseOperation {
	result := make([]DatabaseOperation, len(tsOps))
	for i, op := range tsOps {
		result[i] = DatabaseOperation{
			ServiceName:   op.ServiceName,
			DBName:        op.DBName,
			DBTable:       op.DBTable,
			Operation:     op.Operation,
			DBSystem:      op.DBSystem,
			ServerAddress: op.ServerAddress,
			ServerPort:    op.ServerPort,
		}
	}
	return result
}

// convertOtelDemoOperations converts otel-demo-specific operations to the common type
func convertOtelDemoOperations(otelOps []oteldemodb.DatabaseOperation) []DatabaseOperation {
	result := make([]DatabaseOperation, len(otelOps))
	for i, op := range otelOps {
		result[i] = DatabaseOperation{
			ServiceName:   op.ServiceName,
			DBName:        op.DBName,
			DBTable:       op.DBTable,
			Operation:     op.Operation,
			DBSystem:      op.DBSystem,
			ServerAddress: op.ServerAddress,
			ServerPort:    op.ServerPort,
		}
	}
	return result
}

// convertMediaOperations converts media-specific operations to the common type
func convertMediaOperations(mediaOps []mediadb.DatabaseOperation) []DatabaseOperation {
	result := make([]DatabaseOperation, len(mediaOps))
	for i, op := range mediaOps {
		result[i] = DatabaseOperation{
			ServiceName:   op.ServiceName,
			DBName:        op.DBName,
			DBTable:       op.DBTable,
			Operation:     op.Operation,
			DBSystem:      op.DBSystem,
			ServerAddress: op.ServerAddress,
			ServerPort:    op.ServerPort,
		}
	}
	return result
}

// convertHSOperations converts hs-specific operations to the common type
func convertHSOperations(hsOps []hsdb.DatabaseOperation) []DatabaseOperation {
	result := make([]DatabaseOperation, len(hsOps))
	for i, op := range hsOps {
		result[i] = DatabaseOperation{
			ServiceName:   op.ServiceName,
			DBName:        op.DBName,
			DBTable:       op.DBTable,
			Operation:     op.Operation,
			DBSystem:      op.DBSystem,
			ServerAddress: op.ServerAddress,
			ServerPort:    op.ServerPort,
		}
	}
	return result
}

// convertSNOperations converts sn-specific operations to the common type
func convertSNOperations(snOps []sndb.DatabaseOperation) []DatabaseOperation {
	result := make([]DatabaseOperation, len(snOps))
	for i, op := range snOps {
		result[i] = DatabaseOperation{
			ServiceName:   op.ServiceName,
			DBName:        op.DBName,
			DBTable:       op.DBTable,
			Operation:     op.Operation,
			DBSystem:      op.DBSystem,
			ServerAddress: op.ServerAddress,
			ServerPort:    op.ServerPort,
		}
	}
	return result
}
// convertOBOperations converts ob-specific operations to the common type
func convertOBOperations(obOps []obdb.DatabaseOperation) []DatabaseOperation {
result := make([]DatabaseOperation, len(obOps))
for i, op := range obOps {
result[i] = DatabaseOperation{
ServiceName:   op.ServiceName,
DBName:        op.DBName,
DBTable:       op.DBTable,
Operation:     op.Operation,
DBSystem:      op.DBSystem,
ServerAddress: op.ServerAddress,
ServerPort:    op.ServerPort,
}
}
return result
}

