package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jwalton/go-supportscolor"
	"gopkg.in/yaml.v3"
)

// This struct represents the configuration file
type Config struct {
	CacheFilePath      *string        `yaml:"cache_file,omitempty"`
	DisplayLogo        *bool          `yaml:"display_logo,omitempty"`
	Logo               *string        `yaml:"logo_file,omitempty"`
	Cache              *bool          `yaml:"cache,omitempty"`
	DisplayNerdSymbols *bool          `yaml:"nerd_symbols,omitempty"`
	Items              []string       `yaml:"items,omitempty"`
	Weather            *WeatherConfig `yaml:"weather,omitempty"`
}

type WeatherConfig struct {
	Latitude          *float64 `yaml:"latitude,omitempty"`
	Longitude         *float64 `yaml:"longitude,omitempty"`
	LocationNameEn    *string  `yaml:"location_name_en,omitempty"`
	LocationStateEn   *string  `yaml:"location_state_en,omitempty"`
	LocationCountryEn *string  `yaml:"location_country_en,omitempty"`
	Units             string   `yaml:"units,omitempty"`
	Lang              string   `yaml:"lang,omitempty"`
}

var config = &Config{}

/* ---------- Default Configuration ---------- */
var defaultCacheFilePath = fmt.Sprintf("%s/.cache/minfo/static.json", envHome)
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
	weatherNamedFunc = NamedFunc{
		Id:   "fetchWeatherOpenMeteo",
		Func: fetchWeatherOpenMeteo,
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
	Nerd       string // Nerd Font symbol
	SPDataType *string
	Func       *NamedFunc
	IsCached   bool
}

// All available items we can fetch and display.
var availableItems = map[string]item{
	/* ---------- System Profiler Data (cached data) ---------- */
	"cpu": {
		Title:      "CPU",
		Nerd:       "",
		SPDataType: &SPHardwareDataType,
		IsCached:   true,
	},
	"gpu": {
		Title:      "GPU",
		Nerd:       "",
		SPDataType: &SPDisplaysDataType,
		IsCached:   true,
	},
	"model": {
		Title:      "Model",
		Nerd:       "",
		SPDataType: &SPHardwareDataType,
		IsCached:   true,
	},
	"memory": {
		Title:      "Memory",
		Nerd:       "",
		SPDataType: &SPMemoryDataType,
		IsCached:   true,
	},
	"serial_number": {
		Title:      "Serial",
		Nerd:       "",
		SPDataType: &SPHardwareDataType,
		IsCached:   true,
	},
	/* ---------- System Profiler Data (non-cached data) ---------- */
	"battery": {
		Title:      "Battery",
		Nerd:       "󰂄",
		SPDataType: &SPPowerDataType,
	},
	"disk": {
		Title:      "Disk",
		Nerd:       "󰋊",
		SPDataType: &SPStorageDataType,
	},
	"display": {
		Title:      "Display",
		Nerd:       "",
		SPDataType: &SPDisplaysDataType,
	},
	"hostname": {
		Title:      "Hostname",
		Nerd:       "",
		SPDataType: &SPSoftwareDataType,
	},
	"os": {
		Title:      "OS",
		Nerd:       "",
		SPDataType: &SPSoftwareDataType,
	},
	"system_integrity": {
		Title:      "macOS SIP",
		Nerd:       "",
		SPDataType: &SPSoftwareDataType,
	},
	"uptime": {
		Title:      "Uptime",
		Nerd:       "",
		SPDataType: &SPSoftwareDataType,
	},
	"user": {
		Title:      "User",
		Nerd:       "",
		SPDataType: &SPSoftwareDataType,
	},
	/* ---------- Other Data ---------- */
	"datetime": {
		Title: "Date/Time",
		Nerd:  "",
		Func:  &datetimeNamedFunc,
	},
	"public_ip": {
		Title: "Public IP",
		Nerd:  "󱦂",
		Func:  &publicIpNamedFunc,
	},
	"software": {
		Title: "Software",
		Nerd:  "",
		Func:  &softwareNamedFunc,
	},
	"terminal": {
		Title: "Terminal",
		Nerd:  "",
		Func:  &termProgramNamedFunc,
	},
	"weather": {
		Title: "Weather",
		Nerd:  "󰖙",
		Func:  &weatherNamedFunc,
	},
}

func getDefaultLogoFilePath() (defaultLogoFilePath *string) {
	defaultLogoFilePath = new(string)
	*defaultLogoFilePath = os.Getenv("HOMEBREW_PREFIX")
	if *defaultLogoFilePath != "" {
		if supportscolor.Stdout().Has256 || supportscolor.Stderr().Has16m {
			*defaultLogoFilePath = fmt.Sprintf("%s/share/minfo/apple-256colors", *defaultLogoFilePath)
		} else {
			*defaultLogoFilePath = fmt.Sprintf("%s/share/minfo/apple-16colors", *defaultLogoFilePath)
		}
	}
	return
}

// Load the configuration file and check if the requested items are valid
// If no configuration file is provided, use the default values defined above.
func loadAndCheckConfig(configFilePath string) (err error) {
	if configFilePath == "" {
		config = &Config{
			CacheFilePath:      &defaultCacheFilePath,
			DisplayLogo:        nil,
			Logo:               getDefaultLogoFilePath(),
			Cache:              nil,
			DisplayNerdSymbols: nil,
			Items:              defaultItems,
			Weather: &WeatherConfig{
				Units: "metric",
				Lang:  "en",
			},
		}
		return nil
	}
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
	if config.DisplayLogo == nil {
		config.DisplayLogo = new(bool)
		*config.DisplayLogo = true // This default value might be overridden by the command line
	}
	if config.Cache == nil {
		config.Cache = new(bool)
		*config.Cache = true // This default value might be overridden by the command line
	}
	if config.Weather == nil {
		config.Weather = &WeatherConfig{
			Units: "metric",
			Lang:  "en",
		}
	} else {
		if config.Weather.Units == "" {
			config.Weather.Units = "metric"
		} else if config.Weather.Units != "metric" && config.Weather.Units != "imperial" {
			return fmt.Errorf("invalid weather units: %s", config.Weather.Units)
		}
		if config.Weather.Lang == "" {
			config.Weather.Lang = "en"
		} else if config.Weather.Lang != "en" && config.Weather.Lang != "fr" {
			return fmt.Errorf("invalid language: %s", config.Weather.Lang)
		}
		if config.Weather.LocationNameEn != nil {
			if config.Weather.LocationCountryEn == nil {
				return fmt.Errorf("for weather, you need to provide a country")
			}
			if config.Weather.LocationStateEn == nil {
				config.Weather.LocationStateEn = new(string) // let's initialize it to ""
			}
		}
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
	if config.Logo != nil {
		// Replace '~' with the home directory
		if strings.HasPrefix(*config.Logo, "~") {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("error getting home directory: %w", err)
			}
			*config.Logo = filepath.Join(homeDir, (*config.Logo)[1:])
		}
	} else {
		config.Logo = getDefaultLogoFilePath()
	}
	if config.DisplayNerdSymbols == nil {
		config.DisplayNerdSymbols = new(bool)
		*config.DisplayNerdSymbols = true // This default value might be overridden by the command line
	}

	return nil
}
