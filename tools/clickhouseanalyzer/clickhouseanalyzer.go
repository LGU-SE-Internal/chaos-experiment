package clickhouseanalyzer

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

// Connection parameters
type ClickHouseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// ServiceEndpoint represents a service endpoint with its details
type ServiceEndpoint struct {
	ServiceName    string
	RequestMethod  string
	ResponseStatus string
	Route          string
	ServerAddress  string
	ServerPort     string
}

// DatabaseOperation represents a database operation with its details
type DatabaseOperation struct {
	ServiceName string
	DBName      string
	DBTable     string
	Operation   string
}

// Create materialized view SQL statement
const createMaterializedViewSQL = `
CREATE MATERIALIZED VIEW IF NOT EXISTS otel_traces_mv 
ENGINE = ReplacingMergeTree(version)
PARTITION BY toYYYYMM(Timestamp)
PRIMARY KEY (masked_route, ServiceName, db_sql_table)
ORDER BY (
    masked_route,
    ServiceName,
    db_sql_table,
    SpanKind,
    request_method,
    response_status_code,
    server_address,
    server_port,
	db_name,
    db_operation
)
SETTINGS allow_nullable_key = 1
POPULATE
AS 
WITH 
    replaceRegexpOne(SpanAttributes['url.full'], 'https?://[^/]+(/.*)', '\\1') AS path
SELECT 
    ResourceAttributes['service.name'] AS ServiceName,
    4294967295 - toUnixTimestamp(Timestamp) AS version,
    Timestamp,
    SpanKind,
    SpanAttributes['client.address'] AS client_address,
    SpanAttributes['http.request.method'] AS http_request_method,
    SpanAttributes['http.response.status_code'] AS http_response_status_code,
    SpanAttributes['http.route'] AS http_route,
    SpanAttributes['http.method'] AS http_method,
    SpanAttributes['url.full'] AS url_full,
    SpanAttributes['http.status_code'] AS http_status_code,
    SpanAttributes['http.target'] AS http_target,
    
    CASE 
        WHEN SpanAttributes['http.request.method'] IS NOT NULL AND SpanAttributes['http.request.method'] != '' 
            THEN SpanAttributes['http.request.method']
        WHEN SpanAttributes['http.method'] IS NOT NULL AND SpanAttributes['http.method'] != '' 
            THEN SpanAttributes['http.method']
        ELSE ''
    END AS request_method,
    
    CASE 
        WHEN SpanAttributes['http.response.status_code'] IS NOT NULL AND SpanAttributes['http.response.status_code'] != '' 
            THEN SpanAttributes['http.response.status_code']
        WHEN SpanAttributes['http.status_code'] IS NOT NULL AND SpanAttributes['http.status_code'] != '' 
            THEN SpanAttributes['http.status_code']
        ELSE ''
    END AS response_status_code,
    
    CASE
        WHEN SpanAttributes['http.route'] IS NOT NULL AND SpanAttributes['http.route'] != ''
            THEN replaceRegexpAll(SpanAttributes['http.route'], '/\\{[^}]+\\}', '/*')
            
        WHEN SpanAttributes['http.target'] IS NOT NULL AND SpanAttributes['http.target'] != ''
            THEN 
                CASE
                    WHEN position(SpanAttributes['http.target'], '/api/v1/verifycode/verify/') = 1
                        THEN '/api/v1/verifycode/verify/*'
                    WHEN position(SpanAttributes['http.target'], '/api/v1/cancelservice/cancel/refound/') = 1
                        THEN '/api/v1/cancelservice/cancel/refound/*'
                    WHEN position(SpanAttributes['http.target'], '/api/v1/cancelservice/cancel/') = 1
                        THEN '/api/v1/cancelservice/cancel/*/*'
                    WHEN position(SpanAttributes['http.target'], '/api/v1/consignservice/consigns/account/') = 1
                        THEN '/api/v1/consignservice/consigns/account/*'
                    WHEN position(SpanAttributes['http.target'], '/api/v1/consignservice/consigns/order/') = 1
                        THEN '/api/v1/consignservice/consigns/order/*'
                    WHEN position(SpanAttributes['http.target'], '/api/v1/contactservice/contacts/account/') = 1
                        THEN '/api/v1/contactservice/contacts/account/*'
                    WHEN position(SpanAttributes['http.target'], '/api/v1/foodservice/foods/') = 1
                        THEN '/api/v1/foodservice/foods/*/*/*'
                    WHEN position(SpanAttributes['http.target'], '/api/v1/executeservice/execute/collected/') = 1
                        THEN '/api/v1/executeservice/execute/collected/*'
                    WHEN position(SpanAttributes['http.target'], '/api/v1/executeservice/execute/execute/') = 1
                        THEN '/api/v1/executeservice/execute/execute/*'
                    WHEN position(SpanAttributes['http.target'], '/api/v1/userservice/users/id/') = 1
                        THEN '/api/v1/userservice/users/id/*'
                    WHEN match(SpanAttributes['http.target'], '/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}')
                        THEN replaceRegexpAll(SpanAttributes['http.target'], '/([^/]+/[^/]+/[^/]+/)([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})', '/\\1*')
                    ELSE SpanAttributes['http.target']
                END
                
        WHEN SpanAttributes['url.full'] IS NOT NULL AND SpanAttributes['url.full'] != ''
            THEN 
                CASE
                    WHEN match(SpanAttributes['url.full'], 'https?://[^/]+(/.*)') THEN
                        CASE
                            WHEN position(path, '/api/v1/assuranceservice/assurances/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/assuranceservice/assurances/[^/]+/)[^/]+', '\\1*')
                            WHEN position(path, '/api/v1/consignpriceservice/consignprice/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/consignpriceservice/consignprice/)[^/]+/[^/]+', '\\1*/*')
                            WHEN position(path, '/api/v1/contactservice/contacts/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/contactservice/contacts/)[^/]+', '\\1*')
                            WHEN position(path, '/api/v1/inside_pay_service/inside_payment/drawback/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/inside_pay_service/inside_payment/drawback/)[^/]+/[^/]+', '\\1*/*')
                            WHEN position(path, '/api/v1/securityservice/securityConfigs/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/securityservice/securityConfigs/)[^/]+', '\\1*')
                            WHEN position(path, '/api/v1/travel2service/routes/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/travel2service/routes/)[^/]+', '\\1*')
                            WHEN position(path, '/api/v1/routeservice/routes/') = 1 
                                 AND match(path, '/api/v1/routeservice/routes/[^/]+/[^/]+')
                                THEN replaceRegexpAll(path, '(/api/v1/routeservice/routes/)[^/]+/[^/]+', '\\1*/*')
                            WHEN position(path, '/api/v1/orderservice/order/status/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/orderservice/order/status/)[^/]+(/.*)', '\\1*\\2')
                            WHEN position(path, '/api/v1/orderservice/order/security/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/orderservice/order/security/)[^/]+/[^/]+', '\\1*/*')
                            WHEN position(path, '/api/v1/orderservice/order/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/orderservice/order/)[^/]+$', '\\1*')
                            WHEN position(path, '/api/v1/travelservice/routes/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/travelservice/routes/)[^/]+$', '\\1*')
                            WHEN position(path, '/api/v1/trainfoodservice/trainfoods/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/trainfoodservice/trainfoods/)[^/]+$', '\\1*')
                            WHEN position(path, '/api/v1/trainservice/trains/byName/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/trainservice/trains/byName/)[^/]+$', '\\1*')
                            WHEN position(path, '/api/v1/stationservice/stations/id/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/stationservice/stations/id/)[^/]+$', '\\1*')
                            WHEN position(path, '/api/v1/orderOtherService/orderOther/status/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/orderOtherService/orderOther/status/)[^/]+(/.*)', '\\1*\\2')
                            WHEN position(path, '/api/v1/orderOtherService/orderOther/security/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/orderOtherService/orderOther/security/)[^/]+/[^/]+', '\\1*/*')
                            WHEN position(path, '/api/v1/orderOtherService/orderOther/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/orderOtherService/orderOther/)[^/]+$', '\\1*')
                            WHEN position(path, '/api/v1/routeservice/routes/') = 1 
                                 AND NOT match(path, '/api/v1/routeservice/routes/[^/]+/[^/]+')
                                THEN replaceRegexpAll(path, '(/api/v1/routeservice/routes/)[^/]+$', '\\1*')
                            WHEN position(path, '/api/v1/priceservice/prices/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/priceservice/prices/)[^/]+(/[^/]+)', '\\1*\\2')
                            WHEN position(path, '/api/v1/verifycode/verify/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/verifycode/verify/)[^/]+', '\\1*')
                            WHEN position(path, '/api/v1/userservice/users/id/') = 1
                                THEN replaceRegexpAll(path, '(/api/v1/userservice/users/id/)[^/]+', '\\1*')
                            ELSE path
                        END
                    ELSE SpanAttributes['url.full']
                END
        ELSE ''
    END AS masked_route,
    
    SpanAttributes['server.address'] AS server_address,
    SpanAttributes['server.port'] AS server_port,
    SpanAttributes['db.connection_string'] AS db_connection_string,
    SpanAttributes['db.name'] AS db_name,
    SpanAttributes['db.operation'] AS db_operation,
    SpanAttributes['db.sql.table'] AS db_sql_table, 
    SpanAttributes['db.statement'] AS db_statement,
    SpanAttributes['db.system'] AS db_system,
    SpanAttributes['db.user'] AS db_user
FROM otel_traces
WHERE 
    ResourceAttributes['service.namespace'] = 'ts'
    AND SpanKind IN ('Server', 'Client')
    AND mapExists(
        (k, v) -> (k IS NOT NULL AND k != '') AND (v IS NOT NULL AND v != ''),
        SpanAttributes
    );
`

