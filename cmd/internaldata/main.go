package main

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/CUHK-SE-Group/chaos-experiment/handler"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/networkdependencies"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/serviceendpoints"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "list-services":
		listNetworkServices()
	case "list-dependencies":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a service name")
			return
		}
		listServiceDependencies(os.Args[2])
	case "list-all-dependencies":
		listAllDependencies()
	case "list-jvm-methods":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a service name")
			return
		}
		listJVMMethods(os.Args[2])
	case "list-jvm-services":
		listJVMServices()
	case "list-endpoints":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a service name")
			return
		}
		listServiceEndpoints(os.Args[2])
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  cli list-services                - List all services with network dependencies")
	fmt.Println("  cli list-dependencies <service>  - List dependencies for a specific service")
	fmt.Println("  cli list-all-dependencies        - List all service dependencies")
	fmt.Println("  cli list-jvm-methods <service>   - List JVM methods for a specific service")
	fmt.Println("  cli list-jvm-services            - List all Java services")
	fmt.Println("  cli list-endpoints <service>     - List endpoints for a specific service")
}

func listNetworkServices() {
	services := handler.ListNetworkServiceNames()

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
	dependencies := handler.GetNetworkDependencies(serviceName)

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
	methods := handler.GetJVMMethodsForApp(serviceName)

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
	services := handler.ListJVMServiceNames()

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

	fmt.Printf("Endpoints for service %s:\n", serviceName)

	// Create a tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Method\tRoute\tTarget Address\tTarget Port\tResponse Status")
	fmt.Fprintln(w, "------\t-----\t-------------\t-----------\t--------------")

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
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			method,
			route,
			endpoint.ServerAddress,
			endpoint.ServerPort,
			status)
	}

	w.Flush()
	fmt.Printf("Total: %d endpoints\n", len(endpoints))
}
