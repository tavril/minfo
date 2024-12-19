package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/micromdm/plist"
)

// Fetch information from "system_profiler"
func fetchSystemProfiler(hostInfo *info, haveCache bool) (err error) {
	var spInfo systemProfilerInfo

	args := []string{"-json", "-detailLevel", "basic"}
	// Let's only call system_profiler for the information we need
	for _, requestItem := range config.Items {
		for dataType, items := range spInfoMapping {
			if slices.Contains(items, requestItem) {
				if (dataType == "SPHardwareDataType" && !haveCache) || dataType != "SPHardwareDataType" {
					args = append(args, dataType)
				}
				break
			}
		}
	}

	cmd := exec.Command("/usr/sbin/system_profiler", args...)
	output, err := runCommand(cmd)
	if err != nil {
		return
	}

	if err = json.Unmarshal([]byte(output), &spInfo); err != nil {
		return
	}
	// User
	if slices.Contains(config.Items, "user") {
		re := regexp.MustCompile(`^([\w\s]+)\s\((\w+)\)$`)
		matches := re.FindStringSubmatch(spInfo.Software[0].UserName)

		if len(matches) == 3 {
			hostInfo.User.RealName = matches[1]
			hostInfo.User.Login = matches[2]
		}
	}

	// Hostname
	if slices.Contains(config.Items, "hostname") {
		hostInfo.Hostname = spInfo.Software[0].HostName
	}

	// OS
	if slices.Contains(config.Items, "os") {
		re := regexp.MustCompile(`^(\w+)\s([\d.]+)\s\(([^)]+)\)$`)
		matches := re.FindStringSubmatch(spInfo.Software[0].OsVersion)
		if len(matches) == 4 {
			hostInfo.Os.System = matches[1]        // "macOS"
			hostInfo.Os.SystemVersion = matches[2] // "15.2"
			hostInfo.Os.SystemBuild = matches[3]   // "24C101"
		}
		kernelInfoArr := strings.Split(spInfo.Software[0].Kernel, " ")
		hostInfo.Os.KernelType = kernelInfoArr[0]
		hostInfo.Os.KernelVersion = kernelInfoArr[1]
		majorOsVersion := strings.Split(hostInfo.Os.SystemVersion, ".")[0]
		var osFriendlyNameMap = map[string]string{
			"13": "Ventura",
			"14": "Sonoma",
			"15": "Sequoia",
		}
		var ok bool

		if hostInfo.Os.SystemVersionCodeNname, ok = osFriendlyNameMap[majorOsVersion]; !ok {
			hostInfo.Os.SystemVersionCodeNname = "(Unknown)"
		}
	}

	// If we don't use cache, we need to parse the system_profiler output
	if !haveCache {
		// Model
		if slices.Contains(config.Items, "model") {
			hostInfo.Model.Number = spInfo.Hardware[0].ModelNumber
		}

		// CPU
		if slices.Contains(config.Items, "cpu") {
			cpuCoreInfoArr := strings.Split(strings.Split(spInfo.Hardware[0].NumProc, " ")[1], ":")
			hostInfo.Cpu.Model = spInfo.Displays[0].Name
			hostInfo.Cpu.Cores, _ = strconv.Atoi(cpuCoreInfoArr[0])
			hostInfo.Cpu.PerformanceCores, _ = strconv.Atoi(cpuCoreInfoArr[1])
			hostInfo.Cpu.EfficiencyCores, _ = strconv.Atoi(cpuCoreInfoArr[2])
		}

		// GPU
		if slices.Contains(config.Items, "gpu") {
			hostInfo.GpuCores, _ = strconv.Atoi(spInfo.Displays[0].NumCores)
		}

		// Memory
		if slices.Contains(config.Items, "memory") {
			memUnit := strings.Split(spInfo.Memory[0].Amount, " ")
			hostInfo.Memory.Amount, _ = strconv.Atoi(memUnit[0])
			hostInfo.Memory.Unit = memUnit[1]
			hostInfo.Memory.MemType = spInfo.Memory[0].Type
		}
	}

	// Disk
	if slices.Contains(config.Items, "disk") {
		for _, hd := range spInfo.Storage {
			if hd.MountPoint == "/" {
				hostInfo.Disk.TotalTB = float32(hd.SizeByte) / 1000000000000
				hostInfo.Disk.FreeTB = float32(hd.FreeSpaceByte) / 1000000000000
				hostInfo.Disk.SmartStatus = hd.PhyDrive.SmartStatus
				break
			}
		}
	}

	// Battery
	if slices.Contains(config.Items, "battery") {
		hostInfo.Battery.StatusPercent = spInfo.Power[0].BatteryChargeInfo.StateOfCharge
		hostInfo.Battery.CapacityPercent, _ = strconv.Atoi(strings.TrimSuffix(spInfo.Power[0].BatteryHealthInfo.MaxCapacity, "%"))

		if spInfo.Power[0].BatteryChargeInfo.IsCharging == "FALSE" {
			hostInfo.Battery.Charging = false
		} else {
			hostInfo.Battery.Charging = true
		}
		hostInfo.Battery.Health = spInfo.Power[0].BatteryHealthInfo.Health
	}

	// Displays
	if slices.Contains(config.Items, "display") {
		re := regexp.MustCompile(`^(\d+)\s*x\s*(\d+)\s*@\s*([\d.]+)Hz$`)
		//For some unknown reason, sometime the Display information is empty !
		if len(spInfo.Displays) > 0 {
			for _, displayInfo := range spInfo.Displays[0].Ndrvs {
				dInfo := display{}
				tmpArr := strings.Split(displayInfo.Pixels, " x ")
				dInfo.PixelsWidth, _ = strconv.Atoi(tmpArr[0])
				dInfo.PixelsHeight, _ = strconv.Atoi(tmpArr[1])
				matches := re.FindStringSubmatch(displayInfo.Resolution)
				if len(matches) == 4 {
					dInfo.ResolutionWidth, _ = strconv.Atoi(matches[1])
					dInfo.ResolutionHeight, _ = strconv.Atoi(matches[2])
					dInfo.RefreshRateHz, _ = strconv.ParseFloat(matches[3], 64)
				}
				hostInfo.Displays = append(hostInfo.Displays, dInfo)
			}
		}
	}

	// Uptime and Date
	if slices.Contains(config.Items, "uptime") {
		uptimeInfoArr := strings.Split(strings.Split(spInfo.Software[0].Uptime, " ")[1], ":")
		hostInfo.Uptime = fmt.Sprintf("%s days, %s hours", uptimeInfoArr[0], uptimeInfoArr[1])
	}
	if slices.Contains(config.Items, "datetime") {
		hostInfo.Datetime = time.Now().Format(time.RFC1123)
	}

	return
}

