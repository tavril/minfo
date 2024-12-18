package main

// This struct corresponds to the information
// corresponding to a display (screen).
// It is a subset of the info struct.
// It is defined it because we need to call make([]display, X)
// in the code.
type display struct {
	PixelsWidth      int     `json:"pixels_width"`
	PixelsHeight     int     `json:"pixels_height"`
	ResolutionWidth  int     `json:"resolution_width"`
	ResolutionHeight int     `json:"resolution_height"`
	RefreshRateHz    float64 `json:"refresh_rate_hz"`
}

// This struct corresponds to the information
// that will be cached in the cache file.
// It is a subset of the info struct.
type cachedInfo struct {
	Model struct {
		Name    string `json:"name"`
		SubName string `json:"sub_name"`
		Date    string `json:"date"`
		Number  string `json:"number"`
	} `json:"model"`
	Cpu struct {
		Model            string `json:"model"`
		Cores            int    `json:"cores"`
		PerformanceCores int    `json:"performance_cores"`
		EfficiencyCores  int    `json:"efficiency_cores"`
	} `json:"cpu"`
	GpuCores int `json:"gpu_cores"`
	Memory   struct {
		Amount  int    `json:"amount"`
		Unit    string `json:"unit"`
		MemType string `json:"type"`
	} `json:"memory"`
}
type info struct {
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
	cachedInfo
	Disk struct {
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
	PublicIP string `json:"public_ip"`
}

// systemProfileInfo contains all the information
// we need from system_profiler
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

	Hardware []struct {
		MachineName  string `json:"machine_name"`
		MachineModel string `json:"machine_model"`
		ModelNumber  string `json:"model_number"`
		NumProc      string `json:"number_processors"`
	} `json:"SPHardwareDataType"`

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

	Memory []struct {
		Amount string `json:"SPMemoryDataType"`
		Type   string `json:"dimm_type"`
	} `json:"SPMemoryDataType"`

	Storage []struct {
		FreeSpaceByte int    `json:"free_space_in_bytes"`
		SizeByte      int    `json:"size_in_bytes"`
		MountPoint    string `json:"mount_point"`
		PhyDrive      struct {
			SmartStatus string `json:"smart_status"`
		} `json:"physical_drive"`
	} `json:"SPStorageDataType"`
}
