package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tools "github.com/LGU-SE-Internal/chaos-experiment/tools/javaanalyzer"
)

func main() {
	// Define command-line flags
	servicesPath := flag.String("services", "", "Path to the Java services directory")
	outputPath := flag.String("output", "", "Path for the generated Go file (default: internal/javaclassmethods/javaclassmethods.go)")
	flag.Parse()

	// Validate services path
	if *servicesPath == "" {
		fmt.Println("Error: services path is required")
		flag.Usage()
		os.Exit(1)
	}

	// Set default output path if not specified
	if *outputPath == "" {
		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error determining project root: %v\n", err)
			os.Exit(1)
		}
		*outputPath = filepath.Join(projectRoot, "internal", "javaclassmethods", "javaclassmethods.go")
	}

	// Generate the Java class methods file
	fmt.Printf("Generating Java class methods file...\n")
	fmt.Printf("Services path: %s\n", *servicesPath)
	fmt.Printf("Output path: %s\n", *outputPath)

	if err := tools.GenerateJavaClassMethodsFile(*servicesPath, *outputPath); err != nil {
		fmt.Printf("Error generating Java class methods file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Java class methods file generated successfully!")
}
