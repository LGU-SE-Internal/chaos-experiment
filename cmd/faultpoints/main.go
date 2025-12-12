package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/LGU-SE-Internal/chaos-experiment/internal/grpcoperations"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/resourcelookup"
	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"

	hsdb "github.com/LGU-SE-Internal/chaos-experiment/internal/hs/databaseoperations"
	hsendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/hs/serviceendpoints"
	mediadb "github.com/LGU-SE-Internal/chaos-experiment/internal/media/databaseoperations"
	mediaendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/media/serviceendpoints"
	obdb "github.com/LGU-SE-Internal/chaos-experiment/internal/ob/databaseoperations"
	obendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/ob/serviceendpoints"
	oteldemodb "github.com/LGU-SE-Internal/chaos-experiment/internal/oteldemo/databaseoperations"
	oteldemoendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/oteldemo/serviceendpoints"
	oteldemogrpc "github.com/LGU-SE-Internal/chaos-experiment/internal/oteldemo/grpcoperations"
	oteldemojvm "github.com/LGU-SE-Internal/chaos-experiment/internal/oteldemo/javaclassmethods"
	sndb "github.com/LGU-SE-Internal/chaos-experiment/internal/sn/databaseoperations"
	snendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/sn/serviceendpoints"
	tsdb "github.com/LGU-SE-Internal/chaos-experiment/internal/ts/databaseoperations"
	tsendpoints "github.com/LGU-SE-Internal/chaos-experiment/internal/ts/serviceendpoints"
	tsjvm "github.com/LGU-SE-Internal/chaos-experiment/internal/ts/javaclassmethods"
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

	// Initialize resource lookup caches based on selected system
	initResourceLookupForSystem()

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
	case "list-jvm":
		listJVMMethods()
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
	fmt.Println("  list-jvm           - List all JVM fault injection points")
	fmt.Println("  list-db            - List all database fault injection points")
	fmt.Println("  list-all           - List all fault injection points")
	fmt.Println("  summary            - Show summary of fault injection points")
	fmt.Println()
	fmt.Printf("Current system: %s\n", systemconfig.GetCurrentSystem())
}

// initResourceLookupForSystem initializes the resource lookup with system-specific data
func initResourceLookupForSystem() {
	// Force clearing of any cached data to ensure fresh lookups
	resourcelookup.InvalidateCache()
}

func listHTTPEndpoints() {
	endpoints, err := getHTTPEndpointsForCurrentSystem()
	if err != nil {
		fmt.Printf("Error retrieving HTTP endpoints: %v\n", err)
		return
	}

	if len(endpoints) == 0 {
		fmt.Printf("No HTTP fault injection points found (system: %s)\n", systemconfig.GetCurrentSystem())
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "HTTP Fault Injection Points (system: %s):\n", systemconfig.GetCurrentSystem())
	fmt.Fprintln(w, "Index\tSource Service\tMethod\tRoute\tTarget Service\tPort\tSpanName")
	fmt.Fprintln(w, "-----\t--------------\t------\t-----\t--------------\t----\t--------")

	for i, ep := range endpoints {
		method := ep.Method
		if method == "" {
			method = "N/A"
		}
		spanName := ep.SpanName
		if spanName == "" {
			spanName = "N/A"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			i, ep.AppName, method, ep.Route, ep.ServerAddress, ep.ServerPort, spanName)
	}
	w.Flush()
	fmt.Printf("Total: %d HTTP fault injection points\n", len(endpoints))
}

func listNetworkPairs() {
	pairs, err := getNetworkPairsForCurrentSystem()
	if err != nil {
		fmt.Printf("Error retrieving network pairs: %v\n", err)
		return
	}

	if len(pairs) == 0 {
		fmt.Printf("No network fault injection points found (system: %s)\n", systemconfig.GetCurrentSystem())
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Network Fault Injection Points (system: %s):\n", systemconfig.GetCurrentSystem())
	fmt.Fprintln(w, "Index\tSource Service\tTarget Service\tSpanNames Count")
	fmt.Fprintln(w, "-----\t--------------\t--------------\t---------------")

	for i, pair := range pairs {
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\n",
			i, pair.SourceService, pair.TargetService, len(pair.SpanNames))
	}
	w.Flush()
	fmt.Printf("Total: %d network fault injection points\n", len(pairs))
}

