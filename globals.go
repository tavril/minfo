package main

var (
	hostInfo    = info{}
	GitCommit   string
	GitVersion  string
	colorNormal = "\033[0m"
	// The colors will be defined dependinf on the terminal type
	colorRed    string
	colorGreen  string
	colorYellow string
	colorBlue   string
	colorPurple string
	colorCyan   string
	colorOrange string
)

// For items to be retrieved by calling system_profiler
// we have a mapping of the information we need to fetch
var spInfoMapping = map[string][]string{
	"SPHardwareDataType": {"model", "cpu"},
	"SPSoftwareDataType": {"user", "hostname", "os"},
	"SPDisplaysDataType": {"display", "gpu"},
	"SPPowerDataType":    {"battery"},
	"SPMemoryDataType":   {"memory"},
	"SPStorageDataType":  {"disk"},
}
