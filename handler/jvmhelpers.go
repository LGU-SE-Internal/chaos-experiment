package handler

import (
	"github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/javaclassmethods"
)

// Add a package-level variable for dependency injection
var labelsGetter = client.GetLabels

// selectJVMMethodForService selects a method for a service based on the method index
// and returns the class name, method name, and whether the selection was successful
func selectJVMMethodForService(appName string, methodIndex int) (className, methodName string, ok bool) {
	serviceName := appName
	methodEntry := javaclassmethods.GetMethodByIndexOrRandom(serviceName, methodIndex)

	if methodEntry == nil {
		return "", "", false
	}

	return methodEntry.ClassName, methodEntry.MethodName, true
}

// getAvailableJVMMethodsForApp returns a list of available methods for an app
func getAvailableJVMMethodsForApp(appName string) []string {
	return javaclassmethods.ListAvailableMethods(appName)
}

// getServiceAndMethodForChaosSpec is a helper function that retrieves the label array,
// selects the app name, and finds an appropriate method for a JVM chaos spec
func getServiceAndMethodForChaosSpec(appNameIndex int, methodIndex int) (appName, className, methodName string, ok bool) {
	// Get the app labels using labelsGetter instead of client.GetLabels
	labelArr, err := labelsGetter(TargetNamespace, TargetLabelKey)
	if err != nil || appNameIndex < 0 || appNameIndex >= len(labelArr) {
		return "", "", "", false
	}

	appName = labelArr[appNameIndex]
	className, methodName, ok = selectJVMMethodForService(appName, methodIndex)
	return appName, className, methodName, ok
}

// GetJVMMethods returns all available JVM methods for the given service name
// This function can be exposed as an API to external packages
func GetJVMMethods(serviceName string) []javaclassmethods.ClassMethodEntry {
	return javaclassmethods.GetClassMethodsByService(serviceName)
}

// GetJVMMethodsForApp returns all available JVM methods for the given app name
// This function can be exposed as an API to external packages
func GetJVMMethodsForApp(appName string) []javaclassmethods.ClassMethodEntry {
	serviceName := appName
	return javaclassmethods.GetClassMethodsByService(serviceName)
}

// ListJVMServiceNames returns a list of all available Java service names
func ListJVMServiceNames() []string {
	return javaclassmethods.ListAllServiceNames()
}
