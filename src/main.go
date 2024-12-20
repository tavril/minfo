package main

// compile with: go build -ldflags "-s -w" apple_info_motd.go
// (removes symbols and debug info)

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func help() {
	fmt.Printf(`
%s is a tool to display information about the host system.
It only works on MacOs.

By default, it display an ASCII art logo alonside the information.
You can choose to only display the information (-l=false), or
you can choose to display the information in JSON format (-j).

There is a default set of information to display, but you can
customize this list in a configuration file. To list all available
information to display, use the -i flag.
By default, the configuration file is located at %s.
Example:

---
items:
  - os
  - model

A cache file is used to store the information that is unlikely to change:
computer model, CPU and GPU, and memory. You can change the location of
this file in the configuration file by addind a "cache_file" key.
The default location of the cache file is %s.
You can also decide to not use the cache with the -n flag, or to force refresh
the cache with the -r flag.

`, appName, defaultConfigFile, defaultCacheFilePath)
	fmt.Println("Usage:")
	flag.PrintDefaults()
}

func main() {
	var err error

	/* ---------- Flags ---------- */
	jsonFlag := flag.Bool("j", false, "Output in JSON format instead of displaying logo")
	refreshCacheFlag := flag.Bool("r", false, "Refresh cache (or create it if it doesn't exist)")
	noCacheFlag := flag.Bool("n", false, "Don't use/update cache")
	withLogoFlag := flag.Bool("l", true, "Display the ASCII art logo")
	listItems := flag.Bool("i", false, "Display all available information to display")
	showVersionFlag := flag.Bool("v", false, "Show version")
	configFilePath := flag.String("c", defaultConfigFile, "Path to the configuration file")
	helpFlag := flag.Bool("h", false, "Show help")

	/* ---------- Deal with Flags ---------- */
	flag.Parse()
	haveCache := false

	if *helpFlag {
		help()
		os.Exit(0)
	}
	if *showVersionFlag {
		fmt.Printf("minfo version %s (commit %s)\n", GitVersion, GitCommit)
		os.Exit(0)
	}
	if *noCacheFlag && *refreshCacheFlag {
		log.Fatalf("-r and -n flags are mutually exclusive.")
	}
	if *listItems {
		fmt.Println("Available information to choose from:")
		for k, _ := range itemsConfig {
			fmt.Printf("  %s\n", k)
		}
		os.Exit(0)
	}

	/* ---------- Load Configuration ---------- */
	_, err = os.Stat(*configFilePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal("Cannot read configuration file")
	} else if err == nil {
		if config, err = loadConfig(*configFilePath); err != nil {
			log.Fatalf("Error loading config file: %v", err)
		}
	}
	if config == nil {
		config = &Config{
			CacheFilePath: defaultCacheFilePath,
			Items:         defaultItems,
		}
	} else {
		if len(config.Items) == 0 {
			config.Items = defaultItems
		} else {
			// Check if all requested items are valid
			for _, item := range config.Items {
				if _, exists := itemsConfig[item]; !exists {
					log.Fatalf("Invalid item: %s", item)
				}
			}
			// Make sure there is no duplicate
			config.Items = uniqueStrings(config.Items)
		}
		if config.CacheFilePath == "" {
			config.CacheFilePath = defaultCacheFilePath
		} else {
			// Replace '~' with the home directory
			if strings.HasPrefix(config.CacheFilePath, "~") {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					log.Fatalf("Error getting home directory: %v", err)
				}
				config.CacheFilePath = filepath.Join(homeDir, config.CacheFilePath[1:])
			}
		}
	}

	/* ---------- Deal with cache ---------- */
	//	We cache the following information, which are unlikely to change:
	//	Computer model, CPU and GPU, and memory.
	if !*refreshCacheFlag && !*noCacheFlag {
		if err = readCacheFile(config.CacheFilePath); err != nil {
			if !errors.Is(err, os.ErrNotExist) && err != errEmptyCache {
				log.Fatalf("Error reading cache file: %v", err)
			}
		} else {
			haveCache = true
		}
	}

	/* ---------- Prepare tasks ---------- */
	// Prepare the tasks to execute
	var spErr error

	// the functions to execute to fetch the requested information
	tasks := []func(){}
	// Just a map to easily track which functions we need to run
	toRunFunc := map[string]NamedFunc{}

	// Track which spDataType we will need to fetch from system_profiler
	spDataTypes := map[string]bool{}

	for _, requestedItem := range config.Items {
		item := itemsConfig[requestedItem]

		// do we need to call a function (appart from system_profiler)
		// to retrieve the information?
		if item.retrieveCmd.Id != "" {
			if _, exists := toRunFunc[item.retrieveCmd.Id]; !exists {
				toRunFunc[item.retrieveCmd.Id] = item.retrieveCmd
			}
		}

		if item.SystemProfiler.DataType != "" {
			if item.SystemProfiler.IsCached && haveCache {
				if !spDataTypes[item.SystemProfiler.DataType] {
					spDataTypes[item.SystemProfiler.DataType] = false
				}
			} else {
				if !spDataTypes[item.SystemProfiler.DataType] {
					spDataTypes[item.SystemProfiler.DataType] = true
				}
			}
		}
	}

	// Do we need to run system_profiler ?
	if len(spDataTypes) > 0 {
		tasks = append(tasks, func() { spErr = fetchSystemProfiler(&hostInfo, spDataTypes, haveCache) })
	}
	// Any other functions to run ?
	if len(toRunFunc) > 0 {
		for _, nameFunc := range toRunFunc {
			tasks = append(tasks, func() { nameFunc.Func(&hostInfo) })
		}
	}

	/* ---------- Execute tasks ---------- */
	var wg sync.WaitGroup
	wg.Add(len(tasks))
	for _, task := range tasks {
		go func(t func()) {
			defer wg.Done()
			t()
		}(task)
	}
	wg.Wait()

	if spErr != nil {
		log.Fatalf("Error fetching system profiler: %v", spErr)
	}

	/* ---------- Display information ---------- */
	if *jsonFlag {
		jsonData, err := json.MarshalIndent(hostInfo, "", "  ")
		if err != nil {
			log.Fatalf("Error marshalling JSON: %v", err)
		}
		fmt.Println(string(jsonData))
	} else {
		printInfo(&hostInfo, *withLogoFlag)
	}

	/* ---------- Write cache file ---------- */
	if !haveCache && !*noCacheFlag {
		if err = writeCacheFile(config.CacheFilePath); err != nil {
			log.Fatalf("Error writing cache file: %v", err)
		}
	}
}
