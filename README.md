


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
