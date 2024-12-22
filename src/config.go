package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// This struct represents the configuration file
type Config struct {
	CacheFilePath *string  `yaml:"cache_file,omitempty"`
	Items         []string `yaml:"items,omitempty"`
}

var config = &Config{}

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

/* ---------- AVAILABLE ITEMS ---------- */
// For things to be retrieved from system_profiler,
// We need to know the SPDataType to fetch.
var (
	SPSoftwareDataType = "SPSoftwareDataType"
	SPHardwareDataType = "SPHardwareDataType"
	SPMemoryDataType   = "SPMemoryDataType"
	SPDisplaysDataType = "SPDisplaysDataType"
	SPPowerDataType    = "SPPowerDataType"
	SPStorageDataType  = "SPStorageDataType"
)

// For things that are not retrieved from system_profiler,
// we need to know which function to call to get the information.
// We use a NamedFunc struct because we cannot have pointers
// to functions, and the "item" struct must have pointers,
// because its "Func" field is optional (so we need to be able
// to set it to nil).
type NamedFunc struct {
	Id   string
	Func func(*info)
}

var (
	datetimeNamedFunc = NamedFunc{
		Id:   "fetchDateTime",
		Func: fetchDateTime,
	}
	publicIpNamedFunc = NamedFunc{
		Id:   "fetchPublicIp",
		Func: fetchPublicIp,
	}
	softwareNamedFunc = NamedFunc{
		Id:   "fetchSoftware",
		Func: fetchSoftware,
	}
	termProgramNamedFunc = NamedFunc{
		Id:   "fetchTermProgram",
		Func: fetchTermProgram,
	}
)

// Defines an item that can be fetched and displayed.
// Title: The title of the item to display. Ex. "Public IP"
//   - Required
//
// SPDataType: The SPDataType to fetch from system_profiler. Ex. "SPSoftwareDataType"
//   - Optional (i.e. default = nil)
//
// Func: The function to call to get the information.
//   - Optional (i.e. default = nil)
//
// IsCached: Whether the information should be cached or not.
//   - Optional (i.e. default = false)
type item struct {
	Title      string
	SPDataType *string
	Func       *NamedFunc
	IsCached   bool
}

// All available items we can fetch and display.
var availableItems = map[string]item{
	/* ---------- System Profiler Data (cached data) ---------- */
	"cpu": {
		Title:      "CPU",
		SPDataType: &SPHardwareDataType,
		IsCached:   true,
	},
	"gpu": {
		Title:      "GPU",
		SPDataType: &SPDisplaysDataType,
		IsCached:   true,
	},
	"model": {
		Title:      "Model",
		SPDataType: &SPHardwareDataType,
		IsCached:   true,
	},
	"memory": {
		Title:      "Memory",
		SPDataType: &SPMemoryDataType,
		IsCached:   true,
	},
	"serial_number": {
		Title:      "Serial",
		SPDataType: &SPHardwareDataType,
		IsCached:   true,
	},
	/* ---------- System Profiler Data (non-cached data) ---------- */
	"battery": {
		Title:      "Battery",
		SPDataType: &SPPowerDataType,
	},
	"disk": {
		Title:      "Disk",
		SPDataType: &SPStorageDataType,
	},
	"display": {
		Title:      "Display",
		SPDataType: &SPDisplaysDataType,
	},
	"hostname": {
		Title:      "Hostname",
		SPDataType: &SPSoftwareDataType,
	},
	"os": {
		Title:      "OS",
		SPDataType: &SPSoftwareDataType,
	},
	"system_integrity": {
		Title:      "macOS SIP",
		SPDataType: &SPSoftwareDataType,
	},
	"uptime": {
		Title:      "Uptime",
		SPDataType: &SPSoftwareDataType,
	},
	"user": {
		Title:      "User",
		SPDataType: &SPSoftwareDataType,
	},
	/* ---------- Other Data ---------- */
	"datetime": {
		Title: "Date/Time",
		Func:  &datetimeNamedFunc,
	},
	"public_ip": {
		Title: "Public IP",
		Func:  &publicIpNamedFunc,
	},
	"software": {
		Title: "Software",
		Func:  &softwareNamedFunc,
	},
	"terminal": {
		Title: "Terminal",
		Func:  &termProgramNamedFunc,
	},
}

// Load the configuration file and check if the requested items are valid
// If no configuration file is provided, use the default values defined above.
func loadAndCheckConfig(configFilePath string) (err error) {
	if configFilePath != "" {
		file, err := os.Open(configFilePath)
		if err != nil {
			return fmt.Errorf("failed to open config file: %w", err)
		}
		defer file.Close()

		// Parse the YAML file into the Config structure
		decoder := yaml.NewDecoder(file)
		if err := decoder.Decode(config); err != nil {
			return err
		}

		if config.Items != nil {
			// Check if all requested items are valid
			for _, item := range config.Items {
				if _, exists := availableItems[item]; !exists {
					return fmt.Errorf("invalid item: %s", item)
				}
			}
			// Make sure there is no duplicate
			config.Items = uniqueStrings(config.Items)
		} else {
			config.Items = defaultItems
		}

		if config.CacheFilePath != nil {
			// Replace '~' with the home directory
			if strings.HasPrefix(*config.CacheFilePath, "~") {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("error getting home directory: %w", err)
				}
				*config.CacheFilePath = filepath.Join(homeDir, (*config.CacheFilePath)[1:])
			}
		} else {
			config.CacheFilePath = &defaultCacheFilePath
		}
	} else {
		config = &Config{
			CacheFilePath: &defaultCacheFilePath,
			Items:         defaultItems,
		}
	}

	return nil
}
