package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/LGU-SE-Internal/chaos-experiment/internal/databaseoperations"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/javaclassmethods"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/networkdependencies"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/resourcelookup"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/serviceendpoints"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"
)

func main() {
	// Define global flags
	system := flag.String("system", "ts", "Target system: 'ts' (TrainTicket) or 'otel-demo' (OpenTelemetry Demo)")
	flag.Parse()

	// Set the system type
	systemType, err := systemconfig.ParseSystemType(*system)
	if err != nil {
		fmt.Printf("Invalid system: %s. Must be 'ts' or 'otel-demo'\n", *system)
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
	fmt.Println("  cli [--system ts|otel-demo] <command> [args]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --system <system>                - Target system: 'ts' (TrainTicket) or 'otel-demo' (OpenTelemetry Demo)")
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
	networkPairs, err := resourcelookup.GetAllNetworkPairs()
	if err != nil {
		fmt.Printf("Error retrieving network services: %v\n", err)
		return
	}

	// Extract unique service names
	serviceMap := make(map[string]bool)
	for _, pair := range networkPairs {
		serviceMap[pair.SourceService] = true
		serviceMap[pair.TargetService] = true
	}

	services := make([]string, 0, len(serviceMap))
	for service := range serviceMap {
		services = append(services, service)
	}

	if len(services) == 0 {
		fmt.Println("No services with network dependencies found")
		return
	}

	// Sort the services alphabetically
	sort.Strings(services)

	fmt.Println("Services with network dependencies:")
	for _, service := range services {
		fmt.Printf("- %s\n", service)
	}
	fmt.Printf("Total: %d services\n", len(services))
}

func listServiceDependencies(serviceName string) {
	networkPairs, err := resourcelookup.GetAllNetworkPairs()
	if err != nil {
		fmt.Printf("Error retrieving network dependencies: %v\n", err)
		return
	}

	// Filter dependencies for the given service
	var dependencies []string
	for _, pair := range networkPairs {
		if pair.SourceService == serviceName {
			dependencies = append(dependencies, pair.TargetService)
		}
	}

	if len(dependencies) == 0 {
		fmt.Printf("No dependencies found for service: %s\n", serviceName)
		return
	}

	// Sort the dependencies alphabetically
	sort.Strings(dependencies)

	fmt.Printf("Dependencies for service %s:\n", serviceName)
	for i, dep := range dependencies {
		fmt.Printf("%d. %s\n", i+1, dep)
	}
	fmt.Printf("Total: %d dependencies\n", len(dependencies))
}

func listAllDependencies() {
	// Using original implementation as it requires ConnectionDetails which isn't in resourcelookup
	pairs := networkdependencies.GetAllServicePairs()

	if len(pairs) == 0 {
		fmt.Println("No service dependencies found")
		return
	}

	// Create a tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Source Service\tTarget Service\tConnection Type")
	fmt.Fprintln(w, "-------------\t-------------\t--------------")

	for _, pair := range pairs {
		fmt.Fprintf(w, "%s\t%s\t%s\n", pair.SourceService, pair.TargetService, pair.ConnectionDetails)
	}

	w.Flush()
	fmt.Printf("Total: %d service dependencies\n", len(pairs))
}

func listJVMMethods(serviceName string) {
	// Using original implementation as it requires specific format
	methods := javaclassmethods.GetClassMethodsByService(serviceName)

	if len(methods) == 0 {
		fmt.Printf("No JVM methods found for service: %s\n", serviceName)
		return
	}

	fmt.Printf("JVM methods for service %s:\n", serviceName)

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
	services := javaclassmethods.ListAllServiceNames()

	if len(services) == 0 {
		fmt.Println("No JVM services found")
		return
	}

	// Sort the services alphabetically
	sort.Strings(services)

	fmt.Println("JVM services:")
	for _, service := range services {
		fmt.Printf("- %s\n", service)
	}
	fmt.Printf("Total: %d services\n", len(services))
}

func listServiceEndpoints(serviceName string) {
	endpoints := serviceendpoints.GetEndpointsByService(serviceName)

	if len(endpoints) == 0 {
		fmt.Printf("No endpoints found for service: %s\n", serviceName)
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
		fmt.Println("No services with database operations found")
		return
	}

	// Sort the services alphabetically
	sort.Strings(services)

	fmt.Println("Services with database operations:")
	for _, service := range services {
		fmt.Printf("- %s\n", service)
	}
	fmt.Printf("Total: %d services\n", len(services))
}

func listDatabaseOperations(serviceName string) {
	operations := databaseoperations.GetOperationsByService(serviceName)

	if len(operations) == 0 {
		fmt.Printf("No database operations found for service: %s\n", serviceName)
		return
	}

	fmt.Printf("Database operations for service %s:\n", serviceName)

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
	dbOps, err := resourcelookup.GetAllDatabaseOperations()
	if err != nil {
		fmt.Printf("Error retrieving database operations: %v\n", err)
		return
	}

	// Extract unique table names
	tableMap := make(map[string]bool)
	for _, op := range dbOps {
		tableMap[op.TableName] = true
	}

	tables := make([]string, 0, len(tableMap))
	for table := range tableMap {
		tables = append(tables, table)
	}

	if len(tables) == 0 {
		fmt.Println("No database tables found")
		return
	}

	// Sort the tables alphabetically
	sort.Strings(tables)

	fmt.Println("Database tables:")
	for _, table := range tables {
		fmt.Printf("- %s\n", table)
	}
	fmt.Printf("Total: %d tables\n", len(tables))
}

func listAllDatabaseOperations() {
	dbOps, err := resourcelookup.GetAllDatabaseOperations()
	if err != nil {
		fmt.Printf("Error retrieving database operations: %v\n", err)
		return
	}

	if len(dbOps) == 0 {
		fmt.Println("No database operations found")
		return
	}

	// Create a tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Service\tDatabase\tTable\tOperation")
	fmt.Fprintln(w, "-------\t--------\t-----\t---------")

	for _, op := range dbOps {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			op.AppName,
			op.DBName,
			op.TableName,
			op.OperationType)
	}

	w.Flush()
	fmt.Printf("Total: %d database operations\n", len(dbOps))
}