func listDNSEndpoints() {
	endpoints, err := getDNSEndpointsForCurrentSystem()
	if err != nil {
		fmt.Printf("Error retrieving DNS endpoints: %v\n", err)
		return
	}

	if len(endpoints) == 0 {
		fmt.Printf("No DNS fault injection points found (system: %s)\n", systemconfig.GetCurrentSystem())
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "DNS Fault Injection Points (system: %s):\n", systemconfig.GetCurrentSystem())
	fmt.Fprintln(w, "Index\tSource Service\tTarget Domain\tSpanNames Count")
	fmt.Fprintln(w, "-----\t--------------\t-------------\t---------------")

	for i, ep := range endpoints {
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\n",
			i, ep.AppName, ep.Domain, len(ep.SpanNames))
	}
	w.Flush()
	fmt.Printf("Total: %d DNS fault injection points\n", len(endpoints))
}

func listJVMMethods() {
	methods, err := getJVMMethodsForCurrentSystem()
	if err != nil {
		fmt.Printf("Error retrieving JVM methods: %v\n", err)
		return
	}

	if len(methods) == 0 {
		fmt.Printf("No JVM fault injection points found (system: %s)\n", systemconfig.GetCurrentSystem())
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "JVM Fault Injection Points (system: %s):\n", systemconfig.GetCurrentSystem())
	fmt.Fprintln(w, "Index\tService\tClass\tMethod")
	fmt.Fprintln(w, "-----\t-------\t-----\t------")

	for i, method := range methods {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n",
			i, method.AppName, method.ClassName, method.MethodName)
	}
	w.Flush()
	fmt.Printf("Total: %d JVM fault injection points\n", len(methods))
}

