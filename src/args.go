package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
)

type cmdLineParams struct {
	Json           bool
	RefreshCache   bool
	Cache          bool
	WithLogo       bool
	ListItems      bool
	ShowVersion    bool
	ConfigFilePath string
}

func parseCmdLineArgs(args []string) (*cmdLineParams, error) {
	fs := flag.NewFlagSet("minfo", flag.ContinueOnError)
	/* ---------- Flags ---------- */
	jsonFlag := fs.Bool("j", false, "Output in JSON format instead of displaying logo")
	refreshCacheFlag := fs.Bool("r", false, "Refresh cache (or create it if it doesn't exist)")
	cacheFlag := fs.Bool("n", true, "Don't use/update cache")
	withLogoFlag := fs.Bool("l", true, "Display the ASCII art logo")
	listItems := fs.Bool("i", false, "Display all available information to display")
	showVersionFlag := fs.Bool("v", false, "Show version")
	configFilePath := fs.String("c", "", "Path to the configuration file")
	helpFlag := fs.Bool("h", false, "Show help")

	/* ---------- Deal with Flags ---------- */
	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}
	if *helpFlag {
		help()
		fmt.Println("Usage:")
		fs.PrintDefaults()
		os.Exit(0)
	}
	return &cmdLineParams{
		Json:           *jsonFlag,
		RefreshCache:   *refreshCacheFlag,
		Cache:          *cacheFlag,
		WithLogo:       *withLogoFlag,
		ListItems:      *listItems,
		ShowVersion:    *showVersionFlag,
		ConfigFilePath: *configFilePath,
	}, nil
}

func (cmdLine *cmdLineParams) controlCmdLineParams() {
	if cmdLine.ShowVersion {
		fmt.Printf("minfo %s (commit %s)\n", GitVersion, GitCommit)
		os.Exit(0)
	}
	if !cmdLine.Cache && cmdLine.RefreshCache {
		log.Fatalf("Cannot use -n=false and -r at the same time")
	}
	if cmdLine.ListItems {
		fmt.Println("Available information to choose from:")
		var iArr []string

		for k, _ := range availableItems {
			iArr = append(iArr, k)
		}
		sort.Strings(iArr)
		for _, i := range iArr {
			fmt.Printf("  %s\n", i)
		}
		os.Exit(0)
	}
	if cmdLine.ConfigFilePath != "" {
		if _, err := os.Stat(cmdLine.ConfigFilePath); err != nil {
			log.Fatalf("Error while getting config file stat: %v", err)
		}
	} else {
		// Check default configuration file, that might exit (not mandatory)
		_, err := os.Stat(defaultConfigFile)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Fatalf("error while getting config file stat: %v", err)
		}
	}
}

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
Example

---
cache_file: ~/minfo-cache.json

The default location of the cache file is %s.
You can also decide to not use the cache with the -n flag, or to force refresh
the cache with the -r flag.

`, appName, defaultConfigFile, defaultCacheFilePath)
}
