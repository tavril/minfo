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
// The "haveCache" parameter is used to know if we need to
// parse the output of "system_profiler" for cached items or not.
func fetchSystemProfiler(hostInfo *info, spDataTypes map[string]bool, haveCache bool) (err error) {
	var spInfo systemProfilerInfo

	/* ---------- Call system_profiler with needed SPDataType(s) ---------- */
	args := []string{"-json", "-detailLevel", "basic"}
	for k, v := range spDataTypes {
		if v {
			args = append(args, k)
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

	/* ---------- Parse the output of system_profiler ---------- */

	// If we don't use cache, we need to parse the system_profiler for cached information
	// note: we can run into the case where the first time the program is run, the user
	// only requested a subset of the information, so that the cache file is created
	// with the "cachedInfo" struct that might contains non-set fields (string="", int=0, etc.)
	// --> that why we need to check if the fields are set or not.
	if slices.Contains(config.Items, "model") {
		if !haveCache || hostInfo.Model == nil {
			// We also have to call ioreg to get all the information about the model

			hostInfo.Model = &Model{}
			fetchModelYear(hostInfo.Model)
			(*hostInfo.Model).Number = spInfo.Hardware[0].ModelNumber
		}
	}

	if slices.Contains(config.Items, "cpu") {
		if !haveCache || hostInfo.Cpu == nil {
			cpuCoreInfoArr := strings.Split(strings.Split(spInfo.Hardware[0].NumProc, " ")[1], ":")
			hostInfo.Cpu = &Cpu{}
			(*hostInfo.Cpu).Model = spInfo.Displays[0].Name
			(*hostInfo.Cpu).Cores, _ = strconv.Atoi(cpuCoreInfoArr[0])
			(*hostInfo.Cpu).PerformanceCores, _ = strconv.Atoi(cpuCoreInfoArr[1])
			(*hostInfo.Cpu).EfficiencyCores, _ = strconv.Atoi(cpuCoreInfoArr[2])
		}
	}

	if slices.Contains(config.Items, "gpu") {
		if !haveCache || hostInfo.GpuCores == nil {
			tmp, _ := strconv.Atoi(spInfo.Displays[0].NumCores)
			hostInfo.GpuCores = &tmp
		}
	}

	if slices.Contains(config.Items, "memory") {
		if !haveCache || hostInfo.Memory == nil {
			memUnit := strings.Split(spInfo.Memory[0].Amount, " ")
			hostInfo.Memory = &Memory{}
			(*hostInfo.Memory).Amount, _ = strconv.Atoi(memUnit[0])
			(*hostInfo.Memory).Unit = memUnit[1]
			(*hostInfo.Memory).MemType = spInfo.Memory[0].Type
		}
	}

	if slices.Contains(config.Items, "user") {
		re := regexp.MustCompile(`^([\w\s]+)\s\((\w+)\)$`)
		matches := re.FindStringSubmatch(spInfo.Software[0].UserName)

		if len(matches) == 3 {
			hostInfo.User.RealName = matches[1]
			hostInfo.User.Login = matches[2]
		}
	}

	if slices.Contains(config.Items, "hostname") {
		hostInfo.Hostname = spInfo.Software[0].HostName
	}

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

	if slices.Contains(config.Items, "system_integrity") {
		hostInfo.SystemIntegrity = spInfo.Software[0].SystemIntegrity
	}

	if slices.Contains(config.Items, "serial_number") {
		if !haveCache || hostInfo.SerialNumber == nil {
			hostInfo.SerialNumber = &spInfo.Hardware[0].SerialNumber
		}
	}

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

	if slices.Contains(config.Items, "uptime") {
		uptimeInfoArr := strings.Split(strings.Split(spInfo.Software[0].Uptime, " ")[1], ":")
		hostInfo.Uptime = fmt.Sprintf("%s days, %s hours", uptimeInfoArr[0], uptimeInfoArr[1])
	}

	if slices.Contains(config.Items, "datetime") {
		hostInfo.Datetime = time.Now().Format(time.RFC1123)
	}

	return
}

func fetchDateTime(hostInfo *info) {
	hostInfo.Datetime = time.Now().Format(time.RFC1123)
}

// This functions fetches the number of installed applications:
// - number of directories in /Applications.
// - number of HomeBrew formulae.
// - number of HomeBrew casks.
func fetchSoftware(hostInfo *info) {
	/* ---------- Number of directories in /Applications ---------- */
	entries, err := os.ReadDir("/Applications")
	if err != nil {
		hostInfo.Software.NumApps = -1
	} else {
		for _, entry := range entries {
			if entry.IsDir() {
				hostInfo.Software.NumApps++
			}
		}
	}

	/* ---------- Numner of HomeBrew Formulae/Casks ---------- */
	hostInfo.Software.NumBrewFormulae = -1
	hostInfo.Software.NumBrewCasks = -1
	filePath, err := which("brew")
	if err != nil {
		return
	}

	cmd := exec.Command(filePath, "list", "-1", "--formulae")
	output, err := runCommand(cmd)
	if err != nil {
		return
	}
	hostInfo.Software.NumBrewFormulae = len(strings.Split(output, "\n"))

	cmd = exec.Command(filePath, "list", "-1", "--casks")
	output, err = runCommand(cmd)
	if err != nil {
		return
	}
	hostInfo.Software.NumBrewCasks = len(strings.Split(output, "\n"))

	return
}

// Fetch the terminal program using TERM_PROGRAM env. variable
func fetchTermProgram(hostInfo *info) {
	termProgram := os.Getenv("TERM_PROGRAM")
	if termProgram == "" {
		hostInfo.Terminal = "Unknown"
	}
	hostInfo.Terminal = termProgram
	return
}

// Fetch the model of the Mac.
// It comes in the form "MacBook Pro (16-inch, Nov 2024)".
func fetchModelYear(model *Model) {
	var out bytes.Buffer
	model.Name = "Unknown"
	cmd := exec.Command("/usr/sbin/ioreg", "-arc", "IOPlatformDevice", "-k", "product-name")
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return
	}
	var r io.Reader = strings.NewReader(out.String())

	var mapXml []map[string]interface{}
	var input string
	if err := plist.NewXMLDecoder(r).Decode(&mapXml); err != nil {
		return
	}
	input = string(mapXml[0]["product-name"].([]byte))
	// The last character of the "product-name" is unicode \u0000 (nul char), so let's remove it.
	_, size := utf8.DecodeLastRuneInString(input)
	input = input[:len(input)-size]

	re := regexp.MustCompile(`^([\w\s]+)\s\(([^,]+),\s([^)]+)\)$`)
	matches := re.FindStringSubmatch(input)

	if len(matches) == 4 {
		// "MacBook Pro" "16-inch" "Nov 2024"
		model.Name = matches[1]
		model.SubName = matches[2]
		model.Date = matches[3]
	}
	return
}

// Fetch the public IP address (and its country name)
func fetchPublicIp(hostInfo *info) {
	hostInfo.PublicIp.IP = "Unknown"
	client := &http.Client{
		Timeout: 500 * time.Millisecond,
	}
	resp, err := client.Get("http://ip-api.com/json")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if err = json.Unmarshal(body, &hostInfo.PublicIp); err != nil {
		return
	}

	return
}
