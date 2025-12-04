// Package javaclassmethods provides a system-aware routing layer for Java class method data.
// This package delegates to the appropriate system-specific package based on the current system configuration.
package javaclassmethods

import (
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"

	oteldemojvm "github.com/LGU-SE-Internal/chaos-experiment/internal/oteldemo/javaclassmethods"
	tsjvm "github.com/LGU-SE-Internal/chaos-experiment/internal/ts/javaclassmethods"
)

// ClassMethodEntry represents a class-method pair from Java analysis
type ClassMethodEntry struct {
	ClassName  string
	MethodName string
}

// GetClassMethodsByService returns all class-method pairs for a service based on current system
func GetClassMethodsByService(serviceName string) []ClassMethodEntry {
	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemTrainTicket:
		tsMethods := tsjvm.GetClassMethodsByService(serviceName)
		return convertTSMethods(tsMethods)
	case systemconfig.SystemOtelDemo:
		otelMethods := oteldemojvm.GetClassMethodsByService(serviceName)
		return convertOtelDemoMethods(otelMethods)
	default:
		// Default to TrainTicket
		tsMethods := tsjvm.GetClassMethodsByService(serviceName)
		return convertTSMethods(tsMethods)
	}
}

// GetAllServices returns a list of all available service names based on current system
func GetAllServices() []string {
	system := systemconfig.GetCurrentSystem()
	switch system {
	case systemconfig.SystemTrainTicket:
		return tsjvm.GetAllServices()
	case systemconfig.SystemOtelDemo:
		return oteldemojvm.GetAllServices()
	default:
		// Default to TrainTicket
		return tsjvm.GetAllServices()
	}
}

// convertTSMethods converts ts-specific methods to the common type
func convertTSMethods(tsMethods []tsjvm.ClassMethodEntry) []ClassMethodEntry {
	result := make([]ClassMethodEntry, len(tsMethods))
	for i, m := range tsMethods {
		result[i] = ClassMethodEntry{
			ClassName:  m.ClassName,
			MethodName: m.MethodName,
		}
	}
	return result
}

// convertOtelDemoMethods converts otel-demo-specific methods to the common type
func convertOtelDemoMethods(otelMethods []oteldemojvm.ClassMethodEntry) []ClassMethodEntry {
	result := make([]ClassMethodEntry, len(otelMethods))
	for i, m := range otelMethods {
		result[i] = ClassMethodEntry{
			ClassName:  m.ClassName,
			MethodName: m.MethodName,
		}
	}
	return result
}