func listDatabaseOperations() {
	operations, err := getDatabaseOperationsForCurrentSystem()
	if err != nil {
		fmt.Printf("Error retrieving database operations: %v\n", err)
		return
	}

	if len(operations) == 0 {
		fmt.Printf("No database fault injection points found (system: %s)\n", systemconfig.GetCurrentSystem())
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Database Fault Injection Points (system: %s):\n", systemconfig.GetCurrentSystem())
	fmt.Fprintln(w, "Index\tService\tDatabase\tTable\tOperation")
	fmt.Fprintln(w, "-----\t-------\t--------\t-----\t---------")

	for i, op := range operations {
		table := op.TableName
		if table == "" {
			table = "N/A"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
			i, op.AppName, op.DBName, table, op.OperationType)
	}
	w.Flush()
	fmt.Printf("Total: %d database fault injection points\n", len(operations))
}

func listAllFaultPoints() {
	fmt.Printf("=== All Fault Injection Points (system: %s) ===\n\n", systemconfig.GetCurrentSystem())

	fmt.Println("--- HTTP Endpoints ---")
	listHTTPEndpoints()
	fmt.Println()

	fmt.Println("--- Network Pairs ---")
	listNetworkPairs()
	fmt.Println()

	fmt.Println("--- DNS Endpoints ---")
	listDNSEndpoints()
	fmt.Println()

	fmt.Println("--- JVM Methods ---")
	listJVMMethods()
	fmt.Println()

	fmt.Println("--- Database Operations ---")
	listDatabaseOperations()
}

func showSummary() {
	httpEndpoints, _ := getHTTPEndpointsForCurrentSystem()
	networkPairs, _ := getNetworkPairsForCurrentSystem()
	dnsEndpoints, _ := getDNSEndpointsForCurrentSystem()
	jvmMethods, _ := getJVMMethodsForCurrentSystem()
	dbOperations, _ := getDatabaseOperationsForCurrentSystem()

	fmt.Printf("Fault Injection Points Summary (system: %s):\n", systemconfig.GetCurrentSystem())
	fmt.Println("============================================")
	fmt.Printf("HTTP Endpoints:      %d\n", len(httpEndpoints))
	fmt.Printf("Network Pairs:       %d\n", len(networkPairs))
	fmt.Printf("DNS Endpoints:       %d\n", len(dnsEndpoints))
	fmt.Printf("JVM Methods:         %d\n", len(jvmMethods))
	fmt.Printf("Database Operations: %d\n", len(dbOperations))
	fmt.Println("--------------------------------------------")
	fmt.Printf("Total:               %d\n", len(httpEndpoints)+len(networkPairs)+len(dnsEndpoints)+len(jvmMethods)+len(dbOperations))
}

// ============================================================================
// System-aware helper functions
// ============================================================================

func getHTTPEndpointsForCurrentSystem() ([]resourcelookup.AppEndpointPair, error) {
	// Build the result from system-specific data
	var services []string
	switch systemconfig.GetCurrentSystem() {
	case systemconfig.SystemTrainTicket:
		services = tsendpoints.GetAllServices()
	case systemconfig.SystemOtelDemo:
		services = oteldemoendpoints.GetAllServices()
	case systemconfig.SystemMediaMicroservices:
		services = mediaendpoints.GetAllServices()
	case systemconfig.SystemHotelReservation:
		services = hsendpoints.GetAllServices()
	case systemconfig.SystemSocialNetwork:
		services = snendpoints.GetAllServices()
	case systemconfig.SystemOnlineBoutique:
		services = obendpoints.GetAllServices()
	default:
		return resourcelookup.GetAllHTTPEndpoints()
	}

	result := make([]resourcelookup.AppEndpointPair, 0)
	for _, serviceName := range services {
		var endpoints interface{}
		switch systemconfig.GetCurrentSystem() {
		case systemconfig.SystemTrainTicket:
			endpoints = tsendpoints.GetEndpointsByService(serviceName)
		case systemconfig.SystemOtelDemo:
			endpoints = oteldemoendpoints.GetEndpointsByService(serviceName)
		case systemconfig.SystemMediaMicroservices:
			endpoints = mediaendpoints.GetEndpointsByService(serviceName)
		case systemconfig.SystemHotelReservation:
			endpoints = hsendpoints.GetEndpointsByService(serviceName)
		case systemconfig.SystemSocialNetwork:
			endpoints = snendpoints.GetEndpointsByService(serviceName)
case systemconfig.SystemOnlineBoutique:
endpoints = obendpoints.GetEndpointsByService(serviceName)
		case systemconfig.SystemOnlineBoutique:
			endpoints = obendpoints.GetEndpointsByService(serviceName)
		}

		switch eps := endpoints.(type) {
		case []tsendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.Route != "" && ep.ServerAddress != "ts-rabbitmq" {
					result = append(result, resourcelookup.AppEndpointPair{
						AppName:       serviceName,
						Route:         ep.Route,
						Method:        ep.RequestMethod,
						ServerAddress: ep.ServerAddress,
						ServerPort:    ep.ServerPort,
						SpanName:      ep.SpanName,
					})
				}
			}
		case []oteldemoendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.Route != "" {
					result = append(result, resourcelookup.AppEndpointPair{
						AppName:       serviceName,
						Route:         ep.Route,
						Method:        ep.RequestMethod,
						ServerAddress: ep.ServerAddress,
						ServerPort:    ep.ServerPort,
						SpanName:      ep.SpanName,
					})
				}
			}
		case []mediaendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.Route != "" {
					result = append(result, resourcelookup.AppEndpointPair{
						AppName:       serviceName,
						Route:         ep.Route,
						Method:        ep.RequestMethod,
						ServerAddress: ep.ServerAddress,
						ServerPort:    ep.ServerPort,
						SpanName:      ep.SpanName,
					})
				}
			}
		case []hsendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.Route != "" {
					result = append(result, resourcelookup.AppEndpointPair{
						AppName:       serviceName,
						Route:         ep.Route,
						Method:        ep.RequestMethod,
						ServerAddress: ep.ServerAddress,
						ServerPort:    ep.ServerPort,
						SpanName:      ep.SpanName,
					})
				}
			}
		case []snendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.Route != "" {
					result = append(result, resourcelookup.AppEndpointPair{
						AppName:       serviceName,
						Route:         ep.Route,
						Method:        ep.RequestMethod,
						ServerAddress: ep.ServerAddress,
						ServerPort:    ep.ServerPort,
						SpanName:      ep.SpanName,
					})
				}
			}
		case []obendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.Route != "" {
					result = append(result, resourcelookup.AppEndpointPair{
						AppName:       serviceName,
						Route:         ep.Route,
						Method:        ep.RequestMethod,
						ServerAddress: ep.ServerAddress,
						ServerPort:    ep.ServerPort,
						SpanName:      ep.SpanName,
					})
				}
			}
		}
	}
	return result, nil
}

