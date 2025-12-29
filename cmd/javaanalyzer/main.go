package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LGU-SE-Internal/chaos-experiment/internal/systemconfig"
	tools "github.com/LGU-SE-Internal/chaos-experiment/tools/javaanalyzer"
)

func main() {
	// Define command-line flags
	servicesPath := flag.String("services", "", "Path to the Java services directory")
	outputPath := flag.String("output", "", "Path for the generated Go file")
	system := flag.String("system", "ts", "Target system: 'ts' (TrainTicket) or 'otel-demo' (OpenTelemetry Demo)")
	flag.Parse()

	// Validate and set the system type using systemconfig
	systemType, err := systemconfig.ParseSystemType(*system)
	if err != nil {
		fmt.Printf("Invalid system: %s. Must be 'ts' or 'otel-demo'\n", *system)
		os.Exit(1)
	}
	if err := systemconfig.SetCurrentSystem(systemType); err != nil {
		fmt.Printf("Error setting system type: %v\n", err)
		os.Exit(1)
	}

	// Validate services path
	if *servicesPath == "" {
		fmt.Println("Error: services path is required")
		flag.Usage()
		os.Exit(1)
	}

	// Set default output path based on system type
	if *outputPath == "" {
		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error determining project root: %v\n", err)
			os.Exit(1)
		}

		// Determine system-specific subdirectory
		var systemDir string
		switch systemType {
		case systemconfig.SystemTrainTicket:
			systemDir = "ts"
		case systemconfig.SystemOtelDemo:
			systemDir = "oteldemo"
		case systemconfig.SystemSockShop:
			systemDir = "sockshop"
		case systemconfig.SystemTeaStore:
			systemDir = "teastore"
		default:
			systemDir = string(systemType)
		}

		*outputPath = filepath.Join(projectRoot, "internal", systemDir, "javaclassmethods", "javaclassmethods.go")
	}

	// Generate the Java class methods file
	fmt.Printf("Analyzing system: %s\n", systemconfig.GetCurrentSystem())
	fmt.Printf("Services path: %s\n", *servicesPath)
	fmt.Printf("Output path: %s\n", *outputPath)

	if err := tools.GenerateJavaClassMethodsFile(*servicesPath, *outputPath); err != nil {
		fmt.Printf("Error generating Java class methods file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Java class methods file generated successfully!")
}
