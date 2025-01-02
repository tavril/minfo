package main

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// runCommand runs a command and returns the output
func runCommand(cmd *exec.Cmd) (string, error) {
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

// capitalizeFirstLetter capitalizes the first letter of a string
func capitalizeFirstLetter(s string) string {
	if len(s) == 0 {
		return s // Return an empty string if input is empty
	}
	return strings.ToUpper(string(s[0])) + s[1:]
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

// This function rounds a float to the nearest half.
// (Using this to display temperatures in Celsius to nearest 0.5)
func roundToNearestHalf(x float64) float64 {
	return math.Round(x*2) / 2
}

// FormatFloat intelligently formats float64 to display one decimal if needed
func formatFloat(x float64) string {
	// Format with one decimal place
	s := fmt.Sprintf("%.1f", x)

	// Remove unnecessary trailing ".0"
	if s[len(s)-2:] == ".0" {
		return strconv.FormatFloat(x, 'f', 0, 64)
	}
	return s
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

func ensureDirExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check directory: %v", err)
	}
	return nil
}

func windArrow(deg int) string {
	arrows := []string{"↓", "↙", "←", "↖", "↑", "↗", "→", "↘"}
	return arrows[((deg+22)%360)/45]
}
