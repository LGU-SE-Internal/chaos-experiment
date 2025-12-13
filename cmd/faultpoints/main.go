package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	_ "github.com/LGU-SE-Internal/chaos-experiment/internal/adapter" // Auto-registers all systems
	"github.com/LGU-SE-Internal/chaos-experiment/internal/endpoint"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/registry"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"
)

func main() {
	// Define global flags
	system := flag.String("system", "ts", "Target system: 'ts' (TrainTicket), 'otel-demo' (OpenTelemetry Demo), 'media' (MediaMicroservices), 'hs' (HotelReservation), 'sn' (SocialNetwork), or 'ob' (OnlineBoutique)")
	flag.Parse()

	// Set the system type
	systemType, err := systemconfig.ParseSystemType(*system)
	if err != nil {
		fmt.Printf("Invalid system: %s. Must be 'ts', 'otel-demo', 'media', 'hs', 'sn', or 'ob'\n", *system)
		os.Exit(1)
	}
	if err := systemconfig.SetCurrentSystem(systemType); err != nil {
		fmt.Printf("Error setting system type: %v\n", err)
		os.Exit(1)
	}

	// Verify system is registered
	if !registry.IsRegistered(systemconfig.GetCurrentSystem()) {
		fmt.Printf("System %s is not registered\n", systemconfig.GetCurrentSystem())
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
	case "list-http":
		listHTTPEndpoints()
	case "list-network":
		listNetworkPairs()
	case "list-dns":
		listDNSEndpoints()
	case "list-db":
		listDatabaseOperations()
	case "list-all":
		listAllFaultPoints()
	case "summary":
		showSummary()
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Fault Injection Points Viewer")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  faultpoints [--system ts|otel-demo|media|hs|sn|ob] <command>")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --system <system>  - Target system: 'ts' (TrainTicket), 'otel-demo' (OpenTelemetry Demo), 'media' (MediaMicroservices), 'hs' (HotelReservation), 'sn' (SocialNetwork), or 'ob' (OnlineBoutique)")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list-http          - List all HTTP fault injection points")
	fmt.Println("  list-network       - List all network fault injection points (service pairs)")
	fmt.Println("  list-dns           - List all DNS fault injection points")
	fmt.Println("  list-db            - List all database fault injection points")
	fmt.Println("  list-all           - List all fault injection points")
	fmt.Println("  summary            - Show summary of fault injection points")
	fmt.Println()
	fmt.Printf("Current system: %s\n", systemconfig.GetCurrentSystem())
}

func listHTTPEndpoints() {
	sysData := registry.MustGetCurrent()
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SERVICE\tROUTE\tMETHOD\tTARGET\tPORT")
	fmt.Fprintln(w, "-------\t-----\t------\t------\t----")
	
	var endpoints []endpoint.HTTPEndpointInfo
	for _, service := range sysData.GetAllServices() {
		for _, ep := range sysData.GetHTTPEndpointsByService(service) {
			if ep.Route != "" {
				endpoints = append(endpoints, endpoint.ToHTTPEndpointInfo(ep))
			}
		}
	}
	
	// Sort by service name, then route
	sort.Slice(endpoints, func(i, j int) bool {
		if endpoints[i].ServiceName != endpoints[j].ServiceName {
			return endpoints[i].ServiceName < endpoints[j].ServiceName
		}
		return endpoints[i].Route < endpoints[j].Route
	})
	
	for _, ep := range endpoints {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			ep.ServiceName, ep.Route, ep.Method, ep.ServerAddress, ep.ServerPort)
	}
	w.Flush()
	fmt.Printf("\nTotal HTTP endpoints: %d\n", len(endpoints))
}

