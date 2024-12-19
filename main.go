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
	"slices"
	"strings"
	"sync"
)

func main() {
	var err error
	jsonFlag := flag.Bool("j", false, "Output in JSON format instead of displaying logo")
	refreshCacheFlag := flag.Bool("r", false, "Refresh cache (or create it if it doesn't exist)")
	noCacheFlag := flag.Bool("n", false, "Don't use/update cache")
	withLogoFlag := flag.Bool("l", true, "Display the ASCII art logo")
	listItems := flag.Bool("i", false, "Display all available information to display")
	showVersionFlag := flag.Bool("v", false, "Show version")
	configFilePath := flag.String("c", filepath.Join(os.Getenv("HOME"), ".config", "minfo.yml"), "Path to the configuration file")
	helpFlag := flag.Bool("h", false, "Show help")

	flag.Parse()
	haveCache := false

	if *helpFlag {
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(0)
	}
	if *showVersionFlag {
		if GitCommit != "homebrew" {
			GitCommit = "commit " + GitCommit
		}
		fmt.Printf("minfo version %s (%s)\n", GitVersion, GitCommit)
		os.Exit(0)
	}
	if *noCacheFlag && *refreshCacheFlag {
		log.Fatalf("Can't use both -r and -n flags")
	}
	if *listItems {
		fmt.Println("Here is the list of all available information:")
		slices.Sort(allItems)
		for _, item := range allItems {
			fmt.Printf("- %s\n", item)
		}
		os.Exit(0)
	}

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
				if !slices.Contains(allItems, item) {
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

	/*
		We cache the following information, which are unlikely to change:
			- Model
			- CPU and GPU
			- Memory
		readCacheFiles() reads the cache file and unmarshals it into hostInfo
	*/
	if !*refreshCacheFlag && !*noCacheFlag {
		if err = readCacheFile(config.CacheFilePath); err != nil {
			if !errors.Is(err, os.ErrNotExist) && err != errEmptyCache {
				log.Fatalf("Error reading cache file: %v", err)
			}
		} else {
			haveCache = true
		}
	}

	// Prepare the tasks to execute
	var spErr error
	tasks := []func(){}

	// This is a quick help to know with an item as been requested or not
	// (quicker than parsing config.Items each time)
	reqItems := make(map[string]bool, len(config.Items))
	for _, item := range allItems {
		if slices.Contains(config.Items, item) {
			reqItems[item] = true
		} else {
			reqItems[item] = false
		}
	}

	// system_profiler items:
	// - if !haveCache, then, fetch all SP items that are requested.
	// - if haveCache, then, fetch only the SP items that are not cached.
	alreadyDoneSP := false
	for _, spItem := range spItemsNotCached {
		if reqItems[spItem] {
			tasks = append(tasks, func() { spErr = fetchSystemProfiler(&hostInfo, haveCache) })
			alreadyDoneSP = true
			break
		}
	}
	if !haveCache && !alreadyDoneSP {
		for _, spItem := range spItemsCached {
			if reqItems[spItem] {
				tasks = append(tasks, func() { spErr = fetchSystemProfiler(&hostInfo, haveCache) })
				break
			}
		}
	}

	if reqItems["terminal"] {
		tasks = append(tasks, func() { hostInfo.Terminal = fetchTermProgram() })
	}
	if reqItems["software"] {
		tasks = append(tasks, func() { hostInfo.Software.NumApps = fetchNumApps() })
		tasks = append(tasks, func() { hostInfo.Software.NumBrew = fetchNumHomebrew() })
	}
	if reqItems["public_ip"] {
		tasks = append(tasks, func() { hostInfo.PublicIP = fetchPublicIp() })
	}
	// "model" needs both data from system_profiler and from ioreg
	// system_profiler has been treaded above ...
	if reqItems["model"] && !haveCache {
		tasks = append(tasks, func() { hostInfo.Model.Name, hostInfo.Model.SubName, hostInfo.Model.Date = fetchModelYear() })
	}

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

	if *jsonFlag {
		jsonData, err := json.MarshalIndent(hostInfo, "", "  ")
		if err != nil {
			log.Fatalf("Error marshalling JSON: %v", err)
		}
		fmt.Println(string(jsonData))
	} else {
		printInfo(&hostInfo, *withLogoFlag)
	}
	if !haveCache && !*noCacheFlag {
		if err = writeCacheFile(config.CacheFilePath); err != nil {
			log.Fatalf("Error writing cache file: %v", err)
		}
	}
}
