package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

var (
	weatherHTTPClient  = &http.Client{Timeout: 3 * time.Second}
	geoHTTPClient      = &http.Client{Timeout: 2 * time.Second}
	publicIPHTTPClient = &http.Client{Timeout: 1200 * time.Millisecond}
)

// Custom UnmarshalJSON to handle both "chip_type" and "cpu_type"
func (h *HardwareInfo) UnmarshalJSON(data []byte) error {
	// Parse the input JSON into a temporary map
	var temp map[string]any
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Look for "chip_type" or "cpu_type" and assign to ChipType
	if val, ok := temp["chip_type"].(string); ok {
		h.ChipType = val
	} else if val, ok := temp["cpu_type"].(string); ok {
		h.ChipType = val
	}
	// Assign the rest of the fields
	h.MachineModel = temp["machine_model"].(string)
	h.MachineName = temp["machine_name"].(string)
	if val, ok := temp["model_number"].(string); ok {
		h.ModelNumber = val
	}
	h.NumProc = temp["number_processors"]

	if val, ok := temp["serial_number"].(string); ok {
		h.SerialNumber = val
	}

	return nil
}

func fetchSystemProfiler(hostInfo *info, items []string, spDataTypes map[string]bool, haveCache bool) (err error) {
	var spInfo systemProfilerInfo

	/* ---------- Call system_profiler with needed SPDataType(s) ---------- */
	args := []string{"-json", "-detailLevel", "basic"}
	for k, v := range spDataTypes {
		if v {
			args = append(args, k)
		}
	}
	cmd := exec.Command("/usr/sbin/system_profiler", args...)
	var output string
	output, err = runCommand(cmd)
	if err != nil {
		return
	}

	if err = json.Unmarshal([]byte(output), &spInfo); err != nil {
		return
	}

	/* ---------- Parse the output of system_profiler ---------- */

	if slices.Contains(items, "model") && !haveCache {
		if len(spInfo.Hardware) == 0 {
			return fmt.Errorf("system_profiler returned no hardware information")
		}
		// We also have to call ioreg to get all the information about the model

		hostInfo.Model = &Model{}
		fetchModelYear(hostInfo.Model)
		(*hostInfo.Model).Number = spInfo.Hardware[0].ModelNumber
	}

	if slices.Contains(items, "cpu") && !haveCache {
		if len(spInfo.Hardware) == 0 {
			return fmt.Errorf("system_profiler returned no hardware information")
		}
		// Note: differences between Apple Silicon and Intel CPUs:
		// - field called "chip_type" for Apple Silicon
		// - field called "cpu_type" for Intel
		// - field called "number_processors" is a string for Apple Silicon
		// - field called "number_processors" is an int for Intel
		hostInfo.Cpu = &Cpu{}
		(*hostInfo.Cpu).Model = spInfo.Hardware[0].ChipType
		switch v := spInfo.Hardware[0].NumProc.(type) {
		case string:
			cpuCoreInfoArr := strings.Split(strings.Split(v, " ")[1], ":")
			(*hostInfo.Cpu).Cores, _ = strconv.Atoi(cpuCoreInfoArr[0])
			(*hostInfo.Cpu).PerformanceCores, _ = strconv.Atoi(cpuCoreInfoArr[1])
			(*hostInfo.Cpu).EfficiencyCores, _ = strconv.Atoi(cpuCoreInfoArr[2])
		case int:
			(*hostInfo.Cpu).Cores = int(v)
		}
	}

	if slices.Contains(items, "gpu") && !haveCache {
		if len(spInfo.Displays) == 0 {
			return fmt.Errorf("system_profiler returned no display information")
		}
		tmp, _ := strconv.Atoi(spInfo.Displays[0].NumCores)
		hostInfo.GpuCores = &tmp
	}

	if slices.Contains(items, "memory") && !haveCache {
		if len(spInfo.Memory) == 0 {
			return fmt.Errorf("system_profiler returned no memory information")
		}
		hostInfo.Memory = &Memory{}
		for _, mem := range spInfo.Memory {
			memMap, _ := mem.(map[string]interface{})
			switch arch {
			case "arm64":
				memUnit := strings.Split(memMap["SPMemoryDataType"].(string), " ")
				(*hostInfo.Memory).Amount, _ = strconv.Atoi(memUnit[0])
				(*hostInfo.Memory).Unit = memUnit[1]
				(*hostInfo.Memory).MemType = memMap["dimm_type"].(string)
			case "amd64":
				for _, item := range memMap["Items"].([]interface{}) {
					itemMap, _ := item.(map[string]interface{})
					memUnit := strings.Split(itemMap["dimm_size"].(string), " ")
					tmp, _ := strconv.Atoi(memUnit[0])
					(*hostInfo.Memory).Amount += tmp
					// Unit and MemType are the same for all DIMMs.
					// Let's only fill it once.
					if (*hostInfo.Memory).Unit == "" {
						(*hostInfo.Memory).Unit = memUnit[1]
						(*hostInfo.Memory).MemType = itemMap["dimm_type"].(string)
					}
				}
			}
		}
	}

	if slices.Contains(items, "user") {
		if len(spInfo.Software) == 0 {
			return fmt.Errorf("system_profiler returned no software information")
		}
		re := regexp.MustCompile(`^([\w\s]+)\s\((\w+)\)$`)
		matches := re.FindStringSubmatch(spInfo.Software[0].UserName)

		hostInfo.User = &userInfo{}
		if len(matches) == 3 {
			(*hostInfo.User).RealName = matches[1]
			(*hostInfo.User).Login = matches[2]
		}
	}

	if slices.Contains(items, "hostname") {
		if len(spInfo.Software) == 0 {
			return fmt.Errorf("system_profiler returned no software information")
		}
		hostInfo.Hostname = spInfo.Software[0].HostName
	}

	if slices.Contains(items, "os") {
		if len(spInfo.Software) == 0 {
			return fmt.Errorf("system_profiler returned no software information")
		}
		hostInfo.Os = &osInfo{}
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
			"26": "Tahoe",
		}
		var ok bool

		if hostInfo.Os.SystemVersionCodeNname, ok = osFriendlyNameMap[majorOsVersion]; !ok {
			hostInfo.Os.SystemVersionCodeNname = "(Unknown)"
		}
	}

	if slices.Contains(items, "system_integrity") {
		if len(spInfo.Software) == 0 {
			return fmt.Errorf("system_profiler returned no software information")
		}
		hostInfo.SystemIntegrity = spInfo.Software[0].SystemIntegrity
	}

	if slices.Contains(items, "serial_number") && !haveCache {
		if len(spInfo.Hardware) == 0 {
			return fmt.Errorf("system_profiler returned no hardware information")
		}
		hostInfo.SerialNumber = &spInfo.Hardware[0].SerialNumber
	}

	if slices.Contains(items, "disk") {
		if len(spInfo.Storage) == 0 {
			return fmt.Errorf("system_profiler returned no storage information")
		}
		hostInfo.Disk = &diskInfo{}
		for _, hd := range spInfo.Storage {
			if hd.MountPoint == "/" {
				hostInfo.Disk.TotalTB = float32(hd.SizeByte) / 1000000000000
				hostInfo.Disk.FreeTB = float32(hd.FreeSpaceByte) / 1000000000000
				hostInfo.Disk.SmartStatus = hd.PhyDrive.SmartStatus
				break
			}
		}
	}

	if slices.Contains(items, "battery") {
		if len(spInfo.Power) == 0 {
			return fmt.Errorf("system_profiler returned no power information")
		}
		hostInfo.Battery = &batteryInfo{}
		hostInfo.Battery.StatusPercent = spInfo.Power[0].BatteryChargeInfo.StateOfCharge
		hostInfo.Battery.CapacityPercent, _ = strconv.Atoi(strings.TrimSuffix(spInfo.Power[0].BatteryHealthInfo.MaxCapacity, "%"))

		if spInfo.Power[0].BatteryChargeInfo.IsCharging == "FALSE" {
			hostInfo.Battery.Charging = false
		} else {
			hostInfo.Battery.Charging = true
		}
		hostInfo.Battery.Health = spInfo.Power[0].BatteryHealthInfo.Health
	}

	if slices.Contains(items, "display") {
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

	if slices.Contains(items, "uptime") {
		if len(spInfo.Software) == 0 {
			return fmt.Errorf("system_profiler returned no software information")
		}
		uptimeInfoArr := strings.Split(strings.Split(spInfo.Software[0].Uptime, " ")[1], ":")
		hostInfo.Uptime = fmt.Sprintf("%s days, %s hours", uptimeInfoArr[0], uptimeInfoArr[1])
	}

	if slices.Contains(items, "datetime") {
		hostInfo.Datetime = time.Now().Format(time.RFC1123)
	}

	return
}

