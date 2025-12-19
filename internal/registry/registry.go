// Package registry provides a centralized registration mechanism for system data.
// This eliminates the need for switch-case statements and direct imports of
// system-specific packages, improving extensibility.
package registry

import (
	"fmt"
	"sync"

	"github.com/LGU-SE-Internal/chaos-experiment/internal/model"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"
)

var (
	// Global registry mapping SystemType to SystemData
	systemRegistry = make(map[systemconfig.SystemType]*model.SystemData)
	registryMutex  sync.RWMutex
)

// Register registers system data for a given system type.
// This should be called from init() functions in generated data packages.
func Register(systemType systemconfig.SystemType, data *model.SystemData) error {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	if _, exists := systemRegistry[systemType]; exists {
		return fmt.Errorf("system %s is already registered", systemType)
	}

	systemRegistry[systemType] = data
	return nil
}

// Get retrieves the system data for the given system type.
// Returns nil if the system is not registered.
func Get(systemType systemconfig.SystemType) *model.SystemData {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	return systemRegistry[systemType]
}

// GetCurrent retrieves the system data for the current system (from systemconfig).
// Returns nil if the current system is not registered.
func GetCurrent() *model.SystemData {
	return Get(systemconfig.GetCurrentSystem())
}

// MustGet retrieves the system data for the given system type.
// Panics if the system is not registered.
func MustGet(systemType systemconfig.SystemType) *model.SystemData {
	data := Get(systemType)
	if data == nil {
		panic(fmt.Sprintf("system %s is not registered", systemType))
	}
	return data
}

// MustGetCurrent retrieves the system data for the current system.
// Panics if the current system is not registered.
func MustGetCurrent() *model.SystemData {
	return MustGet(systemconfig.GetCurrentSystem())
}

// IsRegistered checks if a system type is registered.
func IsRegistered(systemType systemconfig.SystemType) bool {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	_, exists := systemRegistry[systemType]
	return exists
}

// ListRegistered returns all registered system types.
func ListRegistered() []systemconfig.SystemType {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	systems := make([]systemconfig.SystemType, 0, len(systemRegistry))
	for systemType := range systemRegistry {
		systems = append(systems, systemType)
	}
	return systems
}

// Clear clears all registered systems. Useful for testing.
func Clear() {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	systemRegistry = make(map[systemconfig.SystemType]*model.SystemData)
}