func listNetworkPairs() {
	sysData := registry.MustGetCurrent()
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SOURCE\tTARGET\tOPERATION TYPES")
	fmt.Fprintln(w, "------\t------\t---------------")
	
	// Build network pairs from all operation types
	pairMap := make(map[string]*endpoint.CallPair)
	
	// Add HTTP endpoints
	for _, service := range sysData.GetAllServices() {
		for _, ep := range sysData.GetHTTPEndpointsByService(service) {
			if ep.ServerAddress != "" && ep.ServerAddress != service {
				key := ep.ServiceName + "->" + ep.ServerAddress
				if pairMap[key] == nil {
					pairMap[key] = &endpoint.CallPair{
						SourceService:  ep.ServiceName,
						TargetService:  ep.ServerAddress,
						OperationTypes: []string{},
					}
				}
				if !contains(pairMap[key].OperationTypes, "http") {
					pairMap[key].OperationTypes = append(pairMap[key].OperationTypes, "http")
				}
			}
		}
	}
	
	// Add RPC operations
	for _, service := range sysData.GetAllRPCServices() {
		for _, op := range sysData.GetRPCOperationsByService(service) {
			if op.ServerAddress != "" && op.ServerAddress != service {
				key := op.ServiceName + "->" + op.ServerAddress
				if pairMap[key] == nil {
					pairMap[key] = &endpoint.CallPair{
						SourceService:  op.ServiceName,
						TargetService:  op.ServerAddress,
						OperationTypes: []string{},
					}
				}
				if !contains(pairMap[key].OperationTypes, "rpc") {
					pairMap[key].OperationTypes = append(pairMap[key].OperationTypes, "rpc")
				}
			}
		}
	}
	
	// Add Database operations
	for _, service := range sysData.GetAllDatabaseServices() {
		for _, op := range sysData.GetDatabaseOperationsByService(service) {
			if op.ServerAddress != "" && op.ServerAddress != service {
				key := op.ServiceName + "->" + op.ServerAddress
				if pairMap[key] == nil {
					pairMap[key] = &endpoint.CallPair{
						SourceService:  op.ServiceName,
						TargetService:  op.ServerAddress,
						OperationTypes: []string{},
					}
				}
				if !contains(pairMap[key].OperationTypes, "db") {
					pairMap[key].OperationTypes = append(pairMap[key].OperationTypes, "db")
				}
			}
		}
	}
	
	// Convert to slice and sort
	var pairs []endpoint.CallPair
	for _, pair := range pairMap {
		sort.Strings(pair.OperationTypes)
		pairs = append(pairs, *pair)
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].SourceService != pairs[j].SourceService {
			return pairs[i].SourceService < pairs[j].SourceService
		}
		return pairs[i].TargetService < pairs[j].TargetService
	})
	
	for _, pair := range pairs {
		opTypes := ""
		for i, t := range pair.OperationTypes {
			if i > 0 {
				opTypes += ", "
			}
			opTypes += t
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", pair.SourceService, pair.TargetService, opTypes)
	}
	w.Flush()
	fmt.Printf("\nTotal network pairs: %d\n", len(pairs))
}