// Fetch the model of the Mac. CALLED BY fetchSystemProfiler()
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
}

func fetchDateTime(hostInfo *info) {
	hostInfo.Datetime = time.Now().Format(time.RFC1123)
}

// This functions fetches the number of installed applications:
// - number of directories in /Applications.
// - number of HomeBrew formulae.
// - number of HomeBrew casks.
func fetchSoftware(hostInfo *info) {
	hostInfo.Software = &softwareInfo{}
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
	hostInfo.Software.NumBrewFormulae = countNonEmptyLines(output)

	cmd = exec.Command(filePath, "list", "-1", "--casks")
	output, err = runCommand(cmd)
	if err != nil {
		return
	}
	hostInfo.Software.NumBrewCasks = countNonEmptyLines(output)
}

// Fetch the terminal program using TERM_PROGRAM env. variable
func fetchTermProgram(hostInfo *info) {
	termProgram := os.Getenv("TERM_PROGRAM")
	if termProgram == "" {
		hostInfo.Terminal = "Unknown"
		return
	}
	hostInfo.Terminal = termProgram
}

func fetchWeatherOpenMeteo(hostInfo *info) {
	hostInfo.Weather = &weather{}

	var latitude, longitude *float64
	var countryCode string
	if config.Weather.LocationNameEn != nil {
		latitude, longitude, countryCode = fetchCoordinatesFromName(
			*config.Weather.LocationNameEn,
			*config.Weather.LocationStateEn,
			*config.Weather.LocationCountryEn,
		)
		if latitude == nil || longitude == nil {
			return
		}
	} else if config.Weather.Latitude != nil && config.Weather.Longitude != nil {
		latitude = config.Weather.Latitude
		longitude = config.Weather.Longitude
	} else {
		if hostInfo.PublicIp == nil {
			fetchPublicIp(hostInfo)
			if hostInfo.PublicIp == nil {
				return
			}
		}
		latitude = &hostInfo.PublicIp.Latitude
		longitude = &hostInfo.PublicIp.Longitude
		countryCode = hostInfo.PublicIp.CountryCode
	}

	// Define the API URL
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,apparent_temperature,weather_code,wind_speed_10m,wind_direction_10m,wind_gusts_10m",
		*latitude,
		*longitude,
	)
	if config.Weather.Units == "imperial" {
		url += "&temperature_unit=fahrenheit&wind_speed_unit=mph&precipitation_unit=inch"
	}

	// Make the HTTP GET request
	response, err := weatherHTTPClient.Get(url)
	if err != nil {
		return
	}
	defer response.Body.Close()

	// Check for successful response
	if response.StatusCode != http.StatusOK {
		return
	}

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	// Parse the JSON response
	var openMeteo openMeteo
	if err = json.Unmarshal(body, &openMeteo); err != nil {
		fmt.Printf("Error fetching weather (unmarshall): %v\n", err)
		return
	}

	hostInfo.Weather.TempUnit = openMeteo.CurrentUnits.Temperature2m
	hostInfo.Weather.WindUnit = openMeteo.CurrentUnits.WindSpeed10m
	hostInfo.Weather.Temperature = openMeteo.Current.Temperature2m
	hostInfo.Weather.FeelsLike = openMeteo.Current.ApparentTemperature
	hostInfo.Weather.WindSpeed = openMeteo.Current.WindSpeed10m
	hostInfo.Weather.WindGusts = openMeteo.Current.WindGusts10m
	hostInfo.Weather.WindDirection = openMeteo.Current.WindDirection10m
	if descByLang, ok := wmoCodesDesc[openMeteo.Current.WeatherCode]; ok {
		if desc, ok := descByLang[config.Weather.Lang]; ok {
			hostInfo.Weather.CurrentWeather = desc
		} else if fallback, ok := descByLang["en"]; ok {
			hostInfo.Weather.CurrentWeather = fallback
		} else {
			hostInfo.Weather.CurrentWeather = "Unknown"
		}
	} else {
		hostInfo.Weather.CurrentWeather = "Unknown"
	}
	hostInfo.Weather.LocationCountryCode = countryCode
	if config.Weather.LocationNameEn != nil {
		hostInfo.Weather.LocationName = *config.Weather.LocationNameEn
	} else {
		hostInfo.Weather.LocationName = ""
	}
	hostInfo.Weather.Latitude = *latitude
	hostInfo.Weather.Longitude = *longitude
}

