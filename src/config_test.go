package main

import (
	"io"
	"os"
	"testing"
)

// Helper function to create temporary configuration files
func createTempConfigFile(content string) (string, error) {
	file, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

// Test valid configuration parsing
func TestLoadConfig_ValidConfig(t *testing.T) {
	content := `
cache_file: /tmp/cache.json
items:
  - public_ip
  - display
  - software
`
	filePath, err := createTempConfigFile(content)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(filePath) // Clean up

	err = loadAndCheckConfig(filePath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if *config.CacheFilePath != "/tmp/cache.json" {
		t.Errorf("Expected CacheFile to be '/tmp/cache.json', got '%s'", *config.CacheFilePath)
	}

	expectedFetchItems := []string{"public_ip", "display", "software"}
	for i, item := range expectedFetchItems {
		if config.Items[i] != item {
			t.Errorf("Expected FetchItems[%d] to be '%s', got '%s'", i, item, config.Items[i])
		}
	}
}

// Test invalid configuration parsing
func TestLoadConfig_InvalidConfig(t *testing.T) {
	content := `
cache_file: /tmp/cache.json
fetch_items: [public_ip, displays, software` // Missing closing bracket

	filePath, err := createTempConfigFile(content)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(filePath) // Clean up

	err = loadAndCheckConfig(filePath)
	if err == nil {
		t.Errorf("Expected an error for invalid YAML, but got nil")
	}
}

// Test empty configuration file
func TestLoadConfig_EmptyConfig(t *testing.T) {
	content := `` // Empty content

	filePath, err := createTempConfigFile(content)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(filePath) // Clean up

	err = loadAndCheckConfig(filePath)
	if err != io.EOF {
		t.Errorf("Unexpected error: %v", err)
	}

}
