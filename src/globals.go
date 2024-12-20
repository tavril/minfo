package main

import (
	"fmt"
	"os"
	"path"
)

var (
	appName           = path.Base(os.Args[0])
	defaultConfigFile = fmt.Sprintf("%s/.config/%s.yml", os.Getenv("HOME"), appName)
	hostInfo          = info{}
	GitCommit         string
	GitVersion        string
	colorNormal       = "\033[0m"
	// The colors will be defined depending on the terminal type
	colorRed    string
	colorGreen  string
	colorYellow string
	colorBlue   string
	colorPurple string
	colorCyan   string
	colorOrange string
)
