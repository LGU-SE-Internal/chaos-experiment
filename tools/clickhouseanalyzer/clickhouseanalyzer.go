package clickhouseanalyzer

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sort"
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
// This is used internally by the analyzer tool
type ServiceEndpoint struct {
	ServiceName    string
	RequestMethod  string
	ResponseStatus string
	Route          string
	ServerAddress  string
	ServerPort     string
	SpanKind       string
	SpanName       string
}

// DatabaseOperation represents a database operation with its details
// This is used internally by the analyzer tool
type DatabaseOperation struct {
	ServiceName   string
	DBName        string
	DBTable       string
	Operation     string
	DBSystem      string
	ServerAddress string
	ServerPort    string
	SpanName      string
}

// GRPCOperation represents a gRPC operation with its details
// This is used internally by the analyzer tool
type GRPCOperation struct {
	ServiceName   string
	RPCSystem     string
	RPCService    string
	RPCMethod     string
	StatusCode    string
	ServerAddress string
	ServerPort    string
	SpanKind      string
	SpanName      string
}

// TrainTicket span name pattern replacements for ts-ui-dashboard and loadgenerator services
// These patterns normalize dynamic URL parameters to template placeholders
var tsSpanNamePatterns = []struct {
	Pattern     *regexp.Regexp
	Replacement string
}{
	{
		regexp.MustCompile(`(.*?)GET (.*?)/api/v1/verifycode/verify/[0-9a-zA-Z]+`),
		"${1}GET ${2}/api/v1/verifycode/verify/{verifyCode}",
	},
	{
		regexp.MustCompile(`(.*?)GET (.*?)/api/v1/foodservice/foods/[0-9]{4}-[0-9]{2}-[0-9]{2}/[a-z]+/[a-z]+/[A-Z0-9]+`),
		"${1}GET ${2}/api/v1/foodservice/foods/{date}/{startStation}/{endStation}/{tripId}",
	},
	{
		regexp.MustCompile(`(.*?)GET (.*?)/api/v1/contactservice/contacts/account/[0-9a-f-]+`),
		"${1}GET ${2}/api/v1/contactservice/contacts/account/{accountId}",
	},
	{
		regexp.MustCompile(`(.*?)GET (.*?)/api/v1/userservice/users/id/[0-9a-f-]+`),
		"${1}GET ${2}/api/v1/userservice/users/id/{userId}",
	},
	{
		regexp.MustCompile(`(.*?)GET (.*?)/api/v1/consignservice/consigns/order/[0-9a-f-]+`),
		"${1}GET ${2}/api/v1/consignservice/consigns/order/{id}",
	},
	{
		regexp.MustCompile(`(.*?)GET (.*?)/api/v1/consignservice/consigns/account/[0-9a-f-]+`),
		"${1}GET ${2}/api/v1/consignservice/consigns/account/{id}",
	},
	{
		regexp.MustCompile(`(.*?)GET (.*?)/api/v1/executeservice/execute/collected/[0-9a-f-]+`),
		"${1}GET ${2}/api/v1/executeservice/execute/collected/{orderId}",
	},
	{
		regexp.MustCompile(`(.*?)GET (.*?)/api/v1/cancelservice/cancel/[0-9a-f-]+/[0-9a-f-]+`),
		"${1}GET ${2}/api/v1/cancelservice/cancel/{orderId}/{loginId}",
	},
	{
		regexp.MustCompile(`(.*?)GET (.*?)/api/v1/cancelservice/cancel/refound/[0-9a-f-]+`),
		"${1}GET ${2}/api/v1/cancelservice/cancel/refound/{orderId}",
	},
	{
		regexp.MustCompile(`(.*?)GET (.*?)/api/v1/executeservice/execute/execute/[0-9a-f-]+`),
		"${1}GET ${2}/api/v1/executeservice/execute/execute/{orderId}",
	},
	{
		regexp.MustCompile(`(.*?)DELETE (.*?)/api/v1/adminorderservice/adminorder/[0-9a-f-]+/[A-Z0-9]+`),
		"${1}DELETE ${2}/api/v1/adminorderservice/adminorder/{orderId}/{trainNumber}",
	},
	{
		regexp.MustCompile(`(.*?)DELETE (.*?)/api/v1/adminrouteservice/adminroute/[0-9a-f-]+`),
		"${1}DELETE ${2}/api/v1/adminrouteservice/adminroute/{routeId}",
	},
}

