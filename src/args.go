package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
)

func usage() {
	fmt.Printf(`
Description:
    %s is a tool to display information about the host system.
    It only works on MacOs.

Usage:
    %s [--config <path>] [-j|--json] [-i|--items] [-v|--version] [-l|--logo <path>]
    %s [-r|--refresh[=false]] [-c|--cache[=false]] [-d|--display-logo[=false]]

Options:
    --config <path>             Path to the configuration file (default: %s).
    -d, --display-logo[=false]  Display the ASCII art logo (default: true).
    -l, --logo[=<path>]         Path to ASCII art logo file
	                            (default: $HOMEBREW_PREFIX/share/minfo/apple or $HOME/.config/minfo/logo).
    -j, --json[=false]          Display information in JSON instead of plain text (default: false).
    -c, --cache[=false]         Use cache file (default: true).
    -r, --refresh[=false]       Refresh the cache file (default: false).
    -i, --items                 Display all available information to display and exit.
    -v, --version               Show version and exit.
    -h, --help                  Show this help message and exit.

Having a configuration file is optional. If you don't provide the --config parameter,
minfo will look for a configuration file at the default path.
If the configuration file does not exist, minfo will use the default values.

Command line argument --cache and --display-logo will override the values in the configuration file.

--refresh=true and --cache=false are mutually exclusive.

If you provide --json=true, then --display-logo will be ignored.

`, appName, appName, appName, defaultConfigFile)
}

type cmdLineParams struct {
	Json           bool
	RefreshCache   bool
	Cache          *bool
	DisplayLogo    *bool
	Logo           *string
	Items          bool
	Version        bool
	ConfigFilePath string
}

func parseCmdLineArgs(args []string) (*cmdLineParams, error) {
	fs := flag.NewFlagSet("minfo", flag.ContinueOnError)
	fs.Usage = usage

	var (
		jsonFlag           bool
		refreshCacheFlag   bool
		itemsFlag          bool
		versionFlag        bool
		configFilePathFlag string
		helpFlag           bool
	)
	displayLogoFlag := new(bool)
	cacheFlag := new(bool)
	logoFlag := new(string)

	fs.BoolVar(&helpFlag, "help", false, "print this help message and exit.")
	fs.BoolVar(&helpFlag, "h", false, "print this help message and exit.")

	fs.BoolVar(&versionFlag, "version", false, "print the version and exit.")
	fs.BoolVar(&versionFlag, "v", false, "print the version and exit.")

	fs.BoolVar(&itemsFlag, "items", false, "display all available items to display and exit.")
	fs.BoolVar(&itemsFlag, "i", false, "display all available items to display and exit.")

	fs.StringVar(&configFilePathFlag, "config", "", "path to the configuration file.")

	fs.BoolVar(&jsonFlag, "json", false, "display information in JSON instead of plain text (default: false).")
	fs.BoolVar(&jsonFlag, "j", false, "display information in JSON instead of plain text (default: false).")

	fs.BoolVar(cacheFlag, "cache", true, "use cache file (default: true).")
	fs.BoolVar(cacheFlag, "c", true, "use cache file (default: true).")

	fs.BoolVar(&refreshCacheFlag, "refresh", false, "refresh the cache file (default: false).")
	fs.BoolVar(&refreshCacheFlag, "r", false, "refresh the cache file (default: false).")

	fs.BoolVar(displayLogoFlag, "display-logo", true, "display the ASCII art logo (default: true).")
	fs.BoolVar(displayLogoFlag, "d", true, "display the ASCII art logo (default: true).")

	fs.StringVar(logoFlag, "logo", "", "path to the logo file")
	fs.StringVar(logoFlag, "l", "", "path to the logo file")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}
	if helpFlag {
		fs.Usage()
		os.Exit(0)
	}

	// Check if flags --display-logo, logo and --cache were explicitly set
	// in which case they would override the values in configuration file.
	displayLogoFlagSet := false
	logoFlagSet := false
	cacheFlagSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "display-logo" || f.Name == "d" {
			displayLogoFlagSet = true
		} else if f.Name == "logo" || f.Name == "l" {
			logoFlagSet = true
		} else if f.Name == "cache" || f.Name == "c" {
			cacheFlagSet = true
		}

	})
	if !displayLogoFlagSet {
		*displayLogoFlag = true
	}
	if !logoFlagSet {
		logoFlag = nil
	}
	if !cacheFlagSet {
		*cacheFlag = true
	}

	return &cmdLineParams{
		Json:           jsonFlag,
		RefreshCache:   refreshCacheFlag,
		Cache:          cacheFlag,
		DisplayLogo:    displayLogoFlag,
		Logo:           logoFlag,
		Items:          itemsFlag,
		Version:        versionFlag,
		ConfigFilePath: configFilePathFlag,
	}, nil
}

func (cmdLine *cmdLineParams) controlCmdLineParams() {
	if cmdLine.Version {
		fmt.Printf("minfo %s (commit %s)\n", GitVersion, GitCommit)
		os.Exit(0)
	}
	if (cmdLine.Cache != nil && !*cmdLine.Cache) && cmdLine.RefreshCache {
		log.Fatalf("--cache=false and --refresh=true are mutually exclusive")
	}
	if cmdLine.Items {
		fmt.Println("Available information to choose from:")
		var iArr []string

		for k := range availableItems {
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
		if _, err := os.Stat(defaultConfigFile); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Fatalf("error while getting default config file stat: %v", err)
			}
		} else {
			cmdLine.ConfigFilePath = defaultConfigFile
		}
	}
}
