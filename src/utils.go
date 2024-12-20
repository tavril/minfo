package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// runCommand runs a command and returns the output
func runCommand(cmd *exec.Cmd) (string, error) {
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

// This function replicates the functionality of the "which" command
func which(command string) (string, error) {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return "", errors.New("PATH environment variable is empty")
	}
	paths := filepath.SplitList(pathEnv)

	for _, dir := range paths {
		fullPath := filepath.Join(dir, command)

		if fileInfo, err := os.Stat(fullPath); err == nil {
			if !fileInfo.IsDir() && (fileInfo.Mode()&0111 != 0) { // Check for executable bit
				return fullPath, nil
			}
		}
	}

	return "", fmt.Errorf("%s: command not found", command)
}

// Helper to get a line from a 2D slice, or an empty line if out of range
func getLine(lines [][]string, index int) string {
	if index >= 0 && index < len(lines) {
		return strings.Join(lines[index], "")
	}
	return ""
}

// This function returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// This function returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// This functions returns only the unique strings in a slice of strings
func uniqueStrings(input []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range input {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
