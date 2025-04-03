package handler

import (
	"github.com/CUHK-SE-Group/chaos-experiment/internal/javaclassmethods"
)

// selectJVMMethodForService selects a method for a service based on the method index
// and returns the class name, method name, and whether the selection was successful
func selectJVMMethodForService(serviceName string, methodIndex int) (className, methodName string, ok bool) {
	entry := javaMethodGetter(serviceName, methodIndex)
	if entry == nil {
		return "", "", false
	}
	return entry.ClassName, entry.MethodName, true
}

// getAvailableJVMMethodsForApp returns a list of available methods for an app
func getAvailableJVMMethodsForApp(appName string) []string {
	return javaMethodsLister(appName)
}

// getServiceAndMethodForChaosSpec is a helper function that retrieves the label array,
// selects the app name, and finds an appropriate method for a JVM chaos spec
func getServiceAndMethodForChaosSpec(appNameIndex int, methodIndex int) (appName, className, methodName string, ok bool) {
	// Get the app labels using labelsGetter
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
	return javaMethodsGetter(serviceName)
}

// GetJVMMethodsForApp returns all available JVM methods for the given app name
// This function can be exposed as an API to external packages
func GetJVMMethodsForApp(appName string) []javaclassmethods.ClassMethodEntry {
	return GetJVMMethods(appName)
}

// ListJVMServiceNames returns a list of all available Java service names
func ListJVMServiceNames() []string {
	return javaServicesGetter()
}
