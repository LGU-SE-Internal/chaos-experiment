package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/CUHK-SE-Group/chaos-experiment/tools/clickhouseanalyzer"
)

func main() {
	// Define command-line flags
	host := flag.String("host", "localhost", "ClickHouse server host")
	port := flag.Int("port", 9000, "ClickHouse server port")
	database := flag.String("database", "default", "ClickHouse database name")
	username := flag.String("username", "default", "ClickHouse username")
	password := flag.String("password", "", "ClickHouse password")
	outputPath := flag.String("output", "", "Path for the generated Go file (default: internal/serviceendpoints/serviceendpoints.go)")
	skipView := flag.Bool("skip-view", false, "Skip creating the materialized view")
	flag.Parse()

	// Set default output path if not specified
	if *outputPath == "" {
		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error determining project root: %v\n", err)
			os.Exit(1)
		}
		*outputPath = filepath.Join(projectRoot, "internal", "serviceendpoints", "serviceendpoints.go")
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

	// Create materialized view if needed
	if !*skipView {
		fmt.Println("Creating materialized view...")
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

	// Combine results
	allEndpoints := append(clientEndpoints, dashboardEndpoints...)

	// Generate Go file
	fmt.Printf("Generating service endpoints file at %s...\n", *outputPath)
	if err := clickhouseanalyzer.GenerateServiceEndpointsFile(allEndpoints, *outputPath); err != nil {
		fmt.Printf("Error generating service endpoints file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Service endpoints file generated successfully!")
}