// NormalizeTrainTicketSpanName applies pattern replacements to normalize
// span names for ts-ui-dashboard and loadgenerator services
func NormalizeTrainTicketSpanName(spanName string, serviceName string) string {
	// Only apply replacements for ts-ui-dashboard and loadgenerator
	if serviceName != "ts-ui-dashboard" && serviceName != "loadgenerator" {
		return spanName
	}

	for _, p := range tsSpanNamePatterns {
		if p.Pattern.MatchString(spanName) {
			return p.Pattern.ReplaceAllString(spanName, p.Replacement)
		}
	}
	return spanName
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
                    -- New patterns first for priority matching
                    -- /api/v1/adminorderservice/adminorder/{uuid}/{id}
                    WHEN match(SpanAttributes['http.target'], '/api/v1/adminorderservice/adminorder/[0-9a-f-]+/[A-Z0-9]+')
                        THEN '/api/v1/adminorderservice/adminorder/*/*'
                    -- /api/v1/users/{uuid}
                    WHEN match(SpanAttributes['http.target'], '^/api/v1/users/[0-9a-f-]+$')
                        THEN '/api/v1/users/*'
                    -- Existing patterns
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
                    -- Generic UUID pattern for remaining cases
                    WHEN match(SpanAttributes['http.target'], '/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}')
                        THEN replaceRegexpAll(SpanAttributes['http.target'], '/([^/]+/[^/]+/[^/]+/)([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})', '/\\1*')
                    ELSE SpanAttributes['http.target']
                END
                
        WHEN SpanAttributes['url.full'] IS NOT NULL AND SpanAttributes['url.full'] != ''
            THEN 
                CASE
                    WHEN match(SpanAttributes['url.full'], 'https?://[^/]+(/.*)') THEN
                        CASE
                            -- New patterns first for priority matching
                            -- /api/v1/adminorderservice/adminorder/{uuid}/{id}
                            WHEN match(path, '/api/v1/adminorderservice/adminorder/[0-9a-f-]+/[A-Z0-9]+')
                                THEN replaceRegexpAll(path, '(/api/v1/adminorderservice/adminorder/)[^/]+/[^/]+', '\\1*/*')
                            -- /api/v1/users/{uuid}
                            WHEN match(path, '^/api/v1/users/[0-9a-f-]+$')
                                THEN '/api/v1/users/*'
                            -- /api/v1/userservice/users/{uuid}
                            WHEN match(path, '^/api/v1/userservice/users/[0-9a-f-]+$')
                                THEN '/api/v1/userservice/users/*'
                            -- Existing patterns
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
    SpanAttributes['db.user'] AS db_user,
    SpanName AS span_name
FROM otel_traces
WHERE 
    ResourceAttributes['service.namespace'] = 'ts0'
    AND SpanKind IN ('Server', 'Client')
    AND mapExists(
        (k, v) -> (k IS NOT NULL AND k != '') AND (v IS NOT NULL AND v != ''),
        SpanAttributes
    );
`

// Create materialized view SQL statement for OpenTelemetry Demo
const createOtelDemoMaterializedViewSQL = `
CREATE MATERIALIZED VIEW IF NOT EXISTS otel_demo_traces_mv 
ENGINE = ReplacingMergeTree(version)
PARTITION BY toYYYYMM(Timestamp)
PRIMARY KEY (masked_route, ServiceName, db_name)
ORDER BY (
    masked_route,
    ServiceName,
    db_name,
    SpanKind,
    request_method,
    response_status_code,
    server_address,
    server_port,
    db_operation,
    db_sql_table,
    rpc_system,
    rpc_service,
    rpc_method,
    grpc_status_code
)
SETTINGS allow_nullable_key = 1
POPULATE
AS 
WITH 
    -- Extract path from url.full (without query string)
    replaceRegexpOne(SpanAttributes['url.full'], 'https?://[^/]+(/[^?]*)?.*', '\\1') AS url_path,
    -- Extract query string from url.full  
    replaceRegexpOne(SpanAttributes['url.full'], 'https?://[^/]+[^?]*(\\?.*)?$', '\\1') AS url_query,
    -- Extract path from http.target (without query string)
    replaceRegexpOne(SpanAttributes['http.target'], '^([^?]*)(\\?.*)?$', '\\1') AS target_path,
    -- Extract query string from http.target
    replaceRegexpOne(SpanAttributes['http.target'], '^[^?]*(\\?.*)?$', '\\1') AS target_query
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
        -- Priority 1: http.route (usually already parameterized like /api/products/{productId})
        WHEN SpanAttributes['http.route'] IS NOT NULL AND SpanAttributes['http.route'] != ''
            THEN 
                -- Replace {param} style with * and product IDs like /XXXXXX with /*
                replaceRegexpAll(
                    replaceRegexpAll(SpanAttributes['http.route'], '\\{[^}]+\\}', '*'),
                    '/[A-Z0-9]{10}',
                    '/*'
                )
        
        -- Priority 2: url.full - need to extract path and mask parameters
        WHEN SpanAttributes['url.full'] IS NOT NULL AND SpanAttributes['url.full'] != ''
            THEN 
                CASE
                    -- /api/products/{productId} - product IDs are 10 char alphanumeric
                    WHEN match(url_path, '^/api/products/[A-Z0-9]+$')
                        THEN '/api/products/*'
                    -- /api/recommendations?productIds={id}
                    WHEN url_path = '/api/recommendations' AND match(url_query, '^\\?productIds=')
                        THEN '/api/recommendations?productIds=*'
                    -- /api/data?contextKeys={key}
                    WHEN url_path = '/api/data' AND match(url_query, '^\\?contextKeys=')
                        THEN '/api/data?contextKeys=*'
                    -- /api/data/?contextKeys={key} (with trailing slash before query)
                    WHEN url_path = '/api/data/' AND match(url_query, '^\\?contextKeys=')
                        THEN '/api/data/?contextKeys=*'
                    -- /ofrep/v1/evaluate/flags/{flagName}
                    WHEN match(url_path, '^/ofrep/v1/evaluate/flags/[^/]+$')
                        THEN '/ofrep/v1/evaluate/flags/*'
                    -- Default: just use the path without query params
                    ELSE 
                        CASE 
                            WHEN url_path != '' THEN url_path 
                            ELSE '/' 
                        END
                END
        
        -- Priority 3: http.target - also need to mask parameters
        WHEN SpanAttributes['http.target'] IS NOT NULL AND SpanAttributes['http.target'] != ''
            THEN 
                CASE
                    -- /api/products/{productId}
                    WHEN match(target_path, '^/api/products/[A-Z0-9]+$')
                        THEN '/api/products/*'
                    -- /api/recommendations?productIds={id}
                    WHEN target_path = '/api/recommendations' AND match(target_query, '^\\?productIds=')
                        THEN '/api/recommendations?productIds=*'
                    -- /api/data?contextKeys={key}
                    WHEN target_path = '/api/data' AND match(target_query, '^\\?contextKeys=')
                        THEN '/api/data?contextKeys=*'
                    -- /api/data/?contextKeys={key}
                    WHEN target_path = '/api/data/' AND match(target_query, '^\\?contextKeys=')
                        THEN '/api/data/?contextKeys=*'
                    -- /ofrep/v1/evaluate/flags/{flagName}
                    WHEN match(target_path, '^/ofrep/v1/evaluate/flags/[^/]+$')
                        THEN '/ofrep/v1/evaluate/flags/*'
                    -- Default: use target_path without query
                    ELSE 
                        CASE 
                            WHEN target_path != '' THEN target_path 
                            ELSE SpanAttributes['http.target']
                        END
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
    SpanAttributes['db.user'] AS db_user,
    SpanAttributes['rpc.system'] AS rpc_system,
    SpanAttributes['rpc.service'] AS rpc_service,
    SpanAttributes['rpc.method'] AS rpc_method,
    SpanAttributes['rpc.grpc.status_code'] AS grpc_status_code,
    SpanName AS span_name
FROM otel_traces
WHERE 
    ResourceAttributes['service.namespace'] = 'otel-demo'
    AND SpanKind IN ('Server', 'Client')
    AND mapExists(
        (k, v) -> (k IS NOT NULL AND k != '') AND (v IS NOT NULL AND v != ''),
        SpanAttributes
    );
`

// Create materialized view SQL for DeathStarBench systems (media, hs, sn)
// These systems use ResourceAttributes['k8s.namespace.name'] for filtering
func createDeathStarBenchMaterializedViewSQL(namespace string, viewName string) string {
	return fmt.Sprintf(`
CREATE MATERIALIZED VIEW IF NOT EXISTS %s 
ENGINE = ReplacingMergeTree(version)
PARTITION BY toYYYYMM(Timestamp)
PRIMARY KEY (masked_route, ServiceName, db_name)
ORDER BY (
    masked_route,
    ServiceName,
    db_name,
    SpanKind,
    request_method,
    response_status_code,
    server_address,
    server_port,
    db_operation,
    db_sql_table,
    rpc_system,
    rpc_service,
    rpc_method,
    grpc_status_code
)
SETTINGS allow_nullable_key = 1
POPULATE
AS 
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
    
    -- Path normalization for DeathStarBench systems - replace IDs with wildcards
    -- Matches: UUIDs (8-4-4-4-12 hex format) and numeric IDs (sequences of digits)
    CASE
        WHEN SpanAttributes['http.route'] IS NOT NULL AND SpanAttributes['http.route'] != ''
            THEN replaceRegexpAll(SpanAttributes['http.route'], '/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|/\\d+', '/*')
        WHEN SpanAttributes['http.target'] IS NOT NULL AND SpanAttributes['http.target'] != ''
            THEN replaceRegexpAll(SpanAttributes['http.target'], '/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|/\\d+', '/*')
        WHEN SpanAttributes['url.full'] IS NOT NULL AND SpanAttributes['url.full'] != ''
            THEN replaceRegexpAll(replaceRegexpOne(SpanAttributes['url.full'], 'https?://[^/]+(/.*)', '\\1'), '/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|/\\d+', '/*')
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
    SpanAttributes['db.user'] AS db_user,
    SpanAttributes['rpc.system'] AS rpc_system,
    SpanAttributes['rpc.service'] AS rpc_service,
    SpanAttributes['rpc.method'] AS rpc_method,
    SpanAttributes['rpc.grpc.status_code'] AS grpc_status_code,
    SpanName AS span_name
FROM otel_traces
WHERE 
    ResourceAttributes['k8s.namespace.name'] = '%s'
    AND SpanKind IN ('Server', 'Client')
    AND mapExists(
        (k, v) -> (k IS NOT NULL AND k != '') AND (v IS NOT NULL AND v != ''),
        SpanAttributes
    );
`, viewName, namespace)
}

// createOnlineBoutiqueMaterializedViewSQL creates SQL for OnlineBoutique materialized view
// Filters out OpenTelemetry collector internal spans
func createOnlineBoutiqueMaterializedViewSQL(namespace string, viewName string) string {
	return fmt.Sprintf(`
CREATE MATERIALIZED VIEW IF NOT EXISTS %s 
ENGINE = ReplacingMergeTree(version)
PARTITION BY toYYYYMM(Timestamp)
PRIMARY KEY (masked_route, ServiceName, db_name)
ORDER BY (
    masked_route,
    ServiceName,
    db_name,
    SpanKind,
    request_method,
    response_status_code,
    server_address,
    server_port,
    db_operation,
    db_sql_table,
    rpc_system,
    rpc_service,
    rpc_method,
    grpc_status_code
)
SETTINGS allow_nullable_key = 1
POPULATE
AS 
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
    
    -- Path normalization for DeathStarBench systems - replace IDs with wildcards
    -- Matches: UUIDs (8-4-4-4-12 hex format) and numeric IDs (sequences of digits)
    CASE
        WHEN SpanAttributes['http.route'] IS NOT NULL AND SpanAttributes['http.route'] != ''
            THEN replaceRegexpAll(SpanAttributes['http.route'], '/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|/\\d+', '/*')
        WHEN SpanAttributes['http.target'] IS NOT NULL AND SpanAttributes['http.target'] != ''
            THEN replaceRegexpAll(SpanAttributes['http.target'], '/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|/\\d+', '/*')
        WHEN SpanAttributes['url.full'] IS NOT NULL AND SpanAttributes['url.full'] != ''
            THEN replaceRegexpAll(replaceRegexpOne(SpanAttributes['url.full'], 'https?://[^/]+(/.*)', '\\1'), '/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|/\\d+', '/*')
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
    SpanAttributes['db.user'] AS db_user,
    SpanAttributes['rpc.system'] AS rpc_system,
    SpanAttributes['rpc.service'] AS rpc_service,
    SpanAttributes['rpc.method'] AS rpc_method,
    SpanAttributes['rpc.grpc.status_code'] AS grpc_status_code,
    SpanName AS span_name
FROM otel_traces
WHERE 
    ResourceAttributes['k8s.namespace.name'] = '%s'
    AND SpanKind IN ('Server', 'Client')
    AND SpanName != 'opentelemetry.proto.collector.trace.v1.TraceService/Export'
    AND mapExists(
        (k, v) -> (k IS NOT NULL AND k != '') AND (v IS NOT NULL AND v != ''),
        SpanAttributes
    );
`, viewName, namespace)
}

// Client query - HTTP endpoints only (excludes database and RPC operations)
const clientTracesQuery = `
SELECT DISTINCT
    ServiceName,
    request_method,
    response_status_code,
    masked_route,
    server_address,
    server_port,
    span_name
FROM otel_traces_mv
FINAL
WHERE SpanKind = 'Client'
  AND (db_system IS NULL OR db_system = '')  -- Exclude database operations
  AND request_method != ''  -- Must have HTTP method
ORDER BY version ASC
`

// Dashboard query - HTTP endpoints only
const dashboardRoutesQuery = `
SELECT DISTINCT
    ServiceName,
    request_method,
    response_status_code,
    masked_route,
    span_name
FROM otel_traces_mv
FINAL
WHERE ServiceName = 'ts-ui-dashboard'
  AND (db_system IS NULL OR db_system = '')  -- Exclude database operations
  AND request_method != ''  -- Must have HTTP method
ORDER BY version ASC
`

// MySQL operations query
const mysqlOperationsQuery = `
SELECT DISTINCT
    ServiceName,
    db_name,
    db_sql_table,
    db_operation,
    span_name
FROM otel_traces_mv
FINAL
WHERE db_system = 'mysql'
ORDER BY version ASC
`

// HTTP Client traces query for OTel Demo - HTTP endpoints only (excludes database and RPC)
const otelDemoHTTPClientTracesQuery = `
SELECT DISTINCT
    ServiceName,
    request_method,
    response_status_code,
    masked_route,
    server_address,
    server_port,
    span_name
FROM otel_demo_traces_mv
FINAL
WHERE SpanKind = 'Client'
  AND (db_system IS NULL OR db_system = '')  -- Exclude database operations
  AND (rpc_system IS NULL OR rpc_system = '')  -- Exclude RPC operations
  AND request_method != ''
  AND masked_route != ''
ORDER BY ServiceName, masked_route
`

// HTTP Server traces query for OTel Demo - include client_address
const otelDemoHTTPServerTracesQuery = `
SELECT DISTINCT
    ServiceName,
    request_method,
    response_status_code,
    masked_route,
    server_address,
    server_port,
    client_address,
    span_name
FROM otel_demo_traces_mv
FINAL
WHERE SpanKind = 'Server'
  AND request_method != ''
  AND masked_route != ''
ORDER BY ServiceName, masked_route
`

// gRPC operations query for OTel Demo
const otelDemoGRPCOperationsQuery = `
SELECT DISTINCT
    ServiceName,
    rpc_system,
    rpc_service,
    rpc_method,
    grpc_status_code,
    server_address,
    server_port,
    SpanKind,
    span_name
FROM otel_demo_traces_mv
FINAL
WHERE rpc_system != ''
  AND rpc_service != ''
ORDER BY ServiceName, rpc_service, rpc_method
`

// Database operations query for OTel Demo
const otelDemoDatabaseOperationsQuery = `
SELECT DISTINCT
    ServiceName,
    db_name,
    db_sql_table,
    db_operation,
    db_system,
    span_name
FROM otel_demo_traces_mv
FINAL
WHERE db_system != ''
ORDER BY ServiceName, db_name
`

// deathStarBenchHTTPClientTracesQuery generates a query for HTTP client traces for DeathStarBench systems
// HTTP endpoints only (excludes database and RPC operations)
func deathStarBenchHTTPClientTracesQuery(viewName string) string {
	return fmt.Sprintf(`
SELECT DISTINCT
    ServiceName,
    request_method,
    response_status_code,
    masked_route,
    server_address,
    server_port,
    span_name
FROM %s
FINAL
WHERE SpanKind = 'Client'
  AND (db_system IS NULL OR db_system = '')  -- Exclude database operations
  AND (rpc_system IS NULL OR rpc_system = '')  -- Exclude RPC operations
  AND request_method != ''
  AND masked_route != ''
ORDER BY ServiceName, masked_route
`, viewName)
}

// deathStarBenchHTTPServerTracesQuery generates a query for HTTP server traces for DeathStarBench systems
func deathStarBenchHTTPServerTracesQuery(viewName string) string {
	return fmt.Sprintf(`
SELECT DISTINCT
    ServiceName,
    request_method,
    response_status_code,
    masked_route,
    server_address,
    server_port,
    client_address,
    span_name
FROM %s
FINAL
WHERE SpanKind = 'Server'
  AND request_method != ''
  AND masked_route != ''
ORDER BY ServiceName, masked_route
`, viewName)
}

// deathStarBenchGRPCOperationsQuery generates a query for gRPC operations for DeathStarBench systems
func deathStarBenchGRPCOperationsQuery(viewName string) string {
	return fmt.Sprintf(`
SELECT DISTINCT
    ServiceName,
    rpc_system,
    rpc_service,
    rpc_method,
    grpc_status_code,
    server_address,
    server_port,
    SpanKind,
    span_name
FROM %s
FINAL
WHERE rpc_system != ''
  AND rpc_service != ''
ORDER BY ServiceName, rpc_service, rpc_method
`, viewName)
}

// deathStarBenchDatabaseOperationsQuery generates a query for database operations for DeathStarBench systems
func deathStarBenchDatabaseOperationsQuery(viewName string) string {
	return fmt.Sprintf(`
SELECT DISTINCT
    ServiceName,
    db_name,
    db_sql_table,
    db_operation,
    db_system,
    span_name
FROM %s
FINAL
WHERE db_system != ''
ORDER BY ServiceName, db_name
`, viewName)
}

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

// CreateOtelDemoMaterializedView creates the materialized view for OTel Demo
func CreateOtelDemoMaterializedView(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, createOtelDemoMaterializedViewSQL); err != nil {
		return fmt.Errorf("error creating OTel Demo materialized view: %w", err)
	}

	return nil
}

// CreateDeathStarBenchMaterializedView creates the materialized view for DeathStarBench systems
// namespace: the k8s namespace (media, hs, sn)
// viewName: the name of the materialized view to create
func CreateDeathStarBenchMaterializedView(db *sql.DB, namespace string, viewName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sql := createDeathStarBenchMaterializedViewSQL(namespace, viewName)

	if _, err := db.ExecContext(ctx, sql); err != nil {
		return fmt.Errorf("error creating DeathStarBench materialized view for namespace %s: %w", namespace, err)
	}

	return nil
}

// CreateOnlineBoutiqueMaterializedView creates the materialized view for OnlineBoutique system
// namespace: the k8s namespace (ob)
// viewName: the name of the materialized view to create
func CreateOnlineBoutiqueMaterializedView(db *sql.DB, namespace string, viewName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sql := createOnlineBoutiqueMaterializedViewSQL(namespace, viewName)

	if _, err := db.ExecContext(ctx, sql); err != nil {
		return fmt.Errorf("error creating OnlineBoutique materialized view for namespace %s: %w", namespace, err)
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
		var serverAddr, serverPort, spanName sql.NullString

		if err := rows.Scan(
			&endpoint.ServiceName,
			&endpoint.RequestMethod,
			&endpoint.ResponseStatus,
			&endpoint.Route,
			&serverAddr,
			&serverPort,
			&spanName,
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

		// Handle span name with normalization for TrainTicket services
		if spanName.Valid {
			endpoint.SpanName = NormalizeTrainTicketSpanName(spanName.String, endpoint.ServiceName)
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
		var spanName sql.NullString

		if err := rows.Scan(
			&endpoint.ServiceName,
			&endpoint.RequestMethod,
			&endpoint.ResponseStatus,
			&endpoint.Route,
			&spanName,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// Handle span name with normalization for TrainTicket services
		if spanName.Valid {
			endpoint.SpanName = NormalizeTrainTicketSpanName(spanName.String, endpoint.ServiceName)
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
		var dbName, dbTable, dbOperation, spanName sql.NullString

		if err := rows.Scan(
			&operation.ServiceName,
			&dbName,
			&dbTable,
			&dbOperation,
			&spanName,
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
		if spanName.Valid {
			operation.SpanName = spanName.String
		}

		// Set fixed MySQL connection info for TrainTicket
		operation.DBSystem = "mysql"
		operation.ServerAddress = "mysql"
		operation.ServerPort = "3306"

		results = append(results, operation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// QueryOtelDemoHTTPClientTraces retrieves HTTP client traces for OTel Demo
func QueryOtelDemoHTTPClientTraces(db *sql.DB) ([]ServiceEndpoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, otelDemoHTTPClientTracesQuery)
	if err != nil {
		return nil, fmt.Errorf("error querying OTel Demo HTTP client traces: %w", err)
	}
	defer rows.Close()

	var results []ServiceEndpoint
	for rows.Next() {
		var endpoint ServiceEndpoint
		var serverAddr, serverPort, spanName sql.NullString

		if err := rows.Scan(
			&endpoint.ServiceName,
			&endpoint.RequestMethod,
			&endpoint.ResponseStatus,
			&endpoint.Route,
			&serverAddr,
			&serverPort,
			&spanName,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		endpoint.SpanKind = "Client"

		if serverAddr.Valid {
			endpoint.ServerAddress = serverAddr.String
		}
		if serverPort.Valid {
			endpoint.ServerPort = serverPort.String
		}
		if spanName.Valid {
			endpoint.SpanName = spanName.String
		}

		// Map empty server address or IP to service based on route
		if endpoint.ServerAddress == "" || isIPAddress(endpoint.ServerAddress) {
			mapOtelDemoRouteToService(&endpoint)
		}

		results = append(results, endpoint)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// QueryOtelDemoHTTPServerTraces retrieves HTTP server traces for OTel Demo
func QueryOtelDemoHTTPServerTraces(db *sql.DB) ([]ServiceEndpoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, otelDemoHTTPServerTracesQuery)
	if err != nil {
		return nil, fmt.Errorf("error querying OTel Demo HTTP server traces: %w", err)
	}
	defer rows.Close()

	var results []ServiceEndpoint
	for rows.Next() {
		var endpoint ServiceEndpoint
		var serverAddr, serverPort, clientAddr, spanName sql.NullString

		if err := rows.Scan(
			&endpoint.ServiceName,
			&endpoint.RequestMethod,
			&endpoint.ResponseStatus,
			&endpoint.Route,
			&serverAddr,
			&serverPort,
			&clientAddr,
			&spanName,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		endpoint.SpanKind = "Server"

		// For Server spans, use the client address as the "caller"
		// The ServerAddress field will represent who is calling this service
		if clientAddr.Valid && clientAddr.String != "" {
			endpoint.ServerAddress = clientAddr.String
			endpoint.ServerPort = "" // Client port is usually dynamic, leave empty
		} else if serverAddr.Valid {
			endpoint.ServerAddress = serverAddr.String
		}
		if serverPort.Valid {
			endpoint.ServerPort = serverPort.String
		}
		if spanName.Valid {
			endpoint.SpanName = spanName.String
		}

		// Map client address (IP) to service name if possible
		mapOtelDemoClientToService(&endpoint)

		results = append(results, endpoint)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// QueryOtelDemoGRPCOperations retrieves gRPC operations for OTel Demo
func QueryOtelDemoGRPCOperations(db *sql.DB) ([]GRPCOperation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, otelDemoGRPCOperationsQuery)
	if err != nil {
		return nil, fmt.Errorf("error querying OTel Demo gRPC operations: %w", err)
	}
	defer rows.Close()

	var results []GRPCOperation
	for rows.Next() {
		var operation GRPCOperation
		var serverAddr, serverPort, grpcStatus, spanName sql.NullString

		if err := rows.Scan(
			&operation.ServiceName,
			&operation.RPCSystem,
			&operation.RPCService,
			&operation.RPCMethod,
			&grpcStatus,
			&serverAddr,
			&serverPort,
			&operation.SpanKind,
			&spanName,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		if serverAddr.Valid {
			operation.ServerAddress = serverAddr.String
		}
		if serverPort.Valid {
			operation.ServerPort = serverPort.String
		}
		if grpcStatus.Valid {
			operation.StatusCode = grpcStatus.String
		}
		if spanName.Valid {
			operation.SpanName = spanName.String
		}

		// Map empty server address or IP to service based on RPC service
		if operation.ServerAddress == "" || isIPAddress(operation.ServerAddress) {
			mapOtelDemoGRPCToService(&operation)
		}

		results = append(results, operation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// QueryOtelDemoDatabaseOperations retrieves database operations for OTel Demo
func QueryOtelDemoDatabaseOperations(db *sql.DB) ([]DatabaseOperation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, otelDemoDatabaseOperationsQuery)
	if err != nil {
		return nil, fmt.Errorf("error querying OTel Demo database operations: %w", err)
	}
	defer rows.Close()

	var results []DatabaseOperation
	for rows.Next() {
		var operation DatabaseOperation
		var dbName, dbTable, dbOperation, dbSystem, spanName sql.NullString

		if err := rows.Scan(
			&operation.ServiceName,
			&dbName,
			&dbTable,
			&dbOperation,
			&dbSystem,
			&spanName,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		if dbName.Valid {
			operation.DBName = dbName.String
		}
		if dbTable.Valid {
			operation.DBTable = dbTable.String
		}
		if dbOperation.Valid {
			operation.Operation = dbOperation.String
		}
		if dbSystem.Valid {
			operation.DBSystem = dbSystem.String
		}
		if spanName.Valid {
			operation.SpanName = spanName.String
		}

		// Map database system to server address and port
		mapOtelDemoDatabaseToService(&operation)

		results = append(results, operation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// QueryDeathStarBenchHTTPClientTraces retrieves HTTP client traces for DeathStarBench systems
func QueryDeathStarBenchHTTPClientTraces(db *sql.DB, viewName string, namespace string) ([]ServiceEndpoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := deathStarBenchHTTPClientTracesQuery(viewName)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying DeathStarBench HTTP client traces: %w", err)
	}
	defer rows.Close()

	var results []ServiceEndpoint
	for rows.Next() {
		var endpoint ServiceEndpoint
		var serverAddr, serverPort, spanName sql.NullString

		if err := rows.Scan(
			&endpoint.ServiceName,
			&endpoint.RequestMethod,
			&endpoint.ResponseStatus,
			&endpoint.Route,
			&serverAddr,
			&serverPort,
			&spanName,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		endpoint.SpanKind = "Client"

		if serverAddr.Valid {
			endpoint.ServerAddress = serverAddr.String
		}
		if serverPort.Valid {
			endpoint.ServerPort = serverPort.String
		}
		if spanName.Valid {
			endpoint.SpanName = spanName.String
		}

		// Map empty server address or IP to service based on route/span name
		// Also map when server address equals service name (frontend calling itself is incorrect)
		if endpoint.ServerAddress == "" || isIPAddress(endpoint.ServerAddress) || endpoint.ServerAddress == endpoint.ServiceName {
			mapDeathStarBenchRouteToService(&endpoint, namespace)
		}

		results = append(results, endpoint)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// QueryDeathStarBenchHTTPServerTraces retrieves HTTP server traces for DeathStarBench systems
func QueryDeathStarBenchHTTPServerTraces(db *sql.DB, viewName string, namespace string) ([]ServiceEndpoint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := deathStarBenchHTTPServerTracesQuery(viewName)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying DeathStarBench HTTP server traces: %w", err)
	}
	defer rows.Close()

	var results []ServiceEndpoint
	for rows.Next() {
		var endpoint ServiceEndpoint
		var serverAddr, serverPort, clientAddr, spanName sql.NullString

		if err := rows.Scan(
			&endpoint.ServiceName,
			&endpoint.RequestMethod,
			&endpoint.ResponseStatus,
			&endpoint.Route,
			&serverAddr,
			&serverPort,
			&clientAddr,
			&spanName,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		endpoint.SpanKind = "Server"

		if clientAddr.Valid && clientAddr.String != "" {
			endpoint.ServerAddress = clientAddr.String
			endpoint.ServerPort = ""
		} else if serverAddr.Valid {
			endpoint.ServerAddress = serverAddr.String
		}
		if serverPort.Valid {
			endpoint.ServerPort = serverPort.String
		}
		if spanName.Valid {
			endpoint.SpanName = spanName.String
		}

		// Map empty server address or IP to service based on route/span name
		// Also map when server address equals service name (frontend calling itself is incorrect)
		if endpoint.ServerAddress == "" || isIPAddress(endpoint.ServerAddress) || endpoint.ServerAddress == endpoint.ServiceName {
			mapDeathStarBenchRouteToService(&endpoint, namespace)
		}

		results = append(results, endpoint)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// QueryDeathStarBenchGRPCOperations retrieves gRPC operations for DeathStarBench systems
func QueryDeathStarBenchGRPCOperations(db *sql.DB, viewName string, namespace string) ([]GRPCOperation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := deathStarBenchGRPCOperationsQuery(viewName)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying DeathStarBench gRPC operations: %w", err)
	}
	defer rows.Close()

	var results []GRPCOperation
	for rows.Next() {
		var operation GRPCOperation
		var serverAddr, serverPort, grpcStatus, spanName sql.NullString

		if err := rows.Scan(
			&operation.ServiceName,
			&operation.RPCSystem,
			&operation.RPCService,
			&operation.RPCMethod,
			&grpcStatus,
			&serverAddr,
			&serverPort,
			&operation.SpanKind,
			&spanName,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		if serverAddr.Valid {
			operation.ServerAddress = serverAddr.String
		}
		if serverPort.Valid {
			operation.ServerPort = serverPort.String
		}
		if grpcStatus.Valid {
			operation.StatusCode = grpcStatus.String
		}
		if spanName.Valid {
			operation.SpanName = spanName.String
		}

		// Map empty server address or IP to service based on RPC service/method
		// Also map when server address equals service name (frontend calling itself is incorrect)
		if operation.ServerAddress == "" || isIPAddress(operation.ServerAddress) || operation.ServerAddress == operation.ServiceName {
			mapDeathStarBenchGRPCToService(&operation, namespace)
		}

		results = append(results, operation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// QueryDeathStarBenchDatabaseOperations retrieves database operations for DeathStarBench systems
func QueryDeathStarBenchDatabaseOperations(db *sql.DB, viewName string) ([]DatabaseOperation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := deathStarBenchDatabaseOperationsQuery(viewName)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying DeathStarBench database operations: %w", err)
	}
	defer rows.Close()

	var results []DatabaseOperation
	for rows.Next() {
		var operation DatabaseOperation
		var dbName, dbTable, dbOperation, dbSystem, spanName sql.NullString

		if err := rows.Scan(
			&operation.ServiceName,
			&dbName,
			&dbTable,
			&dbOperation,
			&dbSystem,
			&spanName,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		if dbName.Valid {
			operation.DBName = dbName.String
		}
		if dbTable.Valid {
			operation.DBTable = dbTable.String
		}
		if dbOperation.Valid {
			operation.Operation = dbOperation.String
		}
		if dbSystem.Valid {
			operation.DBSystem = dbSystem.String
		}
		if spanName.Valid {
			operation.SpanName = spanName.String
		}

		// Map database system to server address and port
		mapDeathStarBenchDatabaseToService(&operation)

		results = append(results, operation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// ConvertDatabaseOperationsToEndpoints converts database operations to service endpoints
// This allows database connections to be included in the service endpoints for network dependency analysis
func ConvertDatabaseOperationsToEndpoints(operations []DatabaseOperation) []ServiceEndpoint {
	// Use a map to deduplicate - one entry per service-database combination
	seen := make(map[string]bool)
	var endpoints []ServiceEndpoint

	for _, op := range operations {
		// Create a unique key for deduplication
		key := fmt.Sprintf("%s-%s-%s", op.ServiceName, op.ServerAddress, op.ServerPort)
		if seen[key] {
			continue
		}
		seen[key] = true

		// Construct a span name for the database operation
		spanName := op.Operation
		if spanName == "" {
			spanName = "DB Access"
		}
		if op.DBTable != "" {
			spanName = fmt.Sprintf("%s %s", spanName, op.DBTable)
		}

		endpoint := ServiceEndpoint{
			ServiceName:    op.ServiceName,
			RequestMethod:  "", // Database operations don't have HTTP methods
			ResponseStatus: "", // Database operations don't have HTTP status
			Route:          "", // Database operations don't have routes
			ServerAddress:  op.ServerAddress,
			ServerPort:     op.ServerPort,
			SpanKind:       "Client", // Database connections are always client-side
			SpanName:       spanName,
		}
		endpoints = append(endpoints, endpoint)
	}

	return endpoints
}

// ConvertGRPCOperationsToEndpoints converts gRPC operations to service endpoints
// This allows gRPC connections to be included in the service endpoints for network dependency analysis
func ConvertGRPCOperationsToEndpoints(operations []GRPCOperation) []ServiceEndpoint {
	// Use a map to deduplicate
	seen := make(map[string]bool)
	var endpoints []ServiceEndpoint

	for _, op := range operations {
		// Only include client-side gRPC operations (outgoing calls)
		if op.SpanKind != "Client" {
			continue
		}

		// Create a unique key for deduplication
		key := fmt.Sprintf("%s-%s-%s-%s", op.ServiceName, op.ServerAddress, op.ServerPort, op.RPCService)
		if seen[key] {
			continue
		}
		seen[key] = true

		// Build the route from RPC service and method
		route := fmt.Sprintf("/%s/%s", op.RPCService, op.RPCMethod)

		// Construct span name (typically Service/Method for gRPC)
		spanName := fmt.Sprintf("%s/%s", op.RPCService, op.RPCMethod)

		endpoint := ServiceEndpoint{
			ServiceName:    op.ServiceName,
			RequestMethod:  "POST", // gRPC uses POST
			ResponseStatus: "",     // gRPC status codes are different from HTTP
			Route:          route,
			ServerAddress:  op.ServerAddress,
			ServerPort:     op.ServerPort,
			SpanKind:       "Client",
			SpanName:       spanName,
		}
		endpoints = append(endpoints, endpoint)
	}

	return endpoints
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

// mapOtelDemoRouteToService maps routes to services for OTel Demo
func mapOtelDemoRouteToService(endpoint *ServiceEndpoint) {
	route := endpoint.Route

	// Map based on route patterns
	routeMap := map[string]struct {
		service string
		port    string
	}{
		"/api/products":            {"product-catalog", "8080"},
		"/api/cart":                {"cart", "8080"},
		"/api/checkout":            {"checkout", "8080"},
		"/api/recommendations":     {"recommendation", "8080"},
		"/api/data":                {"frontend", "8080"},
		"/ship-order":              {"shipping", "8080"},
		"/get-quote":               {"shipping", "8080"},
		"/getquote":                {"quote", "8080"},
		"/send_order_confirmation": {"email", "8080"},
		"/ofrep/v1/evaluate":       {"flagd", "8016"},
		"/status":                  {"image-provider", "8080"},
	}

	for prefix, service := range routeMap {
		if strings.HasPrefix(route, prefix) {
			endpoint.ServerAddress = service.service
			endpoint.ServerPort = service.port
			return
		}
	}

	// Default to frontend-proxy only if address is empty (don't overwrite IP if no match)
	if endpoint.ServerAddress == "" {
		endpoint.ServerAddress = "frontend-proxy"
		endpoint.ServerPort = "8080"
	}
}

// mapOtelDemoGRPCToService maps gRPC services to server addresses for OTel Demo
func mapOtelDemoGRPCToService(operation *GRPCOperation) {
	rpcService := operation.RPCService

	// Map RPC service to actual service
	serviceMap := map[string]struct {
		service string
		port    string
	}{
		"oteldemo.AdService":             {"ad", "8080"},
		"oteldemo.CartService":           {"cart", "8080"},
		"oteldemo.CheckoutService":       {"checkout", "8080"},
		"oteldemo.CurrencyService":       {"currency", "8080"},
		"oteldemo.PaymentService":        {"payment", "8080"},
		"oteldemo.ProductCatalogService": {"product-catalog", "8080"},
		"oteldemo.RecommendationService": {"recommendation", "8080"},
		"flagd.evaluation.v1.Service":    {"flagd", "8013"},
	}

	if service, exists := serviceMap[rpcService]; exists {
		operation.ServerAddress = service.service
		operation.ServerPort = service.port
		return
	}

	// Do not clear existing address (e.g. IP) if no match found
}

// mapOtelDemoDatabaseToService maps database systems to server addresses for OTel Demo
func mapOtelDemoDatabaseToService(operation *DatabaseOperation) {
	switch operation.DBSystem {
	case "postgresql":
		operation.ServerAddress = "postgresql"
		operation.ServerPort = "5432"
	case "redis":
		operation.ServerAddress = "redis"
		operation.ServerPort = "6379"
	case "mysql":
		operation.ServerAddress = "mysql"
		operation.ServerPort = "3306"
	default:
		operation.ServerAddress = ""
		operation.ServerPort = ""
	}
}

// mapDeathStarBenchDatabaseToService maps database systems to server addresses for DeathStarBench systems
// DeathStarBench applications typically use MongoDB, memcached, and Redis
func mapDeathStarBenchDatabaseToService(operation *DatabaseOperation) {
	switch operation.DBSystem {
	case "mongodb":
		operation.ServerAddress = "mongodb"
		operation.ServerPort = "27017"
	case "memcached":
		operation.ServerAddress = "memcached"
		operation.ServerPort = "11211"
	case "redis":
		operation.ServerAddress = "redis"
		operation.ServerPort = "6379"
	case "mysql":
		operation.ServerAddress = "mysql"
		operation.ServerPort = "3306"
	case "postgresql":
		operation.ServerAddress = "postgresql"
		operation.ServerPort = "5432"
	default:
		// Keep the existing values if any, or leave empty
		if operation.ServerAddress == "" {
			operation.ServerAddress = operation.DBSystem
		}
	}
}

// mapOtelDemoClientToService maps client addresses (IPs) to service names for OTel Demo Server spans
func mapOtelDemoClientToService(endpoint *ServiceEndpoint) {
	// For Server spans, the ServerAddress contains the client IP
	// We try to map it to a known service name based on the route pattern
	// Since client IPs are dynamic in Kubernetes, we use route-based inference

	route := endpoint.Route

	// Known callers based on route patterns in OTel Demo
	// These are the services that call specific endpoints
	callerMap := map[string]string{
		"/":                        "load-generator",
		"/api/cart":                "load-generator",
		"/api/checkout":            "load-generator",
		"/api/products":            "load-generator",
		"/api/recommendations":     "load-generator",
		"/api/data":                "load-generator",
		"/getquote":                "shipping",
		"/get-quote":               "checkout",
		"/ship-order":              "checkout",
		"/send_order_confirmation": "checkout",
		"/status":                  "load-generator",
		"/ofrep/v1/evaluate":       "load-generator",
	}

	// Try to find a matching caller
	for prefix, caller := range callerMap {
		if strings.HasPrefix(route, prefix) {
			// Only set if we don't have a better value
			if endpoint.ServerAddress == "" || isIPAddress(endpoint.ServerAddress) {
				endpoint.ServerAddress = caller
			}
			return
		}
	}

	// For gRPC-style routes, infer the caller from route pattern
	if strings.HasPrefix(route, "/oteldemo.CartService/") {
		endpoint.ServerAddress = "frontend"
	} else if strings.HasPrefix(route, "/oteldemo.CheckoutService/") {
		endpoint.ServerAddress = "frontend"
	} else if strings.HasPrefix(route, "/oteldemo.ProductCatalogService/") {
		endpoint.ServerAddress = "frontend"
	} else if strings.HasPrefix(route, "/oteldemo.RecommendationService/") {
		endpoint.ServerAddress = "frontend"
	} else if strings.HasPrefix(route, "/oteldemo.AdService/") {
		endpoint.ServerAddress = "frontend"
	} else if strings.HasPrefix(route, "/oteldemo.CurrencyService/") {
		endpoint.ServerAddress = "checkout"
	} else if strings.HasPrefix(route, "/oteldemo.PaymentService/") {
		endpoint.ServerAddress = "checkout"
	} else if strings.HasPrefix(route, "/flagd.evaluation.v1.Service/") {
		// flagd is called by multiple services
		endpoint.ServerAddress = "multiple"
	}

	// If still an IP address or empty, mark as unknown
	if endpoint.ServerAddress == "" || isIPAddress(endpoint.ServerAddress) {
		endpoint.ServerAddress = "unknown-client"
	}
}

// mapDeathStarBenchRouteToService maps routes/span names to services for DeathStarBench systems
func mapDeathStarBenchRouteToService(endpoint *ServiceEndpoint, namespace string) {
	switch namespace {
	case "media":
		mapMediaMicroservicesRouteToService(endpoint)
	case "sn":
		mapSocialNetworkRouteToService(endpoint)
	case "hs":
		mapHotelReservationRouteToService(endpoint)
	case "ob":
		mapOnlineBoutiqueRouteToService(endpoint)
	}
}

// mapMediaMicroservicesRouteToService maps routes to services for Media Microservices
func mapMediaMicroservicesRouteToService(endpoint *ServiceEndpoint) {
	route := endpoint.Route
	spanName := endpoint.SpanName

	// Service mapping based on route patterns and span names
	serviceMap := map[string]struct {
		service string
		port    string
	}{
		// Cast info service
		"/wrk2-api/cast-info":            {"cast-info-service", "9090"},
		"/wrk2-api/movie/read-cast-info": {"cast-info-service", "9090"},
		"CastInfoHandler":                {"cast-info-service", "9090"},
		"WriteCastInfo":                  {"cast-info-service", "9090"},
		"ReadCastInfo":                   {"cast-info-service", "9090"},
		"/cast-info":                     {"cast-info-service", "9090"},
		// Compose review service
		"/wrk2-api/review/compose": {"compose-review-service", "9090"},
		"/wrk2-api/movie/register": {"compose-review-service", "9090"},
		"ComposeReview":            {"compose-review-service", "9090"},
		"UploadText":               {"compose-review-service", "9090"},
		"UploadRating":             {"compose-review-service", "9090"},
		"UploadMovieId":            {"compose-review-service", "9090"},
		"UploadUniqueId":           {"compose-review-service", "9090"},
		"UploadUserId":             {"compose-review-service", "9090"},
		"/compose":                 {"compose-review-service", "9090"},
		"/register":                {"compose-review-service", "9090"},
		// Movie ID service
		"RegisterMovieId": {"movie-id-service", "9090"},
		"MovieIdHandler":  {"movie-id-service", "9090"},
		"/movie-id":       {"movie-id-service", "9090"},
		// Movie info service
		"/wrk2-api/movie-info":      {"movie-info-service", "9090"},
		"/wrk2-api/movie/read-info": {"movie-info-service", "9090"},
		"MovieInfoHandler":          {"movie-info-service", "9090"},
		"WriteMovieInfo":            {"movie-info-service", "9090"},
		"ReadMovieInfo":             {"movie-info-service", "9090"},
		"/movie-info":               {"movie-info-service", "9090"},
		"/read-info":                {"movie-info-service", "9090"},
		// Movie review service
		"StoreReview":      {"movie-review-service", "9090"},
		"ReadMovieReviews": {"movie-review-service", "9090"},
		"/movie-review":    {"movie-review-service", "9090"},
		"/review":          {"movie-review-service", "9090"},
		// Page service
		"/wrk2-api/page":            {"page-service", "9090"},
		"/wrk2-api/movie/read-page": {"page-service", "9090"},
		"ReadPage":                  {"page-service", "9090"},
		"/read-page":                {"page-service", "9090"},
		// Plot service
		"/wrk2-api/plot":            {"plot-service", "9090"},
		"/wrk2-api/movie/read-plot": {"plot-service", "9090"},
		"PlotHandler":               {"plot-service", "9090"},
		"WritePlot":                 {"plot-service", "9090"},
		"ReadPlot":                  {"plot-service", "9090"},
		"/plot":                     {"plot-service", "9090"},
		"/read-plot":                {"plot-service", "9090"},
		// Rating service
		"StoreRating": {"rating-service", "9090"},
		"ReadRatings": {"rating-service", "9090"},
		"/rating":     {"rating-service", "9090"},
		// Review storage service
		"StoreReviewStorage": {"review-storage-service", "9090"},
		"ReadReviews":        {"review-storage-service", "9090"},
		"/review-storage":    {"review-storage-service", "9090"},
		// Text service
		"TextHandler": {"text-service", "9090"},
		"StoreText":   {"text-service", "9090"},
		"/text":       {"text-service", "9090"},
		// Unique ID service
		"UniqueIdHandler": {"unique-id-service", "9090"},
		"ComposeUniqueId": {"unique-id-service", "9090"},
		"/unique-id":      {"unique-id-service", "9090"},
		// User service
		"/wrk2-api/user": {"user-service", "9090"},
		"UserHandler":    {"user-service", "9090"},
		"RegisterUser":   {"user-service", "9090"},
		"Login":          {"user-service", "9090"},
		"/user":          {"user-service", "9090"},
		// User review service
		"ReadUserReviews": {"user-review-service", "9090"},
		"StoreUserReview": {"user-review-service", "9090"},
		"/user-review":    {"user-review-service", "9090"},
		// Frontend - removed overly broad "/" pattern
		"/wrk2-api/home": {"nginx-web-server", "8080"},
	}

	// Sort patterns by length (longest first) to ensure more specific patterns match first
	patterns := make([]string, 0, len(serviceMap))
	for pattern := range serviceMap {
		patterns = append(patterns, pattern)
	}
	sort.Slice(patterns, func(i, j int) bool {
		return len(patterns[i]) > len(patterns[j])
	})

	// Check route first with sorted patterns
	for _, pattern := range patterns {
		service := serviceMap[pattern]
		if strings.Contains(route, pattern) || strings.Contains(spanName, pattern) {
			endpoint.ServerAddress = service.service
			endpoint.ServerPort = service.port
			return
		}
	}

	// Default to nginx-web-server if no match
	if endpoint.ServerAddress == "" || isIPAddress(endpoint.ServerAddress) {
		endpoint.ServerAddress = "nginx-web-server"
		endpoint.ServerPort = "8080"
	}
}

// mapSocialNetworkRouteToService maps routes to services for Social Network
func mapSocialNetworkRouteToService(endpoint *ServiceEndpoint) {
	route := endpoint.Route
	spanName := endpoint.SpanName

	// Service mapping based on route patterns and span names
	serviceMap := map[string]struct {
		service string
		port    string
	}{
		// Compose post service
		"/wrk2-api/post/compose": {"compose-post-service", "9090"},
		"/wrk2-api/post":         {"compose-post-service", "9090"},
		"ComposePost":            {"compose-post-service", "9090"},
		"UploadText":             {"compose-post-service", "9090"},
		"UploadMedia":            {"compose-post-service", "9090"},
		"UploadUniqueId":         {"compose-post-service", "9090"},
		"UploadCreator":          {"compose-post-service", "9090"},
		"UploadUrls":             {"compose-post-service", "9090"},
		"UploadUserMentions":     {"compose-post-service", "9090"},
		"/compose":               {"compose-post-service", "9090"},
		// Home timeline service
		"/wrk2-api/home-timeline": {"home-timeline-service", "9090"},
		"ReadHomeTimeline":        {"home-timeline-service", "9090"},
		"WriteHomeTimeline":       {"home-timeline-service", "9090"},
		"/home-timeline":          {"home-timeline-service", "9090"},
		// Media service
		"/wrk2-api/media":    {"media-service", "9090"},
		"MediaHandler":       {"media-service", "9090"},
		"UploadMediaHandler": {"media-service", "9090"},
		"StoreMedia":         {"media-service", "9090"},
		"/media":             {"media-service", "9090"},
		// Post storage service
		"StorePost":     {"post-storage-service", "9090"},
		"ReadPost":      {"post-storage-service", "9090"},
		"ReadPosts":     {"post-storage-service", "9090"},
		"/post-storage": {"post-storage-service", "9090"},
		// Social graph service
		"/wrk2-api/user/follow":   {"social-graph-service", "9090"},
		"/wrk2-api/user/unfollow": {"social-graph-service", "9090"},
		"Follow":                  {"social-graph-service", "9090"},
		"Unfollow":                {"social-graph-service", "9090"},
		"GetFollowers":            {"social-graph-service", "9090"},
		"GetFollowees":            {"social-graph-service", "9090"},
		"InsertUser":              {"social-graph-service", "9090"},
		"FollowWithUsername":      {"social-graph-service", "9090"},
		"UnfollowWithUsername":    {"social-graph-service", "9090"},
		"/social-graph":           {"social-graph-service", "9090"},
		// Text service
		"TextHandler": {"text-service", "9090"},
		"ProcessText": {"text-service", "9090"},
		"/text":       {"text-service", "9090"},
		// Unique ID service
		"UniqueIdHandler": {"unique-id-service", "9090"},
		"ComposeUniqueId": {"unique-id-service", "9090"},
		"/unique-id":      {"unique-id-service", "9090"},
		// URL shorten service
		"/wrk2-api/shorten-urls": {"url-shorten-service", "9090"},
		"UrlHandler":             {"url-shorten-service", "9090"},
		"ShortenUrls":            {"url-shorten-service", "9090"},
		"GetExtendedUrls":        {"url-shorten-service", "9090"},
		"/url-shorten":           {"url-shorten-service", "9090"},
		"/shorten":               {"url-shorten-service", "9090"},
		// User mention service
		"UserMentionHandler":  {"user-mention-service", "9090"},
		"ComposeUserMentions": {"user-mention-service", "9090"},
		"/user-mention":       {"user-mention-service", "9090"},
		// User service
		"/wrk2-api/user/register": {"user-service", "9090"},
		"/wrk2-api/user/login":    {"user-service", "9090"},
		"RegisterUser":            {"user-service", "9090"},
		"RegisterUserWithId":      {"user-service", "9090"},
		"Login":                   {"user-service", "9090"},
		"GetUserId":               {"user-service", "9090"},
		"/user":                   {"user-service", "9090"},
		"/register":               {"user-service", "9090"},
		"/login":                  {"user-service", "9090"},
		// User timeline service
		"/wrk2-api/user-timeline": {"user-timeline-service", "9090"},
		"ReadUserTimeline":        {"user-timeline-service", "9090"},
		"WriteUserTimeline":       {"user-timeline-service", "9090"},
		"/user-timeline":          {"user-timeline-service", "9090"},
		// Media frontend
		"/wrk2-api/media-frontend": {"media-frontend", "8081"},
		// Frontend - removed overly broad "/" pattern
		"/wrk2-api/home": {"nginx-thrift", "8080"},
	}

	// Sort patterns by length (longest first) to ensure more specific patterns match first
	patterns := make([]string, 0, len(serviceMap))
	for pattern := range serviceMap {
		patterns = append(patterns, pattern)
	}
	sort.Slice(patterns, func(i, j int) bool {
		return len(patterns[i]) > len(patterns[j])
	})

	// Check route first with sorted patterns
	for _, pattern := range patterns {
		service := serviceMap[pattern]
		if strings.Contains(route, pattern) || strings.Contains(spanName, pattern) {
			endpoint.ServerAddress = service.service
			endpoint.ServerPort = service.port
			return
		}
	}

	// Default to nginx-thrift if no match
	if endpoint.ServerAddress == "" || isIPAddress(endpoint.ServerAddress) {
		endpoint.ServerAddress = "nginx-thrift"
		endpoint.ServerPort = "8080"
	}
}

// mapHotelReservationRouteToService maps routes to services for Hotel Reservation
func mapHotelReservationRouteToService(endpoint *ServiceEndpoint) {
	route := endpoint.Route
	spanName := endpoint.SpanName

	// Service mapping based on route patterns and span names
	serviceMap := map[string]struct {
		service string
		port    string
	}{
		// Attractions service
		"/attractions":            {"attractions", "8089"},
		"GetAttractions":          {"attractions", "8089"},
		"attractions.Attractions": {"attractions", "8089"},
		// Frontend service - removed overly broad "/" pattern
		"/hotels":           {"frontend", "5000"},
		"/recommendations":  {"frontend", "5000"},
		"/user":             {"frontend", "5000"},
		"/reservation":      {"frontend", "5000"},
		"frontend.Frontend": {"frontend", "5000"},
		// Geo service
		"/geo":      {"geo", "8083"},
		"NearbyGeo": {"geo", "8083"},
		"GetGeo":    {"geo", "8083"},
		"geo.Geo":   {"geo", "8083"},
		// Profile service
		"/profile":        {"profile", "8081"},
		"GetProfiles":     {"profile", "8081"},
		"GetProfile":      {"profile", "8081"},
		"profile.Profile": {"profile", "8081"},
		// Rate service
		"/rate":     {"rate", "8084"},
		"GetRates":  {"rate", "8084"},
		"GetRate":   {"rate", "8084"},
		"rate.Rate": {"rate", "8084"},
		// Recommendation service
		"/recommendation":               {"recommendation", "8085"},
		"GetRecommendations":            {"recommendation", "8085"},
		"recommendation.Recommendation": {"recommendation", "8085"},
		// Reservation service
		"/reserve":                {"reservation", "8087"},
		"MakeReservation":         {"reservation", "8087"},
		"CheckAvailability":       {"reservation", "8087"},
		"reservation.Reservation": {"reservation", "8087"},
		// Search service
		"/search":       {"search", "8082"},
		"NearbySearch":  {"search", "8082"},
		"search.Search": {"search", "8082"},
		// User service
		"/login":    {"user", "8086"},
		"/register": {"user", "8086"},
		"Login":     {"user", "8086"},
		"Register":  {"user", "8086"},
		"CheckUser": {"user", "8086"},
		"user.User": {"user", "8086"},
	}

	// Sort patterns by length (longest first) to ensure more specific patterns match first
	patterns := make([]string, 0, len(serviceMap))
	for pattern := range serviceMap {
		patterns = append(patterns, pattern)
	}
	sort.Slice(patterns, func(i, j int) bool {
		return len(patterns[i]) > len(patterns[j])
	})

	// Check route first with sorted patterns
	for _, pattern := range patterns {
		service := serviceMap[pattern]
		if strings.Contains(route, pattern) || strings.Contains(spanName, pattern) {
			endpoint.ServerAddress = service.service
			endpoint.ServerPort = service.port
			return
		}
	}

	// Default to frontend if no match
	if endpoint.ServerAddress == "" || isIPAddress(endpoint.ServerAddress) {
		endpoint.ServerAddress = "frontend"
		endpoint.ServerPort = "5000"
	}
}

// mapOnlineBoutiqueRouteToService maps routes to services for OnlineBoutique
func mapOnlineBoutiqueRouteToService(endpoint *ServiceEndpoint) {
	route := endpoint.Route
	spanName := endpoint.SpanName

	// Service mapping based on route patterns and span names
	serviceMap := map[string]struct {
		service string
		port    string
	}{
		// Frontend service
		"/":         {"frontend", "80"},
		"/product":  {"frontend", "80"},
		"/cart":     {"frontend", "80"},
		"/checkout": {"frontend", "80"},
		"frontend":  {"frontend", "80"},
		// Ad service
		"/hipstershop.AdService": {"adservice", "9555"},
		"AdService":              {"adservice", "9555"},
		"GetAds":                 {"adservice", "9555"},
		// Cart service
		"/hipstershop.CartService": {"cartservice", "7070"},
		"CartService":              {"cartservice", "7070"},
		"AddItem":                  {"cartservice", "7070"},
		"GetCart":                  {"cartservice", "7070"},
		"EmptyCart":                {"cartservice", "7070"},
		// Checkout service
		"/hipstershop.CheckoutService": {"checkoutservice", "5050"},
		"CheckoutService":              {"checkoutservice", "5050"},
		"PlaceOrder":                   {"checkoutservice", "5050"},
		// Currency service
		"/hipstershop.CurrencyService": {"currencyservice", "7000"},
		"CurrencyService":              {"currencyservice", "7000"},
		"GetSupportedCurrencies":       {"currencyservice", "7000"},
		"Convert":                      {"currencyservice", "7000"},
		// Email service
		"/hipstershop.EmailService": {"emailservice", "5000"},
		"EmailService":              {"emailservice", "5000"},
		"SendOrderConfirmation":     {"emailservice", "5000"},
		// Payment service
		"/hipstershop.PaymentService": {"paymentservice", "50051"},
		"PaymentService":              {"paymentservice", "50051"},
		"Charge":                      {"paymentservice", "50051"},
		// Product catalog service
		"/hipstershop.ProductCatalogService": {"productcatalogservice", "3550"},
		"ProductCatalogService":              {"productcatalogservice", "3550"},
		"ListProducts":                       {"productcatalogservice", "3550"},
		"GetProduct":                         {"productcatalogservice", "3550"},
		"SearchProducts":                     {"productcatalogservice", "3550"},
		// Recommendation service
		"/hipstershop.RecommendationService": {"recommendationservice", "8080"},
		"RecommendationService":              {"recommendationservice", "8080"},
		"ListRecommendations":                {"recommendationservice", "8080"},
		// Shipping service
		"/hipstershop.ShippingService": {"shippingservice", "50051"},
		"ShippingService":              {"shippingservice", "50051"},
		"GetQuote":                     {"shippingservice", "50051"},
		"ShipOrder":                    {"shippingservice", "50051"},
	}

	// Sort patterns by length (longest first) to match more specific patterns first
	patterns := make([]string, 0, len(serviceMap))
	for pattern := range serviceMap {
		patterns = append(patterns, pattern)
	}
	sort.Slice(patterns, func(i, j int) bool {
		return len(patterns[i]) > len(patterns[j])
	})

	// Check route and span name with sorted patterns
	for _, pattern := range patterns {
		service := serviceMap[pattern]
		if strings.Contains(route, pattern) || strings.Contains(spanName, pattern) {
			endpoint.ServerAddress = service.service
			endpoint.ServerPort = service.port
			return
		}
	}

	// Default to frontend if no match
	if endpoint.ServerAddress == "" || isIPAddress(endpoint.ServerAddress) {
		endpoint.ServerAddress = "frontend"
		endpoint.ServerPort = "80"
	}
}

// mapDeathStarBenchGRPCToService maps gRPC services to server addresses for DeathStarBench systems
func mapDeathStarBenchGRPCToService(operation *GRPCOperation, namespace string) {
	rpcService := operation.RPCService

	switch namespace {
	case "media":
		mapMediaMicroservicesGRPCToService(operation, rpcService)
	case "sn":
		mapSocialNetworkGRPCToService(operation, rpcService)
	case "hs":
		mapHotelReservationGRPCToService(operation, rpcService)
	case "ob":
		mapOnlineBoutiqueGRPCToService(operation, rpcService)
	}
}

// mapMediaMicroservicesGRPCToService maps gRPC services to server addresses for Media Microservices
func mapMediaMicroservicesGRPCToService(operation *GRPCOperation, rpcService string) {
	// Map based on RPC service name patterns
	serviceMap := map[string]struct {
		service string
		port    string
	}{
		"CastInfoService":      {"cast-info-service", "9090"},
		"ComposeReviewService": {"compose-review-service", "9090"},
		"MovieIdService":       {"movie-id-service", "9090"},
		"MovieInfoService":     {"movie-info-service", "9090"},
		"MovieReviewService":   {"movie-review-service", "9090"},
		"PageService":          {"page-service", "9090"},
		"PlotService":          {"plot-service", "9090"},
		"RatingService":        {"rating-service", "9090"},
		"ReviewStorageService": {"review-storage-service", "9090"},
		"TextService":          {"text-service", "9090"},
		"UniqueIdService":      {"unique-id-service", "9090"},
		"UserService":          {"user-service", "9090"},
		"UserReviewService":    {"user-review-service", "9090"},
	}

	for pattern, service := range serviceMap {
		if strings.Contains(rpcService, pattern) {
			operation.ServerAddress = service.service
			operation.ServerPort = service.port
			return
		}
	}
}

// mapSocialNetworkGRPCToService maps gRPC services to server addresses for Social Network
func mapSocialNetworkGRPCToService(operation *GRPCOperation, rpcService string) {
	// Map based on RPC service name patterns
	serviceMap := map[string]struct {
		service string
		port    string
	}{
		"ComposePostService":  {"compose-post-service", "9090"},
		"HomeTimelineService": {"home-timeline-service", "9090"},
		"MediaService":        {"media-service", "9090"},
		"PostStorageService":  {"post-storage-service", "9090"},
		"SocialGraphService":  {"social-graph-service", "9090"},
		"TextService":         {"text-service", "9090"},
		"UniqueIdService":     {"unique-id-service", "9090"},
		"UrlShortenService":   {"url-shorten-service", "9090"},
		"UserMentionService":  {"user-mention-service", "9090"},
		"UserService":         {"user-service", "9090"},
		"UserTimelineService": {"user-timeline-service", "9090"},
	}

	for pattern, service := range serviceMap {
		if strings.Contains(rpcService, pattern) {
			operation.ServerAddress = service.service
			operation.ServerPort = service.port
			return
		}
	}
}

// mapHotelReservationGRPCToService maps gRPC services to server addresses for Hotel Reservation
func mapHotelReservationGRPCToService(operation *GRPCOperation, rpcService string) {
	// Map based on RPC service name patterns
	serviceMap := map[string]struct {
		service string
		port    string
	}{
		// Standard patterns
		"AttractionsService":    {"attractions", "8089"},
		"FrontendService":       {"frontend", "5000"},
		"GeoService":            {"geo", "8083"},
		"ProfileService":        {"profile", "8081"},
		"RateService":           {"rate", "8084"},
		"RecommendationService": {"recommendation", "8085"},
		"ReservationService":    {"reservation", "8087"},
		"SearchService":         {"search", "8082"},
		"UserService":           {"user", "8086"},
		// DeathStarBench hotelReservation gRPC service patterns
		"attractions.Attractions":       {"attractions", "8089"},
		"frontend.Frontend":             {"frontend", "5000"},
		"geo.Geo":                       {"geo", "8083"},
		"profile.Profile":               {"profile", "8081"},
		"rate.Rate":                     {"rate", "8084"},
		"recommendation.Recommendation": {"recommendation", "8085"},
		"reservation.Reservation":       {"reservation", "8087"},
		"search.Search":                 {"search", "8082"},
		"user.User":                     {"user", "8086"},
	}

	for pattern, service := range serviceMap {
		if strings.Contains(rpcService, pattern) {
			operation.ServerAddress = service.service
			operation.ServerPort = service.port
			return
		}
	}
}

// mapOnlineBoutiqueGRPCToService maps gRPC services to server addresses for OnlineBoutique
func mapOnlineBoutiqueGRPCToService(operation *GRPCOperation, rpcService string) {
	// Map based on RPC service name patterns
	serviceMap := map[string]struct {
		service string
		port    string
	}{
		"hipstershop.AdService":             {"adservice", "9555"},
		"AdService":                         {"adservice", "9555"},
		"hipstershop.CartService":           {"cartservice", "7070"},
		"CartService":                       {"cartservice", "7070"},
		"hipstershop.CheckoutService":       {"checkoutservice", "5050"},
		"CheckoutService":                   {"checkoutservice", "5050"},
		"hipstershop.CurrencyService":       {"currencyservice", "7000"},
		"CurrencyService":                   {"currencyservice", "7000"},
		"hipstershop.EmailService":          {"emailservice", "5000"},
		"EmailService":                      {"emailservice", "5000"},
		"hipstershop.PaymentService":        {"paymentservice", "50051"},
		"PaymentService":                    {"paymentservice", "50051"},
		"hipstershop.ProductCatalogService": {"productcatalogservice", "3550"},
		"ProductCatalogService":             {"productcatalogservice", "3550"},
		"hipstershop.RecommendationService": {"recommendationservice", "8080"},
		"RecommendationService":             {"recommendationservice", "8080"},
		"hipstershop.ShippingService":       {"shippingservice", "50051"},
		"ShippingService":                   {"shippingservice", "50051"},
	}

	for pattern, service := range serviceMap {
		if strings.Contains(rpcService, pattern) {
			operation.ServerAddress = service.service
			operation.ServerPort = service.port
			return
		}
	}
}

// isIPAddress checks if a string looks like an IP address
func isIPAddress(s string) bool {
	// Simple check: if it starts with a digit and contains dots, it's likely an IP
	if len(s) == 0 {
		return false
	}
	if s[0] >= '0' && s[0] <= '9' && strings.Contains(s, ".") {
		return true
	}
	return false
}
