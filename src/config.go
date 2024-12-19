package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config structure to represent the configuration file
type Config struct {
	CacheFilePath string   `yaml:"cache_file"`
	Items         []string `yaml:"items"`
	//NoLogo        bool     `yaml:"no_logo"`
	//NoCache       bool     `yaml:"no_cache"` // don't use cache file
}

var config *Config
var defaultCacheFilePath = fmt.Sprintf("%s/.minfo-cache.json", os.Getenv("HOME"))

// ---------- All the possible items to fetch from system_profiler ---------- //

// First, system_profiler items that can be cached
var spItemsCached = []string{
	"model",
	"cpu",
	"gpu",
	"memory",
}

// Then, system_profiler items that cannot be cached
var spItemsNotCached = []string{
	"user",
	"hostname",
	"os",
	"disk",
	"battery",
	"display",
}
var spItems = append(spItemsCached, spItemsNotCached...)

// All possible items to fetch (including from system_profiler)
var allItems = append(spItems, []string{
	"terminal",
	"software",
	"public_ip",
	"uptime",
	"datetime",
}...)

// By default, the information we fetch from system_profiler
var defaultSpItems = []string{
	"user",
	"hostname",
	"os",
	"model",
	"cpu",
	"gpu",
	"memory",
	"disk",
	"battery",
	"display",
}

// By default, the information we fetch (including from system_profiler)
var defaultItems = append(defaultSpItems, []string{
	"terminal",
	"software",
	"public_ip",
	"uptime",
	"datetime",
}...)

// loadConfig loads and parses the YAML configuration file
func loadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	// Parse the YAML file into the Config structure
	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode YAML file: %w", err)
	}

	return &config, nil
}