func getNetworkPairsForCurrentSystem() ([]resourcelookup.AppNetworkPair, error) {
	// Build network pairs from service endpoints
	var services []string
	switch systemconfig.GetCurrentSystem() {
	case systemconfig.SystemTrainTicket:
		services = tsendpoints.GetAllServices()
	case systemconfig.SystemOtelDemo:
		services = oteldemoendpoints.GetAllServices()
	case systemconfig.SystemMediaMicroservices:
		services = mediaendpoints.GetAllServices()
	case systemconfig.SystemHotelReservation:
		services = hsendpoints.GetAllServices()
	case systemconfig.SystemSocialNetwork:
		services = snendpoints.GetAllServices()
	case systemconfig.SystemOnlineBoutique:
		services = obendpoints.GetAllServices()
	default:
		return resourcelookup.GetAllNetworkPairs()
	}

	// Build unique source->target pairs
	pairMap := make(map[string]map[string][]string) // source -> target -> spanNames

	for _, serviceName := range services {
		var endpoints interface{}
		switch systemconfig.GetCurrentSystem() {
		case systemconfig.SystemTrainTicket:
			endpoints = tsendpoints.GetEndpointsByService(serviceName)
		case systemconfig.SystemOtelDemo:
			endpoints = oteldemoendpoints.GetEndpointsByService(serviceName)
		case systemconfig.SystemMediaMicroservices:
			endpoints = mediaendpoints.GetEndpointsByService(serviceName)
		case systemconfig.SystemHotelReservation:
			endpoints = hsendpoints.GetEndpointsByService(serviceName)
		case systemconfig.SystemSocialNetwork:
			endpoints = snendpoints.GetEndpointsByService(serviceName)
case systemconfig.SystemOnlineBoutique:
endpoints = obendpoints.GetEndpointsByService(serviceName)
		case systemconfig.SystemOnlineBoutique:
			endpoints = obendpoints.GetEndpointsByService(serviceName)
		}

		switch eps := endpoints.(type) {
		case []tsendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.ServerAddress != "" && ep.ServerAddress != serviceName {
					if pairMap[serviceName] == nil {
						pairMap[serviceName] = make(map[string][]string)
					}
					if ep.SpanName != "" {
						pairMap[serviceName][ep.ServerAddress] = appendUnique(pairMap[serviceName][ep.ServerAddress], ep.SpanName)
					} else {
						if pairMap[serviceName][ep.ServerAddress] == nil {
							pairMap[serviceName][ep.ServerAddress] = []string{}
						}
					}
				}
			}
		case []oteldemoendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.ServerAddress != "" && ep.ServerAddress != serviceName {
					if pairMap[serviceName] == nil {
						pairMap[serviceName] = make(map[string][]string)
					}
					if ep.SpanName != "" {
						pairMap[serviceName][ep.ServerAddress] = appendUnique(pairMap[serviceName][ep.ServerAddress], ep.SpanName)
					} else {
						if pairMap[serviceName][ep.ServerAddress] == nil {
							pairMap[serviceName][ep.ServerAddress] = []string{}
						}
					}
				}
			}
		case []mediaendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.ServerAddress != "" && ep.ServerAddress != serviceName {
					if pairMap[serviceName] == nil {
						pairMap[serviceName] = make(map[string][]string)
					}
					if ep.SpanName != "" {
						pairMap[serviceName][ep.ServerAddress] = appendUnique(pairMap[serviceName][ep.ServerAddress], ep.SpanName)
					} else {
						if pairMap[serviceName][ep.ServerAddress] == nil {
							pairMap[serviceName][ep.ServerAddress] = []string{}
						}
					}
				}
			}
		case []hsendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.ServerAddress != "" && ep.ServerAddress != serviceName {
					if pairMap[serviceName] == nil {
						pairMap[serviceName] = make(map[string][]string)
					}
					if ep.SpanName != "" {
						pairMap[serviceName][ep.ServerAddress] = appendUnique(pairMap[serviceName][ep.ServerAddress], ep.SpanName)
					} else {
						if pairMap[serviceName][ep.ServerAddress] == nil {
							pairMap[serviceName][ep.ServerAddress] = []string{}
						}
					}
				}
			}
		case []snendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.ServerAddress != "" && ep.ServerAddress != serviceName {
					if pairMap[serviceName] == nil {
						pairMap[serviceName] = make(map[string][]string)
					}
					if ep.SpanName != "" {
						pairMap[serviceName][ep.ServerAddress] = appendUnique(pairMap[serviceName][ep.ServerAddress], ep.SpanName)
					} else {
						if pairMap[serviceName][ep.ServerAddress] == nil {
							pairMap[serviceName][ep.ServerAddress] = []string{}
						}
					}
				}
			}
		case []obendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.ServerAddress != "" && ep.ServerAddress != serviceName {
					if pairMap[serviceName] == nil {
						pairMap[serviceName] = make(map[string][]string)
					}
					if ep.SpanName != "" {
						pairMap[serviceName][ep.ServerAddress] = appendUnique(pairMap[serviceName][ep.ServerAddress], ep.SpanName)
					} else {
						if pairMap[serviceName][ep.ServerAddress] == nil {
							pairMap[serviceName][ep.ServerAddress] = []string{}
						}
					}
				}
			}
		}
	}

	// Convert to result
	result := make([]resourcelookup.AppNetworkPair, 0)
	for source, targets := range pairMap {
		for target, spanNames := range targets {
			result = append(result, resourcelookup.AppNetworkPair{
				SourceService: source,
				TargetService: target,
				SpanNames:     spanNames,
			})
		}
	}
	return result, nil
}

