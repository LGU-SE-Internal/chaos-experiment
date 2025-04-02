# Internal Data Generator

## Analyzing Java Services
```bash
# Build the generator
go build -o bin/generate-java-methods cmd/javaanalyzer/main.go

# Run the generator with path to Java services
./bin/generate-java-methods --services /path/to/java/services
```

This will generate a file in `internal/javaclassmethods/javaclassmethods.go` with all method information.

## Analyzing service endpoint

```bash
go run cmd/clickhouseanalyzer/main.go --host=10.10.10.58 --username=default --password=password
```
This will generate a file in `internal/serviceendpoints/serviceendpoints.go` with all service endpoint information.


# Example

## Set the namespace and appList to inject chaos
```go
namespace := "onlineboutique"
appList := []string{"checkoutservice", "recommendationservice", "emailservice", "paymentservice", "productcatalogservice"}
```

## Single chaos
- NetworkChaos
    ```go
    appName := "checkoutservice"
    // Example: simple network delay chaos
    controllers.CreateNetworkDelayChaos(k8sClient, namespace, appName, "100ms", "25", "10ms", pointer.String("2m"))

	// Example: Using the simpler helper with target/direction together
	controllers.CreateNetworkDelayChaos(k8sClient, namespace, appName, "100ms", "25", "10ms", pointer.String("2m"),
	   chaos.WithNetworkTargetAndDirection(namespace, "productcatalogservice", v1alpha1.Both))

	// Example: Create network partition with additional options
	controllers.CreateNetworkPartitionChaos(k8sClient, namespace, appName, "productcatalogservice",
	   v1alpha1.Both, pointer.String("3m"),
	   chaos.WithNetworkDevice("eth0")) // Specify network device
    ```
- DNSChaos
    ```go
    appName := "checkoutservice"
	controllers.CreateDnsChaos(k8sClient, namespace, appName, "error", []string{"*"}, pointer.String("2m"))
    controllers.CreateDnsChaos(k8sClient, namespace, appName, "random", []string{"*"}, pointer.String("2m"))
    ```

## JVM Chaos



- JVM Latency Injection
    ```go
    appName := "ts-user-service"
    // Get a dynamic method by index
    controllers.CreateJVMChaos(k8sClient, namespace, appName, 
        chaosmeshv1alpha1.JVMLatencyAction, pointer.String("2m"),
        chaos.WithJVMClass("com.example.UserService"),
        chaos.WithJVMMethod("getUserById"),
        chaos.WithJVMLatencyDuration(1000))
    ```

- JVM Exception Injection
    ```go
    appName := "ts-order-service"
    controllers.CreateJVMChaos(k8sClient, namespace, appName, 
        chaosmeshv1alpha1.JVMExceptionAction, pointer.String("2m"),
        chaos.WithJVMClass("com.example.OrderService"),
        chaos.WithJVMMethod("createOrder"),
        chaos.WithJVMDefaultException())
    ```

- JVM MySQL Latency
    ```go
    appName := "ts-order-service"
    controllers.CreateJVMChaos(k8sClient, namespace, appName, 
        chaosmeshv1alpha1.JVMMySQLAction, pointer.String("2m"),
        chaos.WithJVMMySQLConnector("5"),
        chaos.WithJVMMySQLDatabase("ts"),
        chaos.WithJVMMySQLTable("orders"),
        chaos.WithJVMMySQLType("select"),
        chaos.WithJVMLatencyDuration(1000))
    ```

## Schedule chaos
- StressChaos
    ```go
    stressors := controllers.MakeCPUStressors(100, 5)
    controllers.ScheduleStressChaos(k8sClient, namespace, appList, stressors, "cpu")
    ```
- PodChaos
    ```go
	action := chaosmeshv1alpha1.PodFailureAction
	controllers.SchedulePodChaos(k8sClient, namespace, appList, action)
    ```
- HTTPChaos
    - abort
        ```go
        abort := true
        opts := []chaos.OptHTTPChaos{
            chaos.WithTarget(chaosmeshv1alpha1.PodHttpRequest),
            chaos.WithPort(8080),
            chaos.WithAbort(&abort),
        }
        controllers.ScheduleHTTPChaos(k8sClient, namespace, appList, "request-abort", opts...)
        ```
    - replace
        ```go
        opts := []chaos.OptHTTPChaos{
            chaos.WithTarget(chaosmeshv1alpha1.PodHttpResponse),
            chaos.WithPort(8080),
            chaos.WithReplaceBody([]byte(rand.String(6))),
        }
        controllers.ScheduleHTTPChaos(k8sClient, namespace, appList, "Response-replace", opts...)
        ```

## workflow

```go
namespace := "ts"

appList := []string{"ts-consign-service", "ts-route-service", "ts-train-service", "ts-travel-service", "ts-basic-service", "ts-food-service", "ts-security-service", "ts-seat-service", "ts-routeplan-service", "ts-travel2-service"}
workflowSpec := controllers.NewWorkflowSpec(namespace)
sleepTime := pointer.String("15m")
injectTime := pointer.String("5m")
// Add cpu
stressors := controllers.MakeCPUStressors(100, 5)
controllers.AddStressChaosWorkflowNodes(workflowSpec, namespace, appList, stressors, "cpu", injectTime, sleepTime)
// Add memory
stressors = controllers.MakeMemoryStressors("1GB", 1)
controllers.AddStressChaosWorkflowNodes(workflowSpec, namespace, appList, stressors, "memory", injectTime, sleepTime)
// Add Pod failure
action := chaosmeshv1alpha1.PodFailureAction
controllers.AddPodChaosWorkflowNodes(workflowSpec, namespace, appList, action, injectTime, sleepTime)
// Add abort
abort := true
opts1 := []chaos.OptHTTPChaos{
    chaos.WithTarget(chaosmeshv1alpha1.PodHttpRequest),
    chaos.WithPort(8080),
    chaos.WithAbort(&abort),
}
controllers.AddHTTPChaosWorkflowNodes(workflowSpec, namespace, appList, "request-abort", injectTime, sleepTime, opts1...)
// add replace
opts2 := []chaos.OptHTTPChaos{
    chaos.WithTarget(chaosmeshv1alpha1.PodHttpResponse),
    chaos.WithPort(8080),
    chaos.WithReplaceBody([]byte(rand.String(6))),
}
controllers.AddHTTPChaosWorkflowNodes(workflowSpec, namespace, appList, "response-replace", injectTime, sleepTime, opts2...)
// create workflow
controllers.CreateWorkflow(k8sClient, workflowSpec, namespace)
```

## JVM Method Extraction API

The package provides an API to access extracted Java method information:

```go
// Get all methods for a service name
methods := handler.GetJVMMethods("user-service")

// Get all services names
serviceNames := handler.ListJVMServiceNames()

// Get methods for a specific app
methods := handler.GetJVMMethodsForApp("ts-user-service")
```