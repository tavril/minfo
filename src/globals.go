package main

import (
	"fmt"
	"os"
	"path"
	"runtime"
)

var (
	appName           = path.Base(os.Args[0])
	arch              = runtime.GOARCH
	defaultConfigFile = fmt.Sprintf("%s/.config/%s/config.yaml", os.Getenv("HOME"), appName)
	envHome           = os.Getenv("HOME")
	hostInfo          = info{}
	GitCommit         string
	GitVersion        string
	colorNormal       = "\033[0m"
	// The colors will be defined depending on the terminal type (256 or 16 colors)
	colorRed    string
	colorGreen  string
	colorYellow string
	colorBlue   string
	colorPurple string
	colorCyan   string
	colorOrange string
)
