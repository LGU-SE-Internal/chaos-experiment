package javaanalyzer

import (
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
)

// ClassMethodEntry represents a class-method pair from the Java analysis
type ClassMethodEntry struct {
    ClassName  string `json:"className"`
    MethodName string `json:"methodName"`
}

// PathResult represents the results for a specific path
type PathResult struct {
    PathName string            `json:"pathName"`
    Methods  []ClassMethodEntry `json:"methods"`
}

// AnalyzeJavaPath analyzes a single Java source path and returns the method entries
func AnalyzeJavaPath(sourcePath string, jarPath string) ([]ClassMethodEntry, error) {
    // Create a temporary file for output
    tempFile, err := os.CreateTemp("", "java-analysis-*.json")
    if err != nil {
        return nil, fmt.Errorf("failed to create temp file: %w", err)
    }
    tempFileName := tempFile.Name()
    tempFile.Close()
    defer os.Remove(tempFileName)

    // Run the Java analyzer
    cmd := exec.Command("java", "-jar", jarPath, sourcePath, tempFileName)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("error running Java analyzer: %w", err)
    }

    // Read the results
    data, err := os.ReadFile(tempFileName)
    if err != nil {
        return nil, fmt.Errorf("error reading output file: %w", err)
    }

    var entries []ClassMethodEntry
    if err := json.Unmarshal(data, &entries); err != nil {
        return nil, fmt.Errorf("error parsing JSON: %w", err)
    }

    return entries, nil
}

// AnalyzeJavaPaths analyzes multiple Java source paths and returns path results
func AnalyzeJavaPaths(sourcePaths []string) ([]PathResult, error) {
    // Absolute path to the analyzer JAR
    jarPath := "tools/javaanalyzer/method-extractor.jar"

    // Check if the JAR exists
    if _, err := os.Stat(jarPath); os.IsNotExist(err) {
        return nil, fmt.Errorf("analyzer JAR not found at %s, please build with 'mvn package'", jarPath)
    }

    var results []PathResult

    for _, path := range sourcePaths {
        if path == "" {
            continue
        }

        // Get the last directory name from the path
        pathName := filepath.Base(path)

        // Analyze the path
        entries, err := AnalyzeJavaPath(path, jarPath)
        if err != nil {
            return nil, fmt.Errorf("error analyzing path %s: %w", path, err)
        }

        // Add to results
        result := PathResult{
            PathName: pathName,
            Methods:  entries,
        }
        results = append(results, result)
    }

    return results, nil
}

// SaveResultsToFile saves the analysis results to the specified JSON file
func SaveResultsToFile(results []PathResult, outputFile string) error {
    // Create parent directory if it doesn't exist
    dir := filepath.Dir(outputFile)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("error creating output directory: %w", err)
    }

    // Marshal to JSON
    resultJSON, err := json.MarshalIndent(results, "", "  ")
    if err != nil {
        return fmt.Errorf("error creating JSON: %w", err)
    }

    // Write to file
    if err := os.WriteFile(outputFile, resultJSON, 0644); err != nil {
        return fmt.Errorf("error writing output file: %w", err)
    }

    return nil
}