func getDNSEndpointsForCurrentSystem() ([]resourcelookup.AppDNSPair, error) {
	// Build DNS pairs from service endpoints (similar to network pairs)
	// Note: DNS chaos does NOT work for gRPC-only connections, so we filter those out
	// We use grpcoperations data to identify which service pairs only use gRPC
	var services []string
	switch systemconfig.GetCurrentSystem() {
	case systemconfig.SystemTrainTicket:
		services = tsendpoints.GetAllServices()
	case systemconfig.SystemOtelDemo:
		services = oteldemoendpoints.GetAllServices()
	case systemconfig.SystemMediaMicroservices:
		services = mediaendpoints.GetAllServices()
	case systemconfig.SystemHotelReservation:
		services = hsendpoints.GetAllServices()
	case systemconfig.SystemSocialNetwork:
		services = snendpoints.GetAllServices()
	case systemconfig.SystemOnlineBoutique:
		services = obendpoints.GetAllServices()
	default:
		return resourcelookup.GetAllDNSEndpoints()
	}

	// Build gRPC-only pairs set using grpcoperations data
	grpcOnlyPairs := buildGRPCOnlyPairsForFaultpoints()

	// Build unique source->domain pairs
	pairMap := make(map[string]map[string][]string) // source -> domain -> spanNames

	for _, serviceName := range services {
		var endpoints interface{}
		switch systemconfig.GetCurrentSystem() {
		case systemconfig.SystemTrainTicket:
			endpoints = tsendpoints.GetEndpointsByService(serviceName)
		case systemconfig.SystemOtelDemo:
			endpoints = oteldemoendpoints.GetEndpointsByService(serviceName)
		case systemconfig.SystemMediaMicroservices:
			endpoints = mediaendpoints.GetEndpointsByService(serviceName)
		case systemconfig.SystemHotelReservation:
			endpoints = hsendpoints.GetEndpointsByService(serviceName)
		case systemconfig.SystemSocialNetwork:
			endpoints = snendpoints.GetEndpointsByService(serviceName)
		case systemconfig.SystemOnlineBoutique:
			endpoints = obendpoints.GetEndpointsByService(serviceName)
		}

		switch eps := endpoints.(type) {
		case []tsendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.ServerAddress != "" && ep.ServerAddress != serviceName {
					if pairMap[serviceName] == nil {
						pairMap[serviceName] = make(map[string][]string)
					}
					if ep.SpanName != "" {
						pairMap[serviceName][ep.ServerAddress] = appendUnique(pairMap[serviceName][ep.ServerAddress], ep.SpanName)
					} else {
						if pairMap[serviceName][ep.ServerAddress] == nil {
							pairMap[serviceName][ep.ServerAddress] = []string{}
						}
					}
				}
			}
		case []oteldemoendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.ServerAddress != "" && ep.ServerAddress != serviceName {
					if pairMap[serviceName] == nil {
						pairMap[serviceName] = make(map[string][]string)
					}
					if ep.SpanName != "" {
						pairMap[serviceName][ep.ServerAddress] = appendUnique(pairMap[serviceName][ep.ServerAddress], ep.SpanName)
					} else {
						if pairMap[serviceName][ep.ServerAddress] == nil {
							pairMap[serviceName][ep.ServerAddress] = []string{}
						}
					}
				}
			}
		case []mediaendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.ServerAddress != "" && ep.ServerAddress != serviceName {
					if pairMap[serviceName] == nil {
						pairMap[serviceName] = make(map[string][]string)
					}
					if ep.SpanName != "" {
						pairMap[serviceName][ep.ServerAddress] = appendUnique(pairMap[serviceName][ep.ServerAddress], ep.SpanName)
					} else {
						if pairMap[serviceName][ep.ServerAddress] == nil {
							pairMap[serviceName][ep.ServerAddress] = []string{}
						}
					}
				}
			}
		case []hsendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.ServerAddress != "" && ep.ServerAddress != serviceName {
					if pairMap[serviceName] == nil {
						pairMap[serviceName] = make(map[string][]string)
					}
					if ep.SpanName != "" {
						pairMap[serviceName][ep.ServerAddress] = appendUnique(pairMap[serviceName][ep.ServerAddress], ep.SpanName)
					} else {
						if pairMap[serviceName][ep.ServerAddress] == nil {
							pairMap[serviceName][ep.ServerAddress] = []string{}
						}
					}
				}
			}
		case []snendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.ServerAddress != "" && ep.ServerAddress != serviceName {
					if pairMap[serviceName] == nil {
						pairMap[serviceName] = make(map[string][]string)
					}
					if ep.SpanName != "" {
						pairMap[serviceName][ep.ServerAddress] = appendUnique(pairMap[serviceName][ep.ServerAddress], ep.SpanName)
					} else {
						if pairMap[serviceName][ep.ServerAddress] == nil {
							pairMap[serviceName][ep.ServerAddress] = []string{}
						}
					}
				}
			}
		case []obendpoints.ServiceEndpoint:
			for _, ep := range eps {
				if ep.ServerAddress != "" && ep.ServerAddress != serviceName {
					if pairMap[serviceName] == nil {
						pairMap[serviceName] = make(map[string][]string)
					}
					if ep.SpanName != "" {
						pairMap[serviceName][ep.ServerAddress] = appendUnique(pairMap[serviceName][ep.ServerAddress], ep.SpanName)
					} else {
						if pairMap[serviceName][ep.ServerAddress] == nil {
							pairMap[serviceName][ep.ServerAddress] = []string{}
						}
					}
				}
			}
		}
	}

	// Convert to result, filtering out gRPC-only connections
	result := make([]resourcelookup.AppDNSPair, 0)
	for appName, domains := range pairMap {
		for domain, spanNames := range domains {
			// Check if this service pair is gRPC-only
			pairKey := appName + "->" + domain
			if grpcOnlyPairs[pairKey] {
				// Skip gRPC-only connections - DNS chaos doesn't work for them
				continue
			}
			result = append(result, resourcelookup.AppDNSPair{
				AppName:   appName,
				Domain:    domain,
				SpanNames: spanNames,
			})
		}
	}
	return result, nil
}

