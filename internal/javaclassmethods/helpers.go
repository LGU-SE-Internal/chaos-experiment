package javaclassmethods

import (
	"math/rand"
	"strings"
)

// Function variables that can be replaced during testing
var (
	// GetClassMethodsByServiceFunc is the implementation for GetClassMethodsByService
	GetClassMethodsByServiceFunc = GetClassMethodsByService

	// GetAllServicesFunc is the implementation for GetAllServices
	GetAllServicesFunc = GetAllServices
)

// GetMethodByIndex returns a specific method by index for a service
// If index is out of bounds, it returns either the first method or nil if no methods exist
func GetMethodByIndex(serviceName string, index int) *ClassMethodEntry {
	methods := GetClassMethodsByServiceFunc(serviceName)
	if len(methods) == 0 {
		return nil
	}

	if index >= 0 && index < len(methods) {
		return &methods[index]
	}

	return &methods[0]
}

// GetRandomMethod returns a random method for a service
// If no methods exist, it returns nil
func GetRandomMethod(serviceName string) *ClassMethodEntry {
	methods := GetClassMethodsByServiceFunc(serviceName)
	if len(methods) == 0 {
		return nil
	}

	randomIndex := rand.Intn(len(methods))
	return &methods[randomIndex]
}

// GetMethodByIndexOrRandom returns a method by index, or a random one if index is out of bounds
// If no methods exist, it returns nil
func GetMethodByIndexOrRandom(serviceName string, index int) *ClassMethodEntry {
	methods := GetClassMethodsByServiceFunc(serviceName)
	if len(methods) == 0 {
		return nil
	}

	if index >= 0 && index < len(methods) {
		return &methods[index]
	}

	return GetRandomMethod(serviceName)
}

// CountMethods returns the number of methods available for a service
func CountMethods(serviceName string) int {
	return len(GetClassMethodsByServiceFunc(serviceName))
}

// GetMethodDisplayName returns a short display name for a method (ClassName.methodName)
func GetMethodDisplayName(entry ClassMethodEntry) string {
	// Extract simple class name (without package)
	className := entry.ClassName
	if lastDot := strings.LastIndex(className, "."); lastDot >= 0 {
		className = className[lastDot+1:]
	}

	return className + "." + entry.MethodName
}

// ListAllServiceNames returns a list of all available service names
func ListAllServiceNames() []string {
	return GetAllServicesFunc()
}

// ListAvailableMethods returns a list of display names for all methods in a service
func ListAvailableMethods(serviceName string) []string {
	methods := GetClassMethodsByServiceFunc(serviceName)
	result := make([]string, len(methods))

	for i, method := range methods {
		result[i] = GetMethodDisplayName(method)
	}

	return result
}