// Client query
const clientTracesQuery = `
SELECT 
    ServiceName,
    request_method,
    response_status_code,
    masked_route,
    server_address,
    server_port
FROM otel_traces_mv
FINAL
WHERE SpanKind = 'Client'
ORDER BY version ASC
`

// Dashboard query
const dashboardRoutesQuery = `
SELECT 
    ServiceName,
    request_method,
    response_status_code,
    masked_route
FROM otel_traces_mv
FINAL
WHERE ServiceName = 'ts-ui-dashboard'
ORDER BY version ASC
`

// MySQL operations query
const mysqlOperationsQuery = `
SELECT 
    ServiceName,
    db_name,
    db_sql_table,
    db_operation
FROM otel_traces_mv
FINAL
WHERE db_system = 'mysql'
ORDER BY version ASC
`

// ConnectToDB establishes a connection to ClickHouse
func ConnectToDB(config ClickHouseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("clickhouse://%s:%d/%s?username=%s&password=%s",
		config.Host, config.Port, config.Database, config.Username, config.Password)

	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return db, nil
}

// CreateMaterializedView creates the materialized view if it doesn't exist
func CreateMaterializedView(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, createMaterializedViewSQL); err != nil {
		return fmt.Errorf("error creating materialized view: %w", err)
	}

	return nil
}

