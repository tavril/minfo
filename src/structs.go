package main

/*
This file contains the structs used in the application.
Here are the two main structs
- info: contains all the information that the tool can retrieve.
- systemProfilerInfo: contains all the information retrieved from system_profiler
*/

/* ---------- Structs that hold the fetched information ---------- */
// Information about a display (screen)
// It is a subset of the info struct.
type display struct {
	PixelsWidth      int     `json:"pixels_width,omitempty"`
	PixelsHeight     int     `json:"pixels_height,omitempty"`
	ResolutionWidth  int     `json:"resolution_width,omitempty"`
	ResolutionHeight int     `json:"resolution_height,omitempty"`
	RefreshRateHz    float64 `json:"refresh_rate_hz,omitempty"`
}

// Information that can be cached in file.
// It is a subset of the info struct.
type Model struct {
	Name    string `json:"name,omitempty"`
	SubName string `json:"sub_name,omitempty"`
	Date    string `json:"date,omitempty"`
	Number  string `json:"number,omitempty"`
}
type Cpu struct {
	Model            string `json:"model,omitempty"`
	Cores            int    `json:"cores,omitempty"`
	PerformanceCores int    `json:"performance_cores,omitempty"`
	EfficiencyCores  int    `json:"efficiency_cores,omitempty"`
}
type Memory struct {
	Amount  int    `json:"amount,omitempty"`
	Unit    string `json:"unit,omitempty"`
	MemType string `json:"type,omitempty"`
}

// I use pointers to struct to know if the sub-structs are set or not.
type cachedInfo struct {
	Model        *Model  `json:"model,omitempty"`
	Cpu          *Cpu    `json:"cpu,omitempty"`
	GpuCores     *int    `json:"gpu_cores,omitempty"`
	Memory       *Memory `json:"memory,omitempty"`
	SerialNumber *string `json:"serial_number,omitempty"`
}

type userInfo struct {
	RealName string `json:"real_name,omitempty"`
	Login    string `json:"login,omitempty"`
}

type osInfo struct {
	System                 string `json:"system,omitempty"`
	SystemVersion          string `json:"system_version,omitempty"`
	SystemBuild            string `json:"system_build,omitempty"`
	SystemVersionCodeNname string `json:"system_version_code_name,omitempty"`
	KernelType             string `json:"kernel_type,omitempty"`
	KernelVersion          string `json:"kernel_version,omitempty"`
}

type diskInfo struct {
	TotalTB     float32 `json:"total_tb,omitempty"`
	FreeTB      float32 `json:"free_tb,omitempty"`
	SmartStatus string  `json:"smart_status,omitempty"`
}

type batteryInfo struct {
	StatusPercent   int    `json:"status_percent,omitempty"`
	Charging        bool   `json:"charging,omitempty"`
	CapacityPercent int    `json:"capacity_percent,omitempty"`
	Health          string `json:"health,omitempty"`
}

type softwareInfo struct {
	NumApps         int `json:"num_apps,omitempty"`
	NumBrewFormulae int `json:"num_homebrew_formulae,omitempty"`
	NumBrewCasks    int `json:"num_homebrew_casks,omitempty"`
}

type publicIpInfo struct {
	IP      string `json:"query,omitempty"`
	Country string `json:"country,omitempty"`
}

// info contains all the information that the tool can retrieve.
// Note: I use pointer to struct, so that when the user requests
// JSON output, the output will not contain empty fields.
type info struct {
	cachedInfo
	User            *userInfo     `json:"user,omitempty"`
	Hostname        string        `json:"hostname,omitempty"`
	Os              *osInfo       `json:"os,omitempty"`
	SystemIntegrity string        `json:"system_integrity,omitempty"`
	Disk            *diskInfo     `json:"disk,omitempty"`
	Battery         *batteryInfo  `json:"battery,omitempty"`
	Displays        []display     `json:"displays,omitempty"`
	Software        *softwareInfo `json:"software,omitempty"`
	Terminal        string        `json:"terminal,omitempty"`
	Uptime          string        `json:"uptime,omitempty"`
	Datetime        string        `json:"datetime,omitempty"`
	PublicIp        *publicIpInfo `json:"public_ip,omitempty"`
}

/* ---------- Structs for system_profiler parsing ---------- */

type HardwareInfo struct {
	MachineName  string      `json:"machine_name"`
	MachineModel string      `json:"machine_model"`
	ModelNumber  string      `json:"model_number"`
	NumProc      interface{} `json:"number_processors"` // Can be a string (Apple Silicon) or an int (Intel)
	ChipType     string      `json:"-"`                 // Common field to store "chip_type" (Apple Silicon) or "cpu_type" (Intel)
	SerialNumber string      `json:"serial_number"`
}

type systemProfilerInfo struct {
	Displays []struct {
		Name     string `json:"_name"`
		NumCores string `json:"sppci_cores"`
		Ndrvs    []struct {
			Name       string `json:"_name"`
			Pixels     string `json:"_spdisplays_pixels"`
			Resolution string `json:"_spdisplays_resolution"`
		} `json:"spdisplays_ndrvs"`
	} `json:"SPDisplaysDataType"`

	Software []struct {
		UserName        string `json:"user_name"`
		HostName        string `json:"local_host_name"`
		OsVersion       string `json:"os_version"`
		Uptime          string `json:"uptime"`
		Kernel          string `json:"kernel_version"`
		SystemIntegrity string `json:"system_integrity"`
	} `json:"SPSoftwareDataType"`

	Hardware []HardwareInfo `json:"SPHardwareDataType"`

	Power []struct {
		BatteryChargeInfo struct {
			StateOfCharge int    `json:"sppower_battery_state_of_charge"`
			AtWarnLevel   string `json:"sppower_battery_at_warn_level"`
			FullyCharged  string `json:"sppower_battery_fully_charged"`
			IsCharging    string `json:"sppower_battery_is_charging"`
		} `json:"sppower_battery_charge_info"`

		BatteryHealthInfo struct {
			CycleCount  int    `json:"sppower_battery_cycle_count"`
			MaxCapacity string `json:"sppower_battery_health_maximum_capacity"`
			Health      string `json:"sppower_battery_health"`
		} `json:"sppower_battery_health_info"`
	} `json:"SPPowerDataType"`

	Memory []interface{} `json:"SPMemoryDataType"`

	Storage []struct {
		FreeSpaceByte int    `json:"free_space_in_bytes"`
		SizeByte      int    `json:"size_in_bytes"`
		MountPoint    string `json:"mount_point"`
		PhyDrive      struct {
			SmartStatus string `json:"smart_status"`
		} `json:"physical_drive"`
	} `json:"SPStorageDataType"`
}