// buildGRPCOnlyPairsForFaultpoints builds a set of service pairs that only communicate via gRPC
// Returns a map where key is "source->target" and value is true if gRPC-only
func buildGRPCOnlyPairsForFaultpoints() map[string]bool {
	grpcOnlyPairs := make(map[string]bool)
	
	// Only OtelDemo has gRPC operations
	if systemconfig.GetCurrentSystem() != systemconfig.SystemOtelDemo {
		return grpcOnlyPairs
	}
	
	// Get all gRPC client operations (these represent outgoing gRPC calls)
	grpcOps := oteldemogrpc.GetClientOperations()
	
	// Track which service pairs have gRPC connections
	grpcPairs := make(map[string]bool)
	for _, op := range grpcOps {
		pairKey := op.ServiceName + "->" + op.ServerAddress
		grpcPairs[pairKey] = true
	}
	
	// Get all service endpoints to check which pairs also have HTTP
	services := oteldemoendpoints.GetAllServices()
	httpPairs := make(map[string]bool)
	
	for _, serviceName := range services {
		endpoints := oteldemoendpoints.GetEndpointsByService(serviceName)
		for _, endpoint := range endpoints {
			// HTTP endpoints have non-empty Route that doesn't look like gRPC
			if endpoint.ServerAddress != "" && endpoint.ServerAddress != serviceName {
				if endpoint.Route != "" && !grpcoperations.IsGRPCRoutePattern(endpoint.Route) {
					pairKey := serviceName + "->" + endpoint.ServerAddress
					httpPairs[pairKey] = true
				}
			}
		}
	}
	
	// A pair is gRPC-only if it has gRPC but no HTTP
	for pair := range grpcPairs {
		if !httpPairs[pair] {
			grpcOnlyPairs[pair] = true
		}
	}
	
	return grpcOnlyPairs
}

