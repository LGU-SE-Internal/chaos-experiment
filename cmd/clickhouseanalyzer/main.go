package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"
	"github.com/LGU-SE-Internal/chaos-experiment/tools/clickhouseanalyzer"
)

func main() {
	// Define command-line flags
	host := flag.String("host", "10.10.10.58", "ClickHouse server host")
	port := flag.Int("port", 9000, "ClickHouse server port")
	database := flag.String("database", "default", "ClickHouse database name")
	username := flag.String("username", "default", "ClickHouse username")
	password := flag.String("password", "password", "ClickHouse password")
	system := flag.String("system", "ts", "Target system: 'ts' (TrainTicket), 'otel-demo' (OpenTelemetry Demo), 'media' (mediaMicroservices), 'hs' (hotelReservation), 'sn' (socialNetwork), 'ob' (OnlineBoutique), 'sockshop' (Sock Shop), or 'teastore' (Tea Store)")
	outputEndpoints := flag.String("output", "", "Path for the generated endpoints Go file")
	outputDatabase := flag.String("output-db", "", "Path for the generated database operations Go file")
	outputGRPC := flag.String("output-grpc", "", "Path for the generated gRPC operations Go file (otel-demo only)")
	skipView := flag.Bool("skip-view", false, "Skip creating the materialized view")
	flag.Parse()

	// Validate and set the system type using systemconfig
	systemType, err := systemconfig.ParseSystemType(*system)
	if err != nil {
		fmt.Printf("Invalid system: %s. Must be 'ts', 'otel-demo', 'media', 'hs', 'sn', 'ob', 'sockshop', or 'teastore'\n", *system)
		os.Exit(1)
	}
	if err := systemconfig.SetCurrentSystem(systemType); err != nil {
		fmt.Printf("Error setting system type: %v\n", err)
		os.Exit(1)
	}

	// Set default output paths if not specified
	// Each system has its own directory to allow coexistence
	projectRoot, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error determining project root: %v\n", err)
		os.Exit(1)
	}

	// Determine system-specific subdirectory
	var systemDir string
	switch systemType {
	case systemconfig.SystemTrainTicket:
		systemDir = "ts"
	case systemconfig.SystemOtelDemo:
		systemDir = "oteldemo"
	case systemconfig.SystemMediaMicroservices:
		systemDir = "media"
	case systemconfig.SystemHotelReservation:
		systemDir = "hs"
	case systemconfig.SystemSocialNetwork:
		systemDir = "sn"
	case systemconfig.SystemOnlineBoutique:
		systemDir = "ob"
	case systemconfig.SystemSockShop:
		systemDir = "sockshop"
	case systemconfig.SystemTeaStore:
		systemDir = "teastore"
	default:
		systemDir = string(systemType)
	}

	if *outputEndpoints == "" {
		*outputEndpoints = filepath.Join(projectRoot, "internal", systemDir, "serviceendpoints", "serviceendpoints.go")
	}

	if *outputDatabase == "" {
		*outputDatabase = filepath.Join(projectRoot, "internal", systemDir, "databaseoperations", "databaseoperations.go")
	}

	if *outputGRPC == "" {
		*outputGRPC = filepath.Join(projectRoot, "internal", systemDir, "grpcoperations", "grpcoperations.go")
	}

	// Configure ClickHouse connection
	config := clickhouseanalyzer.ClickHouseConfig{
		Host:     *host,
		Port:     *port,
		Database: *database,
		Username: *username,
		Password: *password,
	}

	// Connect to ClickHouse
	fmt.Println("Connecting to ClickHouse...")
	db, err := clickhouseanalyzer.ConnectToDB(config)
	if err != nil {
		fmt.Printf("Error connecting to ClickHouse: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Printf("Analyzing system: %s\n", systemconfig.GetCurrentSystem())

	switch {
	case systemconfig.IsTrainTicket():
		runTrainTicketAnalysis(db, *outputEndpoints, *outputDatabase, *skipView)
	case systemconfig.IsOtelDemo():
		runOtelDemoAnalysis(db, *outputEndpoints, *outputDatabase, *outputGRPC, *skipView)
	case systemconfig.IsMediaMicroservices():
		runDeathStarBenchAnalysis(db, "media", "media_traces_mv", *outputEndpoints, *outputDatabase, *outputGRPC, *skipView)
	case systemconfig.IsHotelReservation():
		runDeathStarBenchAnalysis(db, "hs", "hs_traces_mv", *outputEndpoints, *outputDatabase, *outputGRPC, *skipView)
	case systemconfig.IsSocialNetwork():
		runDeathStarBenchAnalysis(db, "sn", "sn_traces_mv", *outputEndpoints, *outputDatabase, *outputGRPC, *skipView)
	case systemconfig.IsOnlineBoutique():
		runOnlineBoutiqueAnalysis(db, "ob", "ob_traces_mv", *outputEndpoints, *outputDatabase, *outputGRPC, *skipView)
	case systemconfig.IsSockShop():
		runDeathStarBenchAnalysis(db, "sockshop", "sockshop_traces_mv", *outputEndpoints, *outputDatabase, *outputGRPC, *skipView)
	case systemconfig.IsTeaStore():
		runTeaStoreAnalysis(db, "teastore", "teastore_traces_mv", *outputEndpoints, *outputDatabase, *outputGRPC, *skipView)
	}
}

func runTrainTicketAnalysis(db *sql.DB, outputEndpoints, outputDatabase string, skipView bool) {
	// Create materialized view if needed
	if !skipView {
		fmt.Println("Creating materialized view for TrainTicket...")
		if err := clickhouseanalyzer.CreateMaterializedView(db); err != nil {
			fmt.Printf("Error creating materialized view: %v\n", err)
			os.Exit(1)
		}
	}

	// Query client traces
	fmt.Println("Querying client traces...")
	clientEndpoints, err := clickhouseanalyzer.QueryClientTraces(db)
	if err != nil {
		fmt.Printf("Error querying client traces: %v\n", err)
		os.Exit(1)
	}

	// Query dashboard routes
	fmt.Println("Querying dashboard routes...")
	dashboardEndpoints, err := clickhouseanalyzer.QueryDashboardRoutes(db)
	if err != nil {
		fmt.Printf("Error querying dashboard routes: %v\n", err)
		os.Exit(1)
	}

	// Query MySQL operations
	fmt.Println("Querying MySQL operations...")
	dbOperations, err := clickhouseanalyzer.QueryMySQLOperations(db)
	if err != nil {
		fmt.Printf("Error querying MySQL operations: %v\n", err)
		os.Exit(1)
	}

	// Combine HTTP endpoints only (no conversion from DB operations)
	allEndpoints := append(clientEndpoints, dashboardEndpoints...)

	// Generate service endpoints file - HTTP endpoints only
	fmt.Printf("Generating HTTP endpoints file at %s...\n", outputEndpoints)
	if err := clickhouseanalyzer.GenerateServiceEndpointsFile(allEndpoints, outputEndpoints); err != nil {
		fmt.Printf("Error generating service endpoints file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("HTTP endpoints file generated successfully!")

	// Generate database operations file
	fmt.Printf("Generating database operations file at %s...\n", outputDatabase)
	if err := clickhouseanalyzer.GenerateDatabaseOperationsFile(dbOperations, outputDatabase); err != nil {
		fmt.Printf("Error generating database operations file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Database operations file generated successfully!")
}

func runOtelDemoAnalysis(db *sql.DB, outputEndpoints, outputDatabase, outputGRPC string, skipView bool) {
	// Create materialized view if needed
	if !skipView {
		fmt.Println("Creating materialized view for OTel Demo...")
		if err := clickhouseanalyzer.CreateOtelDemoMaterializedView(db); err != nil {
			fmt.Printf("Error creating materialized view: %v\n", err)
			os.Exit(1)
		}
	}

	// Query HTTP client traces
	fmt.Println("Querying HTTP client traces...")
	clientEndpoints, err := clickhouseanalyzer.QueryOtelDemoHTTPClientTraces(db)
	if err != nil {
		fmt.Printf("Error querying HTTP client traces: %v\n", err)
		os.Exit(1)
	}

	// Query HTTP server traces
	fmt.Println("Querying HTTP server traces...")
	serverEndpoints, err := clickhouseanalyzer.QueryOtelDemoHTTPServerTraces(db)
	if err != nil {
		fmt.Printf("Error querying HTTP server traces: %v\n", err)
		os.Exit(1)
	}

	// Query gRPC operations
	fmt.Println("Querying gRPC operations...")
	grpcOperations, err := clickhouseanalyzer.QueryOtelDemoGRPCOperations(db)
	if err != nil {
		fmt.Printf("Error querying gRPC operations: %v\n", err)
		os.Exit(1)
	}

	// Query database operations
	fmt.Println("Querying database operations...")
	dbOperations, err := clickhouseanalyzer.QueryOtelDemoDatabaseOperations(db)
	if err != nil {
		fmt.Printf("Error querying database operations: %v\n", err)
		os.Exit(1)
	}

	// Combine HTTP endpoints only (no conversion from DB/gRPC operations)
	allEndpoints := append(clientEndpoints, serverEndpoints...)

	// Generate HTTP endpoints file - HTTP endpoints only
	fmt.Printf("Generating HTTP endpoints file at %s...\n", outputEndpoints)
	if err := clickhouseanalyzer.GenerateServiceEndpointsFile(allEndpoints, outputEndpoints); err != nil {
		fmt.Printf("Error generating service endpoints file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("HTTP endpoints file generated successfully!")

	// Generate database operations file
	fmt.Printf("Generating database operations file at %s...\n", outputDatabase)
	if err := clickhouseanalyzer.GenerateDatabaseOperationsFile(dbOperations, outputDatabase); err != nil {
		fmt.Printf("Error generating database operations file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Database operations file generated successfully!")

	// Generate gRPC operations file
	fmt.Printf("Generating gRPC operations file at %s...\n", outputGRPC)
	if err := clickhouseanalyzer.GenerateGRPCOperationsFile(grpcOperations, outputGRPC); err != nil {
		fmt.Printf("Error generating gRPC operations file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("gRPC operations file generated successfully!")
}

func runDeathStarBenchAnalysis(db *sql.DB, namespace, viewName, outputEndpoints, outputDatabase, outputGRPC string, skipView bool) {
	// Create materialized view if needed
	if !skipView {
		fmt.Printf("Creating materialized view for %s (namespace: %s)...\n", viewName, namespace)
		if err := clickhouseanalyzer.CreateDeathStarBenchMaterializedView(db, namespace, viewName); err != nil {
			fmt.Printf("Error creating materialized view: %v\n", err)
			os.Exit(1)
		}
	}

	// Query HTTP client traces
	fmt.Println("Querying HTTP client traces...")
	clientEndpoints, err := clickhouseanalyzer.QueryDeathStarBenchHTTPClientTraces(db, viewName, namespace)
	if err != nil {
		fmt.Printf("Error querying HTTP client traces: %v\n", err)
		os.Exit(1)
	}

	// Query HTTP server traces
	fmt.Println("Querying HTTP server traces...")
	serverEndpoints, err := clickhouseanalyzer.QueryDeathStarBenchHTTPServerTraces(db, viewName, namespace)
	if err != nil {
		fmt.Printf("Error querying HTTP server traces: %v\n", err)
		os.Exit(1)
	}

	// Query gRPC operations
	fmt.Println("Querying gRPC operations...")
	grpcOperations, err := clickhouseanalyzer.QueryDeathStarBenchGRPCOperations(db, viewName, namespace)
	if err != nil {
		fmt.Printf("Error querying gRPC operations: %v\n", err)
		os.Exit(1)
	}

	// Query database operations
	fmt.Println("Querying database operations...")
	dbOperations, err := clickhouseanalyzer.QueryDeathStarBenchDatabaseOperations(db, viewName)
	if err != nil {
		fmt.Printf("Error querying database operations: %v\n", err)
		os.Exit(1)
	}

	// Combine HTTP endpoints only (no conversion from DB/gRPC operations)
	allEndpoints := append(clientEndpoints, serverEndpoints...)

	// Generate HTTP endpoints file - HTTP endpoints only
	fmt.Printf("Generating HTTP endpoints file at %s...\n", outputEndpoints)
	if err := clickhouseanalyzer.GenerateServiceEndpointsFile(allEndpoints, outputEndpoints); err != nil {
		fmt.Printf("Error generating service endpoints file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("HTTP endpoints file generated successfully!")

	// Generate database operations file
	fmt.Printf("Generating database operations file at %s...\n", outputDatabase)
	if err := clickhouseanalyzer.GenerateDatabaseOperationsFile(dbOperations, outputDatabase); err != nil {
		fmt.Printf("Error generating database operations file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Database operations file generated successfully!")

	// Generate gRPC operations file
	fmt.Printf("Generating gRPC operations file at %s...\n", outputGRPC)
	if err := clickhouseanalyzer.GenerateGRPCOperationsFile(grpcOperations, outputGRPC); err != nil {
		fmt.Printf("Error generating gRPC operations file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("gRPC operations file generated successfully!")
}

func runOnlineBoutiqueAnalysis(db *sql.DB, namespace, viewName, outputEndpoints, outputDatabase, outputGRPC string, skipView bool) {
// Create materialized view if needed
if !skipView {
fmt.Printf("Creating materialized view for %s (namespace: %s)...\n", viewName, namespace)
if err := clickhouseanalyzer.CreateOnlineBoutiqueMaterializedView(db, namespace, viewName); err != nil {
fmt.Printf("Error creating materialized view: %v\n", err)
os.Exit(1)
}
}

// Query HTTP client traces
fmt.Println("Querying HTTP client traces...")
clientEndpoints, err := clickhouseanalyzer.QueryDeathStarBenchHTTPClientTraces(db, viewName, namespace)
if err != nil {
fmt.Printf("Error querying HTTP client traces: %v\n", err)
os.Exit(1)
}

// Query HTTP server traces
fmt.Println("Querying HTTP server traces...")
serverEndpoints, err := clickhouseanalyzer.QueryDeathStarBenchHTTPServerTraces(db, viewName, namespace)
if err != nil {
fmt.Printf("Error querying HTTP server traces: %v\n", err)
os.Exit(1)
}

// Query gRPC operations
fmt.Println("Querying gRPC operations...")
grpcOperations, err := clickhouseanalyzer.QueryDeathStarBenchGRPCOperations(db, viewName, namespace)
if err != nil {
fmt.Printf("Error querying gRPC operations: %v\n", err)
os.Exit(1)
}

// Query database operations
fmt.Println("Querying database operations...")
dbOperations, err := clickhouseanalyzer.QueryDeathStarBenchDatabaseOperations(db, viewName)
if err != nil {
fmt.Printf("Error querying database operations: %v\n", err)
os.Exit(1)
}

// Combine HTTP endpoints only (no conversion from DB/gRPC operations)
allEndpoints := append(clientEndpoints, serverEndpoints...)

// Generate HTTP endpoints file - HTTP endpoints only
fmt.Printf("Generating HTTP endpoints file at %s...\n", outputEndpoints)
if err := clickhouseanalyzer.GenerateServiceEndpointsFile(allEndpoints, outputEndpoints); err != nil {
fmt.Printf("Error generating service endpoints file: %v\n", err)
os.Exit(1)
}
fmt.Println("HTTP endpoints file generated successfully!")

// Generate database operations file
fmt.Printf("Generating database operations file at %s...\n", outputDatabase)
if err := clickhouseanalyzer.GenerateDatabaseOperationsFile(dbOperations, outputDatabase); err != nil {
fmt.Printf("Error generating database operations file: %v\n", err)
os.Exit(1)
}
fmt.Println("Database operations file generated successfully!")

// Generate gRPC operations file
fmt.Printf("Generating gRPC operations file at %s...\n", outputGRPC)
if err := clickhouseanalyzer.GenerateGRPCOperationsFile(grpcOperations, outputGRPC); err != nil {
fmt.Printf("Error generating gRPC operations file: %v\n", err)
os.Exit(1)
}
fmt.Println("gRPC operations file generated successfully!")
}

// runTeaStoreAnalysis runs the analysis for TeaStore system with specialized route normalization
// It creates a TeaStore-specific materialized view and queries HTTP, gRPC, and database operations
func runTeaStoreAnalysis(db *sql.DB, namespace, viewName, outputEndpoints, outputDatabase, outputGRPC string, skipView bool) {
	// Create TeaStore-specific materialized view if needed
	if !skipView {
fmt.Printf("Creating TeaStore materialized view for %s (namespace: %s)...\n", viewName, namespace)
if err := clickhouseanalyzer.CreateTeaStoreMaterializedView(db, namespace, viewName); err != nil {
fmt.Printf("Error creating materialized view: %v\n", err)
os.Exit(1)
}
}

// Query HTTP client traces
fmt.Println("Querying HTTP client traces...")
clientEndpoints, err := clickhouseanalyzer.QueryDeathStarBenchHTTPClientTraces(db, viewName, namespace)
if err != nil {
fmt.Printf("Error querying HTTP client traces: %v\n", err)
os.Exit(1)
}

// Query HTTP server traces
fmt.Println("Querying HTTP server traces...")
serverEndpoints, err := clickhouseanalyzer.QueryDeathStarBenchHTTPServerTraces(db, viewName, namespace)
if err != nil {
fmt.Printf("Error querying HTTP server traces: %v\n", err)
os.Exit(1)
}

// Query gRPC operations
fmt.Println("Querying gRPC operations...")
grpcOperations, err := clickhouseanalyzer.QueryDeathStarBenchGRPCOperations(db, viewName, namespace)
if err != nil {
fmt.Printf("Error querying gRPC operations: %v\n", err)
os.Exit(1)
}

// Query database operations
fmt.Println("Querying database operations...")
dbOperations, err := clickhouseanalyzer.QueryDeathStarBenchDatabaseOperations(db, viewName)
if err != nil {
fmt.Printf("Error querying database operations: %v\n", err)
os.Exit(1)
}

// Combine HTTP endpoints
allEndpoints := append(clientEndpoints, serverEndpoints...)
// Convert database operations to service endpoints
dbEndpoints := clickhouseanalyzer.ConvertDatabaseOperationsToEndpoints(dbOperations)
allEndpoints = append(allEndpoints, dbEndpoints...)
// Convert gRPC operations to service endpoints
grpcEndpoints := clickhouseanalyzer.ConvertGRPCOperationsToEndpoints(grpcOperations)
allEndpoints = append(allEndpoints, grpcEndpoints...)

// Generate service endpoints file
fmt.Printf("Generating service endpoints file at %s...\n", outputEndpoints)
if err := clickhouseanalyzer.GenerateServiceEndpointsFile(allEndpoints, outputEndpoints); err != nil {
fmt.Printf("Error generating service endpoints file: %v\n", err)
os.Exit(1)
}
fmt.Println("Service endpoints file generated successfully!")

// Generate database operations file
fmt.Printf("Generating database operations file at %s...\n", outputDatabase)
if err := clickhouseanalyzer.GenerateDatabaseOperationsFile(dbOperations, outputDatabase); err != nil {
fmt.Printf("Error generating database operations file: %v\n", err)
os.Exit(1)
}
fmt.Println("Database operations file generated successfully!")

// Generate gRPC operations file
fmt.Printf("Generating gRPC operations file at %s...\n", outputGRPC)
if err := clickhouseanalyzer.GenerateGRPCOperationsFile(grpcOperations, outputGRPC); err != nil {
fmt.Printf("Error generating gRPC operations file: %v\n", err)
os.Exit(1)
}
fmt.Println("gRPC operations file generated successfully!")
}
