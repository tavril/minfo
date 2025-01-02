package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"
	"sync"
)

func main() {
	/* ---------- Command line parameters ---------- */
	cmdLine, err := parseCmdLineArgs(os.Args[1:])
	if err != nil {
		fmt.Printf("Error parsing arguments: %v\n", err)
		os.Exit(1)
	}
	cmdLine.controlCmdLineParams()

	/* ---------- Load and check configuration ---------- */
	if err := loadAndCheckConfig(cmdLine.ConfigFilePath); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	if cmdLine.Cache != nil {
		config.Cache = cmdLine.Cache
	}
	if cmdLine.DisplayLogo != nil {
		config.DisplayLogo = cmdLine.DisplayLogo
	}
	if cmdLine.Logo != nil {
		config.Logo = cmdLine.Logo
	}

	/* ---------- Deal with cache ---------- */
	// We cache some data which are not going to change:
	// - computer model, CPU, GPU, memory, and serial number.
	// If the cache file does not exist yet, we fetch all the data
	// that can be cached (even if not requested by the suer), and write it,
	// except if explicitly requested by the user to not use the cache.
	if *config.Cache {
		if err := readCacheFile(*config.CacheFilePath); err != nil {
			if !errors.Is(err, os.ErrNotExist) && err != errEmptyCache {
				log.Fatalf("Error reading cache file: %v", err)
			}
			// no cache file (or empty) --> Must populate it, i.e. like '-r' was passed.
			cmdLine.RefreshCache = true
		}
	}
	// User requested to refresh the cache, or cache file does not exist yet.
	if cmdLine.RefreshCache {
		if err := populateCache(*config.CacheFilePath); err != nil {
			log.Fatalf("Error while refreshing cache: %v", err)
		}
	}

	/* ---------- Prepare the tasks to execute ---------- */
	// the functions to execute to fetch the requested information
	tasks := []func(){}

	// Track which spDataType we will need to fetch from system_profiler
	spDataTypes := map[string]bool{}
	writeWeatherCache := true // Do we need to write the weather cache file?

	for _, requestedItem := range config.Items {
		item := availableItems[requestedItem]

		// system_profiler data
		if item.SPDataType != nil {
			if item.IsCached && *config.Cache {
				if _, ok := spDataTypes[*item.SPDataType]; !ok {
					spDataTypes[*item.SPDataType] = false
				}
			} else {
				// Either the spDataType is not cached yet in the map,
				// or it is in the map but set to false,
				// in which case we set it to true.
				_, ok := spDataTypes[*item.SPDataType]
				if !ok || !spDataTypes[*item.SPDataType] {
					spDataTypes[*item.SPDataType] = true
				}
			}
		}

		// other data, each fetched by its own function.
		if item.Func != nil {
			// We have a cache for the weather (default: 15 min)
			if item.Title == "Weather" {
				if isOlder, err := isFileOlderThan(weatherCacheFile, weatherCacheDuration); err != nil || isOlder {
					tasks = append(tasks, func() { (*item.Func).Func(&hostInfo) })
				} else if err := readCacheFile(weatherCacheFile); err != nil || hostInfo.Weather == nil {
					tasks = append(tasks, func() { (*item.Func).Func(&hostInfo) })
				} else {
					writeWeatherCache = false
				}
			} else {
				tasks = append(tasks, func() { (*item.Func).Func(&hostInfo) })
			}
		}
	}

	// Do we need to run system_profiler ?
	var spErr error
	for _, v := range spDataTypes {
		if v {
			tasks = append(tasks,
				func() { spErr = fetchSystemProfiler(&hostInfo, config.Items, spDataTypes, *config.Cache) },
			)
			break
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

	if writeWeatherCache {
		tmpInfo := info{Weather: hostInfo.Weather}
		if err := writeCacheFile(weatherCacheFile, &tmpInfo); err != nil {
			log.Fatalf("Error writing weather cache: %v", err)
		}
	}

	/* ---------- Display information ---------- */
	if cmdLine.Json {
		// We have read information from the cache, so we might
		// have some information that was not requested by the user.
		// We need to filter them out.

		for itemName, item := range availableItems {
			if slices.Contains(config.Items, itemName) || !item.IsCached {
				continue
			}
			switch itemName {
			case "cpu":
				hostInfo.Cpu = nil
			case "gpu":
				hostInfo.GpuCores = nil
			case "model":
				hostInfo.Model = nil
			case "memory":
				hostInfo.Memory = nil
			case "serial_number":
				hostInfo.SerialNumber = nil
			}
		}

		jsonData, err := json.MarshalIndent(hostInfo, "", "  ")
		if err != nil {
			log.Fatalf("Error marshalling JSON: %v", err)
		}
		fmt.Println(string(jsonData))
	} else {
		if err := printInfo(&hostInfo); err != nil {
			log.Fatalf("Error printing info: %v", err)
		}
	}
}