func listDNSEndpoints() {
	sysData := registry.MustGetCurrent()
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SERVICE\tDOMAIN\tTYPES")
	fmt.Fprintln(w, "-------\t------\t-----")
	
	// Build DNS endpoints (HTTP + DB, exclude RPC)
	domainMap := make(map[string]*endpoint.DNSEndpointInfo)
	
	// Add HTTP endpoints
	for _, service := range sysData.GetAllServices() {
		for _, ep := range sysData.GetHTTPEndpointsByService(service) {
			if ep.ServerAddress != "" && ep.ServerAddress != service {
				key := ep.ServiceName + "->" + ep.ServerAddress
				if domainMap[key] == nil {
					domainMap[key] = &endpoint.DNSEndpointInfo{
						ServiceName: ep.ServiceName,
						Domain:      ep.ServerAddress,
					}
				}
				domainMap[key].HasHTTP = true
			}
		}
	}
	
	// Add Database operations
	for _, service := range sysData.GetAllDatabaseServices() {
		for _, op := range sysData.GetDatabaseOperationsByService(service) {
			if op.ServerAddress != "" && op.ServerAddress != service {
				key := op.ServiceName + "->" + op.ServerAddress
				if domainMap[key] == nil {
					domainMap[key] = &endpoint.DNSEndpointInfo{
						ServiceName: op.ServiceName,
						Domain:      op.ServerAddress,
					}
				}
				domainMap[key].HasDB = true
			}
		}
	}
	
	// Convert to slice and sort
	var dnsEndpoints []endpoint.DNSEndpointInfo
	for _, info := range domainMap {
		dnsEndpoints = append(dnsEndpoints, *info)
	}
	sort.Slice(dnsEndpoints, func(i, j int) bool {
		if dnsEndpoints[i].ServiceName != dnsEndpoints[j].ServiceName {
			return dnsEndpoints[i].ServiceName < dnsEndpoints[j].ServiceName
		}
		return dnsEndpoints[i].Domain < dnsEndpoints[j].Domain
	})
	
	for _, ep := range dnsEndpoints {
		types := ""
		if ep.HasHTTP {
			types += "HTTP"
		}
		if ep.HasDB {
			if types != "" {
				types += ", "
			}
			types += "DB"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", ep.ServiceName, ep.Domain, types)
	}
	w.Flush()
	fmt.Printf("\nTotal DNS endpoints: %d\n", len(dnsEndpoints))
}

func listDatabaseOperations() {
	sysData := registry.MustGetCurrent()
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SERVICE\tDB SYSTEM\tDB NAME\tTABLE\tOPERATION")
	fmt.Fprintln(w, "-------\t---------\t-------\t-----\t---------")
	
	var dbOps []endpoint.DatabaseInfo
	for _, service := range sysData.GetAllDatabaseServices() {
		for _, op := range sysData.GetDatabaseOperationsByService(service) {
			dbOps = append(dbOps, endpoint.ToDatabaseInfo(op))
		}
	}
	
	// Sort by service name, then DB name
	sort.Slice(dbOps, func(i, j int) bool {
		if dbOps[i].ServiceName != dbOps[j].ServiceName {
			return dbOps[i].ServiceName < dbOps[j].ServiceName
		}
		if dbOps[i].DBName != dbOps[j].DBName {
			return dbOps[i].DBName < dbOps[j].DBName
		}
		return dbOps[i].TableName < dbOps[j].TableName
	})
	
	for _, op := range dbOps {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			op.ServiceName, "mysql", op.DBName, op.TableName, op.Operation)
	}
	w.Flush()
	fmt.Printf("\nTotal database operations: %d\n", len(dbOps))
}

func listAllFaultPoints() {
	fmt.Println("=== HTTP Endpoints ===")
	listHTTPEndpoints()
	fmt.Println("\n=== Network Pairs ===")
	listNetworkPairs()
	fmt.Println("\n=== DNS Endpoints ===")
	listDNSEndpoints()
	fmt.Println("\n=== Database Operations ===")
	listDatabaseOperations()
}

func showSummary() {
	sysData := registry.MustGetCurrent()
	
	fmt.Printf("System: %s\n", systemconfig.GetCurrentSystem())
	fmt.Printf("Total Services: %d\n", len(sysData.GetAllServices()))
	
	// Count HTTP endpoints
	httpCount := 0
	for _, service := range sysData.GetAllServices() {
		for _, ep := range sysData.GetHTTPEndpointsByService(service) {
			if ep.Route != "" {
				httpCount++
			}
		}
	}
	fmt.Printf("HTTP Endpoints: %d\n", httpCount)
	
	// Count DB operations
	dbCount := 0
	for _, service := range sysData.GetAllDatabaseServices() {
		dbCount += len(sysData.GetDatabaseOperationsByService(service))
	}
	fmt.Printf("Database Operations: %d\n", dbCount)
	
	// Count RPC operations
	rpcCount := 0
	for _, service := range sysData.GetAllRPCServices() {
		rpcCount += len(sysData.GetRPCOperationsByService(service))
	}
	fmt.Printf("RPC Operations: %d\n", rpcCount)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
