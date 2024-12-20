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
	Title          string
	SystemProfiler SystemProfilerItem
	retrieveCmd    NamedFunc
}

// itemsConfig is a map of all the items we can fetch,
// with how to fetch them.
var itemsConfig = map[string]Item{
	"user": {
		Title: "User",
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPSoftwareDataType",
		},
	},
	"hostname": {
		Title: "Hostname",
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPSoftwareDataType",
		},
	},
	"os": {
		Title: "OS",
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPSoftwareDataType",
		},
	},
	"system_integrity": {
		Title: "macOS SIP",
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPSoftwareDataType",
		},
	},
	"serial_number": {
		Title: "Serial",
		SystemProfiler: SystemProfilerItem{
			IsCached: true,
			DataType: "SPHardwareDataType",
		},
	},
	"uptime": {
		Title: "Uptime",
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPSoftwareDataType",
		},
	},
	"model": {
		Title: "Model",
		SystemProfiler: SystemProfilerItem{
			IsCached: true,
			DataType: "SPHardwareDataType",
		},
	},
	"cpu": {
		Title: "CPU",
		SystemProfiler: SystemProfilerItem{
			IsCached: true,
			DataType: "SPHardwareDataType",
		},
	},
	"memory": {
		Title: "Memory",
		SystemProfiler: SystemProfilerItem{
			IsCached: true,
			DataType: "SPMemoryDataType",
		},
	},
	"display": {
		Title: "Display",
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPDisplaysDataType",
		},
	},
	"gpu": {
		Title: "GPU",
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPDisplaysDataType",
		},
	},
	"battery": {
		Title: "Battery",
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPPowerDataType",
		},
	},
	"disk": {
		Title: "Disk",
		SystemProfiler: SystemProfilerItem{
			IsCached: false,
			DataType: "SPStorageDataType",
		},
	},
	"terminal": {
		Title: "Terminal",
		retrieveCmd: NamedFunc{
			Id:   "fetchTermProgram",
			Func: fetchTermProgram,
		},
	},
	"software": {
		Title: "Software",
		retrieveCmd: NamedFunc{
			Id:   "fetchSoftware",
			Func: fetchSoftware,
		},
	},
	"public_ip": {
		Title: "Public IP",
		retrieveCmd: NamedFunc{
			Id:   "fetchPublicIp",
			Func: fetchPublicIp,
		},
	},
	"datetime": {
		Title: "Date/Time",
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
