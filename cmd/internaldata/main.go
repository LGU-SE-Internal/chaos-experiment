package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/LGU-SE-Internal/chaos-experiment/internal/databaseoperations"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/javaclassmethods"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/serviceendpoints"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"
)

func main() {
	// Define global flags
	system := flag.String("system", "ts", "Target system: 'ts' (TrainTicket), 'otel-demo' (OpenTelemetry Demo), 'media' (MediaMicroservices), 'hs' (HotelReservation), 'sn' (SocialNetwork), 'ob' (OnlineBoutique), 'sockshop' (Sock Shop), or 'teastore' (Tea Store)")
	flag.Parse()

	// Set the system type
	systemType, err := systemconfig.ParseSystemType(*system)
	if err != nil {
		fmt.Printf("Invalid system: %s. Must be 'ts', 'otel-demo', 'media', 'hs', 'sn', 'ob', 'sockshop', or 'teastore'\n", *system)
		os.Exit(1)
	}
	if err := systemconfig.SetCurrentSystem(systemType); err != nil {
		fmt.Printf("Error setting system type: %v\n", err)
		os.Exit(1)
	}

	// Get remaining args after flags
	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		return
	}

	command := args[0]

	switch command {
	case "list-services":
		listNetworkServices()
	case "list-dependencies":
		if len(args) < 2 {
			fmt.Println("Please provide a service name")
			return
		}
		listServiceDependencies(args[1])
	case "list-all-dependencies":
		listAllDependencies()
	case "list-jvm-methods":
		if len(args) < 2 {
			fmt.Println("Please provide a service name")
			return
		}
		listJVMMethods(args[1])
	case "list-jvm-services":
		listJVMServices()
	case "list-endpoints":
		if len(args) < 2 {
			fmt.Println("Please provide a service name")
			return
		}
		listServiceEndpoints(args[1])
	case "list-db-services":
		listDatabaseServices()
	case "list-db-operations":
		if len(args) < 2 {
			fmt.Println("Please provide a service name")
			return
		}
		listDatabaseOperations(args[1])
	case "list-db-tables":
		listDatabaseTables()
	case "list-all-db-operations":
		listAllDatabaseOperations()
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  cli [--system ts|otel-demo|media|hs|sn|ob|sockshop|teastore] <command> [args]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --system <system>                - Target system: 'ts' (TrainTicket), 'otel-demo' (OpenTelemetry Demo), 'media' (MediaMicroservices), 'hs' (HotelReservation), 'sn' (SocialNetwork), 'ob' (OnlineBoutique), 'sockshop' (Sock Shop), or 'teastore' (Tea Store)")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list-services                    - List all services with network dependencies")
	fmt.Println("  list-dependencies <service>      - List dependencies for a specific service")
	fmt.Println("  list-all-dependencies            - List all service dependencies")
	fmt.Println("  list-jvm-methods <service>       - List JVM methods for a specific service")
	fmt.Println("  list-jvm-services                - List all Java services")
	fmt.Println("  list-endpoints <service>         - List endpoints for a specific service (with SpanName)")
	fmt.Println("  list-db-services                 - List all services with database operations")
	fmt.Println("  list-db-operations <service>     - List database operations for a specific service")
	fmt.Println("  list-db-tables                   - List all database tables")
	fmt.Println("  list-all-db-operations           - List all database operations")
	fmt.Println()
	fmt.Printf("Current system: %s\n", systemconfig.GetCurrentSystem())
}

func listNetworkServices() {
	// Use system-aware service list via the routing layer
	services := serviceendpoints.GetAllServices()

	if len(services) == 0 {
		fmt.Printf("No services with network dependencies found for system: %s\n", systemconfig.GetCurrentSystem())
		return
	}

	// Sort the services alphabetically
	sort.Strings(services)

	fmt.Printf("Services with network dependencies (system: %s):\n", systemconfig.GetCurrentSystem())
	for _, service := range services {
		fmt.Printf("- %s\n", service)
	}
	fmt.Printf("Total: %d services\n", len(services))
}

