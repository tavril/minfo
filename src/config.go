package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// This struct represents the configuration file
type Config struct {
	CacheFilePath string   `yaml:"cache_file"`
	Items         []string `yaml:"items"`
}

var config *Config

/* ---------- Default Configuration ---------- */
var defaultCacheFilePath = fmt.Sprintf("%s/.minfo-cache.json", os.Getenv("HOME"))
var defaultItems = []string{
	"user",
	"hostname",
	"os",
	"system_integrity",
	"serial_number",
	"model",
	"cpu",
	"gpu",
	"memory",
	"disk",
	"battery",
	"display",
	"terminal",
	"software",
	"public_ip",
	"uptime",
	"datetime",
}

var defaultConfig = Config{
	CacheFilePath: defaultCacheFilePath,
	Items:         defaultItems,
}

// We use this to easily reference functions in map,
// as func() type cannot be used as a map key, we use
// a string instead.
type NamedFunc struct {
	Id   string
	Func func(*info)
}
type SystemProfilerItem struct {
	IsCached bool   // whether the item can be cached
	DataType string // ex. "SPHardwareDataType"
}

// This struct contains information about how to fetch
// information for a given item
// - SystemProfiler: information to fetch from system_profiler (SPDataType)
// - retrieveCmd: a function to retrieve the information
type Item struct {
	SystemProfiler SystemProfilerItem
	retrieveCmd    NamedFunc
}

// itemsConfig is a map of all the items we can fetch,
// with how to fetch them.
var itemsConfig = map[string]Item{
	"user": {
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPSoftwareDataType",
		},
	},
	"hostname": {
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPSoftwareDataType",
		},
	},
	"os": {
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPSoftwareDataType",
		},
	},
	"system_integrity": {
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPSoftwareDataType",
		},
	},
	"serial_number": {
		SystemProfiler: SystemProfilerItem{
			IsCached: true,
			DataType: "SPHardwareDataType",
		},
	},
	"uptime": {
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPSoftwareDataType",
		},
	},
	"model": {
		SystemProfiler: SystemProfilerItem{
			IsCached: true,
			DataType: "SPHardwareDataType",
		},
	},
	"cpu": {
		SystemProfiler: SystemProfilerItem{
			IsCached: true,
			DataType: "SPHardwareDataType",
		},
	},
	"memory": {
		SystemProfiler: SystemProfilerItem{
			IsCached: true,
			DataType: "SPMemoryDataType",
		},
	},
	"display": {
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPDisplaysDataType",
		},
	},
	"gpu": {
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPDisplaysDataType",
		},
	},
	"battery": {
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPPowerDataType",
		},
	},
	"disk": {
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPStorageDataType",
		},
	},
	"terminal": {
		retrieveCmd: NamedFunc{
			Id:   "fetchTermProgram",
			Func: fetchTermProgram,
		},
	},
	"software": {
		retrieveCmd: NamedFunc{
			Id:   "fetchSoftware",
			Func: fetchSoftware,
		},
	},
	"public_ip": {
		retrieveCmd: NamedFunc{
			Id:   "fetchPublicIp",
			Func: fetchPublicIp,
		},
	},
	"datetime": {
		retrieveCmd: NamedFunc{
			Id:   "fetchDateTime",
			Func: fetchDateTime,
		},
	},
}

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