func fetchCoordinatesFromName(locationName, locationState, locationCountry string) (latitude *float64, longitude *float64, countryCode string) {
	// Define the API URL
	reqURL, err := url.Parse("https://geocoding-api.open-meteo.com/v1/search")
	if err != nil {
		return
	}
	query := reqURL.Query()
	query.Set("name", locationName)
	reqURL.RawQuery = query.Encode()

	// Make the HTTP GET request
	resp, err := geoHTTPClient.Get(reqURL.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		return
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// Parse the JSON response
	var geo openMeteoGeo
	if err = json.Unmarshal(body, &geo); err != nil {
		return
	}

	if len(geo.Results) == 0 {
		return
	}

	upLocationName := strings.ToUpper(locationName)
	upLocationState := ""
	if locationState != "" {
		upLocationState = strings.ToUpper(locationState)
	}
	upLocationCountry := strings.ToUpper(locationCountry)
	for _, result := range geo.Results {
		if strings.ToUpper(result.Name) == upLocationName &&
			strings.ToUpper(result.Country) == upLocationCountry &&
			((upLocationState != "" && strings.ToUpper(result.Admin1) == upLocationState) || upLocationState == "") {

			latitude = new(float64)
			longitude = new(float64)
			*latitude = result.Latitude
			*longitude = result.Longitude
			countryCode = result.CountryCode
			return
		}
	}
	return
}

type ipapiResponse struct {
	IP          string  `json:"ip"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	CountryName string  `json:"country_name"`
	CountryCode string  `json:"country"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

// Fetch the public IP address (and its country name)
func fetchPublicIp(hostInfo *info) {

	if hostInfo.PublicIp != nil {
		return
	}
	resp, err := publicIPHTTPClient.Get("https://ipapi.co/json/")
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
	// Unmarshal the JSON response into a tmp Struct,
	// because in case of error we want hostInfo.PublicIp to be nil.
	tmpStruct := ipapiResponse{}
	if err = json.Unmarshal(body, &tmpStruct); err != nil {
		return
	}
	hostInfo.PublicIp = &publicIpInfo{
		IP:          tmpStruct.IP,
		Country:     tmpStruct.CountryName,
		CountryCode: tmpStruct.CountryCode,
		City:        tmpStruct.City,
		State:       tmpStruct.Region,
		Latitude:    tmpStruct.Latitude,
		Longitude:   tmpStruct.Longitude,
	}
}
