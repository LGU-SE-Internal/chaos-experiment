


# Init

```bash
git submodule update --init --depth 1 --recursive
```

# Example

## Set the namespace and appList to inject chaos
```go
namespace := "onlineboutique"
appList := []string{"checkoutservice", "recommendationservice", "emailservice", "paymentservice", "productcatalogservice"}
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