func listServiceDependencies(serviceName string) {
	// Get all endpoints for the service and extract unique target services
	endpoints := serviceendpoints.GetEndpointsByService(serviceName)

	// Extract unique dependencies
	depMap := make(map[string]bool)
	for _, ep := range endpoints {
		if ep.ServerAddress != "" && ep.ServerAddress != serviceName {
			depMap[ep.ServerAddress] = true
		}
	}

	dependencies := make([]string, 0, len(depMap))
	for dep := range depMap {
		dependencies = append(dependencies, dep)
	}

	if len(dependencies) == 0 {
		fmt.Printf("No dependencies found for service: %s (system: %s)\n", serviceName, systemconfig.GetCurrentSystem())
		return
	}

	// Sort the dependencies alphabetically
	sort.Strings(dependencies)

	fmt.Printf("Dependencies for service %s (system: %s):\n", serviceName, systemconfig.GetCurrentSystem())
	for i, dep := range dependencies {
		fmt.Printf("%d. %s\n", i+1, dep)
	}
	fmt.Printf("Total: %d dependencies\n", len(dependencies))
}

func listAllDependencies() {
	// Get all services and build dependency pairs from endpoints
	services := serviceendpoints.GetAllServices()

	type depPair struct {
		Source string
		Target string
	}
	pairMap := make(map[depPair]bool)

	for _, service := range services {
		endpoints := serviceendpoints.GetEndpointsByService(service)
		for _, ep := range endpoints {
			if ep.ServerAddress != "" && ep.ServerAddress != service {
				pairMap[depPair{Source: service, Target: ep.ServerAddress}] = true
			}
		}
	}

	if len(pairMap) == 0 {
		fmt.Printf("No service dependencies found (system: %s)\n", systemconfig.GetCurrentSystem())
		return
	}

	// Convert to slice and sort
	pairs := make([]depPair, 0, len(pairMap))
	for pair := range pairMap {
		pairs = append(pairs, pair)
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Source != pairs[j].Source {
			return pairs[i].Source < pairs[j].Source
		}
		return pairs[i].Target < pairs[j].Target
	})

	// Create a tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Service dependencies (system: %s):\n", systemconfig.GetCurrentSystem())
	fmt.Fprintln(w, "Source Service\tTarget Service\tConnection Type")
	fmt.Fprintln(w, "-------------\t-------------\t--------------")

	for _, pair := range pairs {
		fmt.Fprintf(w, "%s\t%s\t%s\n", pair.Source, pair.Target, "HTTP/gRPC Communication")
	}

	w.Flush()
	fmt.Printf("Total: %d service dependencies\n", len(pairs))
}

func listJVMMethods(serviceName string) {
	methods := javaclassmethods.GetClassMethodsByService(serviceName)

	if len(methods) == 0 {
		fmt.Printf("No JVM methods found for service: %s (system: %s)\n", serviceName, systemconfig.GetCurrentSystem())
		return
	}

	fmt.Printf("JVM methods for service %s (system: %s):\n", serviceName, systemconfig.GetCurrentSystem())

	// Create a tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Index\tClass\tMethod")
	fmt.Fprintln(w, "-----\t-----\t------")

	for i, method := range methods {
		fmt.Fprintf(w, "%d\t%s\t%s\n", i, method.ClassName, method.MethodName)
	}

	w.Flush()
	fmt.Printf("Total: %d methods\n", len(methods))
}

func listJVMServices() {
	services := javaclassmethods.GetAllServices()

	if len(services) == 0 {
		fmt.Printf("No JVM services found (system: %s)\n", systemconfig.GetCurrentSystem())
		return
	}

	// Sort the services alphabetically
	sort.Strings(services)

	fmt.Printf("JVM services (system: %s):\n", systemconfig.GetCurrentSystem())
	for _, service := range services {
		fmt.Printf("- %s\n", service)
	}
	fmt.Printf("Total: %d services\n", len(services))
}

