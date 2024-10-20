


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
        ```
        abort := true
        opts := []chaos.OptHTTPChaos{
            chaos.WithTarget(chaosmeshv1alpha1.PodHttpRequest),
            chaos.WithPort(8080),
            chaos.WithAbort(&abort),
        }
        controllers.ScheduleHTTPChaos(k8sClient, namespace, appList, "request-abort", opts...)
        ```
    - replace
        ```
        opts := []chaos.OptHTTPChaos{
            chaos.WithTarget(chaosmeshv1alpha1.PodHttpResponse),
            chaos.WithPort(8080),
            chaos.WithReplaceBody([]byte(rand.String(6))),
        }
        controllers.ScheduleHTTPChaos(k8sClient, namespace, appList, "Response-replace", opts...)
        ```