package main

/*
This file contains the structs used in the application.
Here are the two main structs
- info: contains all the information that the tool can retrieve.
- systemProfilerInfo: contains all the information retrieved from system_profiler
*/

// Information about a display (screen)
// It is a subset of the info struct.
type display struct {
	PixelsWidth      int     `json:"pixels_width"`
	PixelsHeight     int     `json:"pixels_height"`
	ResolutionWidth  int     `json:"resolution_width"`
	ResolutionHeight int     `json:"resolution_height"`
	RefreshRateHz    float64 `json:"refresh_rate_hz"`
}

// Information that can be cached in file.
// It is a subset of the info struct.
type Model struct {
	Name    string `json:"name"`
	SubName string `json:"sub_name"`
	Date    string `json:"date"`
	Number  string `json:"number"`
}
type Cpu struct {
	Model            string `json:"model"`
	Cores            int    `json:"cores"`
	PerformanceCores int    `json:"performance_cores"`
	EfficiencyCores  int    `json:"efficiency_cores"`
}
type Memory struct {
	Amount  int    `json:"amount"`
	Unit    string `json:"unit"`
	MemType string `json:"type"`
}

// I use pointers to struct to know if the sub-structs are set or not.
type cachedInfo struct {
	Model        *Model  `json:"model,omitempty"`
	Cpu          *Cpu    `json:"cpu,omitempty"`
	GpuCores     *int    `json:"gpu_cores,omitempty"`
	Memory       *Memory `json:"memory,omitempty"`
	SerialNumber *string `json:"serial_number,omitempty"`
}

// info contains all the information that the tool can retrieve.
type info struct {
	cachedInfo
	User struct {
		RealName string `json:"real_name"`
		Login    string `json:"login"`
	}
	Hostname string `json:"hostname"`
	Os       struct {
		System                 string `json:"system"`
		SystemVersion          string `json:"system_version"`
		SystemBuild            string `json:"system_build"`
		SystemVersionCodeNname string `json:"system_version_code_name"`
		KernelType             string `json:"kernel_type"`
		KernelVersion          string `json:"kernel_version"`
	} `json:"os"`
	SystemIntegrity string `json:"system_integrity"`
	Disk            struct {
		TotalTB     float32 `json:"total_tb"`
		FreeTB      float32 `json:"free_tb"`
		SmartStatus string  `json:"smart_status"`
	} `json:"disk"`
	Battery struct {
		StatusPercent   int    `json:"status_percent"`
		Charging        bool   `json:"charging"`
		CapacityPercent int    `json:"capacity_percent"`
		Health          string `json:"health"`
	} `json:"battery"`
	Displays []display `json:"displays"`
	Software struct {
		NumApps         int `json:"num_apps"`
		NumBrewFormulae int `json:"num_homebrew_formulae"`
		NumBrewCasks    int `json:"num_homebrew_casks"`
	} `json:"software"`
	Terminal string `json:"terminal"`
	Uptime   string `json:"uptime"`
	Datetime string `json:"datetime"`
	PublicIp struct {
		IP      string `json:"query"`
		Country string `json:"country"`
	} `json:"public_ip"`
}

// systemProfileInfo contains all the information
// we need from system_profiler

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

	/*
		Memory []struct {
			Amount string `json:"SPMemoryDataType"`
			Type   string `json:"dimm_type"`
		} `json:"SPMemoryDataType"`
	*/
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
