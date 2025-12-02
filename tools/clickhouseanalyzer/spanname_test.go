package clickhouseanalyzer

import (
	"testing"
)

func TestNormalizeTrainTicketSpanName(t *testing.T) {
	tests := []struct {
		name        string
		spanName    string
		serviceName string
		want        string
	}{
		{
			name:        "verify code pattern for ts-ui-dashboard",
			spanName:    "HTTP GET http://caddy/api/v1/verifycode/verify/abc123",
			serviceName: "ts-ui-dashboard",
			want:        "HTTP GET http://caddy/api/v1/verifycode/verify/{verifyCode}",
		},
		{
			name:        "verify code pattern for loadgenerator",
			spanName:    "HTTP GET http://caddy/api/v1/verifycode/verify/xyz789",
			serviceName: "loadgenerator",
			want:        "HTTP GET http://caddy/api/v1/verifycode/verify/{verifyCode}",
		},
		{
			name:        "non-matching service should not be modified",
			spanName:    "HTTP GET http://caddy/api/v1/verifycode/verify/abc123",
			serviceName: "ts-auth-service",
			want:        "HTTP GET http://caddy/api/v1/verifycode/verify/abc123",
		},
		{
			name:        "foodservice foods pattern",
			spanName:    "HTTP GET http://caddy/api/v1/foodservice/foods/2024-01-15/shanghai/beijing/G1234",
			serviceName: "ts-ui-dashboard",
			want:        "HTTP GET http://caddy/api/v1/foodservice/foods/{date}/{startStation}/{endStation}/{tripId}",
		},
		{
			name:        "contactservice contacts account pattern",
			spanName:    "HTTP GET http://caddy/api/v1/contactservice/contacts/account/550e8400-e29b-41d4-a716-446655440000",
			serviceName: "ts-ui-dashboard",
			want:        "HTTP GET http://caddy/api/v1/contactservice/contacts/account/{accountId}",
		},
		{
			name:        "userservice users id pattern",
			spanName:    "HTTP GET http://caddy/api/v1/userservice/users/id/550e8400-e29b-41d4-a716-446655440000",
			serviceName: "loadgenerator",
			want:        "HTTP GET http://caddy/api/v1/userservice/users/id/{userId}",
		},
		{
			name:        "consignservice consigns order pattern",
			spanName:    "HTTP GET http://caddy/api/v1/consignservice/consigns/order/550e8400-e29b-41d4-a716-446655440000",
			serviceName: "ts-ui-dashboard",
			want:        "HTTP GET http://caddy/api/v1/consignservice/consigns/order/{id}",
		},
		{
			name:        "consignservice consigns account pattern",
			spanName:    "HTTP GET http://caddy/api/v1/consignservice/consigns/account/550e8400-e29b-41d4-a716-446655440000",
			serviceName: "ts-ui-dashboard",
			want:        "HTTP GET http://caddy/api/v1/consignservice/consigns/account/{id}",
		},
		{
			name:        "executeservice execute collected pattern",
			spanName:    "HTTP GET http://caddy/api/v1/executeservice/execute/collected/550e8400-e29b-41d4-a716-446655440000",
			serviceName: "loadgenerator",
			want:        "HTTP GET http://caddy/api/v1/executeservice/execute/collected/{orderId}",
		},
		{
			name:        "cancelservice cancel with two IDs pattern",
			spanName:    "HTTP GET http://caddy/api/v1/cancelservice/cancel/550e8400-e29b-41d4-a716-446655440000/660e8400-e29b-41d4-a716-446655440001",
			serviceName: "ts-ui-dashboard",
			want:        "HTTP GET http://caddy/api/v1/cancelservice/cancel/{orderId}/{loginId}",
		},
		{
			name:        "cancelservice cancel refound pattern",
			spanName:    "HTTP GET http://caddy/api/v1/cancelservice/cancel/refound/550e8400-e29b-41d4-a716-446655440000",
			serviceName: "ts-ui-dashboard",
			want:        "HTTP GET http://caddy/api/v1/cancelservice/cancel/refound/{orderId}",
		},
		{
			name:        "executeservice execute execute pattern",
			spanName:    "HTTP GET http://caddy/api/v1/executeservice/execute/execute/550e8400-e29b-41d4-a716-446655440000",
			serviceName: "loadgenerator",
			want:        "HTTP GET http://caddy/api/v1/executeservice/execute/execute/{orderId}",
		},
		{
			name:        "non-matching pattern should not be modified",
			spanName:    "HTTP POST http://caddy/api/v1/travelservice/trips/left",
			serviceName: "ts-ui-dashboard",
			want:        "HTTP POST http://caddy/api/v1/travelservice/trips/left",
		},
		{
			name:        "empty span name",
			spanName:    "",
			serviceName: "ts-ui-dashboard",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeTrainTicketSpanName(tt.spanName, tt.serviceName)
			if got != tt.want {
				t.Errorf("NormalizeTrainTicketSpanName() = %q, want %q", got, tt.want)
			}
		})
	}
}
