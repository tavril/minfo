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
	"strings"
	"sync"
)

var (
	cacheFilePath = fmt.Sprintf("%s/.minfo-cache.json", os.Getenv("HOME"))
	hostInfo      = info{}
	GitCommit     string
	GitVersion    string
)

// Print the information in a human-readable format
func printInfo(hostInfo *info, noLogoFlag bool) {
	var output strings.Builder
	term := os.Getenv("TERM")

	var colorNormal, colorRed, colorGreen, colorYellow string
	var colorBlue, colorPurple, colorCyan, colorOrange string
	if strings.Contains(term, "256") {
		colorNormal = "\033[0m"
		colorRed = "\033[38;5;160m"
		colorGreen = "\033[38;5;028m"
		colorYellow = "\033[38;5;226m"
		colorBlue = "\033[38;5;021m"
		colorPurple = "\033[38;5;054m"
		colorCyan = "\033[38;5;075m"
		colorOrange = "\033[38;5;202m"
	} else {
		colorNormal = "\033[0m"
		colorRed = "\033[00;31m"
		colorGreen = "\033[00;32m"
		colorYellow = "\033[00;33m"
		colorBlue = "\033[00;34m"
		colorPurple = "\033[00;35m"
		colorCyan = "\033[00;36m"
		colorOrange = "\033[00;91m"
	}
	userInfo := fmt.Sprintf("%s (%s)", hostInfo.User.RealName, hostInfo.User.Login)
	osInfo := fmt.Sprintf("%s %s %s (%s) %s %s",
		hostInfo.Os.System,
		hostInfo.Os.SystemVersionCodeNname,
		hostInfo.Os.SystemVersion,
		hostInfo.Os.SystemBuild,
		hostInfo.Os.KernelType,
		hostInfo.Os.KernelVersion,
	)
	modelInfo := fmt.Sprintf("%s %s (%s) %s",
		hostInfo.Model.Name,
		hostInfo.Model.SubName,
		hostInfo.Model.Date,
		hostInfo.Model.Number,
	)
	var cpuCoreInfo string
	if strings.HasPrefix(hostInfo.Cpu.Model, "Apple") {
		cpuCoreInfo = fmt.Sprintf("%s %d cores (%d P and %d E)",
			hostInfo.Cpu.Model,
			hostInfo.Cpu.Cores,
			hostInfo.Cpu.PerformanceCores,
			hostInfo.Cpu.EfficiencyCores,
		)
	} else {
		// Intel CPU: The model also includes the number of cores
		// (Ex: "6-Core Intel Core i7")
		cpuCoreInfo = fmt.Sprintf("%s", hostInfo.Cpu.Model)
	}
	gpuInfo := fmt.Sprintf("%d cores", hostInfo.GpuCores)
	memoryInfo := fmt.Sprintf("%d %s %s",
		hostInfo.Memory.Amount,
		hostInfo.Memory.Unit,
		hostInfo.Memory.MemType,
	)
	diskInfo := fmt.Sprintf("%.2f TB (%.2f TB available)",
		hostInfo.Disk.TotalTB,
		hostInfo.Disk.FreeTB,
	)
	var charging string
	if hostInfo.Battery.Charging {
		charging = "(charging)"
	} else {
		charging = "(discharging)"
	}
	batteryInfo := fmt.Sprintf("%d%% %s | %d%% capacity",
		hostInfo.Battery.StatusPercent,
		charging,
		hostInfo.Battery.CapacityPercent,
	)

	var displayInfo []string
	for _, display := range hostInfo.Displays {
		displayInfo = append(displayInfo, fmt.Sprintf("%d x %d | %d x %d @ %.0f Hz",
			display.PixelsWidth,
			display.PixelsHeight,
			display.ResolutionWidth,
			display.ResolutionHeight,
			display.RefreshRateHz,
		))
	}
	// If we have no secondary display, add a placeholder
	if len(hostInfo.Displays) == 1 {
		displayInfo = append(displayInfo, "N/A")
	}

	softwareInfo := fmt.Sprintf("%d Apps | %d Homebrew packages",
		hostInfo.Software.NumApps,
		hostInfo.Software.NumBrew,
	)

	applelogo := [][]string{
		{colorGreen, "                    ##           "},
		{colorGreen, "                  ####           "},
		{colorGreen, "                #####            "},
		{colorGreen, "               ####              "},
		{colorGreen, "      ########   ############    "},
		{colorGreen, "    ##########################   "},
		{colorYellow, "  ###########################    "},
		{colorYellow, "  ##########################     "},
		{colorOrange, " ##########################      "},
		{colorOrange, " ##########################      "},
		{colorRed, " ###########################     "},
		{colorRed, "  ############################   "},
		{colorPurple, "  #############################  "},
		{colorPurple, "   ############################  "},
		{colorBlue, "     ########################    "},
		{colorBlue, "      ######################     "},
		{colorBlue, "        #######    #######       "},
	}

	info := [][]string{
		{colorCyan, "User", colorNormal, userInfo},
		{colorCyan, "Hostname", colorNormal, hostInfo.Hostname},
		{colorCyan, "OS", colorNormal, osInfo},
		{colorCyan, "Model", colorNormal, modelInfo},
		{colorCyan, "CPU", colorNormal, cpuCoreInfo},
		{colorCyan, "GPU", colorNormal, gpuInfo},
		{colorCyan, "Memory", colorNormal, memoryInfo},
		{colorCyan, "Disk", colorNormal, diskInfo},
		{colorCyan, "Disk SMART", colorNormal, hostInfo.Disk.SmartStatus},
		{colorCyan, "Battery", colorNormal, batteryInfo},
		{colorCyan, "Battery health", colorNormal, hostInfo.Battery.Health},
		{colorCyan, "Main Display", colorNormal, displayInfo[0]},
		{colorCyan, "Add. Display", colorNormal, displayInfo[1]},
		{colorCyan, "Software", colorNormal, softwareInfo},
		{colorCyan, "Public IP", colorNormal, hostInfo.PublicIP},
		{colorCyan, "Uptime", colorNormal, hostInfo.Uptime},
		{colorCyan, "Date/Time", colorNormal, hostInfo.Datetime},
	}

	if !noLogoFlag {
		for i := 0; i < len(applelogo); i++ {
			output.WriteString(fmt.Sprintf("%s%s%s%-15s%s%s\n",
				applelogo[i][0],
				applelogo[i][1],
				info[i][0],
				info[i][1],
				info[i][2],
				info[i][3],
			))
		}
	} else {
		for _, i := range info {
			output.WriteString(fmt.Sprintf("%s%-15s%s%s\n",
				i[0],
				i[1],
				i[2],
				i[3],
			))
		}
	}

	fmt.Printf("%s", output.String())
}