// QueryClientTraces retrieves client traces from the materialized view
func QueryClientTraces(db *sql.DB) ([]ServiceEndpoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, clientTracesQuery)
	if err != nil {
		return nil, fmt.Errorf("error querying client traces: %w", err)
	}
	defer rows.Close()

	var results []ServiceEndpoint
	for rows.Next() {
		var endpoint ServiceEndpoint
		var serverAddr, serverPort sql.NullString

		if err := rows.Scan(
			&endpoint.ServiceName,
			&endpoint.RequestMethod,
			&endpoint.ResponseStatus,
			&endpoint.Route,
			&serverAddr,
			&serverPort,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// Handle null values for server address and port
		if serverAddr.Valid {
			endpoint.ServerAddress = serverAddr.String
		}
		if serverPort.Valid {
			endpoint.ServerPort = serverPort.String
		}

		// If both server address and port are empty, default to RabbitMQ
		if endpoint.ServerAddress == "" && endpoint.ServerPort == "" {
			endpoint.ServerAddress = "ts-rabbitmq"
			endpoint.ServerPort = "5672"
		}

		results = append(results, endpoint)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// QueryDashboardRoutes retrieves routes from the ts-ui-dashboard
func QueryDashboardRoutes(db *sql.DB) ([]ServiceEndpoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, dashboardRoutesQuery)
	if err != nil {
		return nil, fmt.Errorf("error querying dashboard routes: %w", err)
	}
	defer rows.Close()

	var results []ServiceEndpoint
	for rows.Next() {
		var endpoint ServiceEndpoint

		if err := rows.Scan(
			&endpoint.ServiceName,
			&endpoint.RequestMethod,
			&endpoint.ResponseStatus,
			&endpoint.Route,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// Add server information based on route
		mapRouteToService(&endpoint)
		results = append(results, endpoint)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// QueryMySQLOperations retrieves MySQL database operations from the materialized view
func QueryMySQLOperations(db *sql.DB) ([]DatabaseOperation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, mysqlOperationsQuery)
	if err != nil {
		return nil, fmt.Errorf("error querying MySQL operations: %w", err)
	}
	defer rows.Close()

	var results []DatabaseOperation
	for rows.Next() {
		var operation DatabaseOperation
		var dbName, dbTable, dbOperation sql.NullString

		if err := rows.Scan(
			&operation.ServiceName,
			&dbName,
			&dbTable,
			&dbOperation,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// Handle null values
		if dbName.Valid {
			operation.DBName = dbName.String
		}
		if dbTable.Valid {
			operation.DBTable = dbTable.String
		}
		if dbOperation.Valid {
			operation.Operation = dbOperation.String
		}

		results = append(results, operation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// mapRouteToService maps a route to a service based on Caddy rules
func mapRouteToService(endpoint *ServiceEndpoint) {
	// Default to RabbitMQ if we can't determine service
	endpoint.ServerAddress = "ts-rabbitmq"
	endpoint.ServerPort = "5672"

	route := endpoint.Route
	if route == "" {
		return
	}

	// Map route prefixes to services based on Caddy rules
	routeMap := map[string]struct {
		service string
		port    string
	}{
		"/api/v1/adminbasicservice":      {"ts-admin-basic-info-service", "8080"},
		"/api/v1/adminorderservice":      {"ts-admin-order-service", "8080"},
		"/api/v1/adminrouteservice":      {"ts-admin-route-service", "8080"},
		"/api/v1/admintravelservice":     {"ts-admin-travel-service", "8080"},
		"/api/v1/adminuserservice/users": {"ts-admin-user-service", "8080"},
		"/api/v1/assuranceservice":       {"ts-assurance-service", "8080"},
		"/api/v1/auth":                   {"ts-auth-service", "8080"},
		"/api/v1/users":                  {"ts-auth-service", "8080"},
		"/api/v1/avatar":                 {"ts-avatar-service", "8080"},
		"/api/v1/basicservice":           {"ts-basic-service", "8080"},
		"/api/v1/cancelservice":          {"ts-cancel-service", "8080"},
		"/api/v1/configservice":          {"ts-config-service", "8080"},
		"/api/v1/consignpriceservice":    {"ts-consign-price-service", "8080"},
		"/api/v1/consignservice":         {"ts-consign-service", "8080"},
		"/api/v1/contactservice":         {"ts-contacts-service", "8080"},
		"/api/v1/executeservice":         {"ts-execute-service", "8080"},
		"/api/v1/foodservice":            {"ts-food-service", "8080"},
		"/api/v1/inside_pay_service":     {"ts-inside-payment-service", "8080"},
		"/api/v1/notifyservice":          {"ts-notification-service", "8080"},
		"/api/v1/orderOtherService":      {"ts-order-other-service", "8080"},
		"/api/v1/orderservice":           {"ts-order-service", "8080"},
		"/api/v1/paymentservice":         {"ts-payment-service", "8080"},
		"/api/v1/preserveotherservice":   {"ts-preserve-other-service", "8080"},
		"/api/v1/preserveservice":        {"ts-preserve-service", "8080"},
		"/api/v1/priceservice":           {"ts-price-service", "8080"},
		"/api/v1/rebookservice":          {"ts-rebook-service", "8080"},
		"/api/v1/routeplanservice":       {"ts-route-plan-service", "8080"},
		"/api/v1/routeservice":           {"ts-route-service", "8080"},
		"/api/v1/seatservice":            {"ts-seat-service", "8080"},
		"/api/v1/securityservice":        {"ts-security-service", "8080"},
		"/api/v1/stationfoodservice":     {"ts-station-food-service", "8080"},
		"/api/v1/stationservice":         {"ts-station-service", "8080"},
		"/api/v1/trainfoodservice":       {"ts-train-food-service", "8080"},
		"/api/v1/trainservice":           {"ts-train-service", "8080"},
		"/api/v1/travel2service":         {"ts-travel2-service", "8080"},
		"/api/v1/travelplanservice":      {"ts-travel-plan-service", "8080"},
		"/api/v1/travelservice":          {"ts-travel-service", "8080"},
		"/api/v1/userservice/users":      {"ts-user-service", "8080"},
		"/api/v1/verifycode":             {"ts-verification-code-service", "8080"},
		"/api/v1/waitorderservice":       {"ts-wait-order-service", "8080"},
		"/api/v1/fooddeliveryservice":    {"ts-food-delivery-service", "8080"},
	}

	// Find the longest matching prefix
	var longestPrefix string
	for prefix := range routeMap {
		if strings.HasPrefix(route, prefix) && len(prefix) > len(longestPrefix) {
			longestPrefix = prefix
		}
	}

	if longestPrefix != "" {
		service := routeMap[longestPrefix]
		endpoint.ServerAddress = service.service
		endpoint.ServerPort = service.port
	}
}
