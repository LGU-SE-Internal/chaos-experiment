package utils

import "testing"

func TestIsGRPCRoute(t *testing.T) {
	tests := []struct {
		name     string
		route    string
		expected bool
	}{
		// gRPC routes
		{
			name:     "gRPC route with simple package",
			route:    "/oteldemo.CartService/AddItem",
			expected: true,
		},
		{
			name:     "gRPC route with versioned package",
			route:    "/flagd.evaluation.v1.Service/EventStream",
			expected: true,
		},
		{
			name:     "gRPC route with multiple package parts",
			route:    "/oteldemo.CurrencyService/Convert",
			expected: true,
		},
		{
			name:     "gRPC route with ResolveBoolean",
			route:    "/flagd.evaluation.v1.Service/ResolveBoolean",
			expected: true,
		},
		// HTTP routes (not gRPC)
		{
			name:     "empty route",
			route:    "",
			expected: false,
		},
		{
			name:     "simple HTTP route",
			route:    "/api/v1/users",
			expected: false,
		},
		{
			name:     "HTTP route with path params",
			route:    "/api/v1/orders/*/status",
			expected: false,
		},
		{
			name:     "HTTP route get-quote",
			route:    "/get-quote",
			expected: false,
		},
		{
			name:     "HTTP route send_order_confirmation",
			route:    "/send_order_confirmation",
			expected: false,
		},
		{
			name:     "HTTP route ship-order",
			route:    "/ship-order",
			expected: false,
		},
		{
			name:     "HTTP route with numeric version",
			route:    "/api/v1/priceservice/prices/*/DongCheOne",
			expected: false,
		},
		{
			name:     "Single slash",
			route:    "/",
			expected: false,
		},
		{
			name:     "Route with just dots",
			route:    "/some.path.here",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGRPCRoute(tt.route)
			if result != tt.expected {
				t.Errorf("IsGRPCRoute(%q) = %v, want %v", tt.route, result, tt.expected)
			}
		})
	}
}