func listServiceEndpoints(serviceName string) {
	endpoints := serviceendpoints.GetEndpointsByService(serviceName)

	if len(endpoints) == 0 {
		fmt.Printf("No endpoints found for service: %s (system: %s)\n", serviceName, systemconfig.GetCurrentSystem())
		return
	}

	fmt.Printf("Endpoints for service %s (system: %s):\n", serviceName, systemconfig.GetCurrentSystem())

	// Create a tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Method\tRoute\tTarget Address\tTarget Port\tResponse Status\tSpanName")
	fmt.Fprintln(w, "------\t-----\t-------------\t-----------\t--------------\t--------")

	for _, endpoint := range endpoints {
		method := endpoint.RequestMethod
		if method == "" {
			method = "N/A"
		}
		route := endpoint.Route
		if route == "" {
			route = "N/A"
		}
		status := endpoint.ResponseStatus
		if status == "" {
			status = "N/A"
		}
		spanName := endpoint.SpanName
		if spanName == "" {
			spanName = "N/A"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			method,
			route,
			endpoint.ServerAddress,
			endpoint.ServerPort,
			status,
			spanName)
	}

	w.Flush()
	fmt.Printf("Total: %d endpoints\n", len(endpoints))
}

// New functions for database operations

func listDatabaseServices() {
	services := databaseoperations.GetAllDatabaseServices()

	if len(services) == 0 {
		fmt.Printf("No services with database operations found (system: %s)\n", systemconfig.GetCurrentSystem())
		return
	}

	// Sort the services alphabetically
	sort.Strings(services)

	fmt.Printf("Services with database operations (system: %s):\n", systemconfig.GetCurrentSystem())
	for _, service := range services {
		fmt.Printf("- %s\n", service)
	}
	fmt.Printf("Total: %d services\n", len(services))
}

func listDatabaseOperations(serviceName string) {
	operations := databaseoperations.GetOperationsByService(serviceName)

	if len(operations) == 0 {
		fmt.Printf("No database operations found for service: %s (system: %s)\n", serviceName, systemconfig.GetCurrentSystem())
		return
	}

	fmt.Printf("Database operations for service %s (system: %s):\n", serviceName, systemconfig.GetCurrentSystem())

	// Create a tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Database\tTable\tOperation")
	fmt.Fprintln(w, "--------\t-----\t---------")

	for _, op := range operations {
		fmt.Fprintf(w, "%s\t%s\t%s\n", op.DBName, op.DBTable, op.Operation)
	}

	w.Flush()
	fmt.Printf("Total: %d operations\n", len(operations))
}

func listDatabaseTables() {
	services := databaseoperations.GetAllDatabaseServices()

	// Extract unique table names
	tableMap := make(map[string]bool)
	for _, service := range services {
		ops := databaseoperations.GetOperationsByService(service)
		for _, op := range ops {
			if op.DBTable != "" {
				tableMap[op.DBTable] = true
			}
		}
	}

	tables := make([]string, 0, len(tableMap))
	for table := range tableMap {
		tables = append(tables, table)
	}

	if len(tables) == 0 {
		fmt.Printf("No database tables found (system: %s)\n", systemconfig.GetCurrentSystem())
		return
	}

	// Sort the tables alphabetically
	sort.Strings(tables)

	fmt.Printf("Database tables (system: %s):\n", systemconfig.GetCurrentSystem())
	for _, table := range tables {
		fmt.Printf("- %s\n", table)
	}
	fmt.Printf("Total: %d tables\n", len(tables))
}

func listAllDatabaseOperations() {
	services := databaseoperations.GetAllDatabaseServices()

	type dbOpEntry struct {
		AppName       string
		DBName        string
		TableName     string
		OperationType string
	}

	var allOps []dbOpEntry
	for _, service := range services {
		ops := databaseoperations.GetOperationsByService(service)
		for _, op := range ops {
			allOps = append(allOps, dbOpEntry{
				AppName:       service,
				DBName:        op.DBName,
				TableName:     op.DBTable,
				OperationType: op.Operation,
			})
		}
	}

	if len(allOps) == 0 {
		fmt.Printf("No database operations found (system: %s)\n", systemconfig.GetCurrentSystem())
		return
	}

	// Create a tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Database operations (system: %s):\n", systemconfig.GetCurrentSystem())
	fmt.Fprintln(w, "Service\tDatabase\tTable\tOperation")
	fmt.Fprintln(w, "-------\t--------\t-----\t---------")

	for _, op := range allOps {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			op.AppName,
			op.DBName,
			op.TableName,
			op.OperationType)
	}

	w.Flush()
	fmt.Printf("Total: %d database operations\n", len(allOps))
}