// Fetch the number of Homebrew packages installed
func fetchNumHomebrew() (formulaeCount, casksCount int) {
	var filePath string
	var output string
	var err error
	if filePath, err = which("brew"); err != nil {
		return
	}

	cmd := exec.Command(filePath, "list", "-1", "--formulae")
	output, err = runCommand(cmd)
	if err != nil {
		return -1, -1 // "-1" makes it obvious to know there was an error
	}
	formulaeCount = len(strings.Split(output, "\n"))
	cmd = exec.Command(filePath, "list", "-1", "--casks")
	output, err = runCommand(cmd)
	if err != nil {
		return formulaeCount, -1
	}
	casksCount = len(strings.Split(output, "\n"))

	return
}

// Fetch the terminal program using TERM_PROGRAM env. variable
func fetchTermProgram() (termProgram string) {
	termProgram = os.Getenv("TERM_PROGRAM")
	if termProgram == "" {
		termProgram = "Unknown"
	}

	return
}

// Fetch the model of the Mac.
// It comes in the form "MacBook Pro (16-inch, Nov 2024)".
func fetchModelYear() (name, subname, date string) {
	var out bytes.Buffer
	cmd := exec.Command("/usr/sbin/ioreg", "-arc", "IOPlatformDevice", "-k", "product-name")
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		name = "Unknown"
		return
	}
	var r io.Reader = strings.NewReader(out.String())

	var mapXml []map[string]interface{}
	var input string
	if err := plist.NewXMLDecoder(r).Decode(&mapXml); err != nil {
		name = "Unknown"
		return
	}
	input = string(mapXml[0]["product-name"].([]byte))
	// The last character of the "product-name" is unicode \u0000 (nul char), so let's remove it.
	_, size := utf8.DecodeLastRuneInString(input)
	input = input[:len(input)-size]

	re := regexp.MustCompile(`^([\w\s]+)\s\(([^,]+),\s([^)]+)\)$`)
	matches := re.FindStringSubmatch(input)

	if len(matches) == 4 {
		name = matches[1]    // "MacBook Pro"
		subname = matches[2] // "16-inch"
		date = matches[3]    // "Nov 2024"

	} else {
		name = "Unknown"
	}
	return
}

// Fetch the number of applications,
// i.e. the number of directories in /Applications
func fetchNumApps() (numApps int) {
	entries, err := os.ReadDir("/Applications")
	if err != nil {
		return 0
	}

	dirCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			dirCount++
		}
	}
	return dirCount
}

func fetchPublicIp() (ip string) {
	client := &http.Client{
		Timeout: 500 * time.Millisecond,
	}
	resp, err := client.Get("http://ident.me")
	if err != nil {
		return "Unknown"
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Unknown"
	}
	return string(body)
}
