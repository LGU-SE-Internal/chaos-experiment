// Package systemconfig provides a global configuration for the target system type.
// This package allows different systems (TrainTicket, OtelDemo, etc.) to coexist
// with their own metadata and configurations.
package systemconfig

import (
	"fmt"
	"sync"
)

// SystemType represents the type of system being analyzed/targeted
type SystemType string

const (
	// SystemTrainTicket represents the TrainTicket microservice system
	SystemTrainTicket SystemType = "ts"
	// SystemOtelDemo represents the OpenTelemetry Demo system
	SystemOtelDemo SystemType = "otel-demo"
	// SystemMediaMicroservices represents the Media Microservices system
	SystemMediaMicroservices SystemType = "media"
	// SystemHotelReservation represents the Hotel Reservation system
	SystemHotelReservation SystemType = "hs"
	// SystemSocialNetwork represents the Social Network system
	SystemSocialNetwork SystemType = "sn"
	// SystemOnlineBoutique represents the Online Boutique system
	SystemOnlineBoutique SystemType = "ob"
	// SystemSockShop represents the Sock Shop system
	SystemSockShop SystemType = "sockshop"
	// SystemTeaStore represents the Tea Store system
	SystemTeaStore SystemType = "teastore"
)

var (
	// currentSystem holds the current system type
	currentSystem SystemType = SystemTrainTicket

	// mu protects access to currentSystem
	mu sync.RWMutex

	// validSystems contains all valid system types
	validSystems = map[SystemType]bool{
		SystemTrainTicket:        true,
		SystemOtelDemo:           true,
		SystemMediaMicroservices: true,
		SystemHotelReservation:   true,
		SystemSocialNetwork:      true,
		SystemOnlineBoutique:     true,
		SystemSockShop:           true,
		SystemTeaStore:           true,
	}
)

// SetCurrentSystem sets the global system type for the current process.
// This should be called at initialization time before any metadata is accessed.
func SetCurrentSystem(system SystemType) error {
	mu.Lock()
	defer mu.Unlock()

	if !validSystems[system] {
		return fmt.Errorf("invalid system type: %s, valid types are: ts, otel-demo, media, hs, sn, ob, sockshop, teastore", system)
	}

	currentSystem = system
	return nil
}

// GetCurrentSystem returns the current system type.
func GetCurrentSystem() SystemType {
	mu.RLock()
	defer mu.RUnlock()
	return currentSystem
}

// IsTrainTicket returns true if the current system is TrainTicket.
func IsTrainTicket() bool {
	return GetCurrentSystem() == SystemTrainTicket
}

// IsOtelDemo returns true if the current system is OpenTelemetry Demo.
func IsOtelDemo() bool {
	return GetCurrentSystem() == SystemOtelDemo
}

// IsMediaMicroservices returns true if the current system is Media Microservices.
func IsMediaMicroservices() bool {
	return GetCurrentSystem() == SystemMediaMicroservices
}

// IsHotelReservation returns true if the current system is Hotel Reservation.
func IsHotelReservation() bool {
	return GetCurrentSystem() == SystemHotelReservation
}

// IsSocialNetwork returns true if the current system is Social Network.
func IsSocialNetwork() bool {
	return GetCurrentSystem() == SystemSocialNetwork
}

// IsOnlineBoutique returns true if the current system is Online Boutique.
func IsOnlineBoutique() bool {
	return GetCurrentSystem() == SystemOnlineBoutique
}

// IsSockShop returns true if the current system is Sock Shop.
func IsSockShop() bool {
	return GetCurrentSystem() == SystemSockShop
}

// IsTeaStore returns true if the current system is Tea Store.
func IsTeaStore() bool {
	return GetCurrentSystem() == SystemTeaStore
}

// String returns the string representation of the SystemType.
func (s SystemType) String() string {
	return string(s)
}

// GetAllSystemTypes returns all valid system types.
func GetAllSystemTypes() []SystemType {
	return []SystemType{SystemTrainTicket, SystemOtelDemo, SystemMediaMicroservices, SystemHotelReservation, SystemSocialNetwork, SystemOnlineBoutique, SystemSockShop, SystemTeaStore}
}

// ParseSystemType parses a string into a SystemType.
func ParseSystemType(s string) (SystemType, error) {
	st := SystemType(s)
	if !validSystems[st] {
		return "", fmt.Errorf("invalid system type: %s, valid types are: ts, otel-demo, media, hs, sn, ob, sockshop, teastore", s)
	}
	return st, nil
}