func getJVMMethodsForCurrentSystem() ([]resourcelookup.AppMethodPair, error) {
	var services []string
	switch systemconfig.GetCurrentSystem() {
	case systemconfig.SystemTrainTicket:
		services = tsjvm.GetAllServices()
	case systemconfig.SystemOtelDemo:
		services = oteldemojvm.GetAllServices()
	case systemconfig.SystemMediaMicroservices, systemconfig.SystemHotelReservation, systemconfig.SystemSocialNetwork:
		// DeathStarBench systems don't use JVM - return empty list
		return []resourcelookup.AppMethodPair{}, nil
	default:
		return []resourcelookup.AppMethodPair{}, nil
	}

	result := make([]resourcelookup.AppMethodPair, 0)
	for _, serviceName := range services {
		switch systemconfig.GetCurrentSystem() {
		case systemconfig.SystemTrainTicket:
			methods := tsjvm.GetClassMethodsByService(serviceName)
			for _, m := range methods {
				result = append(result, resourcelookup.AppMethodPair{
					AppName:    serviceName,
					ClassName:  m.ClassName,
					MethodName: m.MethodName,
				})
			}
		case systemconfig.SystemOtelDemo:
			methods := oteldemojvm.GetClassMethodsByService(serviceName)
			for _, m := range methods {
				result = append(result, resourcelookup.AppMethodPair{
					AppName:    serviceName,
					ClassName:  m.ClassName,
					MethodName: m.MethodName,
				})
			}
		}
	}
	return result, nil
}

