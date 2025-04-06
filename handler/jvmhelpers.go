package handler

import (
	"fmt"

	"github.com/CUHK-SE-Group/chaos-experiment/internal/javaclassmethods"
)

// selectJVMMethodForService selects a method for a service based on the method index
// and returns the class name, method name, or an error if selection fails
func selectJVMMethodForService(serviceName string, methodIndex int) (className, methodName string, err error) {
	entry := javaMethodGetter(serviceName, methodIndex)
	if entry == nil {
		return "", "", fmt.Errorf("no JVM method found for service %s at index %d",
			serviceName, methodIndex)
	}
	return entry.ClassName, entry.MethodName, nil
}

// getAvailableJVMMethodsForApp returns a list of available methods for an app
func getAvailableJVMMethodsForApp(appName string) ([]string, error) {
	methods := javaMethodsLister(appName)
	if len(methods) == 0 {
		return nil, fmt.Errorf("no JVM methods found for app %s", appName)
	}
	return methods, nil
}

// getServiceAndMethodForChaosSpec is a helper function that retrieves the label array,
// selects the app name, and finds an appropriate method for a JVM chaos spec
func getServiceAndMethodForChaosSpec(appNameIndex int, methodIndex int) (appName, className, methodName string, err error) {
	// Get the app labels using labelsGetter
	labelArr, err := labelsGetter(TargetNamespace, TargetLabelKey)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get service labels: %w", err)
	}

	if appNameIndex < 0 || appNameIndex >= len(labelArr) {
		return "", "", "", fmt.Errorf("app index %d out of range (max: %d)",
			appNameIndex, len(labelArr)-1)
	}

	appName = labelArr[appNameIndex]
	className, methodName, err = selectJVMMethodForService(appName, methodIndex)
	if err != nil {
		return appName, "", "", err
	}

	return appName, className, methodName, nil
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
