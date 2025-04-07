package handler

import (
	"fmt"

	"github.com/CUHK-SE-Group/chaos-experiment/internal/resourcelookup"
)

// NetworkPair represents a source and target service for network chaos
type NetworkPair struct {
	SourceService  string
	TargetService  string
	ConnectionType string
}

// ValidateNetworkPairIndex checks if a network pair index is valid
func ValidateNetworkPairIndex(networkPairIdx int) error {
	networkPairs, err := resourcelookup.GetAllNetworkPairs()
	if err != nil {
		return fmt.Errorf("failed to get network pairs: %w", err)
	}

	if networkPairIdx < 0 || networkPairIdx >= len(networkPairs) {
		return fmt.Errorf("network pair index out of range: %d (max: %d)",
			networkPairIdx, len(networkPairs)-1)
	}

	return nil
}

// GetNetworkPairByIndex returns a network pair by its index
func GetNetworkPairByIndex(networkPairIdx int) (*NetworkPair, error) {
	networkPairs, err := resourcelookup.GetAllNetworkPairs()
	if err != nil {
		return nil, fmt.Errorf("failed to get network pairs: %w", err)
	}

	if networkPairIdx < 0 || networkPairIdx >= len(networkPairs) {
		return nil, fmt.Errorf("network pair index out of range: %d (max: %d)",
			networkPairIdx, len(networkPairs)-1)
	}

	pair := networkPairs[networkPairIdx]
	return &NetworkPair{
		SourceService:  pair.SourceService,
		TargetService:  pair.TargetService,
		ConnectionType: "HTTP/gRPC",
	}, nil
}

// GetAllNetworkPairs returns all network pairs
func GetAllNetworkPairs() ([]NetworkPair, error) {
	networkPairs, err := resourcelookup.GetAllNetworkPairs()
	if err != nil {
		return nil, fmt.Errorf("failed to get network pairs: %w", err)
	}

	result := make([]NetworkPair, 0, len(networkPairs))
	for _, pair := range networkPairs {
		result = append(result, NetworkPair{
			SourceService:  pair.SourceService,
			TargetService:  pair.TargetService,
			ConnectionType: "HTTP/gRPC",
		})
	}

	return result, nil
}

// CountNetworkPairs returns the count of all network pairs
func CountNetworkPairs() (int, error) {
	networkPairs, err := resourcelookup.GetAllNetworkPairs()
	if err != nil {
		return 0, fmt.Errorf("failed to get network pairs: %w", err)
	}
	return len(networkPairs), nil
}

// GetNetworkPairLabel returns a human-readable label for a network pair
func GetNetworkPairLabel(networkPairIdx int) (string, error) {
	pair, err := GetNetworkPairByIndex(networkPairIdx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s → %s", pair.SourceService, pair.TargetService), nil
}

// ValidateChaosDirection validates that a direction code is valid
func ValidateChaosDirection(direction int) error {
	if _, ok := directionMap[direction]; !ok {
		return fmt.Errorf("invalid direction: %d (valid values: 1=to, 2=from, 3=both)", direction)
	}
	return nil
}