func getDatabaseOperationsForCurrentSystem() ([]resourcelookup.AppDatabasePair, error) {
	// Note: DB chaos only supports MySQL, so we filter to only return MySQL operations
	var services []string
	switch systemconfig.GetCurrentSystem() {
	case systemconfig.SystemTrainTicket:
		services = tsdb.GetAllDatabaseServices()
	case systemconfig.SystemOtelDemo:
		services = oteldemodb.GetAllDatabaseServices()
	case systemconfig.SystemMediaMicroservices:
		services = mediadb.GetAllDatabaseServices()
	case systemconfig.SystemHotelReservation:
		services = hsdb.GetAllDatabaseServices()
	case systemconfig.SystemSocialNetwork:
		services = sndb.GetAllDatabaseServices()
	default:
		return resourcelookup.GetAllDatabaseOperations()
	}

	result := make([]resourcelookup.AppDatabasePair, 0)
	for _, serviceName := range services {
		switch systemconfig.GetCurrentSystem() {
		case systemconfig.SystemTrainTicket:
			ops := tsdb.GetOperationsByService(serviceName)
			for _, op := range ops {
				// Only include MySQL operations (DB chaos only supports MySQL)
				if op.DBSystem == "mysql" {
					result = append(result, resourcelookup.AppDatabasePair{
						AppName:       serviceName,
						DBName:        op.DBName,
						TableName:     op.DBTable,
						OperationType: op.Operation,
					})
				}
			}
		case systemconfig.SystemOtelDemo:
			ops := oteldemodb.GetOperationsByService(serviceName)
			for _, op := range ops {
				// Only include MySQL operations (DB chaos only supports MySQL)
				if op.DBSystem == "mysql" {
					result = append(result, resourcelookup.AppDatabasePair{
						AppName:       serviceName,
						DBName:        op.DBName,
						TableName:     op.DBTable,
						OperationType: op.Operation,
					})
				}
			}
		case systemconfig.SystemMediaMicroservices:
			ops := mediadb.GetOperationsByService(serviceName)
			for _, op := range ops {
				// Only include MySQL operations (DB chaos only supports MySQL)
				if op.DBSystem == "mysql" {
					result = append(result, resourcelookup.AppDatabasePair{
						AppName:       serviceName,
						DBName:        op.DBName,
						TableName:     op.DBTable,
						OperationType: op.Operation,
					})
				}
			}
		case systemconfig.SystemHotelReservation:
			ops := hsdb.GetOperationsByService(serviceName)
			for _, op := range ops {
				// Only include MySQL operations (DB chaos only supports MySQL)
				if op.DBSystem == "mysql" {
					result = append(result, resourcelookup.AppDatabasePair{
						AppName:       serviceName,
						DBName:        op.DBName,
						TableName:     op.DBTable,
						OperationType: op.Operation,
					})
				}
			}
		case systemconfig.SystemSocialNetwork:
			ops := sndb.GetOperationsByService(serviceName)
			for _, op := range ops {
				// Only include MySQL operations (DB chaos only supports MySQL)
				if op.DBSystem == "mysql" {
					result = append(result, resourcelookup.AppDatabasePair{
						AppName:       serviceName,
						DBName:        op.DBName,
						TableName:     op.DBTable,
						OperationType: op.Operation,
					})
				}
			}
		}
	}
	return result, nil
}

func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}
