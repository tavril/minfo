package main

import (
	"fmt"
	"os"
)

var (
	cacheFilePath = fmt.Sprintf("%s/.minfo-cache.json", os.Getenv("HOME"))
	hostInfo      = info{}
	GitCommit     string
	GitVersion    string
	colorNormal   = "\033[0m"
	// The colors will be defined dependinf on the terminal type
	colorRed    string
	colorGreen  string
	colorYellow string
	colorBlue   string
	colorPurple string
	colorCyan   string
	colorOrange string
)
