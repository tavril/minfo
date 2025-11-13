package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadCacheFileEmpty(t *testing.T) {
	t.Parallel()

	tempFile := filepath.Join(t.TempDir(), "cache-empty.json")
	if err := os.WriteFile(tempFile, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	var out info
	err := readCacheFile(tempFile, &out)
	if !errors.Is(err, errEmptyCache) {
		t.Fatalf("expected errEmptyCache, got %v", err)
	}
}

func TestWriteAndReadCacheFileRoundTrip(t *testing.T) {
	t.Parallel()

	tempFile := filepath.Join(t.TempDir(), "cache.json")
	gpuCores := 10
	expected := info{
		cachedInfo: cachedInfo{
			Model: &Model{
				Name:    "MacBook Pro",
				SubName: "16-inch",
				Date:    "Nov 2024",
				Number:  "ABC123",
			},
			Cpu: &Cpu{
				Model:            "Apple M4 Max",
				Cores:            16,
				PerformanceCores: 12,
				EfficiencyCores:  4,
			},
			GpuCores:     &gpuCores,
			Memory:       &Memory{Amount: 64, Unit: "GB", MemType: "LPDDR5"},
			SerialNumber: func() *string { s := "SERIAL"; return &s }(),
		},
	}

	if err := writeCacheFile(tempFile, &expected); err != nil {
		t.Fatalf("writeCacheFile failed: %v", err)
	}

	var actual info
	if err := readCacheFile(tempFile, &actual); err != nil {
		t.Fatalf("readCacheFile failed: %v", err)
	}

	if actual.Cpu == nil || actual.Cpu.Model != expected.Cpu.Model {
		t.Fatalf("CPU mismatch: got %+v", actual.Cpu)
	}
	if actual.Model == nil || actual.Model.Name != expected.Model.Name {
		t.Fatalf("Model mismatch: got %+v", actual.Model)
	}
	if actual.GpuCores == nil || *actual.GpuCores != *expected.GpuCores {
		t.Fatalf("GPU cores mismatch: got %v", actual.GpuCores)
	}
	if actual.Memory == nil || actual.Memory.Amount != expected.Memory.Amount {
		t.Fatalf("Memory mismatch: got %+v", actual.Memory)
	}
	if actual.SerialNumber == nil || *actual.SerialNumber != *expected.SerialNumber {
		t.Fatalf("Serial mismatch: got %v", actual.SerialNumber)
	}
}

func TestIsFileOlderThan(t *testing.T) {
	t.Parallel()

	oldFile := filepath.Join(t.TempDir(), "old-weather.json")
	if err := os.WriteFile(oldFile, []byte("old"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	oldTime := time.Now().Add(-2 * weatherCacheDuration)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("failed to touch file: %v", err)
	}

	isOld, err := isFileOlderThan(oldFile, weatherCacheDuration)
	if err != nil {
		t.Fatalf("isFileOlderThan returned error: %v", err)
	}
	if !isOld {
		t.Fatalf("expected file to be considered old")
	}

	newFile := filepath.Join(t.TempDir(), "new-weather.json")
	if err := os.WriteFile(newFile, []byte("new"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	isOld, err = isFileOlderThan(newFile, weatherCacheDuration)
	if err != nil {
		t.Fatalf("isFileOlderThan returned error: %v", err)
	}
	if isOld {
		t.Fatalf("expected file to be considered fresh")
	}
}