func main() {
	var err error
	jsonFlag := flag.Bool("j", false, "Output in JSON format instead of displaying logo")
	refreshCacheFlag := flag.Bool("r", false, "Refresh cache (or create it if it doesn't exist)")
	noCacheFlag := flag.Bool("n", false, "Don't use/update cache")
	noLogoFlag := flag.Bool("l", false, "Don't display the ASCII art logo")
	showVersionFlag := flag.Bool("v", false, "Show version")
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

	/*
		We cache the following information, which are unlikely to change:
			- Model
			- CPU and GPU
			- Memory
		readCacheFiles() reads the cache file and unmarshals it into hostInfo
	*/
	if !*refreshCacheFlag && !*noCacheFlag {
		if err = readCacheFile(); err != nil {
			if !errors.Is(err, os.ErrNotExist) && err != errEmptyCache {
				log.Fatalf("Error reading cache file: %v", err)
			}
		} else {
			haveCache = true
		}
	}

	var spErr error
	tasks := []func(){
		func() { spErr = fetchSystemProfiler(&hostInfo, haveCache) },
		func() { hostInfo.Software.NumBrew = fetchNumHomebrew() },
		func() { hostInfo.Software.NumApps = fetchNumApps() },
		// func() { hostInfo.Terminal = fetchTermProgram() },
		func() { hostInfo.PublicIP = fetchPublicIp() },
	}
	if !haveCache {
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
		printInfo(&hostInfo, *noLogoFlag)
	}
	if !haveCache && !*noCacheFlag {
		if err = writeCacheFile(); err != nil {
			log.Fatalf("Error writing cache file: %v", err)
		}
	}
}
